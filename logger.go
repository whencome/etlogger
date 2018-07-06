package etlogger

import (
	"./filelogger"
	"./redislogger"
	"./common"
	"time"
)

// 日志服务注册列表
var loggerServices map[string]*LoggerService

// 初始化
func init() {
	loggerServices = make(map[string]*LoggerService)
	go startListenExpiration()
}

func NewLoggerConfig(flushCount, flushSize uint, flushDuration, expireDuration time.Duration) common.LoggerConfig {
	return common.LoggerConfig {
		FlushCount 		: flushCount,
		FlushSize 		: flushSize,
		FlushDuration 	: flushDuration,
		ExpireDuration 	: expireDuration,
	}
}

func NewRedisConfig(addr string, dbIdx, maxIdle, maxActive int, idleTimeout time.Duration) redislogger.RedisConfig {
	return redislogger.RedisConfig {
		Addr 			: addr,
		DbIndex 		: dbIdx,
		MaxIdle 		: maxIdle,
		MaxActive 		: maxActive,
		IdleTimeout 	: idleTimeout,
	}
}

// 默认使用filelogger，自动写入目标文件
func NewDefaultLogger(path string, closeExists bool) Logger {
	if service, ok := loggerServices[path]; ok {
		if !closeExists {
			return service.LogHandler
		} else {
			closeLoggerService(service)
		}
	}
	cfg := common.LoggerConfig{
		FlushCount : 200,
		FlushSize : 1024 * 10,
		FlushDuration : 5 * time.Duration(time.Second),
		ExpireDuration : 300 * time.Duration(time.Second),
	}
	flogger := filelogger.NewFileLogger(path, cfg)
	// 注册日志服务，自动flush日志到文件，对于长时间未活跃的服务会自动关闭
	// 如果不注册，则需要自行控制flush以及服务关闭
	Register(path, flogger)
	return flogger
}

// 创建新的filelogger，文件日志只能写本地，所以自定注册
func NewFileLogger(path string, cfg common.LoggerConfig) Logger {
	flogger := filelogger.NewFileLogger(path, cfg)
	Register(path, flogger)
	return flogger
}

// 创建新的redislogger，redislogger主要用于向远端写日志，因此不自行注册
// 向远端写的时候path留空
func NewRedisLogger(path, cacheKey string, cfg common.LoggerConfig, redisCfg redislogger.RedisConfig) Logger {
	rlogger := redislogger.NewRedisLogger(path, cacheKey, cfg, redisCfg)
	return rlogger
}

