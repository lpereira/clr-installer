package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/errors"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/model"
	"github.com/clearlinux/clr-installer/network"
	"github.com/clearlinux/clr-installer/progress"
	"github.com/clearlinux/clr-installer/storage"
	"github.com/clearlinux/clr-installer/swupd"
)

func sortMountPoint(bds []*storage.BlockDevice) []*storage.BlockDevice {
	sort.Slice(bds[:], func(i, j int) bool {
		return filepath.HasPrefix(bds[j].MountPoint, bds[i].MountPoint)
	})

	return bds
}

func verifyRootUser() error {
	// ProgName is the short name of this executable
	progName := path.Base(os.Args[0])

	user, err := user.Current()
	if err != nil {
		return errors.Errorf("%s MUST run as 'root' user to install! (user=%s)",
			progName, "UNKNOWN")
	}

	if user.Uid != "0" {
		return errors.Errorf("%s MUST run as 'root' user to install! (user=%s)",
			progName, user.Uid)
	}

	return nil
}

// Install is the main install controller, this is the entry point for a full
// installation
func Install(rootDir string, model *model.SystemInstall) error {
	var err error
	var version string
	var versionBuf []byte

	// First verify we are running as 'root' user which is required
	// for most of the Installation commands
	if err = verifyRootUser(); err != nil {
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

	prg, err := contentInstall(rootDir, version, model.Bundles)
	if err != nil {
		prg.Done()
		return err
	}

	return nil
}

// use the current host's version to bootstrap the sysroot, then update to the
// latest one and start adding new bundles
// for the bootstrap we huse the hosts's swupd and the following operations are
// executed using the target swupd
func contentInstall(rootDir string, version string, bundles []string) (progress.Progress, error) {
	sw := swupd.New(rootDir)

	prg := progress.NewLoop("Installing the base system")
	if err := sw.Verify(version); err != nil {
		return prg, err
	}

	if err := sw.Update(); err != nil {
		return prg, err
	}
	prg.Done()

	for _, bundle := range bundles {
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

	// 3 attempts to test conectivity
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)

		if err := network.Test(); err == nil {
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

// Cleanup executes post-install cleanups i.e unmount partition, remove
// temporary directory etc.
func Cleanup(rootDir string, umount bool) error {
	var err error

	// Verify we are running as 'root' user which is required
	// for most of the Cleanup commands
	// We probably should not call clean-up if we didn't call Install
	// If we're not running as 'root' then install could not be completed so, log it and
	// consider we've nothing to cleanup
	if err = verifyRootUser(); err != nil {
		log.Warning("Can't cleanup: %s", err)
		return nil
	}

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