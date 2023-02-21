package service

import (
	"context"
	"errors"
	pbservice "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
	pbdao "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/user"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/idGenerator"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/md5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"math/rand"
	"strconv"
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

const (
	userLoginPrefix     = "user_login_"
	userLoginExpireTime = 90 * time.Minute
)

func getUserLoginExpireTime() time.Duration {
	return time.Duration(int64(userLoginExpireTime) + rand.Int63n(int64(30*time.Minute)))
}

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
	//返回已经存在的错误
	if err != nil {
		return nil, err
	}
	//不返回错误
	return &pbservice.UserServiceResp{
		UserId: userInfo.UserID,
	}, nil
}

// GetUserInfo RPC获取用户信息, 有两种方式,一种为通过username和password,另一种为通过queryUserId
func (u *userService) GetUserInfo(ctx context.Context, in *pbservice.UserServicePost) (*pbservice.UserServiceInfoResp, error) {
	username := in.Username
	password := in.Password
	_ = in.LoginUserId
	queryUserId := in.QueryUserId
	if "" != username && "" != password {
		userInfo, err := u.checkUserInfo(username, password)
		if err != nil {
			return nil, err
		}
		return &pbservice.UserServiceInfoResp{
			Id:          userInfo.UserID,
			Name:        userInfo.UserName,
			FollowCnt:   userInfo.FollowCount,
			FollowerCnt: userInfo.FollowerCount,
			IsFollow:    false,
		}, nil
	} else {
		userInfo, err := u.getUserByUserId(queryUserId)
		if err != nil {
			return nil, err
		}
		return &pbservice.UserServiceInfoResp{
			Id:          userInfo.UserID,
			Name:        userInfo.UserName,
			FollowCnt:   userInfo.FollowCount,
			FollowerCnt: userInfo.FollowerCount,
			IsFollow:    false,
		}, nil
	}
}

// GetUserIdByUserName RPC调用，获取Token的阶段会通过Username获取UserId
func (u *userService) GetUserIdByUserName(ctx context.Context, in *pbservice.UserServicePost) (*pbservice.UserServiceResp, error) {
	username := in.Username
	userInfo, err := u.getUserByUserName(username)
	if err != nil {
		return nil, err
	}
	return &pbservice.UserServiceResp{
		UserId: userInfo.UserID,
	}, nil
}

// service层对用户注册请求的内部处理逻辑
func (u *userService) userRegisterInfo(username, password string) (*model.User, error) {
	var err error
	// 先看redis中存不存在
	key := userLoginPrefix + username
	exist, err := redisClient.Exists(key).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, constants.RedisDBErr.Error())
	}
	if exist == 1 { //存在直接返回错误
		return nil, status.Errorf(codes.AlreadyExists, constants.UserAlreadyExistErr.Error())
	}
	address := initialization.RpcSDConf.UserServiceHost + initialization.RpcSDConf.UserServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbdao.NewUserDaoInfoClient(conn)

	// Contact the server and print out its response.
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	userResp, err := c.GetUserInfoByUserName(ctx1, &pbdao.UserDaoPost{Username: username})

	if userResp != nil {
		return nil, status.Errorf(codes.AlreadyExists, constants.UserAlreadyExistErr.Error())
	} else if err != nil {
		if errors.Is(status.Errorf(codes.Internal, constants.InnerDataBaseErr.Error()), err) {
			return nil, err
		}
	}

	userId := idGenerator.GenerateUserId()
	logger.GlobalLogger.Printf("userId = %v", userId)
	user := &model.User{
		UserID:   userId,
		UserName: username,
	}
	if initialization.UserConf.PasswordEncrypted {
		user.PassWord = md5.MD5(password)
	} else {
		user.PassWord = password
	}
	go u.writeUsernameToUserInfoToRedis(username, user.PassWord, userId)
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	result, err := c.AddUser(ctx2, &pbdao.UserDaoPost{
		Username: user.UserName,
		Password: user.PassWord,
		UserId:   user.UserID,
	})

	if !result.Value {
		return nil, err
	}
	return user, nil
}

// 从username,password获得User
func (u *userService) checkUserInfo(username, password string) (*model.User, error) {
	var err error
	if initialization.UserConf.PasswordEncrypted {
		password = md5.MD5(password)
	}
	key := userLoginPrefix + username
	exist, err := redisClient.Exists(key).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, constants.RedisDBErr.Error())
	}
	// 如果存在直接返回
	if exist == 1 {
		userInfo, _ := redisClient.HMGet(key, "UserId", "Password").Result()
		userId := userInfo[0].(int64)
		pwd := userInfo[1].(string)
		if password != pwd {
			return nil, status.Errorf(codes.NotFound, constants.UserNotExistErr.Error())
		}
		redisClient.Expire(key, getUserLoginExpireTime())
		return &model.User{UserID: userId, UserName: username}, nil
	}
	address := initialization.RpcSDConf.UserServiceHost + initialization.RpcSDConf.UserServicePort
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbdao.NewUserDaoInfoClient(conn)

	// Contact the server and print out its response.
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	userResp, err := c.GetUserInfoByUserNameAndPassword(
		ctx1, &pbdao.UserDaoPost{Username: username, Password: password})

	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}
	go u.writeUsernameToUserInfoToRedis(username, password, userResp.Id)
	return &model.User{
		UserID:        userResp.Id,
		UserName:      userResp.Name,
		FollowCount:   userResp.FollowCnt,
		FollowerCount: userResp.FollowerCnt,
	}, nil
}

// 通过userid得到user
func (u *userService) getUserByUserId(userId int64) (*model.User, error) {
	var err error
	//从redis中获取
	key := userLoginPrefix + strconv.FormatInt(userId, 10)
	exist, err := redisClient.Exists(key).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, constants.RedisDBErr.Error())
	}
	// 如果存在直接返回
	if exist == 1 {
		userInfo, _ := redisClient.HMGet(key, "UserName", "FollowCnt", "FollowerCnt").Result()
		redisClient.Expire(key, getUserLoginExpireTime())
		userName := userInfo[0].(string)
		followCnt := userInfo[1].(int64)
		followerCnt := userInfo[2].(int64)
		return &model.User{UserID: userId, UserName: userName, FollowCount: followCnt, FollowerCount: followerCnt}, nil
	}
	address := initialization.RpcSDConf.UserServiceHost + initialization.RpcSDConf.UserServicePort
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbdao.NewUserDaoInfoClient(conn)

	// Contact the server and print out its response.
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	userResp, err := c.GetUserInfoByUserId(
		ctx1, &pbdao.UserDaoPost{UserId: userId})
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}
	userInfo := &model.User{
		UserID:        userResp.Id,
		UserName:      userResp.Name,
		FollowCount:   userResp.FollowCnt,
		FollowerCount: userResp.FollowerCnt,
	}
	u.writeUserIdToUserModelToRedis(userInfo)
	return userInfo, nil
}

// 通过username得到user
func (u *userService) getUserByUserName(username string) (*model.User, error) {
	var err error
	//先看redis中存不存在，存在的话直接返回
	key := userLoginPrefix + username
	exist, err := redisClient.Exists(key).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, constants.RedisDBErr.Error())
	}
	if exist == 1 {
		userInfos, _ := redisClient.HMGet(key, "UserId").Result()
		redisClient.Expire(key, getUserLoginExpireTime())
		userId, _ := userInfos[0].(int64)
		return &model.User{UserID: userId}, nil
	}
	address := initialization.RpcSDConf.UserServiceHost + initialization.RpcSDConf.UserServicePort
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbdao.NewUserDaoInfoClient(conn)

	// Contact the server and print out its response.
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	userResp, err := c.GetUserInfoByUserName(
		ctx1, &pbdao.UserDaoPost{Username: username})
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, 寻找数据失败, err = %s", err.Error())
		return nil, err
	}
	go u.writeUsernameToUserInfoToRedis(username, userResp.Password, userResp.Id)
	return &model.User{
		UserID: userResp.Id,
	}, nil
}

func (u *userService) writeUsernameToUserInfoToRedis(username, password string, userId int64) {
	for {
		key := userLoginPrefix + username
		userName2UserInfo := map[string]interface{}{
			"UserId":   userId,
			"Password": password,
		}
		err := redisClient.HMSet(key, userName2UserInfo).Err()
		err = redisClient.Expire(key, getUserLoginExpireTime()).Err()
		if err == nil {
			break
		}
	}
}

func (u *userService) writeUserIdToUserModelToRedis(user *model.User) {
	for {
		userId := user.UserID
		userIdStr := strconv.FormatInt(userId, 10)
		key := userLoginPrefix + userIdStr
		userId2UserInfo := map[string]interface{}{
			"UserName":    user.UserName,
			"FollowCnt":   user.FollowCount,
			"FollowerCnt": user.FollowerCount,
		}
		err := redisClient.HMSet(key, userId2UserInfo).Err()
		err = redisClient.Expire(key, getUserLoginExpireTime()).Err()
		if err == nil {
			break
		}
	}
}
