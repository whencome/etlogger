package etlogger

import (
	"time"
	"fmt"
)

// 注册日志服务
func Register(cate string, logger Logger) {
	if _, ok := loggerServices[cate]; ok {
		return
	}
	service := LoggerService{
		Category 		: cate,
		LogHandler		: logger,
		LastWriteTime	: time.Now(),
		Config			: logger.GetConfig(),
	}
	loggerServices[cate] = &service
	go StartLogService(cate, loggerServices[cate])
}

// 取消日志服务注册
func Unregister(cate string) {
	if _, ok := loggerServices[cate]; ok {
		delete(loggerServices, cate)
	}
	return
}

// 启动日志服务
func StartLogService(cate string, service *LoggerService) {
	loggerCfg := service.LogHandler.GetConfig()
	for {
		select {
		// 停止服务
		case <-service.LogHandler.Stop():
			closeLoggerService(service)
			service = nil
			// 退出服务
			return
		// 定时
		case <-time.After(loggerCfg.FlushDuration):
			n, err := service.LogHandler.Flush()
			if err == nil && n > 0 {
				service.LastWriteTime = time.Now()
			}
			break
		// 程序控制是否写入
		case <-service.LogHandler.StartFlush():
			n, err := service.LogHandler.Flush()
			if err == nil && n > 0 {
				service.LastWriteTime = time.Now()
			}
			break
		// 超时设置
		case <-time.After(30 * time.Second):
			break
		}
	}
}

// 定时任务，用于过滤过期的服务
func startListenExpiration() {
	// 防止过于频繁调用，10秒钟调用一次
	c := time.Tick(10 * time.Duration(time.Second))
	for {
		<-c
		// 轮询日志服务
		for _, service := range loggerServices {
			go func() {
				now := time.Now()
				if now.Sub(service.LastWriteTime) >= service.Config.ExpireDuration {
					fmt.Printf("%s expired \n", service.Category)
					closeLoggerService(service)
					service.LogHandler.Stop() <- true
				}
			}()
		}
	}
}

// 关闭日志服务
func closeLoggerService(service *LoggerService) {
	if _, ok := loggerServices[service.Category]; !ok {
		return
	}
	if service.LogHandler.IsClosed() {
		return
	}
	// 先写入缓冲区内容
	n, err := service.LogHandler.Flush()
	if err == nil && n > 0 {
		service.LastWriteTime = time.Now()
	}
	// 关闭日志服务
	service.LogHandler.Close()
	// 注销日志服务注册
	Unregister(service.Category)
}

// 关闭指定分类的日志
func CloseCategoryLogger(cate string) {
	if service, ok := loggerServices[cate]; ok {
		closeLoggerService(service)
		service.LogHandler.Stop() <- true
	}
}

// 打印日志服务列表
func StatLoggerServices() {
	servNum := 0
	fmt.Printf("\nSTAT TIME : %s \n", time.Now().String())
	fmt.Println("-------------------------------------")
	fmt.Printf("%-50s\t%s\n", "Category", "Last Time")
	for c, serv := range loggerServices {
		fmt.Printf("%-50s\t%s\n", c, serv.LastWriteTime.String())
		servNum++
	}
	fmt.Println("-------------------------------------")
	fmt.Printf("Total : %d \n", servNum)
}