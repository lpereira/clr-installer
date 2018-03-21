package controller

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"

	"clr-installer/cmd"
	"clr-installer/errors"
	"clr-installer/log"
	"clr-installer/model"
	"clr-installer/progress"
	"clr-installer/storage"
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

	log.Info("Querying clear linux version")

	// in order to avoid issues raised by format bumps between installers image
	// version and the latest released we assume the installers host version
	// in other words we use the same version swupd is based on
	args := []string{
		"cat",
		"/var/lib/swupd/version",
	}

	version := bytes.NewBuffer(nil)
	err = cmd.Run(version, true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	// do we have the minimum required to install a system?
	err = model.Validate()
	if err != nil {
		return err
	}

	mountPoints := []*storage.BlockDevice{}

	// prepare all the target block devices
	for _, curr := range model.TargetMedias {
		// based on the description given, write the partition table
		err = curr.WritePartitionTable()
		if err != nil {
			return err
		}

		// prepare the blockdevice's partitions filesystem
		for _, ch := range curr.Children {
			prg := progress.NewLoop(fmt.Sprintf("Writing %s file system to %s", ch.FsType, ch.Name))
			err = ch.MakeFs()
			if err != nil {
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

		err = curr.Mount(rootDir)
		if err != nil {
			return err
		}
	}

	err = storage.MountMetaFs(rootDir)
	if err != nil {
		return err
	}

	err = contentInstall(rootDir, version.String(), model.Bundles)
	if err != nil {
		return err
	}

	return nil
}

func contentInstall(rootDir string, version string, bundles []string) error {
	prg := progress.NewLoop("Installing the base system")
	args := []string{
		"swupd",
		"verify",
		fmt.Sprintf("--path=%s", rootDir),
		"--install",
		"-m",
		version,
		"--force",
		"--no-scripts",
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}
	prg.Done()

	for _, bundle := range bundles {
		prg = progress.NewLoop(fmt.Sprintf("Installing bundle: %s", bundle))
		args = []string{
			"swupd",
			"bundle-add",
			fmt.Sprintf("--path=%s", rootDir),
			bundle,
		}

		err = cmd.RunAndLog(true, args...)
		if err != nil {
			return errors.Wrap(err)
		}
		prg.Done()
	}

	prg = progress.NewLoop("Installing boot loader")
	args = []string{
		fmt.Sprintf("%s/usr/bin/clr-boot-manager", rootDir),
		"update",
		fmt.Sprintf("--path=%s", rootDir),
	}

	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}
	prg.Done()

	return nil
}

// Cleanup executes post-install cleanups i.e unmount partition, remove
// temporary directory etc.
func Cleanup(rootDir string, umount bool) error {
	var err error

	log.Info("Umounting %s", rootDir)

	// we'll fail to umount only if a device is not mounted
	// then, just log it and move cleaning up
	if umount {
		if storage.UmountAll(rootDir) != nil {
			log.Warning("Failed to umount volumes")
		}
	}

	args := []string{
		"rm",
		"-R",
		"-f",
		rootDir,
	}

	log.Info("Removing rootDir: %s", rootDir)
	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
