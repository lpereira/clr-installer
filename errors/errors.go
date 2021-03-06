// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TraceableError is an internal error used to carry trace details
// to be shared across the multiple layers and reporting facilities
type TraceableError struct {
	Trace string
	When  time.Time
	What  string
}

func getTraceIdx(idx int) (string, string, int) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[idx+1])
	file, line := f.FileLine(pc[idx+1])
	return f.Name(), file, line
}

func formatTraceIdx(idx int) (string, string) {
	funcName, file, line := getTraceIdx(idx)
	fileName := filepath.Base(file)

	fn := strings.Split(funcName, "github.com/clearlinux/clr-installer/")

	if len(fn) > 1 {
		funcName = fn[1]
	} else {
		funcName = fn[0]
	}

	dir := strings.Split(filepath.Dir(file), "/clr-installer/")
	var dirName string
	if len(dir) > 1 {
		dirName = dir[1]
	} else {
		dirName = dir[0]
	}

	return funcName, fmt.Sprintf("%s/%s:%d", dirName, fileName, line)
}

func getTrace() string {
	cfName, cTrace := formatTraceIdx(3)
	caller := fmt.Sprintf("%s()\n     %s\n", cfName, cTrace)

	rfName, rTrace := formatTraceIdx(2)
	raiser := fmt.Sprintf("%s()\n     %s\n", rfName, rTrace)

	return fmt.Sprintf("\n\nError Trace:\n%s%s", raiser, caller)
}

func (e TraceableError) Error() string {
	return fmt.Sprintf("%s%s", e.What, e.Trace)
}

// Errorf Returns a new error with the stack information
func Errorf(format string, a ...interface{}) error {
	return TraceableError{
		Trace: getTrace(),
		When:  time.Now(),
		What:  fmt.Sprintf(format, a...),
	}
}

// Wrap returns an error with the caller stack information
// embedded in the original error message
func Wrap(err error) error {
	return Errorf(err.Error())
}
