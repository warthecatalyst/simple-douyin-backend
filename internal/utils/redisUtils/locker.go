package redisUtils

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"strconv"
	"time"
)

// Locker 可以通过Locker获得一把分布式锁
// 其主要用于配置锁
type Locker struct {
	client          *redis.Client
	script          *redis.Script
	ttl             time.Duration
	tryLockInterval time.Duration
}

// NewDefaultLocker 通过client和默认定义项获取一个Locker
func NewDefaultLocker(client *redis.Client) *Locker {
	return &Locker{
		client:          client,
		script:          redis.NewScript(unlockScript),
		ttl:             ttl,
		tryLockInterval: tryLockInterval,
	}
}

// NewLocker 通过自配置定义项获取Locker
func NewLocker(client *redis.Client, ttl, tryLockInterval time.Duration) *Locker {
	return &Locker{
		client:          client,
		script:          redis.NewScript(unlockScript),
		ttl:             ttl,
		tryLockInterval: tryLockInterval,
	}
}

func (l *Locker) GetLock(resource string) DistributedLock {
	return &Lock{
		client:          l.client,
		script:          l.script,
		resource:        resource,
		randomValue:     strconv.Itoa(int(uuid.New().ID())),
		watchDog:        make(chan struct{}),
		ttl:             l.ttl,
		tryLockInterval: l.tryLockInterval,
	}
}
