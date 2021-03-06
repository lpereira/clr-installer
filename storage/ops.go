// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/errors"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/progress"
)

type blockDeviceOps struct {
	makeFs          func(bd *BlockDevice) error
	makePartCommand func(bd *BlockDevice, start uint64, end uint64) (string, error)
}

var (
	bdOps = map[string]*blockDeviceOps{
		"ext2":  {ext2MakeFs, commonMakePartCommand},
		"ext3":  {ext3MakeFs, commonMakePartCommand},
		"ext4":  {ext4MakeFs, commonMakePartCommand},
		"btrfs": {btrfsMakeFs, commonMakePartCommand},
		"xfs":   {xfsMakeFs, commonMakePartCommand},
		"swap":  {swapMakeFs, swapMakePartCommand},
		"vfat":  {vfatMakeFs, vfatMakePartCommand},
	}

	guidMap = map[string]string{
		"/":     "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
		"/home": "933AC7E1-2EB4-4F13-B844-0E14E2AEF915",
		"/srv":  "3B8F8425-20E0-4F3B-907F-1A25A76F98E8",
		"swap":  "0657FD6D-A4AB-43C4-84E5-0933C84B4F4F",
		"efi":   "C12A7328-F81F-11D2-BA4B-00A0C93EC93B",
	}

	mountedPoints []string
)

// MakeFs runs mkfs.* commands for a BlockDevice definition
func (bd *BlockDevice) MakeFs() error {
	if bd.Type == BlockDeviceTypeDisk {
		return errors.Errorf("Trying to run MakeFs() against a disk, partition required")
	}

	if op, ok := bdOps[bd.FsType]; ok {
		return op.makeFs(bd)
	}

	return errors.Errorf("MakeFs() not implemented for filesystem: %s", bd.FsType)
}

// getGUID determines the partition type guid either based on:
//   + mount point
//   + file system type (i.e swap)
//   + or if it's the "special" efi case
func (bd *BlockDevice) getGUID() (string, error) {
	if guid, ok := guidMap[bd.MountPoint]; ok {
		return guid, nil
	}

	if guid, ok := guidMap[bd.FsType]; ok {
		return guid, nil
	}

	if bd.FsType == "vfat" && bd.MountPoint == "/boot" {
		return guidMap["efi"], nil
	}

	return "", errors.Errorf("Could not determine the guid for: %s", bd.Name)
}

// Mount will mount a block devices bd considering its mount point and the
// root directory
func (bd *BlockDevice) Mount(root string) error {
	if bd.Type == BlockDeviceTypeDisk {
		return errors.Errorf("Trying to run MakeFs() against a disk, partition required")
	}

	targetPath := filepath.Join(root, bd.MountPoint)

	return mountFs(bd.GetDeviceFile(), targetPath, bd.FsType, syscall.MS_RELATIME)
}

// UmountAll unmounts all previously mounted devices
func UmountAll() error {
	var mountError error
	fails := make([]string, 0)

	// Ensure the top level mount point is unmounted last
	sort.Sort(sort.Reverse(sort.StringSlice(mountedPoints)))

	for _, point := range mountedPoints {
		if err := syscall.Unmount(point, syscall.MNT_FORCE); err != nil {
			err = fmt.Errorf("umount %s: %v", point, err)
			log.ErrorError(err)
			fails = append(fails, point)
		} else {
			log.Debug("Unmounted ok: %s", point)
		}
	}

	if len(fails) > 0 {
		mountError = errors.Errorf("Failed to unmount: %v", fails)
	}

	return mountError
}

// WritePartitionTable writes the defined partitions to the actual block device
func (bd *BlockDevice) WritePartitionTable() error {
	if bd.Type != BlockDeviceTypeDisk && bd.Type != BlockDeviceTypeLoop {
		return errors.Errorf("Type is partition, disk required")
	}

	prg := progress.NewLoop(fmt.Sprintf("Writing partition table to: %s", bd.Name))
	args := []string{
		"parted",
		"-s",
		bd.GetDeviceFile(),
		"mklabel",
		"gpt",
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"parted",
		"-a",
		"optimal",
		bd.GetDeviceFile(),
		"--script",
	}

	var start uint64
	bootPartition := -1
	guids := map[int]string{}

	for idx, curr := range bd.Children {
		var cmd string
		var guid string

		op, found := bdOps[curr.FsType]
		if !found {
			return errors.Errorf("No makePartCommand() implementation for: %s",
				curr.FsType)
		}

		end := start + (uint64(curr.Size) >> 20)
		cmd, err = op.makePartCommand(curr, start, end)
		if err != nil {
			return err
		}

		if curr.MountPoint == "/boot" {
			bootPartition = idx + 1
		}

		guid, err = curr.getGUID()
		if err != nil {
			return err
		}

		guids[idx+1] = guid
		args = append(args, cmd)
		start = end
	}

	err = cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"parted",
		bd.GetDeviceFile(),
		fmt.Sprintf("set %d boot on", bootPartition),
	}

	err = cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}
	prg.Success()

	prg = progress.MultiStep(len(guids), "Adjusting filesystem configurations")
	cnt := 1
	for idx, guid := range guids {
		args = []string{
			"sgdisk",
			bd.GetDeviceFile(),
			fmt.Sprintf("--typecode=%d:%s", idx, guid),
		}

		err = cmd.RunAndLog(args...)
		if err != nil {
			return errors.Wrap(err)
		}

		prg.Partial(cnt)
		cnt = cnt + 1
	}

	if err = bd.partProbe(); err != nil {
		prg.Failure()
		return err
	}

	time.Sleep(time.Duration(4) * time.Second)

	prg.Success()

	return nil
}

// MountMetaFs mounts proc, sysfs and devfs in the target installation directory
func MountMetaFs(rootDir string) error {
	err := mountProcFs(rootDir)
	if err != nil {
		return err
	}

	err = mountSysFs(rootDir)
	if err != nil {
		return err
	}

	err = mountDevFs(rootDir)
	if err != nil {
		return err
	}

	return nil
}

func mountFs(device string, mPointPath string, fsType string, flags uintptr) error {
	var err error

	if _, err = os.Stat(mPointPath); os.IsNotExist(err) {
		if err = os.MkdirAll(mPointPath, 0777); err != nil {
			return errors.Errorf("mkdir %s: %v", mPointPath, err)
		}
	}

	if err = syscall.Mount(device, mPointPath, fsType, flags, ""); err != nil {
		return errors.Errorf("mount %s: %v", mPointPath, err)
	}
	log.Debug("Mounted ok: %s", mPointPath)
	// Store the mount point for later unmounting
	mountedPoints = append(mountedPoints, mPointPath)

	return err
}

func mountDevFs(rootDir string) error {
	mPointPath := filepath.Join(rootDir, "dev")

	return mountFs("/dev", mPointPath, "devtmpfs", syscall.MS_BIND)
}

func mountSysFs(rootDir string) error {
	mPointPath := filepath.Join(rootDir, "sys")

	return mountFs("/sys", mPointPath, "sysfs", syscall.MS_BIND)
}

func mountProcFs(rootDir string) error {
	mPointPath := filepath.Join(rootDir, "proc")

	return mountFs("/proc", mPointPath, "proc", syscall.MS_BIND)
}

func commonMakePartCommand(bd *BlockDevice, start uint64, end uint64) (string, error) {
	args := []string{
		"mkpart",
		bd.MountPoint,
		fmt.Sprintf("%dM", start),
		fmt.Sprintf("%dM", end),
	}

	return strings.Join(args, " "), nil
}

func ext4MakeFs(bd *BlockDevice) error {
	args := []string{
		"mkfs.ext4",
		"-v",
		"-F",
		"-b",
		"4096",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func ext2MakeFs(bd *BlockDevice) error {
	args := []string{
		"mkfs.ext2",
		"-v",
		"-F",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func ext3MakeFs(bd *BlockDevice) error {
	args := []string{
		"mkfs.ext3",
		"-v",
		"-F",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func btrfsMakeFs(bd *BlockDevice) error {
	args := []string{
		"mkfs.btrfs",
		"-f",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func xfsMakeFs(bd *BlockDevice) error {
	args := []string{
		"mkfs.xfs",
		"-f",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func swapMakePartCommand(bd *BlockDevice, start uint64, end uint64) (string, error) {
	args := []string{
		"mkpart",
		"linux-swap",
		fmt.Sprintf("%dM", start),
		fmt.Sprintf("%dM", end),
	}

	return strings.Join(args, " "), nil
}

func swapMakeFs(bd *BlockDevice) error {
	args := []string{
		"mkswap",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func vfatMakePartCommand(bd *BlockDevice, start uint64, end uint64) (string, error) {
	args := []string{
		"mkpart",
		"EFI",
		"fat32",
		fmt.Sprintf("%dM", start),
		fmt.Sprintf("%dM", end),
	}

	return strings.Join(args, " "), nil
}

func vfatMakeFs(bd *BlockDevice) error {
	args := []string{
		"mkfs.vfat",
		"-F32",
		bd.GetDeviceFile(),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
