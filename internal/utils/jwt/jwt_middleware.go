package jwt

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/jwt"
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

func LoginResponse(c context.Context, ctx *app.RequestContext, token string) {
	username := ctx.Query("username")
	password := ctx.Query("password")

	userInfo, err := service.GetUserServiceInstance().CheckUserInfo(username, password)

	if err != nil {
		ctx.JSON(consts.StatusOK, api.UserLoginResponse{
			Response: api.Response{
				StatusCode: int32(api.UserNotExistErr),
				StatusMsg:  api.ErrorCodeToMsg[api.UserNotExistErr],
			},
		})
	}

	ctx.JSON(consts.StatusOK, api.UserLoginResponse{
		Response: api.Response{},
		UserId:   userInfo.UserID,
		Token:    token,
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
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var userStruct UserStruct
			if err := c.BindAndValidate(&userStruct); err != nil {
				return nil, err
			}
			userInfo, err := service.GetUserServiceInstance().CheckUserInfo(userStruct.Username, userStruct.Password)

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
