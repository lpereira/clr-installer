// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// BundleListFile the file file for the containing the bundle list definition
	BundleListFile = "bundles.json"

	// ConfigFile is the install descriptor
	ConfigFile = "clr-installer.yaml"

	// DefaultConfigDir is the system wide default configuration directory
	DefaultConfigDir = "/usr/share/defaults/clr-installer"

	// SourcePath is the source path (within the .gopath)
	SourcePath = "src/github.com/clearlinux/clr-installer"
)

func isRunningFromSourceTree() (bool, string, error) {
	src, err := os.Executable()
	if err != nil {
		return false, src, err
	}

	return !strings.HasPrefix(src, "/usr/bin"), src, nil
}

func lookupDefaultFile(file string) (string, error) {
	isSourceTree, sourcePath, err := isRunningFromSourceTree()
	if err != nil {
		return "", err
	}

	// use the config from source code's etc dir if not installed binary
	if isSourceTree {
		sourceRoot := strings.Replace(sourcePath, "bin", filepath.Join(SourcePath, "etc"), 1)
		return filepath.Join(sourceRoot, file), nil
	}

	return filepath.Join(DefaultConfigDir, file), nil
}

// LookupBundleListFile looks up the bundle list definition
// Guesses if we're running from source code our from system, if we're running from
// source code directory then we loads the source default file, otherwise tried to load
// the system installed file
func LookupBundleListFile() (string, error) {
	return lookupDefaultFile(BundleListFile)
}

// LookupDefaultConfig looks up the install descriptor
// Guesses if we're running from source code our from system, if we're running from
// source code directory then we loads the source default file, otherwise tried to load
// the system installed file
func LookupDefaultConfig() (string, error) {
	return lookupDefaultFile(ConfigFile)
}

// FetchRemoteConfigFile given an config url fetches it from the network. This function
// currently supports only http/https protocol. After success return the local file path.
func FetchRemoteConfigFile(url string) (string, error) {
	out, err := ioutil.TempFile("", "clr-installer-yaml-")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = out.Close()
	}()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}
