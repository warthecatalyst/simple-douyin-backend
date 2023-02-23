package redisUtils

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", "127.0.0.1", "6379"),
		Password: "",
		DB:       0,
		PoolSize: 100,
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = rdb.Set("war", "123456", 24*time.Hour).Err()
	if err != nil {
		fmt.Println(err.Error())
	}
	val, err := rdb.Get("war").Result()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("val = ", val)
	val1, err := rdb.Exists("war").Result()
	fmt.Println("val1 = ", val1)
	val1, err = rdb.Exists("war1").Result()
	fmt.Println("val1 = ", val1)
	return
}
