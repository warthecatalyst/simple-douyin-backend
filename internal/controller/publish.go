package controller

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbservice "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/video"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"time"
)

type VideoListResponse struct {
	api.Response
	VideoList []api.Video `json:"video_list"`
}

// Publish check token then save upload file to public directory
func Publish(ctx context.Context, requestContext *app.RequestContext) {
	userId, err := jwt.GetUserId(ctx, requestContext)
	if err != nil {
		requestContext.JSON(consts.StatusOK, api.UserResponse{
			Response: api.Response{
				StatusCode: int32(api.TokenInvalidErr),
				StatusMsg:  api.ErrorCodeToMsg[api.TokenInvalidErr],
			},
		})
		return
	}

	logger.GlobalLogger.Printf("Time = %v,get User From loginUser = %v", time.Now(), userId)
	data, err := requestContext.FormFile("data")
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v,can't get Video Data from post", time.Now())
		requestContext.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.GetDataErr),
			StatusMsg:  api.ErrorCodeToMsg[api.GetDataErr],
		})
		return
	}
	title := requestContext.Query("title")
	address := initialization.RpcCSConf.VideoServiceHost + initialization.RpcCSConf.VideoServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbservice.NewVideoServiceInfoClient(conn)

	content := make([]byte, data.Size)
	src, err := data.Open()
	defer src.Close()
	_, err = src.Read(content)
	if err != nil {
		requestContext.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.UploadFailErr),
			StatusMsg:  api.ErrorCodeToMsg[api.UploadFailErr],
		})
		return
	}

	// Contact the server and print out its response.
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	_, err = c.PublishVideoInfo(ctx1, &pbservice.VideoServicePost{
		UserId:   userId,
		Title:    title,
		FileName: data.Filename,
		FileSize: data.Size,
		Content:  content,
	})
	if err != nil {
		requestContext.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.UploadFailErr),
			StatusMsg:  api.ErrorCodeToMsg[api.UploadFailErr],
		})
		return
	}
	requestContext.JSON(consts.StatusOK, api.Response{
		StatusCode: 0,
	})
}

// PublishList all users have same publish video list
func PublishList(c context.Context, ctx *app.RequestContext) {
	loginUserId, err := jwt.GetUserId(c, ctx)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v,can't get user From token", time.Now())
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

	userStr := ctx.Query("user_id")
	userId, err := strconv.ParseInt(userStr, 10, 64)
	logger.GlobalLogger.Printf("userId = %v", userId)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.Response{
			StatusCode: int32(api.InputFormatCheckErr),
			StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
		})
		return
	}

	videoList, err := service.GetVideoServiceInstance().PublishListInfo(userId, loginUserId)
	ctx.JSON(consts.StatusOK, VideoListResponse{
		Response: api.Response{
			StatusCode: 0,
		},
		VideoList: videoList,
	})
}
