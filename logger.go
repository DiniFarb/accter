package accter

import (
	"fmt"
	"time"
)

type Logger struct {
	timeformat string
	appName    string
	level      Level
}

type Level int

const (
	Error Level = iota
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
