package init

import (
	"fmt"
	"github.com/go-redis/redis"
)

var rdb *redis.Client

func InitRDB() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", RdbHost, RdbPort),
		Password: "",
		DB:       0,
		PoolSize: 100,
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		StdOutLogger.Panic().Caller().Str("Redis启动失败", err.Error())
	}

	return
}

func GetRDB() *redis.Client {
	return rdb
}
