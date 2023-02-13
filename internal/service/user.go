package service

import (
	"errors"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/idGenerator"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/md5"
	"github.com/rs/zerolog/log"
	"sync"
)

// userService 与用户相关的操作使用的结构体
type userService struct{}

var (
	userServiceInstance *userService
	userOnce            sync.Once
)

func GetUserServiceInstance() *userService {
	userOnce.Do(func() {
		userServiceInstance = &userService{}
	})
	return userServiceInstance
}

func (u *userService) UserRegisterInfo(username, password string) error {
	var err error
	userInfo, err := dao.GetUserDaoInstance().GetUserByUsername(username)

	if errors.Is(constants.InnerDataBaseErr, err) {
		logger.GlobalLogger.Error().Caller().Str("用户注册失败", err.Error())
		return err
	}

	if userInfo != nil {
		logger.GlobalLogger.Error().Caller().Str("用户名已存在", err.Error())
		return constants.UserAlreadyExistErr
	}

	userId := idGenerator.GenerateUserId()
	logger.GlobalLogger.Info().Int64("userId ", userId)

	user := &model.User{
		UserID:   userId,
		UserName: username,
	}

	if initialization.UserConf.PasswordEncrypted {
		user.PassWord = md5.MD5(password)
	} else {
		user.PassWord = password
	}
	err = dao.GetUserDaoInstance().CreateUser(user)

	if err != nil {
		log.Error().Caller().Str("用户注册错误", err.Error())
		return constants.CreateDataErr
	}
	return nil
}

//CheckUserInfo 从username,password获得User
func (u *userService) CheckUserInfo(username, password string) (*model.User, error) {
	userInfo, err := dao.GetUserDaoInstance().CheckUserByNameAndPassword(username, password)

	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}

	return userInfo, nil
}

// GetUserByUserId 通过userid得到user
func (u *userService) GetUserByUserId(userId int64) (*model.User, error) {
	userInfo, err := dao.GetUserDaoInstance().GetUserByUserId(userId)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}
	return userInfo, nil
}

// GetUserByUserName 通过username得到user
func (u *userService) GetUserByUserName(username string) (*model.User, error) {
	userInfo, err := dao.GetUserDaoInstance().GetUserByUsername(username)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
	}
	return userInfo, err
}
