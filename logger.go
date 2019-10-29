package logger

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var logger *Logger

type Logger struct {
	RecordCh  chan *Record
	FileName string     // 文件名
	FileDir string     // 文件名
	File      *os.File
	Level     Level
	When      string   // "M", "H", "D", "W"
	WhenInterval int64 // 根据when计算时间间隔
	Ts        int64   // 打印日志的时间戳
	ExpiryTs  int64  // 写此文件的过期时间

	EndCh     chan bool // 文件句柄结束的channel
	ExitCh    chan bool // 程序退出的channel
	BackupCount int
}

type Record struct {
	RLevel  Level
	RTime   string
	RMsg    string
	LineNum int
	File   string
}

func InitLogger(when string, backupCount int, level Level, fileDir, fileName string) (*Logger, error){
	// 合法性校验
	if !IsWhenValid(when) {
		err := fmt.Errorf("init logger, when is invalid")
		return nil, err
	}

	// 初始化logger
	logger = &Logger{
		RecordCh: make(chan *Record, 1024),
		FileDir: fileDir,
		FileName: fileName,
		File:     nil,
		Level:    level,
		When:     when,
		WhenInterval: GetExpiryInterval(when),
		Ts:       0,
		ExpiryTs: 0,
		EndCh:    make(chan bool, 1),
		ExitCh:   make(chan bool, 1),
		BackupCount: backupCount,
	}

	// 初始化轮转机制
	err := logger.InitRotate()
	if err != nil {
		fmt.Printf("%s [%s] %s", logger.GetPreTimeStr(), LevelError.String(), err.Error())
		return nil, err
	}

	// 监控文件句柄结束
	go func() {
		for range logger.EndCh {
			logger.EndFile()
		}
	}()

	// 监控日志程序结束
	go func() {
		for range logger.ExitCh {
			logger.ExitLogger()
		}
	}()

	// 写日志
	go func() {
		for r := range logger.RecordCh {
			// 判断是否需要轮转
			if logger.IsRotate() {
				err := logger.InitRotate()
				if err != nil {
					fmt.Printf("logger.InitRotate err:%v\n", err)
					logger.ExitLogger()
				}
			}

			_, err := fmt.Fprintf(logger.File, r.String())
			if err != nil {
				fmt.Printf("fprintf file err:%v", err)
				logger.ExitLogger()
				return
			}
		}
	}()

	// TODO 删除老的文件
	go func() {



	}()

	return  logger, nil
}

// 获取过期的时间间隔
func GetExpiryInterval(when string) int64 {
	switch when {
	case "M":
		return 60
	case "H":
		return 60*60
	case "D":
		return 60*60*24
	case "W":
		return 60*60*24*7
	default:
		return math.MaxInt64
	}
}

func ( l *Logger) IsRotate() bool {
	t := time.Now().Unix()
	return t > l.ExpiryTs
}

func ( l *Logger) Infof (format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller error")
		return
	}
	r := &Record{
		RLevel: LevelInfo,
		RTime:  l.GetPreTimeStr(),
		RMsg:   fmt.Sprintf(format, args...),
		LineNum:line,
		File: file,
	}
	l.RecordCh <- r
}

// 初始化文件句柄
func ( l *Logger) InitRotate() error {
	// 更新过期时间
	l.UpdateExpiryTs()

	// 创建文件句柄
	file , err := os.OpenFile(l.GetAbsoluteFilePath(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Printf("%s [ERROR] %s\n", l.GetPreTimeStr(), err.Error())
	}
	l.File = file

	return nil
}

func ( l *Logger) GetPreTimeStr() string {
	return  time.Now().Format("2006-01-02 15:04:05")
}

func ( l *Logger) EndFile() {
	err := l.File.Close()
	if err != nil {
		fmt.Printf("%s [%s] %s", l.GetPreTimeStr(), LevelError.String(), err.Error())
	}

	err = l.InitRotate()
	if err != nil {
		fmt.Printf("%s [%s] %s", l.GetPreTimeStr(), LevelError.String(), err.Error())
	}
}

func ( l *Logger) ExitLogger() {
	err := l.File.Close()
	if err != nil {
		fmt.Printf("%s [%s] %s \n", l.GetPreTimeStr(), LevelError.String(), err.Error())
	}

	close(l.EndCh)
	close(l.ExitCh)
	close(l.RecordCh)

	fmt.Printf("%s [%s] logger process is exit！\n", l.GetPreTimeStr(), LevelInfo.String())
	os.Exit(0)
}

func ( l *Logger) UpdateExpiryTs() {
	t := time.Now().Unix()
	l.ExpiryTs = t - t%l.WhenInterval + l.WhenInterval
}

func ( l *Logger) GetFileSuffixName() string{
	switch l.When {
	case "M":
		return time.Unix(l.ExpiryTs, 0).Format("2006-01-02_15-04")
	case "H":
		return time.Unix(l.ExpiryTs, 0).Format("2006-01-02_15")
	case "D":
		return time.Unix(l.ExpiryTs, 0).Format("2006-01-02")
	}
	return ""
}

func ( l *Logger) GetAbsoluteFilePath() string {
	return fmt.Sprintf("%s/%s_%s.log", l.FileDir, l.FileName,  l.GetFileSuffixName())
}

func ( l *Logger) Close() {
	l.ExitCh <- true
}

func IsWhenValid(when string) bool{
	switch when {
	case "M", "H", "D", "W":
		return true
	default:
		return false
	}
}


func (r *Record) String() string {
	return fmt.Sprintf("[%s] %s %s %s.%d\n", r.RTime, r.RLevel.String(), r.RMsg, filepath.Base(r.File), r.LineNum)
}