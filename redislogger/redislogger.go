package redislogger

import (
	"os"
	"sync"
	"../common"
	"io"
	"bytes"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
	"fmt"
)

type RedisConfig struct {
	Addr			string			// redis服务地址
	DbIndex			int				// 使用的db
	MaxIdle			int				// 最大的空闲连接数
	MaxActive 		int				// 最大的激活连接数，0标识没有限制
	IdleTimeout 	time.Duration	// 最大的空闲连接等待时间
}

type RedisLogger struct {
	CacheKey		string			// 缓存的Key
	File 			string 			// 目标存储文件
	redisPool 		*redis.Pool		// redis连接池
	fileHandler 	*os.File		// 文件指针
	count 			uint 			// 记录缓存的数量
	size 			uint 			// 缓存的日志大小
	locker      	sync.RWMutex  	// 锁
	Config  		common.LoggerConfig	// 日志配置
	flushChan		chan bool 		// 是否开始输出日志到文件
	stopChan 		chan bool		// 用于控制是否需要关闭日志对象
	isClosed		bool 			// 判断是否已经关闭
}

// 新建一个redislogger
func NewRedisLogger(filePath, cacheKey string, cfg common.LoggerConfig, redisCfg RedisConfig) *RedisLogger {
	pool := NewPool(redisCfg)
	if pool == nil {
		return nil
	}
	return &RedisLogger{
		CacheKey		: cacheKey,
		File 			: filePath,
		fileHandler 	: nil,
		redisPool		: pool,
		count			: 0,
		size 			: 0,
		locker 			: sync.RWMutex{},
		Config 			: cfg,
		flushChan 		: make(chan bool),
		stopChan 		: make(chan bool),
		isClosed 		: false,
	}
}

// 获取日志配置
func (logger *RedisLogger) GetConfig() common.LoggerConfig {
	return logger.Config
}

// 开始写入存储介质
func (logger *RedisLogger) StartFlush() chan bool {
	return logger.flushChan
}

// 判断日志服务是否已经关闭
func (logger *RedisLogger) IsClosed() bool {
	return logger.isClosed
}

// 停止日志服务
func (logger *RedisLogger) Stop() chan bool {
	return logger.stopChan
}

// 获取文件指针
func (logger *RedisLogger) GetWriter() (io.Writer, error) {
	// 如果已经存在Writer，直接返回
	if nil != logger.fileHandler {
		return logger.fileHandler, nil
	}
	// 如果没有设置写入文件，则直接返回错误
	if strings.TrimSpace(logger.File) == "" {
		return nil, common.ErrOutputNotSet
	}
	fd, err := os.OpenFile(logger.File, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	logger.fileHandler = fd
	return logger.fileHandler, nil
}

// 关闭日志对象
func (logger *RedisLogger) Close() (error) {
	// 设置关闭标识
	logger.isClosed = true
	// 先写入文件
	logger.Flush()
	// 关闭日志
	if nil != logger.fileHandler {
		logger.fileHandler.Close()
	}
	// 删除redis缓存
	conn := logger.redisPool.Get()
	conn.Do("DEL", logger.CacheKey)
	// 关闭连接池
	logger.redisPool.Close()
	return nil
}

// 追加一条日志
func (logger *RedisLogger) Append(content string) (int, error) {
	logger.locker.Lock()
	defer logger.locker.Unlock()

	if logger.isClosed {
		fmt.Printf("**CLOSED** %s \n", content)
		return 0, common.ErrLoggerClosed
	}

	byteLog := []byte(content)
	conn := logger.redisPool.Get()
	conn.Do("RPUSH", logger.CacheKey, content)
	conn.Do("EXPIRE", logger.CacheKey, 7200)
	logger.count += 1
	logger.size += uint(len([]byte(content)))

	fmt.Printf("append : %s \n", content)

	return len(byteLog), nil
}

// 读取日志数量，可以指定最大值，如果maxNum < 1, 表示读取全部日志
func (logger *RedisLogger) Read() (int, []byte, error) {
	// 从redis中读取全部值
	conn := logger.redisPool.Get()
	// 一次只读取指定的数量
	logs, err := redis.Strings(conn.Do("LRANGE", logger.CacheKey, 0, logger.Config.FlushCount))
	if err != nil {
		return 0, nil, err
	}

	readNum := len(logs)
	if readNum <= 0 {
		return 0, nil, nil
	}

	buffer := &bytes.Buffer{}
	var i int
	for i = 0; i < readNum; i++ {
		n, err := buffer.Write([]byte(logs[i]))
		if err != nil {
			return 0, nil, err
		}
		logger.count -= 1
		logger.size -= uint(n)
	}

	logger.locker.Lock()
	defer logger.locker.Unlock()
	val, _ := redis.Int(conn.Do("LTRIM", logger.CacheKey, logger.Config.FlushCount + 1, -1))
	if val > 0 {
		logger.count = 0
		logger.size = 0
	}

	bytesData := buffer.Bytes()
	return len(bytesData), bytesData, nil
}

// 写入存储介质
func (logger *RedisLogger) Flush() (int, error) {
	fmt.Println("*** flush ***")
	fd, err := logger.GetWriter()
	if err != nil {
		return 0, err
	}
	n, logs, err := logger.Read()
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, nil
	}
	n, err = fd.Write(logs)
	if err != nil {
		return 0, err
	}
	return n, err
}

// 获取缓存的日志数量
func (logger *RedisLogger) GetCount() uint {
	return logger.count
}

// 获取缓存日志大小
func (logger *RedisLogger) GetSize() uint {
	return logger.size
}
