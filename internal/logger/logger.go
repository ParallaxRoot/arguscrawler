package logger

import "log"

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Info(msg string) {
	log.Println(msg)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
