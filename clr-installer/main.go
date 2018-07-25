// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/clearlinux/clr-installer/args"
	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/conf"
	"github.com/clearlinux/clr-installer/crypt"
	"github.com/clearlinux/clr-installer/frontend"
	"github.com/clearlinux/clr-installer/keyboard"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/massinstall"
	"github.com/clearlinux/clr-installer/model"
	"github.com/clearlinux/clr-installer/swupd"
	"github.com/clearlinux/clr-installer/tui"
)

var (
	frontEndImpls []frontend.Frontend
)

func fatal(err error) {
	log.ErrorError(err)
	panic(err)
}

func initFrontendList() {
	frontEndImpls = []frontend.Frontend{
		massinstall.New(),
		tui.New(),
	}
}

func main() {
	var options args.Args

	if err := options.ParseArgs(); err != nil {
		fatal(err)
	}

	if options.DemoMode {
		model.Version = "X.Y.Z"
	}
	if err := log.SetLogLevel(options.LogLevel); err != nil {
		fatal(err)
	}

	if options.PamSalt != "" {
		hashed, err := crypt.Crypt(options.PamSalt)
		if err != nil {
			panic(err)
		}

		fmt.Println(hashed)
		return
	}

	if options.Version {
		fmt.Println(path.Base(os.Args[0]) + ": " + model.Version)
		return
	}

	f, err := log.SetOutputFilename(options.LogFile)
	if err != nil {
		fatal(err)
	}
	defer func() {
		_ = f.Close()
	}()

	initFrontendList()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	rootDir, err := ioutil.TempDir("", "install-")
	if err != nil {
		fatal(err)
	}

	var md *model.SystemInstall
	cf := options.ConfigFile

	if options.ConfigFile == "" {
		if cf, err = conf.LookupDefaultConfig(); err != nil {
			fatal(err)
		}
	}

	log.Debug("Loading config file: %s", cf)
	if md, err = model.LoadFile(cf); err != nil {
		fatal(err)
	}

	if options.RebootSet {
		md.PostReboot = options.Reboot
	}

	if options.ArchiveSet {
		md.PostArchive = options.Archive
	}

	// Command line overrides the configuration file
	if options.SwupdMirror != "" {
		md.SwupdMirror = options.SwupdMirror
	}
	// Now validate the mirror from the config or command line
	if md.SwupdMirror != "" {
		var url string
		url, err = swupd.SetHostMirror(md.SwupdMirror)
		if err != nil {
			fatal(err)
		} else {
			log.Debug("Using Swupd Mirror value: %q", url)
		}
	}

	if md.Keyboard != nil {
		if err = keyboard.Apply(md.Keyboard); err != nil {
			fatal(err)
		}
	}

	installReboot := false

	go func() {
		for _, fe := range frontEndImpls {
			if !fe.MustRun(&options) {
				continue
			}

			installReboot, err = fe.Run(md, rootDir)
			if err != nil {
				fatal(err)
			}

			break
		}

		done <- true
	}()

	go func() {
		<-sigs
		fmt.Println("Leaving...")
		done <- true
	}()

	<-done

	// Stop the signal handlers
	// or we get a SIGTERM from reboot
	signal.Reset()

	if options.Reboot && installReboot {
		if err := cmd.RunAndLog("reboot"); err != nil {
			fatal(err)
		}
	}
}
