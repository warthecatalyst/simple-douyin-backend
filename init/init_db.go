package init

import (
	"fmt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

func InitDB() {
	StdOutLogger.Print("In InitDataBase")
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassWord,
		dbHost,
		dbPort,
		dbName)

	var err error
	logLevelMap := map[string]logger.LogLevel{
		"silent": logger.Silent,
		"error":  logger.Error,
		"warn":   logger.Warn,
		"info":   logger.Info,
	}

	db, err = gorm.Open(mysql.Open(dns), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, //关闭外键
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,       //默认在表的后面加s
			TablePrefix:   "t_douyin_", // 表名前缀
		},
		SkipDefaultTransaction: true, // 禁用默认事务
		Logger:                 logger.Default.LogMode(logLevelMap[dbLogLevel]),
	})

	if err != nil {
		StdOutLogger.Panic().Caller().Str("数据库初始化失败", err.Error())
	}

	err = db.AutoMigrate(&model.Video{}, &model.User{}, &model.Follow{}, &model.Comment{}, &model.Favourite{}, &model.Message{}) //数据库自动迁移

	if err != nil {
		StdOutLogger.Panic().Caller().Str("数据库自动迁移失败", err.Error())
	}

	sqlDb, _ := db.DB()

	sqlDb.SetMaxIdleConns(50)                  // 连接池中的最大闲置连接数
	sqlDb.SetMaxOpenConns(100)                 // 数据库的最大连接数量
	sqlDb.SetConnMaxLifetime(10 * time.Second) // 连接的最大可复用时间
}
