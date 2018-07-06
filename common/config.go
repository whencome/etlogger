package common

import (
	"time"
)

// 日志配置
type LoggerConfig struct {
	FlushCount		uint		// Flush数量，达到此数量时开始写入目标存储对象
	FlushSize		uint		// Flush大小，达到此大小时开始写入目标存储对象
	FlushDuration	time.Duration	// flush周期，每个一个周期遍历一下日志对象
	ExpireDuration	time.Duration	// 过期周期，如果超过这个时间没有任何操作则关闭该日志对象
}
