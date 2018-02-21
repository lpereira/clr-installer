package log

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
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
	SetLogLevel(LogLevelDebug)

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

	if !strings.Contains(str, "[ERR] log_test.go") {
		t.Fatalf("Expected to have trace marks and the error tag - entry: %s", str)
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
		SetLogLevel(curr.mutedLevel)
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
