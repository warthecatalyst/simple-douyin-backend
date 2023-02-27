package dao

import (
	"context"
	"errors"
	pbdao "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/video"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"
	"io"
	"sync"
	"time"
)

//videoDao 与video相关的数据库操作集合
type videoDao struct {
	pbdao.UnimplementedVideoDaoInfoServer
}

var (
	videoDaoInstance *videoDao
	videoOnce        sync.Once
)

// GetVideoDaoInstance 获取一个VideoDao的实例
func GetVideoDaoInstance() *videoDao {
	videoOnce.Do(func() {
		videoDaoInstance = &videoDao{}
	})
	return videoDaoInstance
}

// AddVideo RPC远程调用添加一条Video信息
func (v *videoDao) AddVideo(ctx context.Context, post *pbdao.VideoDaoPost) (*wrapperspb.BoolValue, error) {
	video := &model.Video{
		VideoID:       post.VideoId,
		VideoName:     post.VideoName,
		UserID:        post.UserId,
		FavoriteCount: 0,
		CommentCount:  0,
		PlayURL:       post.PlayURL,
		CoverURL:      post.CoverURL,
	}
	err := v.createVideo(video)
	if err != nil {
		return &wrapperspb.BoolValue{Value: false}, status.Errorf(codes.Internal, constants.InnerDataBaseErr.Error())
	}
	return &wrapperspb.BoolValue{Value: true}, nil
}

// GetPublishIdList RPC远程调用获取publish的IdList
func (v *videoDao) GetPublishIdList(in *wrapperspb.Int64Value, stream pbdao.VideoDaoInfo_GetPublishIdListServer) error {
	userId := in.Value
	videoInfos, err := v.GetPublishListInfo(userId)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	if len(videoInfos) == 0 {
		return status.Errorf(codes.NotFound, constants.RecordNotExistErr.Error())
	}
	for _, video := range videoInfos {
		if err := stream.Send(&wrapperspb.Int64Value{Value: video.VideoID}); err != nil {
			return err
		}
	}
	return nil
}

// GetVideoByVideoId RPC通过Video获得VideoId
func (v *videoDao) GetVideoByVideoId(ctx context.Context, in *wrapperspb.Int64Value) (*pbdao.VideoDaoMsg, error) {
	videoId := in.Value
	videoInfo, err := v.GetVideoByVideoIdInfo(videoId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pbdao.VideoDaoMsg{
		VideoId:       videoId,
		VideoName:     videoInfo.VideoName,
		UserId:        videoInfo.UserID,
		FavoriteCount: videoInfo.FavoriteCount,
		CommentCount:  videoInfo.CommentCount,
		PlayURL:       videoInfo.PlayURL,
		CoverURL:      videoInfo.CoverURL,
	}, status.New(codes.OK, "").Err()
}

// GetVideoListByVideoIdList 通过VideoIdList获取VideoList
func (v *videoDao) GetVideoListByVideoIdList(stream pbdao.VideoDaoInfo_GetVideoListByVideoIdListServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		videoId := in.Value
		videoInfo, err := v.GetVideoByVideoIdInfo(videoId)
		if err != nil {
			if errors.Is(constants.RecordNotExistErr, err) {
				return status.Errorf(codes.NotFound, err.Error())
			} else {
				return status.Errorf(codes.Internal, err.Error())
			}
		}
		if err = stream.Send(&pbdao.VideoDaoMsg{
			VideoId:       videoId,
			VideoName:     videoInfo.VideoName,
			UserId:        videoInfo.UserID,
			FavoriteCount: videoInfo.FavoriteCount,
			CommentCount:  videoInfo.CommentCount,
			PlayURL:       videoInfo.PlayURL,
			CoverURL:      videoInfo.CoverURL,
		}); err != nil {
			return err
		}
	}
}

// createVideo 在数据库中通过事务插入一条Video数据
func (v *videoDao) createVideo(video *model.Video) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		if err := tx.Create(video).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		// 返回 nil 提交事务
		return nil
	})
}

// GetPublishListInfo 在数据库中获得该user发表过的所有的视频
func (v *videoDao) GetPublishListInfo(userId int64) ([]*model.Video, error) {
	videoInfos := make([]*model.Video, 0)
	if err := db.Where("user_id = ?", userId).Find(&videoInfos).Error; err != nil {
		return nil, constants.InnerDataBaseErr
	}
	return videoInfos, nil
}

// GetFeedList 在数据库中得到时间戳在latestTime前的一系列视频
func (v *videoDao) GetFeedList(latestTime time.Time) ([]*model.Video, error) {
	videoInfos := make([]*model.Video, 0)
	if err := db.Where("created_at < ?", latestTime).
		Order("created_at desc").Limit(initialization.FeedListLength).Find(&videoInfos).Error; err != nil {
		if err != nil {
			return nil, constants.InnerDataBaseErr

		} else if 0 == len(videoInfos) {
			return nil, constants.RecordNotExistErr
		}
	}
	return videoInfos, nil
}

// GetVideoByVideoIdInfo 通过VideoId查找Video
func (v *videoDao) GetVideoByVideoIdInfo(videoId int64) (*model.Video, error) {
	videoInfos := make([]*model.Video, 0)
	if err := db.Where("video_id = ?", videoId).Find(&videoInfos).Error; err != nil {
		if err != nil || 1 < len(videoInfos) {
			return nil, constants.InnerDataBaseErr
		} else if 0 == len(videoInfos) {
			return nil, constants.RecordNotExistErr
		}
	}
	return videoInfos[0], nil
}
