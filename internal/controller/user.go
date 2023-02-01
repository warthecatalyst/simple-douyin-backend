package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
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

	err = service.GetUserServiceInstance().UserRegisterInfo(user.Username, user.Password)
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

func UserInfo(c context.Context, ctx *app.RequestContext) {
	user, _ := ctx.Get(jwt.IdentityKey)
	ctx.JSON(consts.StatusOK, utils.H{
		"message": fmt.Sprintf("username:%v", user.(*model.User).UserName),
	})
}
