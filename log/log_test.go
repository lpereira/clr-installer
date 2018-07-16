// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package log

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/clearlinux/clr-installer/errors"
)

func setLog(t *testing.T) *os.File {
	var handle *os.File

	tmpfile, err := ioutil.TempFile("", "writeLog")
	if err != nil {
		t.Fatalf("could not make tempfile: %v", err)
	}
	_ = tmpfile.Close()

	if handle, err = SetOutputFilename(tmpfile.Name()); err != nil {
		t.Fatal("Could not set Log file")
	}

	return handle
}

func readLog(t *testing.T) *bytes.Buffer {
	tmpfile, err := ioutil.TempFile("", "readLog")
	if err != nil {
		t.Fatalf("could not make tempfile: %v", err)
	}
	_ = tmpfile.Close()
	defer func() { _ = os.Remove(tmpfile.Name()) }() // clean up

	_ = ArchiveLogFile(tmpfile.Name())

	var contents []byte
	contents, err = ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("could not read tempfile: %v %q", err, tmpfile.Name())
	} else {
		return bytes.NewBuffer(contents)
	}

	return nil
}

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

	fh := setLog(t)
	defer func() { _ = fh.Close() }()

	if err := SetLogLevel(LogLevelDebug); err != nil {
		t.Fatal("Should not fail with a valid level")
	}

	for _, curr := range tests {
		curr.fc(curr.msg)

		str := readLog(t).String()
		if str == "" {
			t.Fatal("No log written to output")
		}

		if !strings.Contains(str, curr.tag) {
			t.Fatalf("Log generated an entry without the expected tag: %s - entry: %s",
				curr.tag, str)
		}
	}
}

func TestErrorError(t *testing.T) {
	fh := setLog(t)
	defer func() { _ = fh.Close() }()

	ErrorError(fmt.Errorf("testing log with error"))

	str := readLog(t).String()
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

	fh := setLog(t)
	defer func() { _ = fh.Close() }()

	for _, curr := range tests {
		if err := SetLogLevel(curr.mutedLevel); err != nil {
			t.Fatal("Should not fail with a valid log level")
		}
		curr.fc(curr.msg)

		if readLog(t).String() != "" {
			t.Fatalf("Shouldn't produce any log with level: %d", curr.mutedLevel)
		}
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
		t.Fatal("Should fail with an invalid log level")
	}
}

func TestLogTraceableError(t *testing.T) {
	fh := setLog(t)
	defer func() { _ = fh.Close() }()

	ErrorError(errors.Errorf("Traceable error"))

	if !strings.Contains(readLog(t).String(), "log_test.go") {
		t.Fatal("Traceable should contain the source name")
	}
}

func TestLogOut(t *testing.T) {
	fh := setLog(t)
	defer func() { _ = fh.Close() }()
	Out("command output")

	if !strings.Contains(readLog(t).String(), "[OUT]") {
		t.Fatal("Out logs should contain the tag [OUT]")
	}
}
