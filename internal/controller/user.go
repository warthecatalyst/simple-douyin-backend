package controller

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"strconv"
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
