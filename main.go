package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seelog/seelog"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	fsnotify "gopkg.in/fsnotify.v1"
)

const (
	DebugLog = "debug.log"
	ErrLog   = "err.log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//创建一个监控对象
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watch.Close()
	logfile := os.Getenv("LOG_FILE")
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	err = watch.Add(logfile)
	if err != nil {
		log.Fatal(err)
	}
	//获取当前目录下的所有文件或目录信息
	filepath.Walk(logfile, func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path) //打印path信息
		//fmt.Println(info.Name()) //打印文件或目录名

		if strings.Contains(path, "backup") != true {
			seelog.See(info.Name(), path)
		}

		return nil
	})

	go func() {
		for {
			select {
			case ev := <-watch.Events:
				{
					fname := filepath.Base(ev.Name)
					if ev.Op == fsnotify.Remove {
						seelog.Remove(fname)

						// fmt.Println(fname + "-------------")

					} else if ev.Op == fsnotify.Create {
						seelog.See(fname, ev.Name)
						seelog.Addmonitor(fname, ev.Name)
						//fmt.Println(fname + "+++++++++++++")

					}
				}
			case err := <-watch.Errors:
				{
					log.Println("error : ", err)
					return
				}
			}
		}
	}()
	// 测试
	//seelog.See("错误日志",ErrLog)
	//seelog.See("调试日志",DebugLog)

	seelog.Serve(port, "password")

	// 模拟服务输出日志
	//go printLog("调试日志",DebugLog)
	//go printLog("错误日志",ErrLog)
	select {}
}

func printLog(name, path string) {
	// 模拟日志输出
	err := os.Remove(path)
	if err != nil {
		log.Println(err)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return
	}

	for t := range time.Tick(time.Second * 1) {
		testLog := fmt.Sprintf("「%s」[%s]\n", name, t.String())
		_, err := f.WriteString(testLog)
		if err != nil {
			log.Println(err.Error())
		}
	}
}
