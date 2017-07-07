package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 实现io.Writer
// 主要做日志分割
type LogFile struct {
	fd     *os.File
	option *FileOption
	rsize  uint32
	name   string
}

// FileOption 日志文件配置信息
type FileOption struct {
	Dir          string
	MaxFileCount int
	MaxFileSize  uint64 // 单位MB
}

func NewLogFile(opt *FileOption) (*LogFile, error) {
	if opt.MaxFileCount < 0 {
		opt.MaxFileCount = 0
	}
	if opt.MaxFileSize != 0 {
		opt.MaxFileSize = opt.MaxFileSize * sizeMB
	}
	opt.Dir, _ = filepath.Abs(strings.TrimSuffix(opt.Dir, "/"))
	f := &LogFile{
		option: opt,
		name:   opt.Dir + "/" + ServiceName + ".log",
	}
	if err := os.MkdirAll(opt.Dir, 0777); err != nil {
		return nil, err
	}
	return f, f.rotate()
}

func (f *LogFile) Write(data []byte) (int, error) {
	n, err := f.fd.Write(data)
	if err != nil {
		return n, err
	}
	f.rsize += uint32(n)
	if f.option.MaxFileSize > 0 && uint64(f.rsize) > f.option.MaxFileSize {
		err = f.rotate()
	}
	return n, err
}

// 获取目录下指定前缀的所有日志文件
func (f *LogFile) removeFiles() {
	fs, err := filepath.Glob(fmt.Sprintf("%s/%s.log.*", f.option.Dir, ServiceName))
	if err != nil {
		return
	}
	sort.Strings(fs)
	x := len(fs) - (f.option.MaxFileCount - 1)
	if f.option.MaxFileCount > 0 && x > 0 {
		dels := fs[:x]
		for _, v := range dels {
			os.Remove(v)
		}
	}
}

// 分割
func (f *LogFile) rotate() error {
	f.removeFiles()
	if f.fd != nil {
		f.fd.Sync()
		f.fd.Close()
		os.Rename(f.name, f.name+time.Now().Format(".20060102150405"))
	}
	fmt.Println("log rotate")
	// 创建最新的日志文件
	fd, err := os.OpenFile(f.name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	fi, err := fd.Stat()
	if err != nil {
		return err
	}
	f.fd = fd
	f.rsize = uint32(fi.Size())
	return nil
}
