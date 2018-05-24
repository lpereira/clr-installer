// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/clearlinux/clr-installer/errors"
)

// MkdirAll similar to go's standard os.MkdirAll() this function creates a directory
// named path, along with any necessary parents but also checks if path exists and
// takes no action if that's true.
func MkdirAll(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(path, 0777); err != nil {
		return errors.Errorf("mkdir %s: %v", path, err)
	}

	return nil
}

// CopyFile copies src file to dest
func CopyFile(src string, dest string) error {
	var err error
	destDir := filepath.Dir(dest)

	if _, err = os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("no such file: %s", src)
		}
		return errors.Wrap(err)
	}

	if _, err = os.Stat(destDir); err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("no such dest directory: %s", destDir)
		}
		return errors.Wrap(err)
	}

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(dest, data, 0644); err != nil {
		return err
	}

	return nil
}
