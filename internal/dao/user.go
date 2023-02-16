package dao

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbdao "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/user"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
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
func (u *userDao) AddUser(ctx context.Context, in *pbdao.UserDaoPost) (*pbdao.BaseDaoResp, error) {
	user := &model.User{}
	user.UserID = in.GetUserId()
	user.PassWord = in.GetPassword()
	user.UserName = in.GetUsername()
	err := u.CreateUser(user)
	if err != nil {
		return &pbdao.BaseDaoResp{
			StatusCode: int32(api.InnerDataBaseErr),
			StatusMsg:  api.ErrorCodeToMsg[api.InnerDataBaseErr],
		}, nil
	}
	return &pbdao.BaseDaoResp{
		StatusCode: 0,
		StatusMsg:  "",
	}, nil
}

// GetUserByUsername 通过用户名查找在数据库中的User
func (u *userDao) GetUserByUsername(username string) (*model.User, error) {
	userInfos := make([]*model.User, 0)
	if err := db.Where("user_name = ?", username).First(&userInfos).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.UserNotExistErr
		}
		return nil, constants.InnerDataBaseErr
	}

	// 理论上来说userInfos不应当>1, 因为username是唯一索引
	if len(userInfos) > 1 {
		return nil, constants.InnerDataBaseErr
	}

	return userInfos[0], nil
}

// GetUserByUserId 通过userId查找在数据库中的User
func (u *userDao) GetUserByUserId(userId int64) (*model.User, error) {
	userInfos := make([]*model.User, 0)
	if err := db.Where("user_id = ?", userId).Find(&userInfos).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.UserNotExistErr
		}
		return nil, constants.InnerDataBaseErr
	}

	// 理论上来说userInfos不应当>1, 因为userId是唯一索引
	if len(userInfos) > 1 {
		return nil, constants.InnerDataBaseErr
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
