package logger

import (
	"log"
	"os"
)

type Logger struct {
	l *log.Logger
}

func New() *Logger {
	return &Logger{
		l: log.New(os.Stdout, "[ArgusCrawler] ", log.LstdFlags),
	}
}

func (lg *Logger) Infof(format string, v ...interface{}) {
	lg.l.Printf(format, v...)
}

func (lg *Logger) Errorf(format string, v ...interface{}) {
	lg.l.Printf("[ERROR] "+format, v...)
}
