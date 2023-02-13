package dao

import (
	"errors"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"gorm.io/gorm"
	"sync"
	"time"
)

//videoDao 与video相关的数据库操作集合
type videoDao struct{}

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

// CreateVideo 在数据库中通过事务插入一条Video数据
func (v *videoDao) CreateVideo(video *model.Video) error {
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

// GetPublishList 在数据库中获得该user发表过的所有的视频
func (v *videoDao) GetPublishList(userId int64) ([]*model.Video, error) {
	videoInfos := make([]*model.Video, 0)
	if err := db.Where("user_id = ?", userId).Find(&videoInfos).Error; err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return nil, constants.RecordNotExistErr
		} else {
			return nil, constants.InnerDataBaseErr
		}
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

// GetVideoByVideoId 通过VideoId查找Video
func (v *videoDao) GetVideoByVideoId(videoId int64) (*model.Video, error) {
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
