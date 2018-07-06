package redislogger

import (
	"time"
	"github.com/gomodule/redigo/redis"
)

// 创建一个redis连接
func NewPool(redisCfg RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle		: redisCfg.MaxIdle,
		MaxActive	: redisCfg.MaxActive,
		IdleTimeout : redisCfg.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisCfg.Addr)
			if err != nil {
				return nil, err
			}
			// 选择db
			c.Do("SELECT", redisCfg.DbIndex)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
