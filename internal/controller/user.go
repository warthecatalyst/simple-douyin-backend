package controller

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbuser "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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

// Register 处理用户登录请求的RPC远程调用
func Register(content context.Context, requestContext *app.RequestContext) {
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
	address := initialization.RpcCSConf.UserServiceHost + initialization.RpcCSConf.UserServicePort
	logger.GlobalLogger.Printf("address = %v", address)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pbuser.NewUserServiceInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = grpcClient.UserRegister(ctx, &pbuser.UserServicePost{
		Username: user.Username,
		Password: user.Password,
	})
	if err != nil {
		logger.GlobalLogger.Printf("Got error from RPC ,error = %v", err)
		if errors.Is(status.Errorf(codes.AlreadyExists, constants.UserAlreadyExistErr.Error()), err) {
			requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
				Response: api.Response{
					StatusCode: int32(api.UserAlreadyExistErr),
					StatusMsg:  api.ErrorCodeToMsg[api.UserAlreadyExistErr],
				},
			})
		} else if errors.Is(status.Errorf(codes.Internal, constants.InnerDataBaseErr.Error()), err) {
			requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
				Response: api.Response{
					StatusCode: int32(api.InnerDataBaseErr),
					StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
				},
			})
		} else {
			requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
				Response: api.Response{
					StatusCode: int32(api.InnerConnectionErr),
					StatusMsg:  api.ErrorCodeToMsg[api.InnerConnectionErr],
				},
			})
		}
		return
	}
	jwt.JwtMiddleware.LoginHandler(content, requestContext)
}

func UserInfo(content context.Context, requestContext *app.RequestContext) {
	var err error
	loginUserId, err := jwt.GetUserId(content, requestContext)
	if err != nil {
		requestContext.JSON(consts.StatusOK, api.UserResponse{
			Response: api.Response{
				StatusCode: int32(api.TokenInvalidErr),
				StatusMsg:  api.ErrorCodeToMsg[api.TokenInvalidErr],
			},
		})
		return
	}
	userIdStr := requestContext.Query("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	address := initialization.RpcCSConf.UserServiceHost + initialization.RpcCSConf.UserServicePort
	logger.GlobalLogger.Printf("address = %v", address)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pbuser.NewUserServiceInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := grpcClient.GetUserInfo(ctx, &pbuser.UserServicePost{
		QueryUserId: userId,
		LoginUserId: loginUserId,
	})
	if err != nil {
		logger.GlobalLogger.Printf("Failed to remotely access UserService ,error = %v", err)
		requestContext.JSON(consts.StatusOK, api.UserResponse{
			Response: api.Response{
				StatusCode: int32(api.InnerConnectionErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InnerConnectionErr],
			},
		})
		return
	}
	requestContext.JSON(consts.StatusOK, api.UserResponse{
		Response: api.Response{
			StatusCode: 0,
			StatusMsg:  "",
		},
		User: api.User{
			Id:            result.Id,
			Name:          result.Name,
			FollowCount:   result.FollowCnt,
			FollowerCount: result.FollowerCnt,
			IsFollow:      result.IsFollow,
		},
	})
}
