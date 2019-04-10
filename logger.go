package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

type (
	TConfig struct {
		Level  int    `json:"Level"`
		Prefix string `json:"Prefix"`
	}

	IWriter interface {
		Init(config string) error
		Destroy()
		//Flush()
		Write(level int, msg string) error
	}

	// 创建新Writer类型函数接口
	IWriterType func() IWriter

	TWriterMsg struct {
		level int
		msg   string
	}

	// Manager all object and API function
	TWriterManager struct {
		//prefix string // prefix to write at beginning of each line
		flag int // properties
		//level      int
		writer       map[string]IWriter // destination for output
		level_writer map[int]IWriter    // destination for output
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
		GetLevel() int
		SetLevel(l int)

		Panicf(format string, v ...interface{})
		Dbgf(format string, v ...interface{})
		Atkf(format string, v ...interface{})
		Errf(format string, v ...interface{}) error
		Warnf(format string, v ...interface{})
		Infof(format string, v ...interface{})

		//Panic(v ...interface{})
		Dbg(v ...interface{})
		Atk(v ...interface{})
		Err(v ...interface{})
		Warn(v ...interface{})
		Info(v ...interface{})
	}

	// Supply API to user
	TLogger struct {
		manager *TWriterManager
	}
)

var (
	creaters = make(map[string]IWriterType) // 注册的Writer类型函数接口
	Logger   = NewLogger("")
)

// 断言如果结果和条件不一致就错误
func Assert(cnd bool, format string, args ...interface{}) {
	if !cnd {
		panic(fmt.Sprintf(format, args...))
	}
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

func Err(err ...interface{}) {
	Logger.Err(err...)
}

func Errf(fmt string, arg ...interface{}) {
	Logger.Errf(fmt, arg...)
}

func Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func PanicErr(err error, title ...string) bool {
	if err != nil {
		Logger.Dbg(err)
		panic(err)
		//panic("[" + title[0] + "] " + err.Error())
		return true
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
func Register(aName string, aWriterCreater IWriterType) {
	aName = strings.ToLower(aName)
	if aWriterCreater == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := creaters[aName]; dup {
		panic("logs: Register called twice for provider " + aName)
	}
	creaters[aName] = aWriterCreater
}

func (self *TWriterManager) writeDown(msg string, level int) {
	for name, wt := range self.writer {
		err := wt.Write(level, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to Write message to adapter:%v,error:%v\n", name, err)
		}
	}
}
func (self *TWriterManager) write(aLevel int, aMsg string) error {
	if aLevel > self.config.Level {
		return nil
	}

	if self.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(self.loggerFuncCallDepth)
		if ok {
			_, filename := path.Split(file)
			aMsg = fmt.Sprintf("[%s:%d] %s", filename, line, aMsg)
		}
	}

	// 异步执行
	if self.asynchronous {
		lWM := self.msgPool.Get().(*TWriterMsg)
		lWM.level = aLevel
		lWM.msg = aMsg
		self.msg <- lWM
	} else {

		self.writeDown(aMsg, aLevel)
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

func NewLogger(aConfig string) *TLogger {
	lConfig := new(TConfig)
	lConfig.Level = LevelDebug
	lConfig.Prefix = "vectors"

	if aConfig != "" { // 空字符串会导致错误
		err := json.Unmarshal([]byte(aConfig), lConfig)
		if err != nil {
			fmt.Println("NewMemorySession Unmarshal", err, lConfig)
			return nil
		}
	}
	lLogger := &TLogger{}
	lLogger.manager = &TWriterManager{
		writer:              make(map[string]IWriter),
		level_writer:        make(map[int]IWriter),
		config:              lConfig,
		msg:                 make(chan *TWriterMsg, 10000), //10000 means the number of messages in chan.
		loggerFuncCallDepth: 2,
	}

	//go lLogger.manager.listen()

	lLogger.manager.writer["Console"] = NewConsoleWriter()
	lLogger.manager.writerName = "Console"

	// 缓存池新建对象函数
	lLogger.manager.msgPool = &sync.Pool{
		New: func() interface{} {
			return &TWriterMsg{}
		},
	}

	return lLogger
}

/*
func (self *TLogger) Request(hd *web.THandler) {

}


func (self *TLogger) Response(hd *web.THandler) {

}
*/
// SetLogger provides a given logger creater into Logger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (self *TLogger) SetWriter(aName string, aConfig string) error {
	var wt IWriter
	var has bool
	aName = strings.ToLower(aName)
	self.manager.lock.Lock()
	defer self.manager.lock.Unlock()

	if wt, has = self.manager.writer[aName]; !has {
		if creater, has := creaters[aName]; has {
			wt = creater()
		} else {
			return fmt.Errorf("Logger.SetLogger: unknown creater %q (forgotten Register?)", aName)
		}
	}

	err := wt.Init(aConfig)
	if err != nil {
		fmt.Println("Logger.SetLogger: " + err.Error())
		return err
	}
	self.manager.writer[aName] = wt
	self.manager.writerName = aName
	return nil
}

// 设置不同等级使用不同警报方式
func (self *TLogger) SetLevelWriter(level int, writer IWriter) {
	if level > -1 && writer != nil {
		self.manager.level_writer[level] = writer
	}
}

// remove a logger adapter in BeeLogger.
func (self *TLogger) RemoveWriter(aName string) error {
	self.manager.lock.Lock()
	defer self.manager.lock.Unlock()
	if wt, has := self.manager.writer[aName]; has {
		wt.Destroy()
		delete(self.manager.writer, aName)
	} else {
		return fmt.Errorf("Logger.RemoveWriter: unknown writer %q (forgotten Register?)", self)
	}
	return nil
}

func (self *TLogger) GetLevel() int {
	return self.manager.config.Level
}

func (self *TLogger) SetLevel(aLevel int) {
	self.manager.config.Level = aLevel
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
	msg := fmt.Sprintf("Info: "+format, v...)
	self.manager.write(LevelInfo, msg)
}

func (self *TLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

// Log WARNING level message.
func (self *TLogger) Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf("Warn: "+format, v...)
	self.manager.write(LevelWarn, msg)
}

// Log ERROR level message.
func (self *TLogger) Errf(format string, v ...interface{}) error {
	msg := fmt.Errorf("Err: "+format, v...)
	self.manager.write(LevelError, msg.Error())
	return msg
}

// Log DEBUG level message.
func (self *TLogger) Dbgf(format string, v ...interface{}) {
	msg := fmt.Sprintf("Dbg: "+format, v...)
	self.manager.write(LevelDebug, msg)
}

// Log Attack level message.
func (self *TLogger) Atkf(format string, v ...interface{}) {
	msg := fmt.Sprintf("Atk: "+format, v...)
	self.manager.write(LevelAttack, msg)
}

// Log INFORMATIONAL level message.
func (self *TLogger) Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelInfo, "[I] "+msg)
}

// Log WARNING level message.
func (self *TLogger) Warn(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelWarn, "[W] "+msg)
}

// Log ERROR level message.
func (self *TLogger) Err(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelError, "[E] "+msg)
}

// Log DEBUG level message.
func (self *TLogger) Dbg(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelDebug, "[D] "+msg)
}

// Log Attack level message.
func (self *TLogger) Atk(v ...interface{}) {
	msg := fmt.Sprint(v...)
	self.manager.write(LevelAttack, "[D] "+msg)
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
