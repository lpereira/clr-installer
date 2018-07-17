// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/conf"
	"github.com/clearlinux/clr-installer/errors"
	"github.com/clearlinux/clr-installer/hostname"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/model"
	"github.com/clearlinux/clr-installer/network"
	"github.com/clearlinux/clr-installer/progress"
	"github.com/clearlinux/clr-installer/storage"
	"github.com/clearlinux/clr-installer/swupd"
	cuser "github.com/clearlinux/clr-installer/user"
	"github.com/clearlinux/clr-installer/utils"
)

func sortMountPoint(bds []*storage.BlockDevice) []*storage.BlockDevice {
	sort.Slice(bds[:], func(i, j int) bool {
		return filepath.HasPrefix(bds[j].MountPoint, bds[i].MountPoint)
	})

	return bds
}

// Install is the main install controller, this is the entry point for a full
// installation
func Install(rootDir string, model *model.SystemInstall) error {
	var err error
	var version string
	var versionBuf []byte

	// First verify we are running as 'root' user which is required
	// for most of the Installation commands
	if err = utils.VerifyRootUser(); err != nil {
		return err
	}

	log.Info("Querying Clear Linux version")

	// in order to avoid issues raised by format bumps between installers image
	// version and the latest released we assume the installers host version
	// in other words we use the same version swupd is based on
	if versionBuf, err = ioutil.ReadFile("/usr/lib/os-release"); err != nil {
		return errors.Errorf("Read version file /usr/lib/os-release: %v", err)
	}
	versionExp := regexp.MustCompile(`VERSION_ID=([0-9][0-9]*)`)
	match := versionExp.FindSubmatch(versionBuf)

	if len(match) < 2 {
		return errors.Errorf("Version not found in /usr/lib/os-release")
	}

	version = string(match[1])
	log.Debug("Clear Linux version: %s", version)

	// do we have the minimum required to install a system?
	if err = model.Validate(); err != nil {
		return err
	}

	if err = ConfigureNetwork(model); err != nil {
		return err
	}

	mountPoints := []*storage.BlockDevice{}

	// prepare all the target block devices
	for _, curr := range model.TargetMedias {
		// based on the description given, write the partition table
		if err = curr.WritePartitionTable(); err != nil {
			return err
		}

		// prepare the blockdevice's partitions filesystem
		for _, ch := range curr.Children {
			prg := progress.NewLoop("Writing %s file system to %s", ch.FsType, ch.Name)
			if err = ch.MakeFs(); err != nil {
				return err
			}
			prg.Done()

			// if we have a mount point set it for future mounting
			if ch.MountPoint != "" {
				mountPoints = append(mountPoints, ch)
			}
		}
	}

	// mount all the prepared partitions
	for _, curr := range sortMountPoint(mountPoints) {
		log.Info("Mounting: %s", curr.MountPoint)

		if err = curr.Mount(rootDir); err != nil {
			return err
		}
	}

	err = storage.MountMetaFs(rootDir)
	if err != nil {
		return err
	}

	if model.Telemetry.Enabled {
		model.AddBundle("telemetrics")
	}

	if model.KernelCMDLine != "" {
		cmdlineDir := filepath.Join(rootDir, "etc", "kernel")
		cmdlineFile := filepath.Join(cmdlineDir, "cmdline")
		cmdline := model.KernelCMDLine

		if err = utils.MkdirAll(cmdlineDir); err != nil {
			return err
		}

		if err = ioutil.WriteFile(cmdlineFile, []byte(cmdline), 0644); err != nil {
			return err
		}
	}

	prg, err := contentInstall(rootDir, version, model.Bundles, model.Kernel.Bundle, model.SwupdMirror)
	if err != nil {
		prg.Done()
		return err
	}

	if err := cuser.Apply(rootDir, model.Users); err != nil {
		return err
	}

	if model.Hostname != "" {
		if err := hostname.SetTargetHostname(rootDir, model.Hostname); err != nil {
			return err
		}
	}

	return nil
}

// use the current host's version to bootstrap the sysroot, then update to the
// latest one and start adding new bundles
// for the bootstrap we huse the hosts's swupd and the following operations are
// executed using the target swupd
func contentInstall(rootDir string, version string, bundles []string, kernel string, mirror string) (progress.Progress, error) {
	sw := swupd.New(rootDir)

	prg := progress.NewLoop("Installing the base system")
	if err := sw.Verify(version, mirror); err != nil {
		return prg, err
	}

	if err := sw.Update(); err != nil {
		return prg, err
	}
	prg.Done()

	bundles = append(bundles, kernel)
	for _, bundle := range bundles {
		// swupd will fail (return exit code 18) if we try to "re-install" a bundle
		// already installed - with that we need to prevent doing bundle-add for bundles
		// previously installed by verify operation
		if swupd.IsCoreBundle(bundle) {
			log.Debug("Bundle %s was already installed with the core bundles, skipping")
			continue
		}

		prg = progress.NewLoop("Installing bundle: %s", bundle)
		if err := sw.BundleAdd(bundle); err != nil {
			return prg, err
		}

		prg.Done()
	}

	prg = progress.NewLoop("Installing boot loader")
	args := []string{
		fmt.Sprintf("%s/usr/bin/clr-boot-manager", rootDir),
		"update",
		fmt.Sprintf("--path=%s", rootDir),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return prg, errors.Wrap(err)
	}
	prg.Done()

	return nil, nil
}

// ConfigureNetwork applies the model/configured network interfaces
func ConfigureNetwork(model *model.SystemInstall) error {
	prg, err := configureNetwork(model)
	if err != nil {
		prg.Done()
		return err
	}
	return nil
}

func configureNetwork(model *model.SystemInstall) (progress.Progress, error) {
	if model.HTTPSProxy != "" {
		cmd.SetHTTPSProxy(model.HTTPSProxy)
	}

	if len(model.NetworkInterfaces) > 0 {
		prg := progress.NewLoop("Applying network settings")
		if err := network.Apply("/", model.NetworkInterfaces); err != nil {
			return prg, err
		}
		prg.Done()

		prg = progress.NewLoop("Restarting network interfaces")
		if err := network.Restart(); err != nil {
			return prg, err
		}
		prg.Done()
	}

	prg := progress.NewLoop("Testing connectivity")
	ok := false

	// 3 attempts to test connectivity
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)

		if err := network.VerifyConnectivity(); err == nil {
			ok = true
			break
		}
	}

	if !ok {
		return prg, errors.Errorf("Failed, network is not working.")
	}

	prg.Done()

	return nil, nil
}

// SaveInstallResults saves the results of the installation process
// onto the target media
func SaveInstallResults(rootDir string, model *model.SystemInstall) error {
	var err error
	var errMsg string

	if !model.PostArchive {
		log.Info("Skipping archiving of Installation results")
		return nil
	}

	log.Info("Saving Installation results to %s", rootDir)

	saveDir := filepath.Join(rootDir, "root")
	if err = utils.MkdirAll(saveDir); err != nil {
		// Fallback in the unlikely case we can't use root's home
		saveDir = rootDir
	}

	confFile := filepath.Join(saveDir, conf.ConfigFile)

	if err := model.WriteFile(confFile); err != nil {
		log.Error("Failed to write YAML file (%v) %q", err, confFile)
		errMsg = "Failed to write YAML file"
	}

	logFile := filepath.Join(saveDir, conf.LogFile)

	if err := log.ArchiveLogFile(logFile); err != nil {
		if errMsg != "" {
			errMsg = errMsg + "; "
		}
		errMsg = errMsg + "Failed to archive log file"
	}

	if errMsg != "" {
		return errors.Errorf("%s", errMsg)
	}
	return nil
}

// Cleanup executes post-install cleanups i.e unmount partition, remove
// temporary directory etc.
func Cleanup(rootDir string, umount bool) error {
	var err error

	log.Info("Cleaning up %s", rootDir)

	// we'll fail to umount only if a device is not mounted
	// then, just log it and move cleaning up
	if umount {
		if storage.UmountAll() != nil {
			log.Warning("Failed to umount volumes")
		}
	}

	log.Info("Removing rootDir: %s", rootDir)
	if err = os.RemoveAll(rootDir); err != nil {
		return errors.Errorf("Failed to remove all in %s: %v", rootDir, err)
	}

	return nil
}
