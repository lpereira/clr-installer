// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package args

// Arguments which influence how this program executes
// Order of Precedence
// 1. Command Line Arguments -- Highest Priority
// 2. Kernel Command Line Arguments
// 3. Program defaults -- Lowest Priority

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/clearlinux/clr-installer/conf"
	"github.com/clearlinux/clr-installer/log"
	flag "github.com/spf13/pflag"
)

const (
	kernelCmdlineConf = "clri.descriptor"
	kernelCmdlineDemo = "clri.demo"
)

var (
	kernelCmdlineFile = "/proc/cmdline"
)

// Args represents the user provided arguments
type Args struct {
	Version    bool
	Reboot     bool
	LogFile    string
	ConfigFile string
	PamSalt    string
	LogLevel   int
	ForceTUI   bool
	DemoMode   bool
}

func (args *Args) setKernelArgs() (err error) {
	var (
		kernelCmd string
		url       string
	)

	if kernelCmd, err = args.readKernelCmd(); err != nil {
		return err
	}

	// Parse the kernel command for relevant installer options
	for _, curr := range strings.Split(kernelCmd, " ") {
		if strings.HasPrefix(curr, kernelCmdlineConf+"=") {
			url = strings.Split(curr, "=")[1]
		} else if strings.HasPrefix(curr, kernelCmdlineDemo) {
			args.DemoMode = true
		}
	}

	if url != "" {
		var ffile string

		if ffile, err = conf.FetchRemoteConfigFile(url); err != nil {
			return err
		}

		args.ConfigFile = ffile
	}

	return nil
}

// readKernelCmd returns the kernel command line
func (args *Args) readKernelCmd() (string, error) {
	if _, err := os.Stat(kernelCmdlineFile); os.IsNotExist(err) {
		return "", err
	}

	content, err := ioutil.ReadFile(kernelCmdlineFile)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (args *Args) setCommandLineArgs() (err error) {
	flag.BoolVar(
		&args.Version, "version", false, "Version of the Installer",
	)

	flag.BoolVar(
		&args.Reboot, "reboot", true, "Reboot after finishing",
	)

	flag.BoolVar(
		&args.ForceTUI, "tui", false, "Use TUI frontend",
	)

	flag.StringVar(
		&args.ConfigFile, "config", args.ConfigFile, "Installation configuration file",
	)

	flag.StringVar(
		&args.PamSalt, "genpass", "", "Generates a PAM compatible password hash based on the provided salt string",
	)

	flag.IntVar(
		&args.LogLevel,
		"log-level",
		log.LogLevelDebug,
		fmt.Sprintf("%d (debug), %d (info), %d (warning), %d (error)",
			log.LogLevelDebug, log.LogLevelInfo, log.LogLevelWarning, log.LogLevelError),
	)

	flag.BoolVar(
		&args.DemoMode, "demo", args.DemoMode, "Demonstration mode for documentation generation",
	)
	//flag.MarkHidden("demo")

	usr, err := user.Current()
	if err != nil {
		return err
	}

	defaultLogFile := filepath.Join(usr.HomeDir, "clr-installer.log")
	flag.StringVar(
		&args.LogFile, "log-file", defaultLogFile, "The log file path",
	)

	flag.Parse()

	return nil
}

// ParseArgs will both parse the command line arguments to the program
// and read any options set on the kernel command line from boot-time
// setting the results into the Args member variables.
func (args *Args) ParseArgs() (err error) {
	err = args.setKernelArgs()
	if err != nil {
		return err
	}

	err = args.setCommandLineArgs()
	if err != nil {
		return err
	}

	return nil
}
