package logger

import (
	"fmt"
	"log"
)

type Logger struct {
	name string
}

func New(loggerName string) *Logger {
	return &Logger{
		name: fmt.Sprintf("[ %s ]", loggerName),
	}
}

func (l *Logger) Info(format string, i ...interface{}) {
	log.Printf(l.name+" [INFO] "+format, i...)
}

func (l *Logger) Err(format string, i ...interface{}) {
	log.Printf(l.name+" [ERROR] "+format, i...)
}

func (l *Logger) Debug(format string, i ...interface{}) {
	log.Printf(l.name+" [DEBUG] "+format, i...)
}
