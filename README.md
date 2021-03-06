# log
new log for skynet


# 获取

`go get -u github.com/go-irain/log`


# 日志格式

`2017-07-06 15:45:41 INFO 14e0e960525d46c7b1223cf09160cb38 login main:32 used time 0.0005s`

* 日期 `2017-07-06`
* 时间 `15:45:41`
* 等级 `INFO`
* 编号 `14e0e960525d46c7b1223cf09160cb38` (默认-)
* 标签 `login` (默认-) 刻录本条日志的事件名称或分类等。
* 代码 `main:32` main.go文件的32行
* 内容 `used time 0.0005s`

所有分段使用`空格`隔开,分段无内容使用`-`替代。最后的消息内容随意使用空格。


# 输出等级

分为四级，可以通过`SetLevel(log.L_Debug)`设置输出级别，只有当日志级别大于等于设置的级别才会输出

* Error 错误级别日志
* Warnning 警告
* Info 信息
* Debug 调试信息 建议生产环境关闭
* Alarm 报警 此日志触发报警

# 基本使用

默认输出到os.StdErr 也就是控制台
```go

import "github.com/go-irain/log"

func main(){
    log.Info("hello,world")
    log.Infof("my name is %s","afocus")
}

```



# 高级使用 LogItem
一个适合从入口函数开始传到各个处理函数的日志块  
很多时候由于并行处理，导致日志循序混乱无法定位具体日志关联性

demo
```go
// 生成一个随机字符串当作日志的ID
logi:=log.ID(log.CreateID())
// 设置tag
logi.Tag("login")
logi.Info("username:????,pwd:???????")
// 释放对象
logi.Free()


// 支持链式调用
log.ID(log.CreateID()).Tag("request").Info("POST","http://127.0.0.1").Free()
```

并发使用
```go
func runTask(logi *log.LogItem){
    for i:=0;i<1000;i++{
        time.Sleep(time.Second)
        logi.Info("do",i)
    }
    logi.Free()
}


func main(){
    for i:=0;i<4;i++{
        logi:=log.ID(log.CreateID())
        logi.Info("start",i)
        go runTask(logi)
    }
    time.Sleep(time.Hour)
}

```


# 存到文件并进行日志自动分割

任何实现了`io.Writer`接口的都可以被当作日志的输出源  


## LogFile
内部已经实现一个可以分割文件的`io.Writer` `LogFile`  
如果需要记录到文件并分割的话，请使用他

demo
```go
opt:=&log.FileOption{
    Dir : "./", // 日志路径，可以绝对，可以相对
    MaxFileCount : 10, // 最大保存文件数量
    MaxFileSize : 200, // 单个文件最大大小 单位MB
}


file,err:=log.NewLogFile(opt)
// 设置out输出为logfile
log.SetOutput(file)

// 正常使用
log.Info(".....")
```

## 自定义输出地

```
// 文件
f,err:=os.Open("....")
log.SetOutput(f)

// 写到net.Conn
conn,err:=net.Dial("tcp","127.0.0.1:8888")
log.SetOutput(conn)
```


