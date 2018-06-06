// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

const (
	kernelCmdlineFile = "/proc/cmdline"
	kernelCmdlineArg  = "clri.descriptor"
)

// ReadKernelCmdline read the kernel cmdline and seeks for the clr installer
// descriptor url argument, if set returns the value
func ReadKernelCmdline() (string, error) {
	if _, err := os.Stat(kernelCmdlineFile); os.IsNotExist(err) {
		return "", nil
	}

	content, err := ioutil.ReadFile(kernelCmdlineFile)
	if err != nil {
		return "", err
	}

	for _, curr := range strings.Split(string(content), " ") {
		if strings.HasPrefix(curr, kernelCmdlineArg) {
			return strings.Split(curr, "=")[1], nil
		}
	}

	return "", nil
}
