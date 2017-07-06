package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"regexp"
	"strconv"
	"strings"
)

// 实现io.Writer
// 主要做日志分割
type LogFile struct {
	fd     *os.File
	option *FileOption
	rsize  uint32
	name   string
	reg    *regexp.Regexp
}

// FileOption 日志文件配置信息
type FileOption struct {
	Dir          string
	MaxFileCount int
	MaxFileSize  uint64
}

func NewLogFile(opt *FileOption) (*LogFile, error) {
	opt.Dir, _ = filepath.Abs(strings.TrimSuffix(opt.Dir, "/"))
	f := &LogFile{
		option: opt,
	}
	if err := os.MkdirAll(opt.Dir, 0777); err != nil {
		return nil, err
	}
	f.reg = regexp.MustCompile(ServiceName + `_(\d+).log`)
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
func (f *LogFile) getFiles() []string {
	fs, err := filepath.Glob(fmt.Sprintf("%s/%s_*.log", f.option.Dir, ServiceName))
	if err != nil {
		return []string{}
	}
	sort.Strings(fs)
	x := len(fs) - f.option.MaxFileCount
	if f.option.MaxFileCount > 0 && x > 0 {
		dels := fs[:x]
		for _, v := range dels {
			os.Remove(v)
		}
		fs = fs[x:]
	}
	return fs
}

// 分割
func (f *LogFile) rotate() error {
	fs := f.getFiles()
	var index int
	if flen := len(fs); flen > 0 {
		lastfile := fs[len(fs)-1]
		index = f.matchNumber(filepath.Base(lastfile))
		if f.fd != nil {
			f.fd.Sync()
			f.fd.Close()
			index++
		} else {
			info, err := os.Stat(lastfile)
			if err != nil {
				return err
			}
			if info.Size() < int64(f.option.MaxFileSize) {
				f.rsize = uint32(info.Size())
				fd, err := f.createFile(index)
				if err == nil {
					f.fd = fd
				}
				return err
			}
		}
	}
	return f.mustCreate(index)
}

func (f *LogFile) matchNumber(fname string) int {
	fs := f.reg.FindStringSubmatch(fname)
	var index int
	if len(fs) == 2 {
		index, _ = strconv.Atoi(fs[1])
	}
	return index
}

func (f *LogFile) mustCreate(index int) error {
	var fd *os.File
	var err error
	for {
		if fd, err = f.createFile(index); err != nil {
			if err == os.ErrExist {
				index++
				continue
			}
			break
		} else {
			f.fd = fd
			f.rsize = 0
			return nil
		}
	}
	return err
}

func (f *LogFile) createFile(index int) (*os.File, error) {
	name := fmt.Sprintf(
		"%s/%s_%06d.log",
		f.option.Dir, ServiceName, index,
	)
	return os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
}
