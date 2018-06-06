// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"io"
	"log"

	"github.com/clearlinux/clr-installer/errors"
)

const (
	// LogLevelDebug specified the log level as: DEBUG
	LogLevelDebug = 1

	// LogLevelInfo specified the log level as: INFO
	LogLevelInfo = 2

	// LogLevelWarning specified the log level as: WARNING
	LogLevelWarning = 3

	// LogLevelError specified the log level as: ERROR
	LogLevelError = 4
)

var (
	level    = LogLevelInfo
	levelMap = map[int]string{}
)

func init() {
	levelMap[LogLevelDebug] = "LogLevelDebug"
	levelMap[LogLevelInfo] = "LogLevelInfo"
	levelMap[LogLevelWarning] = "LogLevelWarning"
	levelMap[LogLevelError] = "LogLevelError"
}

// SetLogLevel sets the default log level to l
func SetLogLevel(l int) error {
	if l < LogLevelDebug || l > LogLevelError {
		return errors.Errorf("Invalid log level: %d", l)
	}

	level = l
	return nil
}

// SetOutput sets the default log output to w instead of stdout/stderr
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// LevelStr converts level to its text equivalent, if level is invalid
// an error is returned
func LevelStr(level int) (string, error) {
	for k, v := range levelMap {
		if k == level {
			return v, nil
		}
	}

	return "", fmt.Errorf("Invalid log level: %d", level)
}

func logTag(tag string, format string, a ...interface{}) {
	f := fmt.Sprintf("[%s] %s\n", tag, format)
	str := fmt.Sprintf(f, a...)
	log.Printf(str)
}

// Debug prints a debug log entry with DBG tag
func Debug(format string, a ...interface{}) {
	if level > LogLevelDebug {
		return
	}

	logTag("DBG", format, a...)
}

// Error prints an error log entry with ERR tag
func Error(format string, a ...interface{}) {
	logTag("ERR", format, a...)
}

// ErrorError prints an error log entry with ERR tag, it takes an
// error instead of format and args, if a TraceableError is provided
// then we also include the trace information in the error message
func ErrorError(err error) {
	msg := err.Error()

	if e, ok := err.(errors.TraceableError); ok {
		msg = fmt.Sprintf("%s %s", e.Trace, e.What)
	}

	logTag("ERR", msg)
}

// Info prints an info log entry with INF tag
func Info(format string, a ...interface{}) {
	if level > LogLevelInfo {
		return
	}

	logTag("INF", format, a...)
}

// Out is an special logging function, it's used for command output
// and has no log level restrictions
func Out(format string, a ...interface{}) {
	logTag("OUT", format, a...)
}

// Warning prints an warning log entry with WRN tag
func Warning(format string, a ...interface{}) {
	if level > LogLevelWarning {
		return
	}

	logTag("WRN", format, a...)
}
