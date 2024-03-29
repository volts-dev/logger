package logger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

type (
	IWriter interface {
		Init(config string) error
		Destroy()
		//Flush()
		Write(level Level, msg string) error
	}

	// 创建新Writer类型函数接口
	IWriterType func() IWriter

	TWriterMsg struct {
		level Level
		msg   string
	}

	// Manager all object and API function
	TWriterManager struct {
		//prefix string // prefix to write at beginning of each line
		flag int // properties
		//level      int
		writer       map[string]IWriter // destination for output
		level_writer map[Level]IWriter  // destination for output
		config       *TConfig
		writerName   string       // 现在使用的Writer
		buf          bytes.Buffer // for accumulating text to write
		levelStats   [6]int64

		enableFuncCallDepth bool //report detail of the path,row
		loggerFuncCallDepth int  // 1:function 2:path

		// 异步
		asynchronous bool
		msg          chan *TWriterMsg // 消息通道
		msgPool      *sync.Pool       // 缓存池

		lock sync.Mutex // ensures atomic writes; protects the following fields
	}

	ILogger interface {
		GetLevel() Level
		SetLevel(l Level)

		Assert(cnd bool, format string, args ...interface{})

		Panicf(format string, v ...interface{})
		Dbgf(format string, v ...interface{})
		Atkf(format string, v ...interface{})
		Errf(format string, v ...interface{}) error
		Warnf(format string, v ...interface{})
		Infof(format string, v ...interface{})

		//Panic(v ...interface{})
		Dbg(v ...interface{})
		Atk(v ...interface{})
		Err(v ...interface{}) error
		Warn(v ...interface{})
		Info(v ...interface{})
	}

	// Supply API to user
	TLogger struct {
		manager *TWriterManager
	}
)

var (
	creators = make(map[string]IWriterType) // 注册的Writer类型函数接口
	Logger   = NewLogger()
)

// 断言如果结果和条件不一致就错误
func Assert(cnd bool, format string, args ...interface{}) {
	if !cnd {
		panic(fmt.Sprintf(format, args...))
	}
}

// Returns true if the given level is at or lower the current logger level
func Lvl(level Level, log ...ILogger) bool {
	var l ILogger = Logger
	if len(log) > 0 {
		l = log[0]
	}
	return l.GetLevel() <= level
}
func Atkf(fmt string, arg ...interface{}) {
	Logger.Atkf(fmt, arg...)
}

func Info(err ...interface{}) {
	Logger.Info(err...)
}

func Infof(fmt string, arg ...interface{}) {
	Logger.Infof(fmt, arg...)
}

func Warn(err ...interface{}) {
	Logger.Warn(err...)
}

func Warnf(fmt string, arg ...interface{}) {
	Logger.Warnf(fmt, arg...)
}

func Dbg(err ...interface{}) {
	Logger.Dbg(err...)
}

func Err(err ...interface{}) error {
	return Logger.Err(err...)
}

func Errf(fmt string, arg ...interface{}) error {
	return Logger.Errf(fmt, arg...)
}

func Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func PanicErr(err error, title ...string) bool {
	if err != nil {
		Logger.Dbg(err)
		panic(err)
		//panic("[" + title[0] + "] " + err.Error())
	}
	return false
}

func LogErr(err error, title ...string) bool {
	if err != nil {
		//Logger.ErrorLn(err)
		if len(title) > 0 {
			Logger.Err("[" + title[0] + "] " + err.Error())
		} else {
			Logger.Err(err.Error())
		}

		return true
	}
	return false
}

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, aWriterCreater IWriterType) {
	name = strings.ToLower(name)
	if aWriterCreater == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := creators[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	creators[name] = aWriterCreater
}

func (self *TWriterManager) writeDown(msg string, level Level) {
	for name, wt := range self.writer {
		err := wt.Write(level, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to Write message to adapter:%v,error:%v\n", name, err)
		}
	}
}
func (self *TWriterManager) write(level Level, msg string) error {
	if level > self.config.Level {
		return nil
	}

	if self.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(self.loggerFuncCallDepth)
		if ok {
			_, filename := path.Split(file)
			msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
		}
	}

	msg = "[" + self.config.Prefix + "]" + msg

	// 异步执行
	if self.asynchronous {
		wm := self.msgPool.Get().(*TWriterMsg)
		wm.level = level
		wm.msg = msg
		self.msg <- wm
	} else {
		self.writeDown(msg, level)
	}

	return nil
}

// start logger chan reading.
// when chan is not empty, write logs.
func (self *TWriterManager) listen() {
	for {
		// 如果不是异步则退出监听
		if !self.asynchronous {
			return
		}

		select {
		case wm := <-self.msg:
			// using level writer first
			if wt, has := self.level_writer[wm.level]; has {
				err := wt.Write(wm.level, wm.msg)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg:", err)
				}
			} else {
				for _, wt := range self.writer {
					err := wt.Write(wm.level, wm.msg)
					if err != nil {
						fmt.Println("ERROR, unable to WriteMsg:", err)
					}
				}
			}

		}
	}
}

func NewLogger(opts ...Option) *TLogger {
	config := &TConfig{
		Level:  LevelDebug,
		Prefix: "LOG",
	}
	// init options
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}

	log := &TLogger{}
	log.manager = &TWriterManager{
		writer:              make(map[string]IWriter),
		level_writer:        make(map[Level]IWriter),
		config:              config,
		msg:                 make(chan *TWriterMsg, 10000), //10000 means the number of messages in chan.
		loggerFuncCallDepth: 2,
	}

	//go log.manager.listen()

	log.manager.writer["Console"] = NewConsoleWriter()
	log.manager.writerName = "Console"

	// 缓存池新建对象函数
	log.manager.msgPool = &sync.Pool{
		New: func() interface{} {
			return &TWriterMsg{}
		},
	}

	return log
}

// SetLogger provides a given logger creater into Logger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (self *TLogger) SetWriter(name string, config string) error {
	var wt IWriter
	var has bool
	name = strings.ToLower(name)
	self.manager.lock.Lock()
	defer self.manager.lock.Unlock()

	if wt, has = self.manager.writer[name]; !has {
		if creater, has := creators[name]; has {
			wt = creater()
		} else {
			return fmt.Errorf("Logger.SetLogger: unknown creater %q (forgotten Register?)", name)
		}
	}

	err := wt.Init(config)
	if err != nil {
		fmt.Println("Logger.SetLogger: " + err.Error())
		return err
	}
	self.manager.writer[name] = wt
	self.manager.writerName = name
	return nil
}

// 设置不同等级使用不同警报方式
func (self *TLogger) SetLevelWriter(level Level, writer IWriter) {
	if level > -1 && writer != nil {
		self.manager.level_writer[level] = writer
	}
}

// remove a logger adapter in BeeLogger.
func (self *TLogger) RemoveWriter(name string) error {
	self.manager.lock.Lock()
	defer self.manager.lock.Unlock()
	if wt, has := self.manager.writer[name]; has {
		wt.Destroy()
		delete(self.manager.writer, name)
	} else {
		return fmt.Errorf("Logger.RemoveWriter: unknown writer %v (forgotten Register?)", self)
	}
	return nil
}

func (self *TLogger) GetLevel() Level {
	return self.manager.config.Level
}

func (self *TLogger) SetLevel(level Level) {
	self.manager.config.Level = level
}

// Async set the log to asynchronous and start the goroutine
func (self *TLogger) Async(aSwitch ...bool) *TLogger {
	if len(aSwitch) > 0 {
		self.manager.asynchronous = aSwitch[0]
	} else {
		self.manager.asynchronous = true
	}

	// 避免多次运行 Go 程
	if self.manager.asynchronous {
		go self.manager.listen()
	}

	return self
}

// 断言如果结果和条件不一致就错误
func (self *TLogger) Assert(cnd bool, format string, args ...interface{}) {
	if !cnd {
		panic(fmt.Sprintf(format, args...))
	}
}

// enable log funcCallDepth
func (self *TLogger) EnableFuncCallDepth(b bool) {
	self.manager.lock.Lock()
	defer self.manager.lock.Unlock()
	self.manager.enableFuncCallDepth = b
}

// set log funcCallDepth
func (self *TLogger) SetLogFuncCallDepth(aDepth int) {
	self.manager.loggerFuncCallDepth = aDepth
}

/*
// Log EMERGENCY level message.
func (self *TLogger) Emergency(format string, v ...interface{}) {
	msg := fmt.Sprintf("[M] "+format, v...)
	self.manager.writerMsg(LevelEmergency, msg)
}

// Log ALERT level message.
func (self *TLogger) Alert(format string, v ...interface{}) {
	msg := fmt.Sprintf("[A] "+format, v...)
	self.manager.writerMsg(LevelAlert, msg)
}

// Log CRITICAL level message.
func (self *TLogger) Critical(format string, v ...interface{}) {
	msg := fmt.Sprintf("[C] "+format, v...)
	self.manager.writerMsg(LevelCritical, msg)
}
*/

// Log INFORMATIONAL level message.
func (self *TLogger) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf("[Info]: "+format, v...)
	self.manager.write(LevelInfo, msg)
}

func (self *TLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

// Log WARNING level message.
func (self *TLogger) Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf("[WARM]: "+format, v...)
	self.manager.write(LevelWarn, msg)
}

// Log ERROR level message.
func (self *TLogger) Errf(format string, v ...interface{}) error {
	msg := fmt.Errorf("[ERR]: "+format, v...)
	self.manager.write(LevelError, msg.Error())
	return msg
}

// Log DEBUG level message.
func (self *TLogger) Dbgf(format string, v ...interface{}) {
	msg := fmt.Sprintf("[DBG]: "+format, v...)
	self.manager.write(LevelDebug, msg)
}

// Log Attack level message.
func (self *TLogger) Atkf(format string, v ...interface{}) {
	msg := fmt.Sprintf("[ATK]: "+format, v...)
	self.manager.write(LevelAttack, msg)
}

// Log INFORMATIONAL level message.
func (self *TLogger) Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelInfo, "[INFO] "+msg)
}

// Log WARNING level message.
func (self *TLogger) Warn(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelWarn, "[WARM] "+msg)
}

// Log ERROR level message.
func (self *TLogger) Err(v ...interface{}) error {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelError, "[ERR] "+msg)
	return errors.New(msg)
}

// Log DEBUG level message.
func (self *TLogger) Dbg(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelDebug, "[DBG] "+msg)
}

// Log Attack level message.
func (self *TLogger) Atk(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelAttack, "[ATK] "+msg)
}

/*
// flush all chan data.
func (self *TLogger) Flush() {
	for _, l := range self.manager.writer {
		l.Flush()
	}
}

// close logger, flush all chan data and destroy all adapters in TLogger.
func (self *TLogger) Close() {
	for {
		if len(self.msg) > 0 {
			bm := <-self.msg
			for _, l := range self.outputs {
				err := l.write(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to write (while closing logger):", err)
				}
			}
		} else {
			break
		}
	}
	for _, l := range self.outputs {
		l.Flush()
		l.Destroy()
	}
}
*/
