package controller

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbuser "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
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

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin
var usersLoginInfo = map[string]api.User{
	"zhangleidouyin": {
		Id:            1,
		Name:          "zhanglei",
		FollowCount:   10,
		FollowerCount: 5,
		IsFollow:      true,
	},
}

func Register(c context.Context, ctx *app.RequestContext) {
	var err error
	var user jwt.UserStruct
	if err = ctx.BindAndValidate(&user); err != nil {
		ctx.JSON(consts.StatusOK, api.UserLoginResponse{
			Response: api.Response{
				StatusCode: int32(api.InputFormatCheckErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
			},
		})
	}

	_, err = service.GetUserServiceInstance().UserRegisterInfo(user.Username, user.Password)
	if err != nil {
		if errors.Is(errors.New(api.ErrorCodeToMsg[api.UserAlreadyExistErr]), err) {
			ctx.JSON(consts.StatusOK, api.UserLoginResponse{
				Response: api.Response{
					StatusCode: int32(api.UserAlreadyExistErr),
					StatusMsg:  api.ErrorCodeToMsg[api.UserAlreadyExistErr],
				},
			})
		}
	}

	jwt.JwtMiddleware.LoginHandler(c, ctx)
}

// Register_RPC 处理用户登录请求的RPC远程调用
func Register_RPC(content context.Context, requestContext *app.RequestContext) {
	var err error
	var user jwt.UserStruct
	if err = requestContext.BindAndValidate(&user); err != nil {
		requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
			Response: api.Response{
				StatusCode: int32(api.InputFormatCheckErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
			},
		})
	}
	address := initialization.RpcCSConf.Host + initialization.RpcCSConf.UserServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pbuser.NewUserInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := grpcClient.UserRegister(ctx, &pbuser.UserPost{
		Username: user.Username,
		Password: user.Password,
	})
	if err != nil {
		logger.GlobalLogger.Printf("Failed to remotely access UserService")
	}

	if result.BaseResp.StatusCode != 0 {
		switch result.BaseResp.StatusCode {
		case int32(api.UserAlreadyExistErr):
			requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
				Response: api.Response{
					StatusCode: int32(api.UserAlreadyExistErr),
					StatusMsg:  api.ErrorCodeToMsg[api.UserAlreadyExistErr],
				},
			})
		}
	}
	jwt.JwtMiddleware.LoginHandler(content, requestContext)
}

func UserInfo(c context.Context, ctx *app.RequestContext) {
	var err error
	_, err = jwt.GetUserId(c, ctx)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.UserResponse{
			Response: api.Response{
				StatusCode: int32(api.TokenInvalidErr),
				StatusMsg:  api.ErrorCodeToMsg[api.TokenInvalidErr],
			},
		})
		return
	}

	userIdStr := ctx.Query("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		ctx.JSON(consts.StatusOK, api.UserResponse{
			Response: api.Response{
				StatusCode: int32(api.InputFormatCheckErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InputFormatCheckErr],
			},
		})
		return
	}

	queryUser, err := service.GetUserServiceInstance().GetUserByUserId(userId)
	if errors.Is(constants.UserNotExistErr, err) {
		ctx.JSON(consts.StatusOK, api.UserResponse{
			Response: api.Response{
				StatusCode: int32(api.UserNotExistErr),
				StatusMsg:  api.ErrorCodeToMsg[api.UserNotExistErr],
			},
		})
		return
	}

	ctx.JSON(consts.StatusOK, api.UserResponse{
		Response: api.Response{
			StatusCode: 0,
			StatusMsg:  "",
		},
		User: api.User{
			Id:            queryUser.UserID,
			Name:          queryUser.UserName,
			FollowCount:   queryUser.FollowCount,
			FollowerCount: queryUser.FollowerCount,
			IsFollow:      false,
		},
	})
}
