package service

import (
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"sync"
)

type favoriteService struct{}

var (
	favoriteServiceInstance *favoriteService
	favoriteOnce            sync.Once
)

// GetFavoriteServiceInstance 获取一个favoriteService的实例
func GetFavoriteServiceInstance() *favoriteService {
	favoriteOnce.Do(func() {
		favoriteServiceInstance = &favoriteService{}
	})
	return favoriteServiceInstance
}

// FavoriteInfo service层处理用户点赞或者取消点赞
// 可能返回的错误类型：InnerDataBaseError, RecordNotMatch, RecordNotExist,UnknownActionTypeErr
func (f *favoriteService) FavoriteInfo(userId, videoId int64, actionType int) error {
	var err error
	var favoriteCount int32
	if actionType == api.FavoriteAction {
		//1、插入点赞信息
		//2、点赞数++
		//这边不能反过来!!!
		err = dao.GetFavoriteDaoInstance().Add(userId, videoId)
		if err != nil {
			return err
		}
		favoriteCount, err = dao.GetFavoriteDaoInstance().GetFavoriteCount(videoId)
		if err != nil {
			return err
		}
		// logger.GlobalLogger.Printf("favoriteCount = %v", favoriteCount)
		favoriteCount++
		// logger.GlobalLogger.Printf("favoriteCount = %v", favoriteCount)
		err = dao.GetFavoriteDaoInstance().SetFavoriteCount(videoId, favoriteCount)
	} else if actionType == api.UnFavoriteAction {
		//1、删除点赞信息
		//2、点赞数--
		//同样不能够反过来
		err = dao.GetFavoriteDaoInstance().Del(userId, videoId)
		if err != nil {
			return err
		}
		favoriteCount, err = dao.GetFavoriteDaoInstance().GetFavoriteCount(videoId)
		if err != nil {
			return err
		}
		favoriteCount--
		err = dao.GetFavoriteDaoInstance().SetFavoriteCount(videoId, favoriteCount)
	} else {
		return constants.UnKnownActionTypeErr
	}
	return err
}

//FavoriteListInfo service层查找用户点赞过的所有视频
func (f favoriteService) FavoriteListInfo(loginUserId, userId int64) (*[]api.Video, error) {
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
