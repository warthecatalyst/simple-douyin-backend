package jwt

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"time"
)

func GetUserId(c context.Context, ctx *app.RequestContext) (int64, error) {
	user, exists := ctx.Get(IdentityKey)
	if !exists {
		return 0, constants.InvalidTokenErr
	}

	loginUserInfo := user.(*model.User)
	logger.GlobalLogger.Printf("Time = %v, In UserInfo, Got Login Username =%v", time.Now(), loginUserInfo.UserName)
	loginUserInfo, err := service.GetUserServiceInstance().GetUserByUserName(loginUserInfo.UserName)
	return loginUserInfo.UserID, err
}
