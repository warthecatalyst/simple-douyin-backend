package dao

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"gorm.io/gorm"
	"sync"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

// DaoInitialization 初始化Dao层的服务，包括获取DB以及Kafka
func DaoInitialization() {
	dbOnce.Do(func() {
		db = initialization.GetDB()
		initKafkaClient()
	})
}
