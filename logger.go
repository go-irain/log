package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu     sync.RWMutex
	out    io.Writer
	level  Level
	upload bool
}

var bufpool = &sync.Pool{New: func() interface{} {
	return bytes.NewBuffer([]byte{})
}}

func NewLogger() *Logger {
	return &Logger{
		out: os.Stderr,
	}
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *bytes.Buffer, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	buf.Write(b[bp:])
}

func formatHeader(buf *bytes.Buffer, t time.Time, level Level, tag, id, file string, line int) {
	year, month, day := t.Date()
	itoa(buf, year, 4)
	buf.WriteByte('-')
	itoa(buf, int(month), 2)
	buf.WriteByte('-')
	itoa(buf, day, 2)
	buf.WriteByte(' ')

	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	buf.WriteByte(':')
	itoa(buf, min, 2)
	buf.WriteByte(':')
	itoa(buf, sec, 2)
	buf.WriteByte(' ')

	buf.WriteString(level.String())
	buf.WriteByte(' ')

	if id == "" {
		buf.WriteByte('-')
	} else {
		buf.WriteString(id)
	}
	buf.WriteByte(' ')

	if tag == "" {
		buf.WriteByte('-')
	} else {
		buf.WriteString(tag)
	}
	buf.WriteByte(' ')

	buf.WriteString(file)
	buf.WriteByte(':')
	itoa(buf, line, -1)
	buf.WriteByte(' ')
}

// 格式化消息
func (log *Logger) Output(calldept int, level Level, tag, id, msg string) error {
	log.mu.RLock()
	iswrite := level >= log.level
	// 是否允许上传日志
	isupload := log.upload
	log.mu.RUnlock()
	// 等级不足以输出
	if !iswrite {
		return nil
	}
	buf := bufpool.Get().(*bytes.Buffer)
	buf.Reset()
	_, file, line, ok := runtime.Caller(calldept)
	if !ok {
		file = "???"
		line = 0
	}
	length := len(file) - 1
	for i := length; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1 : length-2]
			break
		}
	}
	formatHeader(buf, time.Now(), level, tag, id, file, line)
	buf.WriteString(msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		buf.WriteByte('\n')
	}

	log.mu.Lock()
	_, err := log.out.Write(buf.Bytes())
	log.mu.Unlock()
	// 需要上传的话异步发送上传队列
	if isupload {
		//todo
	}
	bufpool.Put(buf)
	return err

}

func (log *Logger) SetOutput(out io.Writer) {
	log.mu.Lock()
	log.out = out
	log.mu.Unlock()
}

// SetLevel 设置过滤等级
func (log *Logger) SetLevel(l Level) {
	if l > L_Error {
		l = L_Error
	}
	if l < L_Debug {
		l = L_Debug
	}
	log.mu.Lock()
	log.level = l
	log.mu.Unlock()
}

func (log *Logger) Debug(s ...interface{}) {
	log.Output(2, L_Debug, "", "", fmt.Sprintln(s...))
}

func (log *Logger) Info(s ...interface{}) {
	log.Output(2, L_Info, "", "", fmt.Sprintln(s...))
}

func (log *Logger) Warnning(s ...interface{}) {
	log.Output(2, L_Warnning, "", "", fmt.Sprintln(s...))
}

func (log *Logger) Error(s ...interface{}) {
	log.Output(2, L_Error, "", "", fmt.Sprintln(s...))
}

//

func (log *Logger) Debugf(s string, args ...interface{}) {
	log.Output(2, L_Debug, "", "", fmt.Sprintf(s, args...))
}

func (log *Logger) Infof(s string, args ...interface{}) {
	log.Output(2, L_Info, "", "", fmt.Sprintf(s, args...))
}

func (log *Logger) Warnningf(s string, args ...interface{}) {
	log.Output(2, L_Warnning, "", "", fmt.Sprintf(s, args...))
}

func (log *Logger) Errorf(s string, args ...interface{}) {
	log.Output(2, L_Error, "", "", fmt.Sprintf(s, args...))
}

//

type LogItem struct {
	id, tag string
	log     *Logger
	t       time.Time
	mu      sync.RWMutex
}

var logItemsPool = sync.Pool{
	New: func() interface{} {
		return new(LogItem)
	},
}

func (log *Logger) ID(id string) *LogItem {
	item := logItemsPool.Get().(*LogItem)
	item.log = log
	item.id = id
	item.t = time.Now()
	return item
}

func (o *LogItem) Tag(tag string) *LogItem {
	tag = strings.TrimSpace(tag)
	if tag != "" {
		o.mu.Lock()
		o.tag = tag
		o.mu.Unlock()
	}
	return o
}

func (o *LogItem) Free() {
	o.output(L_Info, fmt.Sprintf("used time %.4fs", time.Now().Sub(o.t).Seconds()))
	logItemsPool.Put(o)
}

func (o *LogItem) output(lvl Level, s string) {
	o.mu.RLock()
	tag := o.tag
	o.mu.RUnlock()
	o.log.Output(3, lvl, tag, o.id, s)
}

func (o *LogItem) Debug(s ...interface{}) *LogItem {
	o.output(L_Debug, fmt.Sprintln(s...))
	return o
}

func (o *LogItem) Info(s ...interface{}) *LogItem {
	o.output(L_Info, fmt.Sprintln(s...))
	return o
}

func (o *LogItem) Warnning(s ...interface{}) *LogItem {
	o.output(L_Warnning, fmt.Sprintln(s...))
	return o
}

func (o *LogItem) Error(s ...interface{}) *LogItem {
	o.output(L_Error, fmt.Sprintln(s...))
	return o
}

func (o *LogItem) Debugf(s string, args ...interface{}) *LogItem {
	o.output(L_Debug, fmt.Sprintf(s, args...))
	return o
}

func (o *LogItem) Infof(s string, args ...interface{}) *LogItem {
	o.output(L_Info, fmt.Sprintf(s, args...))
	return o
}

func (o *LogItem) Warnningf(s string, args ...interface{}) *LogItem {
	o.output(L_Warnning, fmt.Sprintf(s, args...))
	return o
}

func (o *LogItem) Errorf(s string, args ...interface{}) *LogItem {
	o.output(L_Error, fmt.Sprintf(s, args...))
	return o
}
