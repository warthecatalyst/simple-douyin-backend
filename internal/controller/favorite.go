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
	"strconv"
)

// FavoriteAction 视频点赞接口
func FavoriteAction(c context.Context, ctx *app.RequestContext) {
	var err error
	loginUserId, err := jwt.GetUserId(c, ctx)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.TokenInvalidErr),
			StatusMsg:  api.ErrorCodeToMsg[api.TokenInvalidErr],
		})
		return
	}
	videoIdStr := ctx.Query("video_id")
	videoId, err := strconv.ParseInt(videoIdStr, 10, 64)
	actionTypeStr := ctx.Query("action_type")
	actionType, err := strconv.ParseInt(actionTypeStr, 10, 32)
	logger.GlobalLogger.Printf("videoId = %v, actionType = %v", videoId, actionType)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.InputFormatCheckErr),
			StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
		})
		return
	}
	err = service.GetFavoriteServiceInstance().FavoriteInfo(loginUserId, videoId, int(actionType))
	if err != nil {
		if errors.Is(constants.RecordNotExistErr, err) {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.RecordNotExistErr),
				StatusMsg:  api.ErrorCodeToMsg[api.RecordNotExistErr],
			})
		} else if errors.Is(constants.RecordNotMatchErr, err) {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.RecordNotMatchErr),
				StatusMsg:  api.ErrorCodeToMsg[api.RecordNotMatchErr],
			})
		} else if errors.Is(constants.InnerDataBaseErr, err) {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.InnerDataBaseErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
			})
		} else {
			ctx.JSON(consts.StatusOK, api.Response{
				StatusCode: int32(api.UnKnownActionType),
				StatusMsg:  api.ErrorCodeToMsg[api.UnKnownActionType],
			})
		}
		return
	}
	ctx.JSON(consts.StatusOK, api.Response{
		StatusCode: 0,
	})
}

// FavoriteList 用户的喜欢列表
func FavoriteList(c context.Context, ctx *app.RequestContext) {
	var err error
	loginUserId, err := jwt.GetUserId(c, ctx)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.TokenInvalidErr),
			StatusMsg:  api.ErrorCodeToMsg[api.TokenInvalidErr],
		})
		return
	}
	userIdStr := ctx.Query("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.InputFormatCheckErr),
			StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
		})
		return
	}
	videoList, err := service.GetFavoriteServiceInstance().FavoriteListInfo(loginUserId, userId)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.InnerDataBaseErr),
			StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
		})
		return
	}
	ctx.JSON(consts.StatusOK, VideoListResponse{
		Response:  api.Response{StatusCode: 0},
		VideoList: *videoList,
	})
}
