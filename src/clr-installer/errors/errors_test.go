package errors

import (
	"fmt"
	"testing"
)

func testErrorf(t *testing.T) {
	err := Errorf("traceable error")

	e, ok := err.(TraceableError)
	if !ok {
		t.Fatal("Errorf() should return a TraceableError")
	}

	if e.Trace == "" {
		t.Fatal("Traceable error should contain trace info")
	}

	if e.Error() != e.What {
		t.Fatal("Error() should return the content of What member")
	}
}

func TestErrorf(t *testing.T) {
	testErrorf(t)
}

func TestWrapp(t *testing.T) {
	err := Wrap(fmt.Errorf("wrapper error"))

	e, ok := err.(TraceableError)
	if !ok {
		t.Fatal("Wrap() should return a TraceableError")
	}

	if e.Trace == "" {
		t.Fatal("Traceable error should contain trace info")
	}

	if e.Error() != e.What {
		t.Fatal("Error() should return the content of What member")
	}
}

func TestFakeExternalTrace(t *testing.T) {
	fakeExternalTrace = true
	testErrorf(t)
	fakeExternalTrace = false
}
