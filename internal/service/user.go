package service

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pb "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
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
type userService struct {
	pb.UnimplementedUserInfoServer
}

var (
	userServiceInstance *userService
	userOnce            sync.Once
)

// GetUserServiceInstance 单例模式，获得一个userService的实例
func GetUserServiceInstance() *userService {
	initRedis()
	initKafka()
	userOnce.Do(func() {
		userServiceInstance = &userService{}
	})
	return userServiceInstance
}

// UserRegisterAction RPC对用户注册的请求
func (u *userService) UserRegisterAction(ctx context.Context, in *pb.UserPost) (*pb.UserResp, error) {
	username := in.Username
	password := in.Password
	userInfo, err := u.UserRegisterInfo(username, password)
	if err != nil {
		if errors.Is(constants.UserAlreadyExistErr, err) {
			return &pb.UserResp{
				UserId: 0,
				Token:  "",
				BaseResp: &pb.BaseResp{
					StatusCode: int32(api.UserAlreadyExistErr),
					StatusMsg:  api.ErrorCodeToMsg[api.UserAlreadyExistErr],
				},
			}, nil
		} else {
			return &pb.UserResp{
				UserId: 0,
				Token:  "",
				BaseResp: &pb.BaseResp{
					StatusCode: int32(api.InnerDataBaseErr),
					StatusMsg:  err.Error(),
				},
			}, err
		}
	}
	return &pb.UserResp{
		UserId: userInfo.UserID,
		Token:  "",
		BaseResp: &pb.BaseResp{
			StatusCode: 0,
			StatusMsg:  "",
		},
	}, nil
}

// GetUserInfo RPC获取用户信息
func (u *userService) GetUserInfo(ctx context.Context, in *pb.UserPost) (*pb.UserInfoResp, error) {
	username := in.Username
	password := in.Password
	userInfo, err := u.CheckUserInfo(username, password)
	if err != nil {
		if errors.Is(constants.UserNotExistErr, err) {
			return &pb.UserInfoResp{
				BaseResp: &pb.BaseResp{
					StatusCode: int32(api.UserNotExistErr),
					StatusMsg:  api.ErrorCodeToMsg[api.UserNotExistErr],
				},
			}, err
		} else {
			return &pb.UserInfoResp{BaseResp: &pb.BaseResp{
				StatusCode: int32(api.InnerDataBaseErr),
				StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
			}}, err
		}
	}
	return &pb.UserInfoResp{
		Id:          userInfo.UserID,
		Name:        userInfo.UserName,
		FollowCnt:   userInfo.FollowCount,
		FollowerCnt: userInfo.FollowerCount,
		BaseResp:    &pb.BaseResp{
			StatusCode: 0,
			StatusMsg:  "",
		},
	},nil
}

// UserRegisterInfo 用户登录请求的内部处理逻辑
func (u *userService) UserRegisterInfo(username, password string) (*model.User, error) {
	var err error
	userInfo, err := dao.GetUserDaoInstance().GetUserByUsername(username)

	if errors.Is(constants.InnerDataBaseErr, err) {
		logger.GlobalLogger.Error().Caller().Str("用户注册失败", err.Error())
		return nil, err
	}

	if userInfo != nil {
		logger.GlobalLogger.Error().Caller().Str("用户名已存在", err.Error())
		return nil, constants.UserAlreadyExistErr
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
		return nil, constants.CreateDataErr
	}
	return user, nil
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
