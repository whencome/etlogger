package etlogger

import (
	"./common"
	"time"
)

// 定义日志接口，便于后续扩展
type Logger interface {
	// 获取日志配置
	GetConfig() common.LoggerConfig
	// 追加一条日志
	Append(content string) (int, error)
	// 写入存储介质
	Flush() (int, error)
	// 开始写入存储介质
	StartFlush() chan bool
	// 停止
	Stop() chan bool
	// 关闭日志服务
	Close() error
	// 判断日志服务示范已经关闭
	IsClosed() bool
}

// 日志服务结构
type LoggerService struct {
	Category 		string
	LogHandler		Logger
	LastWriteTime	time.Time
	Config 			common.LoggerConfig
}