package redisUtils

import (
	"context"
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/go-redis/redis/v8"
	"time"
)

// DistributedLock 分布式锁
type DistributedLock interface {
	//Lock 尝试加锁
	Lock(context.Context) error
	//TryLock 在duration过后，自动放弃锁
	TryLock(context.Context) error
	//UnLock 解锁
	UnLock(context.Context) error
}

type Lock struct {
	client          *redis.Client // redis客户端
	script          *redis.Script // 解锁脚本
	resource        string        // 锁定的资源
	randomValue     string        //随机值
	watchDog        chan struct{} //看门狗
	ttl             time.Duration // 过期时间
	tryLockInterval time.Duration // 重新获取锁间隔
}

func (l *Lock) Lock(ctx context.Context) error {
	// 尝试加锁
	err := l.TryLock(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(constants.LockFailedErr, err) {
		return err
	}
	//加锁失败不断尝试
	ticker := time.NewTicker(l.tryLockInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// 超时
			return constants.TimeOutErr
		case <-ticker.C:
			err := l.TryLock(ctx)
			if err != nil {
				return nil
			}
			if !errors.Is(constants.LockFailedErr, err) {
				return err
			}
		}
	}
}

func (l *Lock) TryLock(ctx context.Context) error {
	success, err := l.client.SetNX(ctx, l.resource, l.randomValue, l.ttl).Result()
	if err != nil {
		return err
	}
	// 加锁失败
	if !success {
		return constants.LockFailedErr
	}
	go l.startWatchDog()
	return nil
}

func (l *Lock) startWatchDog() {
	ticker := time.NewTicker(l.ttl / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 延长锁的过期时间
			ctx, cancel := context.WithTimeout(context.Background(), l.ttl/3*2)
			ok, err := l.client.Expire(ctx, l.resource, l.ttl).Result()
			cancel()
			// 异常或锁已经不存在则不再续期
			if err != nil || !ok {
				return
			}
		case <-l.watchDog:
			// 已经解锁
			return
		}
	}
}

func (l *Lock) UnLock(ctx context.Context) error {
	err := l.script.Run(ctx, l.client, []string{l.resource}, l.randomValue).Err()
	// 关闭看门狗
	close(l.watchDog)
	return err
}
