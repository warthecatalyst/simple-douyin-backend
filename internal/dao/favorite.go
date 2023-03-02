package dao

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"sync"
)

//userDao 与favorite相关的数据库操作
type favoriteDao struct{}

var (
	favoriteDaoInstance *favoriteDao
	favoriteOnce        sync.Once
)

// GetFavoriteDaoInstance 获取一个Dao层与Favorite操作有关的Instance
func GetFavoriteDaoInstance() *favoriteDao {
	favoriteOnce.Do(func() {
		favoriteDaoInstance = &favoriteDao{}
	})
	return favoriteDaoInstance
}

// GetFavoriteCount 通过videoId获取点赞数
func (f *favoriteDao) GetFavoriteCount(videoId int64) (int32, error) {
	var video model.Video
	if err := db.Where("video_id = ?", videoId).First(&video).Error; err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return 0, constants.RecordNotExistErr
		} else {
			return -1, constants.InnerDataBaseErr
		}
	}
	return video.FavoriteCount, nil
}

// SetFavoriteCount 通过videoId设置点赞数
func (f favoriteDao) SetFavoriteCount(videoId int64, favoriteCount int32) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Video{}).
			Where("video_id = ?", videoId).Update("favorite_count", favoriteCount).Error; err != nil {
			return constants.InnerDataBaseErr
		}
		return nil
	})
}

//FavoriteAction 向数据库中插入一条点赞记录，若已有点赞记录，将该点赞记录设置为1
func (f *favoriteDao) FavoriteAction(userId, videoId int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var err error
		var favor model.Favourite
		err = tx.Where("video_id = ? And user_id = ?", videoId, userId).First(&favor).Error
		if errors.Is(gorm.ErrRecordNotFound, err) {
			favor.UserID = userId
			favor.VideoID = videoId
			favor.IsFavor = 1
			if err = tx.Create(&favor).Error; err != nil {
				return constants.InnerDataBaseErr
			}
			return nil
		} else if err != nil {
			return constants.InnerDataBaseErr
		}
		if favor.IsFavor == 1 {
			return constants.RecordNotMatchErr
		}
		err = tx.Model(&favor).Update("is_favor", 1).Error
		if err != nil {
			return constants.InnerDataBaseErr
		}
		return nil
	})
}

//UnfavoriteAction 从数据库中软删除一条点赞记录，也即将点赞的记录设置为0
func (f *favoriteDao) UnfavoriteAction(userId, videoId int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var err error
		var favor model.Favourite
		err = tx.Where("video_id = ? And user_id = ?", videoId, userId).First(&favor).Error
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return constants.RecordNotExistErr
		} else if err != nil {
			return constants.InnerDataBaseErr
		}
		if favor.IsFavor == 0 {
			return constants.RecordNotMatchErr
		}
		err = tx.Model(&favor).Update("is_favor", 0).Error
		if err != nil {
			return constants.InnerDataBaseErr
		}
		return nil
	})
}

// GetFavoriteList 从数据库中获得userId点赞过的所有video
func (f *favoriteDao) GetFavoriteList(userId int64) ([]*model.Video, error) {
	favors := make([]*model.Favourite, 0)
	err := db.Where("user_id = ? And is_favor = ?", userId, 1).Find(&favors).Error
	if err != nil {
		return nil, constants.InnerDataBaseErr
	}
	n := len(favors)
	videos := make([]*model.Video, n)
	for i, fav := range favors {
		videos[i], err = GetVideoDaoInstance().GetVideoByVideoIdInfo(fav.VideoID)
		if err != nil {
			return nil, err
		}
	}
	return videos, nil
}

// CheckFavorite 查看一个用户是否点赞过一个视频
func (f *favoriteDao) CheckFavorite(userId, videoId int64) (bool, error) {
	var favor model.Favourite
	err := db.Where("video_id = ? And user_id = ? And is_favor = ?", videoId, userId, 1).First(&favor).Error
	if errors.Is(gorm.ErrRecordNotFound, err) {
		return false, nil
	} else if err != nil {
		return false, constants.InnerDataBaseErr
	}
	return true, nil
}

// HardDeleteUnFavorite 在数据库中删除所有软删除的点赞条目
func (f *favoriteDao) HardDeleteUnFavorite() error {
	err := db.Where("is_favor = ?", 0).Delete(&model.Favourite{}).Error
	if err != nil {
		return constants.InnerDataBaseErr
	}
	return nil
}

// getFromMessageQueue 从消息队列中异步获取点赞信息，然后将信息写入数据库
func (f *favoriteDao) getFromMessageQueue() error {
	partitionList, err := kafkaClient.Partitions(constants.KafkaTopicPrefix + "favorite") // 根据topic取到所有的分区
	if err != nil {
		logger.GlobalLogger.Printf("fail to get list of partition:err%v\n", err)
		return constants.KafkaClientErr
	}
	logger.GlobalLogger.Printf("partitionList = %v", partitionList)
	for _, partition := range partitionList { // 遍历所有的分区
		logger.GlobalLogger.Printf("partition = %v", partition)
		// 针对每个分区创建一个对应的分区消费者
		pc, err2 := kafkaClient.ConsumePartition(constants.KafkaTopicPrefix+"favorite", partition, sarama.OffsetNewest)
		if err2 != nil {
			logger.GlobalLogger.Printf("failed to start consumer for partition %d,err:%v\n", partition, err)
			return constants.KafkaClientErr
		}
		defer pc.AsyncClose()
		for msg := range pc.Messages() {
			// 异步从每个分区消费信息
			msg1 := msg
			go func() {
				logger.GlobalLogger.Print("Got messageFrom 消息队列")
				key := string(msg1.Key)
				value := string(msg1.Value)
				logger.GlobalLogger.Printf("Partition:%d Offset:%d Key:%v Value:%v\n", msg1.Partition, msg1.Offset, key, value)
				idx := strings.Index(value, ":")
				userId, _ := strconv.ParseInt(value[0:idx], 10, 64)
				videoId, _ := strconv.ParseInt(value[idx+1:], 10, 64)
				logger.GlobalLogger.Printf("userId:%d videoId:%d", userId, videoId)
				if key == "Favorite" {
					for {
						err1 := f.FavoriteAction(userId, videoId)
						if err1 == nil {
							break
						}
					}
				} else {
					for {
						err1 := f.UnfavoriteAction(userId, videoId)
						if err1 == nil {
							break
						}
					}
				}
			}()
		}
	}
	return nil
}
