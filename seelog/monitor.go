package seelog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hpcloud/tail"
)

type msg struct {
	LogName string `json:"logName"`
	Data    string `json:"data"`
}

// 监控日志文件
func monitor() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
	defer cancel()
	for _, sl := range slogs {

		go func(sl slog) {
			defer func() {
				if err := recover(); err != nil {
					printError(errors.New("monitor() panic"))
				}
			}()

			// 等待文件
			//fileInfo, err := os.Stat(sl.Path)

			// if err != nil {
			// 	printInfo(fmt.Sprintf("等待文件 %s 生成", sl.Path))
			// ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
			// ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
			// defer cancel()
			fileInfo, err := BlockUntilExists(ctx, sl.Path)
			if err != nil {
				printError(err)
				return
			}
			// }
			//fmt.Println(sl.Path)
			printInfo(fmt.Sprintf("开始监控文件 %s", sl.Path))

			// t, _ := tail.TailFile(sl.Path, tail.Config{ReOpen: true, Poll: true, Follow: true, Location: &tail.SeekInfo{
			t, _ := tail.TailFile(sl.Path, tail.Config{Poll: true, Follow: true, Location: &tail.SeekInfo{
				Offset: fileInfo.Size(),
				Whence: 0,
			}})

			for line := range t.Lines {
				manager.broadcast <- msg{sl.Name, line.Text}
			}
		}(sl)
	}

}

func Addmonitor(name, path string) {
	s := slog{name, path}
	// ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
	// defer cancel()
	go func(s slog) {
		defer func() {
			if err := recover(); err != nil {
				printError(errors.New("monitor() panic"))
			}
		}()
		fileInfo, err := os.Stat(s.Path)

		// if err != nil {
		// 	printInfo(fmt.Sprintf("等待文件 %s 生成", s.Path))
		// ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

		//fileInfo, err := BlockUntilExists(ctx, s.Path)
		if err != nil {
			printError(err)
			return
		}
		// }
		fmt.Println(s.Path)
		printInfo(fmt.Sprintf("开始监控文件 %s", s.Path))

		// t, _ := tail.TailFile(path, tail.Config{ReOpen: true, Poll: true, Follow: true, Location: &tail.SeekInfo{
		t, _ := tail.TailFile(s.Path, tail.Config{Poll: true, Follow: true, Location: &tail.SeekInfo{
			Offset: fileInfo.Size(),
			Whence: 0,
		}})

		for line := range t.Lines {
			manager.broadcast <- msg{s.Name, line.Text}
		}
	}(s)

}

func BlockUntilExists(ctx context.Context, fileName string) (os.FileInfo, error) {

	for {
		f, err := os.Stat(fileName)
		if err == nil {
			return f, err
		}

		select {
		case <-time.After(time.Millisecond * 200):
			continue
		case <-ctx.Done():
			return nil, errors.New(fmt.Sprintf("等待 %s 超时", fileName))
		}
	}
}
