package service

import (
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"strconv"
)

//通过model.Video构造api.Video切片, userId是当前登录的userId
func getVideoListByModel(userId int64, videos []*model.Video) ([]api.Video, error) {
	videoList := make([]api.Video, len(videos))
	for i, v := range videos {
		userInfo, err := GetUserServiceInstance().getUserByUserId(v.UserID)
		isFavor, err := dao.GetFavoriteDaoInstance().CheckFavorite(userId, v.VideoID)
		if err != nil {
			return nil, constants.InnerDataBaseErr
		}
		videoList[i] = api.Video{
			Id: v.VideoID,
			Author: api.User{
				Id:            userInfo.UserID,
				Name:          userInfo.UserName,
				FollowCount:   userInfo.FollowCount,
				FollowerCount: userInfo.FollowerCount,
				IsFollow:      false,
			},
			PlayUrl:       v.PlayURL,
			CoverUrl:      v.CoverURL,
			FavoriteCount: int64(v.FavoriteCount),
			CommentCount:  int64(v.CommentCount),
			IsFavorite:    isFavor,
		}
	}
	return videoList, nil
}

func getVideoListByID(userId int64, videoIds []string) ([]api.Video, error) {
	videoList := make([]api.Video, len(videoIds))
	for i, videoIdstr := range videoIds {
		videoId, _ := strconv.ParseInt(videoIdstr, 10, 64)
		videoInfo, err := dao.GetVideoDaoInstance().GetVideoByVideoIdInfo(videoId)
		userInfo, err := GetUserServiceInstance().getUserByUserId(userId)
		isFavor, err := dao.GetFavoriteDaoInstance().CheckFavorite(userId, videoId)
		if err != nil {
			return nil, constants.InnerDataBaseErr
		}
		videoList[i] = api.Video{
			Id: videoId,
			Author: api.User{
				Id:            userInfo.UserID,
				Name:          userInfo.UserName,
				FollowCount:   userInfo.FollowCount,
				FollowerCount: userInfo.FollowerCount,
				IsFollow:      false,
			},
			PlayUrl:       videoInfo.PlayURL,
			CoverUrl:      videoInfo.CoverURL,
			FavoriteCount: int64(videoInfo.FavoriteCount),
			CommentCount:  int64(videoInfo.CommentCount),
			IsFavorite:    isFavor,
		}
	}
	return videoList, nil
}
