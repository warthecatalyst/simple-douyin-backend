package dao

import (
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"gorm.io/gorm"
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
	dataBaseInitialization()
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

//Add 向数据库中插入一条点赞记录，将已有记录设置为1
func (f *favoriteDao) Add(userId, videoId int64) error {
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

//Del 从数据库中删除一条点赞记录，软删除，将点赞的记录设置为0
func (f *favoriteDao) Del(userId, videoId int64) error {
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
		videos[i], err = GetVideoDaoInstance().GetVideoByVideoId(fav.VideoID)
		if err != nil {
			return nil, err
		}
	}
	return videos, nil
}

// CheckFavorite 查看一个用户是否点赞过一个视频
func (f favoriteDao) CheckFavorite(userId, videoId int64) (bool, error) {
	var favor model.Favourite
	err := db.Where("video_id = ? And user_id = ? And is_favor = ?", videoId, userId, 1).First(&favor).Error
	if errors.Is(gorm.ErrRecordNotFound, err) {
		return false, nil
	} else if err != nil {
		return false, constants.InnerDataBaseErr
	}
	return true, nil
}
