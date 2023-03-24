package accter

import (
	"fmt"
	"sync"
	"time"
)

type Logger struct {
	timeformat string
	appName    string
	level      Level
}

type Level int

const (
	_ Level = iota
	Error
	Warn
	Info
	Debug
	Trace
)

func (l *Logger) trace(message string, args ...interface{}) {
	if l.level >= Trace {
		l.log(fmt.Sprintf(" %s |  TRACE | %s", l.createTimeStamp(), fmt.Sprintf(message, args...)))
	}
}

func (l *Logger) debug(message string, args ...interface{}) {
	if l.level >= Debug {
		l.log(fmt.Sprintf(" %s |  DEBUG | %s", l.createTimeStamp(), fmt.Sprintf(message, args...)))
	}
}

func (l *Logger) info(message string, args ...interface{}) {
	if l.level >= Info {
		l.log(fmt.Sprintf(" %s |  INFO  | %s", l.createTimeStamp(), fmt.Sprintf(message, args...)))
	}
}

func (l *Logger) warn(message string, args ...interface{}) {
	if l.level >= Warn {
		l.log(fmt.Sprintf(" %s |  WARN  | %s", l.createTimeStamp(), fmt.Sprintf(message, args...)))
	}
}

func (l *Logger) error(message string, args ...interface{}) {
	if l.level >= Error {
		l.log(fmt.Sprintf(" %s |  ERROR | %s", l.createTimeStamp(), fmt.Sprintf(message, args...)))
	}
}

func (l *Logger) createTimeStamp() string {
	return time.Now().Format(l.timeformat)
}

func (l *Logger) log(message string) {
	fmt.Printf("[%s]%s\n", l.appName, message)
}

var logger *Logger

var once sync.Once

func CreateLogger(level Level) {
	once.Do(func() {
		logger = &Logger{
			timeformat: "2006-01-02 15:04:05.000",
			appName:    "ACCTER",
			level:      level,
		}
	})
}
