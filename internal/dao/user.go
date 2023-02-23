package dao

import (
	"context"
	pbdao "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/user"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"
	"sync"
)

//userDao 与user相关的数据库操作
type userDao struct {
	pbdao.UnimplementedUserDaoInfoServer
}

var (
	userDaoInstance *userDao
	userOnce        sync.Once
)

func GetUserDaoInstance() *userDao {
	userOnce.Do(func() {
		userDaoInstance = &userDao{}
	})
	return userDaoInstance
}

// AddUser RPC调用在数据库中添加一个用户
func (u *userDao) AddUser(ctx context.Context, in *pbdao.UserDaoPost) (*wrapperspb.BoolValue, error) {
	user := &model.User{}
	user.UserID = in.GetUserId()
	user.PassWord = in.GetPassword()
	user.UserName = in.GetUsername()
	err := u.CreateUser(user)
	if err != nil {
		return &wrapperspb.BoolValue{Value: false},
			status.Errorf(codes.Internal, constants.InnerDataBaseErr.Error())
	}
	return &wrapperspb.BoolValue{Value: true}, status.New(codes.OK, "").Err()
}

// GetUserInfoByUserName RPC远程调用通过Username得到User
func (u *userDao) GetUserInfoByUserName(ctx context.Context, in *pbdao.UserDaoPost) (*pbdao.UserDaoInfoResp, error) {
	username := in.Username
	userInfo, err := u.GetUserByUsername(username)
	return returnUserDaoRPC(userInfo, err)
}

// GetUserInfoByUserId RPC远程调用通过userId得到User
func (u *userDao) GetUserInfoByUserId(ctx context.Context, in *pbdao.UserDaoPost) (*pbdao.UserDaoInfoResp, error) {
	userId := in.UserId
	userInfo, err := u.GetUserByUserId(userId)
	return returnUserDaoRPC(userInfo, err)
}

// GetUserInfoByUserNameAndPassword RPC远程调用检查username和password
func (u *userDao) GetUserInfoByUserNameAndPassword(ctx context.Context, in *pbdao.UserDaoPost) (*pbdao.UserDaoInfoResp, error) {
	username := in.Username
	password := in.Password
	userInfo, err := u.CheckUserByNameAndPassword(username, password)
	return returnUserDaoRPC(userInfo, err)
}

// GetUserByUsername 通过用户名查找在数据库中的User
func (u *userDao) GetUserByUsername(username string) (*model.User, error) {
	userInfos := make([]*model.User, 0)
	if err := db.Where("user_name = ?", username).Find(&userInfos).Error; err != nil {
		return nil, constants.InnerDataBaseErr
	}

	// 理论上来说userInfos不应当>1, 因为username是唯一索引
	if len(userInfos) > 1 {
		return nil, constants.InnerDataBaseErr
	}
	if len(userInfos) == 0 {
		return nil, constants.UserNotExistErr
	}

	return userInfos[0], nil
}

// GetUserByUserId 通过userId查找在数据库中的User
func (u *userDao) GetUserByUserId(userId int64) (*model.User, error) {
	userInfos := make([]*model.User, 0)
	if err := db.Where("user_id = ?", userId).Find(&userInfos).Error; err != nil {
		return nil, constants.InnerDataBaseErr
	}

	// 理论上来说userInfos不应当>1, 因为userId是唯一索引
	if len(userInfos) > 1 {
		return nil, constants.InnerDataBaseErr
	}
	if len(userInfos) == 0 {
		return nil, constants.UserNotExistErr
	}

	return userInfos[0], nil
}

// CheckUserByNameAndPassword 通过username与password查找在数据库中的User
func (u *userDao) CheckUserByNameAndPassword(username string, password string) (*model.User, error) {
	userInfos := make([]*model.User, 0)
	if err := db.Where("user_name = ?", username).Where("pass_word = ?", password).Find(&userInfos).Error; err != nil {
		return nil, constants.InnerDataBaseErr
	}

	// 理论上来说userInfos不应当>1, 因为username是唯一索引
	if len(userInfos) > 1 {
		return nil, constants.InnerDataBaseErr
	}
	if len(userInfos) == 0 {
		return nil, constants.UserNotExistErr
	}

	return userInfos[0], nil
}

// CreateUser 在数据库中通过事务创建一个新用户,所有的写操作都通过事务完成
func (u *userDao) CreateUser(user *model.User) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		return nil
	})
}

func returnUserDaoRPC(userInfo *model.User, err error) (*pbdao.UserDaoInfoResp, error) {
	if err != nil {
		switch err {
		case constants.UserNotExistErr:
			return nil, status.Errorf(codes.NotFound, constants.UserNotExistErr.Error())
		case constants.InnerDataBaseErr:
			return nil, status.Errorf(codes.Internal, constants.InnerDataBaseErr.Error())
		}
	}
	return &pbdao.UserDaoInfoResp{
		Id:          userInfo.UserID,
		Name:        userInfo.UserName,
		Password:    userInfo.PassWord,
		FollowCnt:   userInfo.FollowCount,
		FollowerCnt: userInfo.FollowerCount,
	}, nil
}
