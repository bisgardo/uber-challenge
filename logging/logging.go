package logging

import (
	"fmt"
	"time"
	"log"
)

// TODO Add facade with functions for setting logger level

// TODO Make thread-safe! Instead of globally wrapping logger, a request method should have it's own wrapper of both the recorder and the actual logger

// Copied from App Engine Context.
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

// TODO Make sure that "DEBUG(init)" etc. is also included in the recorded log.

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
	Entries []string
	wrapped Logger
}

func (logger * RecordingLogger) Wrap(wrapped Logger) {
	logger.wrapped = wrapped
}
func (logger * RecordingLogger) Unwrap() {
	logger.wrapped = nil
}

func (logger *RecordingLogger) add(kind string, format string, args ...interface{}) {
	ts := time.Now().String()
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("%v(%v): %v", kind, ts, msg)
	logger.Entries = append(logger.Entries, entry)
}

func (logger *RecordingLogger) Debugf(format string, args ...interface{}) {
	logger.add("DEBUG", format, args...)
	if logger.wrapped != nil {
		logger.wrapped.Debugf(format, args...)
	}
}
func (logger *RecordingLogger) Infof(format string, args ...interface{}) {
	logger.add("INFO", format, args...)
	if logger.wrapped != nil {
		logger.wrapped.Infof(format, args...)
	}
}
func (logger *RecordingLogger) Warningf(format string, args ...interface{}) {
	logger.add("WARN", format, args...)
	if logger.wrapped != nil {
		logger.wrapped.Warningf(format, args...)
	}
}
func (logger *RecordingLogger) Errorf(format string, args ...interface{}) {
	logger.add("ERROR", format, args...)
	if logger.wrapped != nil {
		logger.wrapped.Errorf(format, args...)
	}
}
func (logger *RecordingLogger) Criticalf(format string, args ...interface{}) {
	logger.add("CRITICAL", format, args...)
	if logger.wrapped != nil {
		logger.wrapped.Criticalf(format, args...)
	}
}

func (logger *RecordingLogger) Clear() {
	logger.Entries = nil 
}
