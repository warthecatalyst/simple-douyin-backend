package redisUtils

import "time"

const (
	// 过期时间
	ttl = time.Second * 30
	// 重置过期时间间隔
	resetTTLInterval = ttl / 3
	// 重新获取锁间隔
	tryLockInterval = time.Second
	//解锁脚本
	unlockScript = `
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end`
)
