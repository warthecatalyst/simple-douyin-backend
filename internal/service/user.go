package service

import (
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/idGenerator"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/md5"
	"github.com/rs/zerolog/log"
)

// UserService 与用户相关的操作使用的结构体
type UserService struct{}

var (
	userService = &UserService{}
)

func GetUserServiceInstance() *UserService {
	return userService
}

func (u *UserService) UserRegisterInfo(username, password string) error {
	var err error
	userInfo, err := dao.NewUserDaoInstance().GetUserByUsername(username)

	if err != nil {
		initialization.StdOutLogger.Error().Caller().Str("UserRegisterInfoError", err.Error())
		return err
	}

	if userInfo != nil {
		return errors.New(api.ErrorCodeToMsg[api.UserAlreadyExistErr])
	}

	userId := idGenerator.GenerateUserId()
	initialization.StdOutLogger.Info().Int64("userId ", userId)

	user := &model.User{
		UserID:   userId,
		UserName: username,
	}

	if initialization.UserConf.PasswordEncrpted {
		user.PassWord = md5.MD5(password)
	} else {
		user.PassWord = password
	}
	err = dao.NewUserDaoInstance().CreateUser(user)

	if err != nil {
		log.Error().Caller().Str("UserDaoError", err.Error())
	}
	return nil
}

func (u *UserService) CheckUserInfo(username, password string) (*model.User, error) {
	userInfo, err := dao.NewUserDaoInstance().CheckUserByNameAndPassword(username, password)

	if err != nil {
		initialization.StdOutLogger.Error().Caller().Str("CheckUserInfoError", err.Error())
		return nil, err
	}

	return userInfo, nil
}
