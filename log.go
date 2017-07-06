// package main

// import "github.com/go-irain/log"

// func main() {

// 	log.Info("基本使用")

// 	// 带有logid的日志块
// 	// 可以每次输出前设置tag，也可以只设置一次tag，后续继承
// 	logi := log.ID(log.CreateID())
// 	logi.Tag("login").Info("username:afocus,password:1234")
// 	logi.Tag("request").Info("used some times")
// 	logi.Debug("继承上面的tag:usedtime")

// 	logi.Tag("respone").Info("usps","")
// 	// 释放logi对象
// 	logi.Free()
//
//
// 	// 存到文件中
// 	fileio, err := log.NewLogFile(&log.FileOption{
// 		Dir:          "./logs",
// 		MaxFileCount: 4,
// 		MaxFileSize:  10 * log.MB,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	// 设置默认的log写入到fileio
// 	// 任何实现io.writer的接口均可设置
// 	log.SetOutput(fileio)
// 	log.Info("我存到文件里了")
// 	logi.Info("我也存到文件了")

// }

package log

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// 默认的日志对象
var std = NewLogger()

// ServiceName 当前进程cmd名字
var ServiceName = filepath.Base(os.Args[0])

// Level 日志等级
type Level uint8

const (
	L_Debug Level = iota
	L_Info
	L_Warnning
	L_Error
)

const (
	_  uint64 = iota
	KB uint64 = 1 << (10 * iota)
	MB
)

func (l Level) String() string {
	switch l {
	case L_Debug:
		return "DBUG"
	case L_Info:
		return "INFO"
	case L_Warnning:
		return "WARN"
	case L_Error:
		return "ERRO"
	default:
		return "UNKN"
	}
}

// CreateID 简单的返回一个随机字符串id
func CreateID() string {
	x := make([]byte, 16)
	io.ReadFull(rand.Reader, x)
	return fmt.Sprintf("%x", x)
}

// ID 返回一个设置了id的日志块
func ID(id string) *LogItem {
	return std.ID(id)
}

// SetOutput 设置默认日志的输出模式
func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

// SetLevel 设置默认日志的日志等级
func SetLevel(lvl Level) {
	std.SetLevel(lvl)
}

func Debug(s ...interface{}) {
	std.Output(2, L_Debug, "", "", fmt.Sprintln(s...))
}

func Info(s ...interface{}) {
	std.Output(2, L_Info, "", "", fmt.Sprintln(s...))
}

func Warnning(s ...interface{}) {
	std.Output(2, L_Warnning, "", "", fmt.Sprintln(s...))
}

func Error(s ...interface{}) {
	std.Output(2, L_Error, "", "", fmt.Sprintln(s...))
}

//

func Debugf(s string, args ...interface{}) {
	std.Output(2, L_Debug, "", "", fmt.Sprintf(s, args...))
}

func Infof(s string, args ...interface{}) {
	std.Output(2, L_Info, "", "", fmt.Sprintf(s, args...))
}

func Warnningf(s string, args ...interface{}) {
	std.Output(2, L_Warnning, "", "", fmt.Sprintf(s, args...))
}

func Errorf(s string, args ...interface{}) {
	std.Output(2, L_Error, "", "", fmt.Sprintf(s, args...))
}
