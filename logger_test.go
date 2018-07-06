package etlogger

import (
	"testing"
	"time"
	"fmt"
	"sync"
	"os"
)

func TestNewDefaultLogger(t *testing.T) {
	wg := &sync.WaitGroup{}
	numChan := make(chan int)
	sizeChan := make(chan int)
	stopChan := make(chan bool)
	logNum := 0
	logSize := 0
	// 统计数据
	go func(){
		for {
			select {
			case <-stopChan:
				return
			case n := <-numChan:
				logNum += n
				break
			case s := <-sizeChan:
				logSize += s
				break
			}
		}
	}()
	// 打印日志服务统计信息
	go func() {
		t := time.Tick(1 * time.Second)
		for {
			<-t
			StatLoggerServices()
		}
	}()
	// 输出日志
	for j := 0; j<20; j++ {
		wg.Add(1)
		go func(j int) {
			logger := NewDefaultLogger(fmt.Sprintf("/home/logs/test/etlogger_test_%d.log", j), true)
			for i := 0; i < 100; i++ {
				numChan <- 1
				n, _ := logger.Append(fmt.Sprintf("%d 【%d】 [%s]: golang 中的 slice 非常强大，让数组操作非常方便高效。在开发中不定长度表示的数组全部都是 slice 。但是很多同学对 slice 的模糊认识，造成认为golang中的数组是引用类型，结果就是在实际开发中碰到很多坑，以至于出现一些莫名奇妙的问题，数组中的数据丢失了。 \n", logNum, j, time.Now().String()))
				sizeChan <- n
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	stopChan <- true
	time.Sleep(360 * time.Second)

	fmt.Println("test finished")
	fmt.Printf("total log num : %d \n", logNum)
	fmt.Printf("total log size : %d \n", logSize)
}

func TestRepeatLogger(t *testing.T) {
	wg := &sync.WaitGroup{}
	numChan := make(chan int)
	sizeChan := make(chan int)
	stopChan := make(chan bool)
	logNum := 0
	logSize := 0
	// 统计数据
	go func(){
		for {
			select {
			case <-stopChan:
				return
			case n := <-numChan:
				logNum += n
				break
			case s := <-sizeChan:
				logSize += s
				break
			}
		}
	}()
	// 打印日志服务统计信息
	go func() {
		t := time.Tick(1 * time.Second)
		for {
			<-t
			StatLoggerServices()
		}
	}()
	// 输出日志
	for j := 0; j<20; j++ {
		wg.Add(1)
		go func(j int) {
			for i := 0; i < 100; i++ {
				logger := NewDefaultLogger(fmt.Sprintf("/home/logs/test/etlogger_test_%d.log", j), true)
				numChan <- 1
				n, _ := logger.Append(fmt.Sprintf("%d 【%d】 [%s]: golang 中的 slice 非常强大，让数组操作非常方便高效。在开发中不定长度表示的数组全部都是 slice 。但是很多同学对 slice 的模糊认识，造成认为golang中的数组是引用类型，结果就是在实际开发中碰到很多坑，以至于出现一些莫名奇妙的问题，数组中的数据丢失了。 \n", logNum, j, time.Now().String()))
				sizeChan <- n
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	stopChan <- true
	time.Sleep(360 * time.Duration(time.Second))

	fmt.Println("test finished")
	fmt.Printf("total log num : %d \n", logNum)
	fmt.Printf("total log size : %d \n", logSize)
}

func TestRedisLogger(t *testing.T) {
	// 打印日志服务统计信息
	go func() {
		t := time.Tick(3 * time.Second)
		for {
			<-t
			StatLoggerServices()
		}
	}()
	redisCfg := NewRedisConfig("127.0.0.1:6379", 0, 3, 10, 60*time.Second)
	loggerCfg := NewLoggerConfig(100, 10 * 1024, 5 * time.Second, 60 * time.Second)
	cacheKey := "test_remote_log"
	logFile := "/home/logs/test/test_remote.log"
	// 用于向redis写日志
	lgrW := NewRedisLogger("", cacheKey, loggerCfg, redisCfg)
	lgrW.Append(fmt.Sprintf("logW [%s]: golang 中的 slice 非常强大，让数组操作非常方便高效。在开发中不定长度表示的数组全部都是 slice 。但是很多同学对 slice 的模糊认识，造成认为golang中的数组是引用类型，结果就是在实际开发中碰到很多坑，以至于出现一些莫名奇妙的问题，数组中的数据丢失了。 \n", time.Now().String()))
	// 用于从redis读取日志并写入本地文件
	lgrR := NewRedisLogger(logFile, cacheKey, loggerCfg, redisCfg)
	Register(logFile, lgrR)
	time.Sleep(120 * time.Second)
	_, err := os.Stat(logFile)
	if err != nil {
		t.Error("failed to read and write log")
	}
	fmt.Println("test finished")
}


func TestBatchRedisLogger(t *testing.T) {
	wg := &sync.WaitGroup{}
	numChan := make(chan int)
	sizeChan := make(chan int)
	stopChan := make(chan bool)
	logNum := 0
	logSize := 0
	// 统计数据
	go func(){
		for {
			select {
			case <-stopChan:
				return
			case n := <-numChan:
				logNum += n
				break
			case s := <-sizeChan:
				logSize += s
				break
			}
		}
	}()
	// 打印日志服务统计信息
	go func() {
		t := time.Tick(3 * time.Second)
		for {
			<-t
			StatLoggerServices()
		}
	}()

	redisCfg := NewRedisConfig("127.0.0.1:6379", 0, 3, 0, 60*time.Second)
	loggerCfg := NewLoggerConfig(100, 10 * 1024, 5 * time.Second, 60 * time.Second)
	fmtCacheKey := "test_remote_log_%d"
	fmtLogFile := "/home/logs/test/test_remote_%d.log"


	go func() {
		for j := 0; j<5; j++ {
			path := fmt.Sprintf(fmtLogFile, j)
			lgrR := NewRedisLogger(path, fmt.Sprintf(fmtCacheKey, j), loggerCfg, redisCfg)
			Register(path, lgrR)
		}
	}()

	// 输出日志
	for j := 0; j<5; j++ {
		wg.Add(1)
		lgrW := NewRedisLogger("", fmt.Sprintf(fmtCacheKey, j), loggerCfg, redisCfg)
		go func(j int) {
			for i := 0; i < 2000; i++ {
				numChan <- 1
				n, _ := lgrW.Append(fmt.Sprintf("%d 【%d】 [%s]: golang 中的 slice 非常强大，让数组操作非常方便高效。在开发中不定长度表示的数组全部都是 slice 。但是很多同学对 slice 的模糊认识，造成认为golang中的数组是引用类型，结果就是在实际开发中碰到很多坑，以至于出现一些莫名奇妙的问题，数组中的数据丢失了。 \n", logNum, j, time.Now().String()))
				sizeChan <- n
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	stopChan <- true
	time.Sleep(120 * time.Second)

	fmt.Println("test finished")
	fmt.Printf("total log num : %d \n", logNum)
	fmt.Printf("total log size : %d \n", logSize)
}