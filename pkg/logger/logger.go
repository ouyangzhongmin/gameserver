// 基于logrus封装的log打印
package logger

import (
	"log"
	"os"

	logrus "github.com/sirupsen/logrus"
)

var Log = logrus.StandardLogger()

var (
	Panic      = Log.Panic
	Fatal      = Log.Fatal
	Error      = Log.Error
	Errorln    = Log.Errorln
	Errorf     = Log.Errorf
	Warn       = Log.Warn
	Warnf      = Log.Warnf
	Warnln     = Log.Warnln
	Warningln  = Log.Warningln
	Info       = Log.Info
	Infof      = Log.Infof
	Infoln     = Log.Infoln
	Debug      = Log.Debug
	Debugln    = Log.Debugln
	Debugf     = Log.Debugf
	Trace      = Log.Trace
	Traceln    = Log.Traceln
	WithFields = Log.WithFields
	Println    = Log.Println
	Printf     = Log.Printf
	Print      = Log.Print
	Panicf     = Log.Panicf
	Fatalf     = Log.Fatalf
)

// var Log = log.New(os.Stderr, "", log.LstdFlags)
// "panic","fatal","error","warn","info","debug","trace"
type LogConf struct {
	FileName     string
	Level        string
	ReportCaller bool
	ForceColors  bool
}

func Init(cnf *LogConf) {
	if len(cnf.FileName) > 0 {
		file, err := os.OpenFile(cnf.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			//设置标准Log库的日志文件到logrus同一份里
			log.SetOutput(file)
			Log.SetOutput(file)
		}
	}
	Log.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		ForceColors:     cnf.ForceColors,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	if cnf.Level != "" {
		level, err := logrus.ParseLevel(cnf.Level)
		if err == nil {
			Log.SetLevel(level)
		} else {
			Log.Println("logrus.ParseLevel error:", err)
		}
	}
	if cnf.ReportCaller {
		Log.SetReportCaller(cnf.ReportCaller)
	}
}
