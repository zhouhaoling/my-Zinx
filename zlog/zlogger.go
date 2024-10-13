package zlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const LOG_MAX_BUF = 1024 * 1024

// 日志头部信息标记位，采用 bitmap 方式，
// 用户可以选择头部需要那些标记位被打印
const (
	BitDate         = 1 << iota                            // 日期标志位 2019/01/23
	BitTime                                                // 时间标记位 01:23:12
	BitMicroSeconds                                        // 微秒级标记位 01:23:12.111222
	BitLongFile                                            // 完整文件名称 /home/go/src/zinx/server.go
	BitShortFile                                           // 最后文件名 server.go
	BitLevel                                               // 当前日志级别
	BitStdFlag      = BitDate | BitTime                    // 标准头部日志格式
	BitDefault      = BitLevel | BitShortFile | BitStdFlag // 默认日志头部格式
)

// 日志级别
const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
	LogPanic
	LogFatal
)

// 日志级别对应的显示字符串
var levels = []string{
	"[DEBUG]",
	"[INFO]",
	"[WARN]",
	"[ERROR]",
	"[PANIC]",
	"[FATAL]",
}

type ZinxLogger struct {
	// 确保多协程读写的线程安全
	mu sync.Mutex
	// 每行 log 日志的前缀字符串，拥有日志标记
	prefix string
	// 日志标记位
	flag int
	// 日志输出的文件描述符
	out io.Writer
	// 输出的缓冲区
	buf bytes.Buffer
	// 当前日志绑定的输出文件
	file *os.File
	// 是否打印调试的 debug 信息
	debugClose bool
	// 获取日志文件名和代码上述的 runtime.call 的函数调用层数
	calledDepth int
}

// 将一个整形转换成一个固定长度的字符串，字符串宽度应该是大于0的
// 要确保buffer是有容量空间的
func itoa(buf *bytes.Buffer, val int, width int) {
	u := uint(val)
	if u == 0 && width <= 1 {
		buf.WriteByte('0')
		return
	}

	//
	var b [32]byte
	bp := len(b)
	for ; u > 0 || width > 0; u /= 10 {
		bp--
		width--
		b[bp] = byte(u%10) + '0'
	}

	for bp < len(b) {
		buf.WriteByte(b[bp])
		bp++
	}
}

// 格式头信息
func (log *ZinxLogger) formatHeader(t time.Time, file string, line int, level int) {
	var buf = &log.buf
	// 如果当前前缀字符串不为空，那么需要先写前缀
	if log.prefix != "" {
		buf.WriteByte('<')
		buf.WriteString(log.prefix)
		buf.WriteByte('>')
	}

	// 已经设置了时间相关的标识位,那么需要加时间信息在日志头部
	if log.flag&(BitDate|BitTime|BitMicroSeconds) != 0 {
		// 日期位被标记
		if log.flag&BitDate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			buf.WriteByte('/')
			itoa(buf, int(month), 2)
			buf.WriteByte('/')
			itoa(buf, day, 2)
			buf.WriteByte(' ')
		}

		// 时钟位被标记
		if log.flag&(BitTime|BitMicroSeconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			buf.WriteByte(':')
			itoa(buf, min, 2)
			buf.WriteByte(':')
			itoa(buf, sec, 2)
			if log.flag&BitMicroSeconds != 0 {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			buf.WriteByte(' ')
		}
	}

	// 日志级别位被标记
	if log.flag&BitLevel != 0 {
		buf.WriteString(levels[level])
		buf.WriteByte(' ')
	}

	// 日志当前代码调用文件名称位被标记
	if log.flag&(BitShortFile|BitLongFile) != 0 {
		// 短文件名称
		if log.flag&BitShortFile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		buf.WriteString(file)
		buf.WriteByte(':')
		itoa(buf, line, -1) // 行数
		buf.WriteString(": ")
	}
}

// 输出日志文件
func (log *ZinxLogger) OutPut(level int, s string) error {
	now := time.Now()
	var file string
	var line int
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.flag&(BitShortFile|BitLongFile) != 0 {
		log.mu.Unlock()
		var ok bool
		// 得到当前调用者的文件名称和执行到的代码行数
		_, file, line, ok = runtime.Caller(log.calledDepth)
		if !ok {
			file = "unknown-file"
			line = 0
		}
		log.mu.Lock()
	}

	// 清空 buf
	log.buf.Reset()
	// 写日志头
	log.formatHeader(now, file, line, level)
	// 写日志内容
	log.buf.WriteString(s)
	// 补充换行符
	if len(s) > 0 && s[len(s)-1] != '\n' {
		log.buf.WriteByte('\n')
	}

	// 将填充好的 buf 写到 IO 输出上
	_, err := log.out.Write(log.buf.Bytes())
	return err
}

func (log *ZinxLogger) Debugf(format string, v ...interface{}) {
	if log.debugClose {
		return
	}
	log.OutPut(LogDebug, fmt.Sprintf(format, v...))
}

func (log *ZinxLogger) Debug(v ...interface{}) {
	if log.debugClose {
		return
	}
	log.OutPut(LogDebug, fmt.Sprintln(v...))
}

func (log ZinxLogger) Infof(format string, v ...interface{}) {
	log.OutPut(LogInfo, fmt.Sprintf(format, v...))
}

func (log ZinxLogger) Info(v ...interface{}) {
	log.OutPut(LogInfo, fmt.Sprintln(v...))
}

// ====> Warn <====
func (log *ZinxLogger) Warnf(format string, v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintf(format, v...))
}

func (log *ZinxLogger) Warn(v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintln(v...))
}

// ====> Error <====
func (log *ZinxLogger) Errorf(format string, v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintf(format, v...))
}

func (log *ZinxLogger) Error(v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintln(v...))
}

// ====> Fatal 需要终止程序 <====
func (log *ZinxLogger) Fatalf(format string, v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (log *ZinxLogger) Fatal(v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintln(v...))
	os.Exit(1)
}

// ====> Panic  <====
func (log *ZinxLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

func (log *ZinxLogger) Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

func (log *ZinxLogger) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, LOG_MAX_BUF)
	n := runtime.Stack(buf, true)
	s += string(buf[:n])
	s += "\n"
	log.OutPut(LogError, s)
}

// 获取当前日志bitmap标记
func (log *ZinxLogger) Flags() int {
	log.mu.Lock()
	defer log.mu.Unlock()

	return log.flag
}

// 重新设置日志 Flags bitMap 标记位
func (log *ZinxLogger) ResetFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()

	log.flag = flag
}

// 添加flag标记
func (log *ZinxLogger) AddFlag(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()

	log.flag |= flag
}

// 设置日志的 用户自定义前缀字符串
func (log *ZinxLogger) SetPrefix(prefix string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.prefix = prefix
}

// 判断日志文件是否存在
func (log *ZinxLogger) checkFileExist(fileName string) bool {
	exist := true
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// 关闭日志绑定的文件
func (log *ZinxLogger) closeFile() {
	if log.file != nil {
		log.file.Close()
		log.file = nil
		log.out = os.Stderr
	}
}

// 创建日志文件夹
func mkdirLog(dir string) (e error) {
	_, err := os.Stat(dir)
	b := err == nil || os.IsExist(err)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

// 设置日志文件输出
func (log *ZinxLogger) SetLogFile(fileDir, fileName string) {
	var file *os.File
	// 创建日志文件夹
	mkdirLog(fileDir)

	fullPath := fileDir + "/" + fileName
	if log.checkFileExist(fullPath) {
		// 文件存在,打开
		file, _ = os.OpenFile(fullPath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		// 文件不存在,创建
		file, _ = os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}

	log.mu.Lock()
	defer log.mu.Unlock()

	// 关闭之前绑定的文件
	log.closeFile()
	log.file = file
	log.out = file
}

/*
回收日志处理
*/
func CleanZinxLog(log *ZinxLogger) {
	log.closeFile()
}

func (log *ZinxLogger) CloseDebug() {
	log.debugClose = true
}

func (log *ZinxLogger) OpenDebug() {
	log.debugClose = false
}

func NewZinxLog(out io.Writer, prefix string, flag int) *ZinxLogger {
	// 默认 debug 打开，calledDepth 深度为2，
	// ZinxLogger 对象调用日志打印方法最多调用两层到达 output 函数
	zlog := &ZinxLogger{
		out:         out,
		prefix:      prefix,
		flag:        flag,
		file:        nil,
		debugClose:  false,
		calledDepth: 2,
	}
	// 设置 log 对象，回收资源，析构方法(不设置也可以，go的Gc会自动回收)
	runtime.SetFinalizer(zlog, CleanZinxLog)
	return zlog
}
