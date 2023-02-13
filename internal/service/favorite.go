package service

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pb "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/favorite/route"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/cronUtils"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type favoriteService struct {
	pb.UnimplementedFavoriteInfoServer
	videoKeySet map[string]interface{}
}

var (
	favoriteServiceInstance *favoriteService
	favoriteOnce            sync.Once
	deleteOnce              sync.Once
	writeOnce               sync.Once
)

const (
	videoFavoriteExpireTime = 24 * time.Hour
	userFavoriteExpireTime  = 90 * time.Minute
	videoFavoritePrefix     = "video_favorite_"
	userFavoritePrefix      = "user_favorite_"
)

// 获取视频点赞的持续时间
func getVideoFavoriteExpireTime() time.Duration {
	return time.Duration(int64(videoFavoriteExpireTime) + rand.Int63n(int64(12*time.Hour)))
}

// 获取用户点赞的持续时间
func getUserFavoriteExpireTime() time.Duration {
	return time.Duration(int64(userFavoriteExpireTime) + rand.Int63n(int64(30*time.Minute)))
}

// GetFavoriteServiceInstance 获取一个favoriteService的实例
func GetFavoriteServiceInstance() *favoriteService {
	initRedis()
	initKafka()
	favoriteOnce.Do(func() {
		favoriteServiceInstance = &favoriteService{
			videoKeySet: make(map[string]interface{}, 0),
		}
	})
	return favoriteServiceInstance
}

// FavoriteAction GRPC调用，调用本地点赞信息
func (f *favoriteService) FavoriteAction(ctx context.Context, in *pb.FavoriteAction) (*pb.BaseResp, error) {
	userId := in.UserId
	videoId := in.VideoId
	actionType := in.ActionType
	err := f.FavoriteInfo(userId, videoId, actionType)
	if err != nil {
		return nil, err
	}
	return &pb.BaseResp{
		StatusCode: 0,
		StatusMsg:  "",
	}, nil
}

func (f *favoriteService) FavoriteList(userFavorite *pb.UserFavorite, stream pb.FavoriteInfo_FavoriteListServer) error {
	loginUserId := userFavorite.LoginUserId
	queryUserId := userFavorite.QueryUserId
	videoList, err := f.FavoriteListInfo(loginUserId, queryUserId)
	if err != nil {
		return err
	}

	for _, video := range *videoList {
		err := stream.Send(&pb.VideoResp{
			Id: video.Id,
			Author: &pb.UserResp{
				Id:            video.Author.Id,
				Name:          video.Author.Name,
				FollowCount:   video.Author.FollowCount,
				FollowerCount: video.Author.FollowerCount,
				IsFollow:      video.Author.IsFollow,
			},
			PlayURL:       video.PlayUrl,
			CoverURL:      video.CoverUrl,
			FavoriteCount: video.FavoriteCount,
			CommentCount:  video.CommentCount,
			IsFavorite:    video.IsFavorite,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// FavoriteInfo service层处理用户点赞或者取消点赞
// 当有点赞消息传入时，通过go协程启动分布式定时任务，删除Favorite点赞记录数据库中被软删除的部分
// 可能返回的错误类型：InnerDataBaseError, RecordNotMatch, RecordNotExist,UnknownActionTypeErr
func (f *favoriteService) FavoriteInfo(userId, videoId int64, actionType int32) error {
	//定时删除点赞消息
	go deleteOnce.Do(func() {
		for {
			_, err := cronUtils.CronLab.AddFunc("@every 30m", func() {
				logger.GlobalLogger.Printf("In DeleteDatabaseRegularly")
				err := f.DeleteDatabaseRegularly()
				if err != nil {
					logger.GlobalLogger.Printf("error occurs in DeleteDatabaseRegularly")
				}
			})
			if err == nil {
				break
			}
		}
	})

	//定时将redis中的点赞数写入数据库
	go writeOnce.Do(func() {
		for {
			_, err := cronUtils.CronLab.AddFunc("@every 15m", func() {
				logger.GlobalLogger.Printf("In WriteToDataBaseRegularly")
				err := f.WriteToDataBaseRegularly()
				if err != nil {
					logger.GlobalLogger.Printf("error occurs in WriteToDataBaseRegularly")
				}
			})
			if err == nil {
				break
			}
		}
	})

	go f.writeToKafkaAsyn(userId, videoId, actionType)
	go func() {
		for {
			err := f.writeToRedis(userId, videoId, actionType)
			if err == nil {
				break
			}
		}
	}()
	return nil
}

// 异步写入消息队列
func (f *favoriteService) writeToKafkaAsyn(userId, videoId int64, actionType int32) {
	//将点赞消息写入Kafka
	for {
		favoriteMsg := &sarama.ProducerMessage{}
		favoriteMsg.Topic = constants.KafkaTopicPrefix + "favorite"
		if actionType == api.FavoriteAction {
			favoriteMsg.Key = sarama.StringEncoder("Favorite")
		} else if actionType == api.UnFavoriteAction {
			favoriteMsg.Key = sarama.StringEncoder("Unfavorite")
		}
		favoriteMsg.Value = sarama.StringEncoder(strconv.FormatInt(userId, 10) + ":" + strconv.FormatInt(videoId, 10))
		pid, offset, err := kafkaServer.SendMessage(favoriteMsg)
		if err == nil {
			logger.GlobalLogger.Printf("pid:%v offset:%v\n", pid, offset)
			break
		}
	}
}

//将点赞信息写入redis
func (f *favoriteService) writeToRedis(userId, videoId int64, actionType int32) error {
	//判断redis中是否存在
	videoKey := videoFavoritePrefix + strconv.FormatInt(videoId, 10)
	exists, err := redisClient.Exists(videoKey).Result()
	if err != nil {
		return constants.RedisDBErr
	}
	//写入videoKeySet
	f.videoKeySet[videoKey] = nil
	if exists == 0 {
		favoriteCount, err := dao.GetFavoriteDaoInstance().GetFavoriteCount(videoId)
		if err != nil && errors.Is(constants.RecordNotExistErr, err) {
			//放入空缓存
			err = redisClient.Set(videoKey, emptyCache, getEmptyCacheExpireTime()).Err()
			if err != nil {
				return constants.RedisDBErr
			}
		} else {
			return err
		}
		if actionType == api.FavoriteAction {
			favoriteCount++
		} else if actionType == api.UnFavoriteAction {
			favoriteCount--
		}
		err = redisClient.Set(videoKey, strconv.Itoa(int(favoriteCount)), getVideoFavoriteExpireTime()).Err()
		if err != nil {
			return constants.RedisDBErr
		}
	} else {
		redisClient.Expire(videoKey, videoFavoriteExpireTime)
		if emptyCache == redisClient.Get(videoKey).Val() {
			//直接返回
			return nil
		}
		if actionType == api.FavoriteAction {
			err = redisClient.Incr(videoKey).Err()
		} else if actionType == api.UnFavoriteAction {
			err = redisClient.Decr(videoKey).Err()
		}
		if err != nil {
			return constants.RedisDBErr
		}
	}
	userKey := userFavoritePrefix + strconv.FormatInt(userId, 10)
	exists, err = redisClient.Exists(userKey).Result()
	if err != nil {
		return constants.RedisDBErr
	}
	if exists == 0 {
		//从数据库中获得点赞列表，放入redis中
		videos, errdb := dao.GetFavoriteDaoInstance().GetFavoriteList(userId)
		for errdb != nil {
			videos, err = dao.GetFavoriteDaoInstance().GetFavoriteList(userId)
		}
		for _, video := range videos {
			redisClient.LPush(userKey, video.VideoID)
		}
	}
	redisClient.LPush(userKey, videoId)
	redisClient.Expire(userKey, getUserFavoriteExpireTime())
	return nil
}

//FavoriteListInfo service层查找用户点赞过的所有视频
func (f *favoriteService) FavoriteListInfo(loginUserId, userId int64) (*[]api.Video, error) {
	_, err := GetUserServiceInstance().GetUserByUserId(userId)
	if errors.Is(constants.UserNotExistErr, err) {
		return nil, err
	}
	userKey := userFavoritePrefix + strconv.FormatInt(userId, 10)
	exists, err := redisClient.Exists(userKey).Result()
	if err != nil {
		return nil, constants.RedisDBErr
	}
	if exists == 0 {
		//从数据库中获得点赞列表，放入redis中
		videos, errdb := dao.GetFavoriteDaoInstance().GetFavoriteList(userId)
		for errdb != nil {
			videos, err = dao.GetFavoriteDaoInstance().GetFavoriteList(userId)
		}
		for _, video := range videos {
			redisClient.LPush(userKey, video.VideoID)
		}
	}
	videoIds, err := redisClient.LRange(userKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	videoList, err := getVideoListByID(loginUserId, videoIds)
	return &videoList, err
}

// DeleteDatabaseRegularly 定时删除数据库
func (f *favoriteService) DeleteDatabaseRegularly() error {
	err := dao.GetFavoriteDaoInstance().HardDeleteUnFavorite()
	return err
}

// WriteToDataBaseRegularly 定时从redis中获取点赞记录写入数据库
func (f *favoriteService) WriteToDataBaseRegularly() error {
	for key, _ := range f.videoKeySet {
		logger.GlobalLogger.Printf("key = %v", key)
		val, err := redisClient.Get(key).Result()
		if err != nil {
			return constants.RedisDBErr
		}
		videoId, err := strconv.ParseInt(key[len(videoFavoritePrefix):], 10, 64)
		favoriteCnt, err := strconv.ParseInt(val, 10, 32)
		logger.GlobalLogger.Printf("videoId = %v, favoriteCnt = %v", videoId, favoriteCnt)
		if err != nil {
			return constants.RedisDBErr
		}
		err = dao.GetFavoriteDaoInstance().SetFavoriteCount(videoId, int32(favoriteCnt))
		if err != nil {
			return constants.InnerDataBaseErr
		}
	}
	//删除videoKeySet
	f.videoKeySet = make(map[string]interface{}, 0)
	return nil
}
