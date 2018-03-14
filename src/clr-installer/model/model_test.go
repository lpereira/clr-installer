package model

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
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
		{"basic-valid-descriptor.json", true},
		{"basic-invalid-descriptor.json", false},
		{"malformed-descriptor.json", false},
		{"no-bootable-descriptor.json", false},
		{"no-root-partition-descriptor.json", false},
	}

	for _, curr := range tests {
		path := filepath.Join(testsDir, curr.file)
		model, err := LoadFile(path)

		if curr.valid && err != nil {
			t.Fatalf("%s is a valid tests and shouldn't return an error", curr.file)
		}

		err = model.Validate()
		if curr.valid && err != nil {
			t.Fatalf("%s is a valid tests and shouldn't return an error", curr.file)
		}
	}
}

func TestUnreadable(t *testing.T) {
	file, err := ioutil.TempFile("", "test-")
	if err != nil {
		t.Fatal("Could not create a temp file")
	}

	if file.Chmod(0111) != nil {
		t.Fatal("Failed to change tmp file mod")
	}

	_, err = LoadFile(file.Name())
	if err == nil {
		t.Fatal("Should have failed to read")
	}

	if os.Remove(file.Name()) != nil {
		t.Fatal("Failed to cleanup test file")
	}
}