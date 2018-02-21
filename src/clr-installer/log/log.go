package log

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
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
func SetLogLevel(l int) {
	level = l
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

func getTrace() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[1])
	file, line := f.FileLine(pc[1])
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
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
// error instead of format and args
func ErrorError(err error) {
	logTag("ERR", fmt.Sprintf("%s %s", getTrace(), err))
}

// Info prints an info log entry with INF tag
func Info(format string, a ...interface{}) {
	if level > LogLevelInfo {
		return
	}

	logTag("INF", format, a...)
}

// Warning prints an warning log entry with WRN tag
func Warning(format string, a ...interface{}) {
	if level > LogLevelWarning {
		return
	}

	logTag("WRN", format, a...)
}
