package dao

import (
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"gorm.io/gorm"
	"sync"
)

type UserDao struct{}

var (
	userDao  *UserDao
	userOnce sync.Once
)

func NewUserDaoInstance() *UserDao {
	userOnce.Do(func() {
		userDao = &UserDao{}
	})
	return userDao
}

// GetUserByUsername 通过用户名查找是否有是否在数据库中已经存在User
func (u *UserDao) GetUserByUsername(username string) (*model.User, error) {
	userInfo := &model.User{}
	if err := db.Where("user_name = ?", username).First(userInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userInfo, nil
}

// CreateUser 在数据库中通过事务创建一个新用户
func (u *UserDao) CreateUser(user *model.User) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		if err := tx.Create(user).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		// 返回 nil 提交事务
		return nil
	})
}

// CheckUserByNameAndPassword 登录时候查找数据库中是否存在该用户以及密码是否正确
func (u *UserDao) CheckUserByNameAndPassword(username string, password string) (*model.User, error) {
	userInfos := make([]*model.User, 0)
	if err := db.Where("user_name = ?", username).Where("pass_word = ?", password).Find(&userInfos).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	// 理论上来说userInfos不应当>1, 因为username是唯一索引
	if len(userInfos) > 1 {
		return nil, errors.New(api.ErrorCodeToMsg[api.InnerDataBaseErr])
	}

	return userInfos[0], nil
}
