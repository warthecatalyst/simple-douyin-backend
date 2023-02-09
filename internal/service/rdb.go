package service

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/go-redis/redis"
	"math/rand"
	"sync"
	"time"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func initRedis() {
	redisOnce.Do(func() {
		redisClient = initialization.GetRDB()
	})
}

func getFavoriteRandomTime() time.Duration {
	return time.Duration(int64(videoFavoriteExpireTime) + rand.Int63n(int64(12*time.Hour)))
}
