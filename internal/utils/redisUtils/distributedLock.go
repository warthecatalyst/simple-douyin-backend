package redisUtils

import (
	"github.com/go-redis/redis"
	"math"
	"math/rand"
	"time"
)

// DistributedLock 分布式锁
type DistributedLock interface {
	//Lock 尝试添加分布式锁
	Lock()
	//TryLock 在duration过后，自动放弃锁
	TryLock(duration time.Duration)
	//UnLock 解锁
	UnLock()
}

type myDistributedLock struct {
	redisClient  *redis.Client
	key          string
	maxRetryTime int
	curTryTime   int
}

// GetDistributedLock 获取一个分布式锁
func GetDistributedLock(rdb *redis.Client, key string, ops ...int) DistributedLock {
	maxRetryTime := 15
	if ops != nil && ops[0] != 0 {
		maxRetryTime = ops[0]
	}
	return &myDistributedLock{
		redisClient:  rdb,
		key:          key,
		maxRetryTime: maxRetryTime,
		curTryTime:   0,
	}
}

func (m myDistributedLock) Lock() {
	for m.curTryTime < m.maxRetryTime {
		res := m.tryGetDistributedLock()
		if res {
			break
		}
	}
}

func (m myDistributedLock) TryLock(duration time.Duration) {
	start := time.Now()
	for m.curTryTime < m.maxRetryTime {
		if start.Add(duration).After(time.Now()) {
			return
		}
		res := m.tryGetDistributedLock()
		if res {
			break
		}
	}
}

func (m myDistributedLock) UnLock() {
	for {
		err := m.redisClient.Del(m.key).Err()
		if err == nil {
			break
		}
	}
}

func (m myDistributedLock) tryGetDistributedLock() bool {
	//如果尝试次数>0，先休眠一下
	if m.curTryTime > 0 {
		sleepTime := int64(math.Pow(2, float64(m.curTryTime)))
		time.Sleep(time.Millisecond * time.Duration(sleepTime*(1000+rand.Int63n(1000))))
	}
	ok, err := m.redisClient.SetNX(m.key, nil, 15*time.Second).Result() //最多保存15s
	if err == nil && ok {
		return true
	}
	m.curTryTime++
	return false
}
