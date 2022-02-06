package logger

import (
	"fmt"
	"log"
)

type Logger struct {
	name  string
	debug bool
}

func New(loggerName string, debug bool) *Logger {
	return &Logger{
		name:  fmt.Sprintf("[%s]", loggerName),
		debug: debug,
	}
}

func (l *Logger) Info(format string, i ...interface{}) {
	log.Printf(l.name+" [INFO] "+format, i...)
}

func (l *Logger) Err(format string, i ...interface{}) {
	log.Printf(l.name+" [ERROR] "+format, i...)
}

func (l *Logger) Debug(format string, i ...interface{}) {
	if !l.debug {
		return
	}
	log.Printf(l.name+" [DEBUG] "+format, i...)
}
