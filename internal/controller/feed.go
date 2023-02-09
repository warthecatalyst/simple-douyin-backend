package controller

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"net/http"
	"strconv"
	"time"
)

type FeedResponse struct {
	api.Response
	VideoList []api.Video `json:"video_list,omitempty"`
	NextTime  int64       `json:"next_time,omitempty"`
}

// Feed 推送视频流
func Feed(c context.Context, ctx *app.RequestContext) {
	token := ctx.Query("token")
	var userId int64
	var err error
	if token != "" {
		jwt.JwtMiddleware.MiddlewareFunc()(c, ctx)
		userId, err = jwt.GetUserId(c, ctx)
		if err != nil {
			logger.GlobalLogger.Printf("Time = %v ,can't get user From token", time.Now())
			if errors.Is(constants.InvalidTokenErr, err) {
				ctx.JSON(consts.StatusOK, api.Response{
					StatusCode: int32(api.TokenInvalidErr),
					StatusMsg:  api.ErrorCodeToMsg[api.TokenInvalidErr],
				})
			} else {
				ctx.JSON(consts.StatusOK, api.Response{
					StatusCode: int32(api.InnerDataBaseErr),
					StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
				})
			}
			return
		}
	}
	latestTimeStr := ctx.Query("latest_time")
	logger.GlobalLogger.Printf("Time = %v, latestTime = %v", time.Now(), latestTimeStr)
	var latestTime time.Time
	if latestTimeStr == "" {
		latestTime = time.Now()
	} else {
		latestTimeInt, err := strconv.ParseInt(latestTimeStr, 10, 64)
		if err != nil {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.InputFormatCheckErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
			})
			return
		}
		latestTime = time.UnixMilli(latestTimeInt)
	}

	nextTime, videoList, err := service.GetFeedServiceInstance().Feed(userId, latestTime)
	if err != nil {
		if errors.Is(constants.RecordNotExistErr, err) {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.RecordNotExistErr),
				StatusMsg:  api.ErrorCodeToMsg[api.RecordNotExistErr],
			})
		} else if errors.Is(constants.NoVideoErr, err) {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.NoVideoErr),
				StatusMsg:  api.ErrorCodeToMsg[api.NoVideoErr],
			})
		} else {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.InnerDataBaseErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, FeedResponse{
		Response:  api.Response{StatusCode: 0},
		VideoList: videoList,
		NextTime:  nextTime,
	})
}
