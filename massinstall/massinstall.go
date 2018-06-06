// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package massinstall

import (
	"fmt"
	"strings"
	"time"

	"github.com/clearlinux/clr-installer/controller"
	"github.com/clearlinux/clr-installer/frontend"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/model"
	"github.com/clearlinux/clr-installer/progress"
)

// MassInstall is the frontend implementation for the "mass installer" it also
// implements the progress interface: progress.Client
type MassInstall struct {
	prgDesc  string
	prgIndex int
}

// New creates a new instance of MassInstall frontend implementation
func New() *MassInstall {
	return &MassInstall{}
}

// Step is the progress step implementation for progress.Client interface
func (mi *MassInstall) Step() {
	elms := []string{"|", "-", "\\", "|", "/", "-", "\\"}

	fmt.Printf("%s [%s]\r", mi.prgDesc, elms[mi.prgIndex])

	if mi.prgIndex+1 == len(elms) {
		mi.prgIndex = 0
	} else {
		mi.prgIndex = mi.prgIndex + 1
	}
}

// LoopWaitDuration is part of the progress.Client implementation and returns the
// duration each loop progress step should wait
func (mi *MassInstall) LoopWaitDuration() time.Duration {
	return 50 * time.Millisecond
}

// Desc is part of the implementation for ProgresIface and is used to adjust the progress bar
// label content
func (mi *MassInstall) Desc(desc string) {
	mi.prgDesc = desc
}

// Partial is part of the progress.Client implementation and sets the progress bar based
// on actuall progression
func (mi *MassInstall) Partial(total int, step int) {
	line := fmt.Sprintf("%s %.0f%%\r", mi.prgDesc, (float64(step)/float64(total))*100)
	fmt.Printf("%s", line)
}

// Done is part of the progress.Client implementation and represents the progress task "done"
// notification
func (mi *MassInstall) Done() {
	mi.prgIndex = 0
	fmt.Printf("%s [done]\n", mi.prgDesc)
}

// MustRun is part of the Frontend implementation and tells the core implementation that this
// frontend wants or should be executed
func (mi *MassInstall) MustRun(args *frontend.Args) bool {
	return args.ConfigFile != "" && !args.ForceTUI
}

func shouldReboot() (bool, bool, error) {
	var answer string
	va := map[string]bool{
		"y":   true,
		"yes": true,
		"n":   false,
		"no":  false,
	}

	fmt.Printf("reboot?[Y|n]: ")
	_, err := fmt.Scanf("%s", &answer)
	if err != nil {
		return false, false, err
	}

	reboot := false
	valid := false
	answer = strings.ToLower(answer)

	for k, v := range va {
		if k == answer {
			valid = true
			reboot = v
			break
		}
	}

	return valid, reboot, nil
}

// Run is part of the Frontend implementation and is the actual entry point for the
// "mass installer" frontend
func (mi *MassInstall) Run(md *model.SystemInstall, rootDir string) (bool, error) {
	var instError error

	progress.Set(mi)

	log.Debug("Starting install")

	instError = controller.Install(rootDir, md)

	prg := progress.NewLoop("Cleaning up install environment")
	if err := controller.Cleanup(rootDir, true); err != nil {
		log.ErrorError(err)
	}
	prg.Done()

	if instError != nil {
		return false, instError
	}

	var reboot bool

	for {
		var valid bool
		var err error

		if valid, reboot, err = shouldReboot(); err != nil {
			panic(err)
		}

		if !valid {
			fmt.Printf("Invalid answer...\n")
			continue
		}

		break
	}

	return reboot, nil
}
