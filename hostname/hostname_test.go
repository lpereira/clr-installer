// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package hostname

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestEmptyHostname(t *testing.T) {

	var host string
	var err string

	host = ""
	if err = IsValidHostname(host); err == "" {
		t.Fatalf("Empty hostname %q should fail", host)
	}
}

func TestInvalidHostnames(t *testing.T) {

	var host string
	var err string

	host = "-nogood"
	if err = IsValidHostname(host); err == "" {
		t.Fatalf("Hostname %q should fail", host)
	}

	host = "no@good"
	if err = IsValidHostname(host); err == "" {
		t.Fatalf("Hostname %q should fail", host)
	}
}

func TestTooLongHostname(t *testing.T) {

	var host string
	var err string

	host = "1234567890123456789012345678901234567890123456789012345678901234567890"
	if err = IsValidHostname(host); err == "" {
		t.Fatalf("Hostname %q should fail", host)
	}
}

func TestGoodHostnames(t *testing.T) {

	var host string
	var err string

	host = "clear-linux-host"
	if err = IsValidHostname(host); err != "" {
		t.Fatalf("Hostname %q should pass: %q", host, err)
	}

	host = "c"
	if err = IsValidHostname(host); err != "" {
		t.Fatalf("Hostname %q should pass: %q", host, err)
	}

	host = "clear01"
	if err = IsValidHostname(host); err != "" {
		t.Fatalf("Hostname %q should pass: %q", host, err)
	}

	host = "1"
	if err = IsValidHostname(host); err != "" {
		t.Fatalf("Hostname %q should pass: %q", host, err)
	}
}

func TestSaveHostname(t *testing.T) {

	rootDir, err := ioutil.TempDir("", "testhost-")
	if err != nil {
		t.Fatalf("Could not make temp dir for testing hostname: %q", err)
	}

	defer func() { _ = os.RemoveAll(rootDir) }()

	host := "hello"
	if err = SetTargetHostname(rootDir, host); err != nil {
		t.Fatalf("Could not SetTargetHostname to %q: %q", host, err)
	}
}
