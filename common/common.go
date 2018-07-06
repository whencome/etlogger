package common

import (
	"errors"
)

// 日志关闭错误
var ErrLoggerClosed = errors.New("logger closed")
// 文件为设置
var ErrOutputNotSet = errors.New("output target not set")