package service

import (
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/cronUtils"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"strconv"
	"sync"
	"time"
)

type favoriteService struct {
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
	videoFavoritePrefix     = "video_favorite_"
)

// GetFavoriteServiceInstance 获取一个favoriteService的实例
func GetFavoriteServiceInstance() *favoriteService {
	initRedis()
	favoriteOnce.Do(func() {
		favoriteServiceInstance = &favoriteService{}
	})
	return favoriteServiceInstance
}

// FavoriteInfo service层处理用户点赞或者取消点赞
// 当有点赞消息传入时，通过go协程启动分布式定时任务，删除Favorite点赞记录数据库中被软删除的部分
// 可能返回的错误类型：InnerDataBaseError, RecordNotMatch, RecordNotExist,UnknownActionTypeErr
func (f *favoriteService) FavoriteInfo(userId, videoId int64, actionType int) error {
	//定时删除点赞消息
	go deleteOnce.Do(func() {
		for {
			_, err := cronUtils.CronLab.AddFunc("@every 30m", func() {
				logger.GlobalLogger.Printf("Adding DeleteDatabaseRegularly")
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
				logger.GlobalLogger.Printf("Adding WriteToDataBaseRegularly")
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

	//执行业务逻辑处理
	err := dao.GetFavoriteDaoInstance().FavoriteAction(userId, videoId)
	if err != nil {
		return constants.RedisDBErr
	}
	//判断redis中是否存在，在本地的videoIdSet中存在不一定存在于redis中，以为可能过期
	videoKey := videoFavoritePrefix + strconv.FormatInt(videoId, 10)
	exists, err := redisClient.Exists(videoKey).Result()
	if err != nil {
		return constants.RedisDBErr
	}
	if exists == 0 {
		//写入videoKeySet
		f.videoKeySet[videoKey] = nil
		favoriteCount, err := dao.GetFavoriteDaoInstance().GetFavoriteCount(videoId)
		if err != nil {
			return constants.RedisDBErr
		}
		favoriteCount++
		redisClient.Set(videoKey, string(favoriteCount), getFavoriteRandomTime())
	} else {
		err = redisClient.Incr(videoKey).Err()
		if err != nil {
			return constants.RedisDBErr
		}
	}
	return nil
}

//FavoriteListInfo service层查找用户点赞过的所有视频
func (f *favoriteService) FavoriteListInfo(loginUserId, userId int64) (*[]api.Video, error) {
	videos, err := dao.GetFavoriteDaoInstance().GetFavoriteList(userId)
	if err != nil {
		return nil, err
	}
	videoList, err := newVideoList(loginUserId, videos)
	if err != nil {
		return nil, err
	}
	return &videoList, nil
}

// DeleteDatabaseRegularly 定时删除数据库
func (f *favoriteService) DeleteDatabaseRegularly() error {
	err := dao.GetFavoriteDaoInstance().HardDeleteUnFavorite()
	return err
}

// WriteToDataBaseRegularly 定时从redis中获取点赞记录写入数据库
func (f *favoriteService) WriteToDataBaseRegularly() error {
	for key, _ := range f.videoKeySet {
		val, err := redisClient.Get(key).Result()
		if err != nil {
			return constants.RedisDBErr
		}
		videoId, err := strconv.ParseInt(key[len(videoFavoritePrefix):], 10, 64)
		favoriteCnt, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return constants.RedisDBErr
		}
		err = dao.GetFavoriteDaoInstance().SetFavoriteCount(videoId, int32(favoriteCnt))
		if err != nil {
			return constants.InnerDataBaseErr
		}
	}
	return nil
}
