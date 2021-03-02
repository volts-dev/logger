package logger

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	log := NewLogger("")
	fmt.Println("Strart")

	log.SetLevel(LevelDebug)
	log.Async()
	log.EnableFuncCallDepth(true)
	log.Dbgf("%s", "Test logger!")
	log.Errf("%s", "Test logger!")
	log.Infof("%s", "Test logger!")
	log.Warnf("%s", "Test logger!")

	fmt.Println("")
	log.SetLevel(LevelError)
	log.Dbgf("%s", "Test logger!")
	log.Errf("%s", "Test logger!")
	log.Infof("%s", "Test logger!")
	log.Warnf("%s", "Test logger!")

	fmt.Println("")
	log.SetLevel(LevelWarn)
	log.Dbgf("%s", "Test logger!")
	log.Errf("%s", "Test logger!")
	log.Infof("%s", "Test logger!")
	log.Warnf("%s", "Test logger!")

	fmt.Println("")
	log.SetLevel(LevelInfo)
	log.Dbgf("%s", "Test logger!")
	log.Errf("%s", "Test logger!")
	log.Infof("%s", "Test logger!")
	log.Warnf("%s", "Test logger!")

	fmt.Println("end")
	//<-make(chan int)
}
