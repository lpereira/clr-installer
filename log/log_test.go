package log

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/clearlinux/clr-installer/errors"
)

func TestTag(t *testing.T) {
	tests := []struct {
		msg string
		tag string
		fc  func(fmt string, args ...interface{})
	}{
		{"debug tag test", "[DBG]", Debug},
		{"info tag test", "[INF]", Info},
		{"warning tag test", "[WRN]", Warning},
		{"error tag test", "[ERR]", Error},
	}

	w := bytes.NewBuffer(nil)
	SetOutput(w)
	if err := SetLogLevel(LogLevelDebug); err != nil {
		t.Fatal("Shoul not fail with a valid level")
	}

	for _, curr := range tests {
		curr.fc(curr.msg)

		str := w.String()
		if str == "" {
			t.Fatal("No log written to output")
		}

		if !strings.Contains(str, curr.tag) {
			t.Fatalf("Log generated an entry without the expected tag: %s - entry: %s",
				curr.tag, str)
		}

		w.Reset()
	}
}

func TestErrorError(t *testing.T) {
	w := bytes.NewBuffer(nil)
	SetOutput(w)

	ErrorError(fmt.Errorf("testing log with error"))

	str := w.String()
	if str == "" {
		t.Fatal("No log written to output")
	}
}

func TestLogLevel(t *testing.T) {
	tests := []struct {
		mutedLevel int
		msg        string
		fc         func(fmt string, args ...interface{})
	}{
		{LogLevelInfo, "Debug() log with LogLevelInfo", Debug},
		{LogLevelWarning, "Info() with LogLevelWarning", Info},
		{LogLevelError, "Warning() with LogLevelError", Warning},
	}

	w := bytes.NewBuffer(nil)
	SetOutput(w)

	for _, curr := range tests {
		if err := SetLogLevel(curr.mutedLevel); err != nil {
			t.Fatal("Should not fail with a valid log level")
		}
		curr.fc(curr.msg)

		if w.String() != "" {
			t.Fatalf("Shouldn't produce any log with level: %d", curr.mutedLevel)
		}

		w.Reset()
	}
}

func TestLogLevelStr(t *testing.T) {
	tests := []struct {
		level int
		str   string
	}{
		{LogLevelDebug, "LogLevelDebug"},
		{LogLevelInfo, "LogLevelInfo"},
		{LogLevelWarning, "LogLevelWarning"},
		{LogLevelError, "LogLevelError"},
	}

	for _, curr := range tests {
		str, err := LevelStr(curr.level)
		if err != nil {
			t.Fatalf(fmt.Sprintf("%s", err))
		}

		if str != curr.str {
			t.Fatalf("Expected string %s, but got: %s", curr.str, str)
		}
	}
}

func TestInvalidLogLevelStr(t *testing.T) {
	_, err := LevelStr(-1)
	if err == nil {
		t.Fatal("Should have failed to format an invalid log level")
	}
}

func TestInvalidLogLevel(t *testing.T) {
	if err := SetLogLevel(999); err == nil {
		t.Fatal("Shold fail with an invalid log level")
	}
}

func TestLogTraceableError(t *testing.T) {
	w := bytes.NewBuffer(nil)
	SetOutput(w)
	ErrorError(errors.Errorf("Traceable error"))

	if !strings.Contains(w.String(), "log_test.go") {
		t.Fatal("Traceable should contain the source name")
	}
}

func TestLogOut(t *testing.T) {
	w := bytes.NewBuffer(nil)
	SetOutput(w)
	Out("command output")

	if !strings.Contains(w.String(), "[OUT]") {
		t.Fatal("Out logs should contain the tag [OUT]")
	}
}
