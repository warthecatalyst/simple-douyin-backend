package jwt

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbuser "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"time"
)

var (
	JwtMiddleware *jwt.HertzJWTMiddleware
	IdentityKey   = "identity"
)

type UserStruct struct {
	Username string `form:"username" json:"username" query:"username" vd:"(len($) > 0 && len($) < 128); msg:'Illegal format'"`
	Password string `form:"password" json:"password" query:"password" vd:"(len($) > 0 && len($) < 128); msg:'Illegal format'"`
}

func LoginResponse(content context.Context, requestContext *app.RequestContext, token string) {
	username := requestContext.Query("username")
	password := requestContext.Query("password")

	address := initialization.RpcCSConf.Host + initialization.RpcCSConf.UserServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pbuser.NewUserInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userInfoResp, err := grpcClient.GetUserInfo(ctx, &pbuser.UserPost{
		Username: username,
		Password: password,
	})

	if err != nil {
		logger.GlobalLogger.Printf("Can't get RPC From UserService")
	}

	if userInfoResp.BaseResp.StatusCode != 0 {
		requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
			Response: api.Response{
				StatusCode: int32(api.UserNotExistErr),
				StatusMsg:  api.ErrorCodeToMsg[api.UserNotExistErr],
			},
		})
	}

	requestContext.JSON(consts.StatusOK, api.UserLoginResponse{
		Response: api.Response{
			StatusCode: 0,
		},
		UserId: userInfoResp.Id,
		Token:  token,
	})
}

func InitJwt() {
	var err error
	JwtMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{
		Realm:         "test zone",
		Key:           []byte("secret key"),
		Timeout:       time.Hour,
		MaxRefresh:    time.Hour,
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			LoginResponse(ctx, c, token)
		},
		Authenticator: func(content context.Context, requestContext *app.RequestContext) (interface{}, error) {
			var userStruct UserStruct
			if err := requestContext.BindAndValidate(&userStruct); err != nil {
				return nil, err
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
			userInfoResp, err := grpcClient.GetUserInfo(ctx, &pbuser.UserPost{
				Username: userStruct.Username,
				Password: userStruct.Password,
			})
			if err != nil {
				return nil, err
			}
			userInfo := &model.User{
				UserID:        userInfoResp.Id,
				UserName:      userInfoResp.Name,
				FollowCount:   userInfoResp.FollowCnt,
				FollowerCount: userInfoResp.FollowerCnt,
			}
			if err != nil {
				return nil, err
			}
			return userInfo, nil
		},
		IdentityKey: IdentityKey,
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			return &model.User{
				UserName: claims[IdentityKey].(string),
			}
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					IdentityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		HTTPStatusMessageFunc: func(e error, ctx context.Context, c *app.RequestContext) string {
			hlog.CtxErrorf(ctx, "jwt biz err = %+v", e.Error())
			return e.Error()
		},
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(http.StatusOK, utils.H{
				"code":    code,
				"message": message,
			})
		},
	})

	if err != nil {
		logger.GlobalLogger.Error().Str("JWT初始化错误", err.Error())
	}
}
