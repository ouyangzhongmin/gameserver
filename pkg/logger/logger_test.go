package logger

import (
	"testing"
	"time"
)

func Test_log(t *testing.T) {
	Init(&LogConf{
		FileName: "",
		Level:    "trace",
	})

	Log.Trace("hello")
	Log.Error("error")
	Log.Println("print")
}

func Test_level_log(t *testing.T) {
	Init(&LogConf{
		FileName: "",
		Level:    "debug",
	})

	defer func() {
		if err := recover(); err != nil {
			//Warn("Panic")
		}
	}()

	Log.Trace("Trace")
	time.Sleep(time.Duration(time.Millisecond * 500))
	Log.Debug("Debug")
	time.Sleep(time.Duration(time.Millisecond * 500))
	Log.Info("Info")
	time.Sleep(time.Duration(time.Millisecond * 500))
	Log.Warn("Warn")
	Log.Error("Error")
	Log.Panic("Panic")
	t.Fatal("Fatal")

}
