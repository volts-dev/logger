package logger

import (
	//	"encoding/json"
	"log"
	"os"
	"runtime"
)

type Brush func(string) string

func NewBrush(color string) Brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []Brush{
	NewBrush("1;31"), // LevelAttack red
	NewBrush("1;31"), // LevelCritical red
	NewBrush("1;31"), // LevelAlert red
	NewBrush("1;35"), // LevelEmergency magenta
	NewBrush("1;37"), // LevelNone white
	NewBrush("1;37"), // LevelInfo white
	NewBrush("1;33"), // LevelWarn yellow
	NewBrush("1;31"), // LevelError red
	NewBrush("1;34"), // LevelDebug blue
}

// ConsoleWriter implements LoggerInterface and writes messages to terminal.
type ConsoleWriter struct {
	log *log.Logger
	//Level int `json:"level"`
}

// create ConsoleWriter returning as LoggerInterface.
func NewConsoleWriter() *ConsoleWriter {
	cw := new(ConsoleWriter)
	cw.log = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	//cw.Level = LevelDebug
	return cw
}

/*
// init console logger.
// jsonconfig like '{"level":LevelTrace}'.
func (c *ConsoleWriter) Init(jsonconfig string) error {
	if len(jsonconfig) == 0 {
		return nil
	}
	err := json.Unmarshal([]byte(jsonconfig), c)
	if err != nil {
		return err
	}
	return nil
}
*/

func (self *ConsoleWriter) Init(config string) error {
	return nil
}

// write message in console.
func (self *ConsoleWriter) Write(level Level, msg string) error {
	if goos := runtime.GOOS; goos == "windows" {
		self.log.Println(msg)
	} else {
		self.log.Println(colors[level](msg))
	}
	return nil
}

// implementing method. empty.
func (self *ConsoleWriter) Destroy() {
}

// implementing method. empty.
func (self *ConsoleWriter) Flush() {
}
