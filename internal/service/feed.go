package service

import (
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"sync"
	"time"
)

type feedService struct{}

var (
	feedServiceInstance *feedService
	feedOnce            sync.Once
)

func GetFeedServiceInstance() *feedService {
	feedOnce.Do(func() {
		feedServiceInstance = &feedService{}
	})
	return feedServiceInstance
}

//Feed service层获取视频流
func (f *feedService) Feed(userId int64, latestTime time.Time) (int64, []api.Video, error) {
	videos, err := dao.GetVideoDaoInstance().GetFeedList(latestTime)
	logger.GlobalLogger.Printf("get Videos From FeedList")
	if err != nil {
		logger.GlobalLogger.Printf("dao.NewVideoDaoInstance().GetLatest error: %s", err)
		return -1, nil, err
	}
	if len(videos) == 0 {
		logger.GlobalLogger.Printf("没有早于当前时间的视频")
		return -1, nil, constants.NoVideoErr
	}
	videoList, err := newVideoList(userId, videos)
	if err != nil {
		return -1, nil, err
	}
	return videos[len(videos)-1].CreatedAt.Unix(), videoList, nil
}
