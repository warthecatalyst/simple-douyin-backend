package jwt

import (
	"context"
	pbuser "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

func GetUserId(content context.Context, requestContext *app.RequestContext) (int64, error) {
	user, exists := requestContext.Get(IdentityKey)
	if !exists {
		return 0, constants.InvalidTokenErr
	}

	loginUserInfo := user.(*model.User)
	logger.GlobalLogger.Printf("Time = %v, In GetUserId, Got Login Username =%v", time.Now(), loginUserInfo.UserName)
	address := initialization.RpcCSConf.UserServiceHost + initialization.RpcCSConf.UserServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pbuser.NewUserServiceInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userResp, err := grpcClient.GetUserIdByUserName(ctx, &pbuser.UserServicePost{
		Username: loginUserInfo.UserName,
	})

	if err != nil {
		logger.GlobalLogger.Printf("Can't get RPC From UserService, err = %v", err)
		return 0, err
	}
	logger.GlobalLogger.Printf("Time = %v, In GetUserId, Got Login UserId =%v", time.Now(), userResp.UserId)
	return userResp.UserId, err
}
