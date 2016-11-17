package logging

import (
	"fmt"
	"time"
	"log"
)

// TODO Add facade with functions for setting logging output level.

// Copied from `appengine.Context`.
type Logger interface {
	// Debugf formats its arguments according to the format, analogous to fmt.Printf,
	// and records the text as a log message at Debug level.
	Debugf(format string, args ...interface{})
	
	// Infof is like Debugf, but at Info level.
	Infof(format string, args ...interface{})
	
	// Warningf is like Debugf, but at Warning level.
	Warningf(format string, args ...interface{})
	
	// Errorf is like Debugf, but at Error level.
	Errorf(format string, args ...interface{})
	
	// Criticalf is like Debugf, but at Critical level.
	Criticalf(format string, args ...interface{})
}

type InitLogger struct {
}

func (logger *InitLogger) Debugf(format string, args ...interface{}) {
	log.Printf("DEBUG(init): " + format + "\n", args...)
}
func (logger *InitLogger) Infof(format string, args ...interface{}) {
	log.Printf("INFO(init): " + format + "\n", args...)
}

func (logger *InitLogger) Warningf(format string, args ...interface{}) {
	log.Printf("WARN(init): " + format + "\n", args...)
}

func (logger *InitLogger) Errorf(format string, args ...interface{}) {
	log.Printf("ERROR(init): " + format + "\n", args...)
}

func (logger *InitLogger) Criticalf(format string, args ...interface{}) {
	log.Printf("CRITICAL(init): " + format + "\n", args...)
}

type RecordingLogger struct {
	wrapped Logger
	init    bool
	Entries []string
}

func NewRecordingLogger(wrapped Logger, init bool) *RecordingLogger {
	return &RecordingLogger{wrapped: wrapped, init: init}
}

func (l *RecordingLogger) add(kind string, format string, args ...interface{}) string {
	ts := time.Now().String()
	msg := fmt.Sprintf(format, args...)
	
	initStr := ""
	if l.init {
		initStr = "[init] "
	}
	fullMsg := fmt.Sprintf("(%s): %s%s", ts, initStr, msg)
	
	l.Entries = append(l.Entries, kind + fullMsg)
	return fullMsg
}

func (l *RecordingLogger) Debugf(format string, args ...interface{}) {
	msg := l.add("DEBUG", format, args...)
	if l.wrapped != nil {
		l.wrapped.Debugf(msg)
	}
}

func (l *RecordingLogger) Infof(format string, args ...interface{}) {
	msg := l.add("INFO", format, args...)
	if l.wrapped != nil {
		l.wrapped.Infof(msg)
	}
}

func (l *RecordingLogger) Warningf(format string, args ...interface{}) {
	msg := l.add("WARN", format, args...)
	if l.wrapped != nil {
		l.wrapped.Warningf(msg)
	}
}

func (l *RecordingLogger) Errorf(format string, args ...interface{}) {
	msg := l.add("ERROR", format, args...)
	if l.wrapped != nil {
		l.wrapped.Errorf(msg)
	}
}

func (l *RecordingLogger) Criticalf(format string, args ...interface{}) {
	msg := l.add("CRITICAL", format, args...)
	if l.wrapped != nil {
		l.wrapped.Criticalf(msg)
	}
}
