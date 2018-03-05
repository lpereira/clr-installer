package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	fakeExternalTrace = false
)

// TraceableError is an internal error used to carry trace details
// to be shared across the multiple layers and reporting facilities
type TraceableError struct {
	Trace string
	When  time.Time
	What  string
}

func getTraceIdx(idx int) (string, int) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[idx+1])
	file, line := f.FileLine(pc[idx+1])

	if fakeExternalTrace {
		file = "/external/source/file.go"
	}

	return file, line
}

func getTrace() string {
	split := "/src/"
	file, line := getTraceIdx(1)

	if !strings.Contains(file, split) {
		idx := 2

		if fakeExternalTrace {
			idx = 1
			split = ""
		}

		file, line = getTraceIdx(idx)
	}

	dir := strings.Split(filepath.Dir(file), split)
	fName := filepath.Base(file)

	return fmt.Sprintf("%s/%s:%d", dir[1], fName, line)
}

func (e TraceableError) Error() string {
	return e.What
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
