package filelogger

import (
	"sync"
	"os"
	"io"
	"bytes"
	"../common"
	"strings"
)

// FileLogger对象，用于存储日志到文件
type FileLogger struct {
	File		string		// 日志文件，全路径
	fileHandler *os.File	// 文件指针
	Caches		[][]byte	// 缓存内容
	count 		uint 		// 记录缓存的数量
	size 		uint 		// 缓存的日志大小
	locker      sync.RWMutex  // 锁
	Config  	common.LoggerConfig	// 日志配置
	flushChan	chan bool 	// 是否开始输出日志到文件
	stopChan 	chan bool	// 用于控制是否需要关闭日志对象
	isClosed	bool 		// 判断是否已经关闭
}

// 创建一个FileLogger
func NewFileLogger(path string, cfg common.LoggerConfig) *FileLogger {
	return &FileLogger{
		File 	: path,
		Config  : cfg,
		Caches 	: make([][]byte, 0),
		count 	: 0,
		size 	: 0,
		locker 	: sync.RWMutex{},
		flushChan : make(chan bool),
		stopChan : make(chan bool),
		isClosed : false,
	}
}

// 获取日志配置
func (logger *FileLogger) GetConfig() common.LoggerConfig {
	return logger.Config
}

// 开始写入存储介质
func (logger *FileLogger) StartFlush() chan bool {
	return logger.flushChan
}

// 判断日志服务是否已经关闭
func (logger *FileLogger) IsClosed() bool {
	return logger.isClosed
}


// 停止日志服务
func (logger *FileLogger) Stop() chan bool {
	return logger.stopChan
}

// 获取文件指针
func (logger *FileLogger) GetWriter() (io.Writer, error) {
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
func (logger *FileLogger) Close() (error) {
	// 设置关闭标识
	logger.isClosed = true
	// 先写入文件
	// logger.Flush()
	// 关闭日志
	if nil != logger.fileHandler {
		logger.fileHandler.Close()
	}
	return nil
}

// 追加一条日志
func (logger *FileLogger) Append(content string) (int, error) {
	if logger.isClosed {
		return 0, common.ErrLoggerClosed
	}
	logger.locker.Lock()
	byteLog := []byte(content)
	logger.Caches = append(logger.Caches, byteLog)
	logger.count += 1
	logger.size += uint(len(byteLog))
	logger.locker.Unlock()
	// 判断是否需要写入存储介质
	if logger.count >= logger.Config.FlushCount || logger.size >= logger.Config.FlushSize {
		logger.flushChan <- true
	}
	return len(byteLog), nil
}

// 读取日志数量，可以指定最大值，如果maxNum < 1, 表示读取全部日志
func (logger *FileLogger) Read(maxNum uint) (int, []byte, error) {
	logger.locker.RLock()
	readNum := maxNum
	if maxNum == 0 || logger.count < maxNum {
		readNum = logger.count
	}
	buffer := &bytes.Buffer{}
	var i uint
	for i = 0; i < readNum; i++ {
		n, err := buffer.Write(logger.Caches[i])
		if err != nil {
			logger.locker.RUnlock()
			return 0, nil, err
		}
		logger.count -= 1
		logger.size -= uint(n)
	}
	logger.locker.RUnlock()

	logger.locker.Lock()
	logger.Caches = logger.Caches[readNum:]
	logger.locker.Unlock()

	bytesData := buffer.Bytes()
	return len(bytesData), bytesData, nil
}

// 写入存储介质
func (logger *FileLogger) Flush() (int, error) {
	fd, err := logger.GetWriter()
	if err != nil {
		return 0, err
	}
	n, logs, err := logger.Read(0)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	n, err = fd.Write(logs)
	return n, err
}

// 获取缓存的日志数量
func (logger *FileLogger) GetCount() uint {
	return logger.count
}

// 获取缓存日志大小
func (logger *FileLogger) GetSize() uint {
	return logger.size
}
