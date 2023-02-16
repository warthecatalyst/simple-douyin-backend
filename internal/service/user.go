package service

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbservice "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
	pbdao "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/user"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/idGenerator"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/md5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

// userService 与用户相关的操作使用的结构体
type userService struct {
	pbservice.UnimplementedUserServiceInfoServer
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

// UserRegister RPC对用户注册的请求
func (u *userService) UserRegister(ctx context.Context, in *pbservice.UserServicePost) (*pbservice.UserServiceResp, error) {
	username := in.Username
	password := in.Password
	logger.GlobalLogger.Printf("username = %v, password = %v", username, password)
	userInfo, err := u.userRegisterInfo(username, password)
	if err != nil {
		if errors.Is(constants.UserAlreadyExistErr, err) {
			return &pbservice.UserServiceResp{
				UserId: 0,
				Token:  "",
				BaseResp: &pbservice.BaseServiceResp{
					StatusCode: int32(api.UserAlreadyExistErr),
					StatusMsg:  api.ErrorCodeToMsg[api.UserAlreadyExistErr],
				},
			}, nil
		} else {
			return &pbservice.UserServiceResp{
				UserId: 0,
				Token:  "",
				BaseResp: &pbservice.BaseServiceResp{
					StatusCode: int32(api.InnerDataBaseErr),
					StatusMsg:  err.Error(),
				},
			}, err
		}
	}
	return &pbservice.UserServiceResp{
		UserId: userInfo.UserID,
		Token:  "",
		BaseResp: &pbservice.BaseServiceResp{
			StatusCode: 0,
			StatusMsg:  "",
		},
	}, nil
}

// GetUserInfo RPC获取用户信息, 有两种方式,一种为通过username和password,另一种为通过queryUserId
func (u *userService) GetUserInfo(ctx context.Context, in *pbservice.UserServicePost) (*pbservice.UserServiceInfoResp, error) {
	username := in.Username
	password := in.Password
	queryUserId := in.QueryUserId
	if "" != username && "" != password {
		userInfo, err := u.checkUserInfo(username, password)
		if err != nil {
			if errors.Is(constants.UserNotExistErr, err) {
				return &pbservice.UserServiceInfoResp{
					BaseResp: &pbservice.BaseServiceResp{
						StatusCode: int32(api.UserNotExistErr),
						StatusMsg:  api.ErrorCodeToMsg[api.UserNotExistErr],
					},
				}, err
			} else {
				return &pbservice.UserServiceInfoResp{BaseResp: &pbservice.BaseServiceResp{
					StatusCode: int32(api.InnerDataBaseErr),
					StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
				}}, err
			}
		}
		return &pbservice.UserServiceInfoResp{
			Id:          userInfo.UserID,
			Name:        userInfo.UserName,
			FollowCnt:   userInfo.FollowCount,
			FollowerCnt: userInfo.FollowerCount,
			IsFollow:    false,
			BaseResp: &pbservice.BaseServiceResp{
				StatusCode: 0,
				StatusMsg:  "",
			},
		}, nil
	} else {
		userInfo, err := u.getUserByUserId(queryUserId)
		if err != nil {
			if errors.Is(constants.UserNotExistErr, err) {
				return &pbservice.UserServiceInfoResp{
					BaseResp: &pbservice.BaseServiceResp{
						StatusCode: int32(api.UserNotExistErr),
						StatusMsg:  api.ErrorCodeToMsg[api.UserNotExistErr],
					},
				}, err
			} else {
				return &pbservice.UserServiceInfoResp{BaseResp: &pbservice.BaseServiceResp{
					StatusCode: int32(api.InnerDataBaseErr),
					StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
				}}, err
			}
		}
		return &pbservice.UserServiceInfoResp{
			Id:          userInfo.UserID,
			Name:        userInfo.UserName,
			FollowCnt:   userInfo.FollowCount,
			FollowerCnt: userInfo.FollowerCount,
			IsFollow:    false,
			BaseResp: &pbservice.BaseServiceResp{
				StatusCode: 0,
				StatusMsg:  "",
			},
		}, nil
	}
}

// GetUserIdByUserName RPC调用，获取Token的阶段会通过Username获取UserId
func (u *userService) GetUserIdByUserName(ctx context.Context, in *pbservice.UserServicePost) (*pbservice.UserServiceResp, error) {
	username := in.Username
	userInfo, err := u.getUserByUserName(username)
	if err != nil {
		return &pbservice.UserServiceResp{
			UserId: -1,
			Token:  "",
			BaseResp: &pbservice.BaseServiceResp{
				StatusCode: 101,
				StatusMsg:  err.Error(),
			},
		}, err
	}
	return &pbservice.UserServiceResp{
		UserId: userInfo.UserID,
		BaseResp: &pbservice.BaseServiceResp{
			StatusCode: 0,
			StatusMsg:  "",
		},
	}, nil
}

// 用户登录请求的内部处理逻辑
func (u *userService) userRegisterInfo(username, password string) (*model.User, error) {
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
	address := initialization.RpcSDConf.Host + initialization.RpcSDConf.UserServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbdao.NewUserDaoInfoClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := c.AddUser(ctx, &pbdao.UserDaoPost{
		Username: user.UserName,
		Password: user.PassWord,
		UserId:   user.UserID,
	})

	if err != nil {
		logger.GlobalLogger.Printf("Can't Access Dao for adding a user, err = %v", err)
		return nil, constants.InnerConnectionErr
	}
	if result.StatusCode != 0 {
		return nil, constants.InnerDataBaseErr
	}
	return user, nil
}

// 从username,password获得User
func (u *userService) checkUserInfo(username, password string) (*model.User, error) {
	userInfo, err := dao.GetUserDaoInstance().CheckUserByNameAndPassword(username, password)

	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}

	return userInfo, nil
}

// 通过userid得到user
func (u *userService) getUserByUserId(userId int64) (*model.User, error) {
	userInfo, err := dao.GetUserDaoInstance().GetUserByUserId(userId)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}
	return userInfo, nil
}

// 通过username得到user
func (u *userService) getUserByUserName(username string) (*model.User, error) {
	userInfo, err := dao.GetUserDaoInstance().GetUserByUsername(username)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
	}
	return userInfo, err
}
