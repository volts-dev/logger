package logger

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	lLog := NewLogger("")
	fmt.Println("Strart")

	lLog.SetLevel(LevelDebug)
	lLog.Async()
	lLog.EnableFuncCallDepth(true)
	lLog.Dbg("%s", "Test logger!")
	lLog.Err("%s", "Test logger!")
	lLog.Info("%s", "Test logger!")
	lLog.Warn("%s", "Test logger!")

	fmt.Println("")
	lLog.SetLevel(LevelError)
	lLog.Dbg("%s", "Test logger!")
	lLog.Err("%s", "Test logger!")
	lLog.Info("%s", "Test logger!")
	lLog.Warn("%s", "Test logger!")

	fmt.Println("")
	lLog.SetLevel(LevelWarn)
	lLog.Dbg("%s", "Test logger!")
	lLog.Err("%s", "Test logger!")
	lLog.Info("%s", "Test logger!")
	lLog.Warn("%s", "Test logger!")

	fmt.Println("")
	lLog.SetLevel(LevelInfo)
	lLog.Dbg("%s", "Test logger!")
	lLog.Err("%s", "Test logger!")
	lLog.Info("%s", "Test logger!")
	lLog.Warn("%s", "Test logger!")

	fmt.Println("end")
	//<-make(chan int)
}
