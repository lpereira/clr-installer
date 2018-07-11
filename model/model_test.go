// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/clearlinux/clr-installer/utils"
)

var (
	testsDir string
)

func init() {
	testsDir = os.Getenv("TESTS_DIR")
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		file  string
		valid bool
	}{
		{"basic-valid-descriptor.yaml", true},
		{"basic-invalid-descriptor.yaml", false},
		{"malformed-descriptor.yaml", false},
		{"no-bootable-descriptor.yaml", false},
		{"no-root-partition-descriptor.yaml", false},
		{"invalid-no-keyboard.yaml", false},
		{"invalid-no-language.yaml", false},
		{"valid-network.yaml", true},
		{"real-example.yaml", true},
	}

	for _, curr := range tests {
		path := filepath.Join(testsDir, curr.file)
		model, err := LoadFile(path)

		if curr.valid && err != nil {
			t.Fatalf("%s is a valid tests and shouldn't return an error: %v", curr.file, err)
		}

		err = model.Validate()
		if curr.valid && err != nil {
			t.Fatalf("%s is a valid tests and shouldn't return an error: %v", curr.file, err)
		}
	}
}

func TestUnreadable(t *testing.T) {
	file, err := ioutil.TempFile("", "test-")
	if err != nil {
		t.Fatal("Could not create a temp file")
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if file.Chmod(0111) != nil {
		t.Fatal("Failed to change tmp file mod")
	}

	if utils.IsRoot() {
		t.Log("Not running as 'root', not checking read permission")
	} else {
		_, err = LoadFile(file.Name())
		if err == nil {
			t.Fatal("Should have failed to read")
		}
	}
	if os.Remove(file.Name()) != nil {
		t.Fatal("Failed to cleanup test file")
	}
}

func TestBundle(t *testing.T) {
	si := &SystemInstall{}

	if si.ContainsBundle("test-bundle") {
		t.Fatal("Should return false since test-bundle wasn't added to si")
	}

	si.AddBundle("test-bundle")
	si.AddBundle("test-bundle-2")
	if !si.ContainsBundle("test-bundle") {
		t.Fatal("Should return true since test-bundle was added to si")
	}

	si.RemoveBundle("test-bundle")
	if si.ContainsBundle("test-bundle") {
		t.Fatal("Should return false since test-bundle was removed from si")
	}

	si.RemoveBundle("test-bundle-2")

	// duplicated
	si.AddBundle("test-bundle")
	si.AddBundle("test-bundle")
	if len(si.Bundles) > 1 {
		t.Fatal("We should have handled the duplication")
	}
}

func TestAddTargetMedia(t *testing.T) {
	path := filepath.Join(testsDir, "basic-valid-descriptor.yaml")
	loaded, err := LoadFile(path)

	if err != nil {
		t.Fatal("Failed to load a valid descriptor")
	}

	nm := &SystemInstall{}
	nm.AddTargetMedia(loaded.TargetMedias[0])
	if len(nm.TargetMedias) != 1 {
		t.Fatal("Failed to add target media to model")
	}
}

func TestAddNetworkInterface(t *testing.T) {
	path := filepath.Join(testsDir, "valid-network.yaml")
	loaded, err := LoadFile(path)

	if err != nil {
		t.Fatal("Failed to load a valid descriptor")
	}

	nm := &SystemInstall{}
	nm.AddNetworkInterface(loaded.NetworkInterfaces[0])
	if len(nm.NetworkInterfaces) != 1 {
		t.Fatal("Failed to add network interface to model")
	}
}

func TestWriteFile(t *testing.T) {
	path := filepath.Join(testsDir, "basic-valid-descriptor.yaml")
	loaded, err := LoadFile(path)

	if err != nil {
		t.Fatal("Failed to load a valid descriptor")
	}

	tmpFile, err := ioutil.TempFile("", "test-")
	if err != nil {
		t.Fatal("Could not create a temp file")
	}
	path = tmpFile.Name()
	if err = tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := loaded.WriteFile(path); err != nil {
		t.Fatal("Failed to write descriptor, should be valid")
	}
}
