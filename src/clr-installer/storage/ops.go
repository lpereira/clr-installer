package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clr-installer/cmd"
	"clr-installer/errors"
	"clr-installer/log"
	"clr-installer/progress"
)

type blockDeviceOps struct {
	makeFs          func(bd *BlockDevice) error
	makePartCommand func(bd *BlockDevice, start uint64, end uint64) (string, error)
}

var (
	bdOps = map[string]*blockDeviceOps{
		"ext4": {ext4MakeFs, ext4MakePartCommand},
		"swap": {swapMakeFs, swapMakePartCommand},
		"vfat": {vfatMakeFs, vfatMakePartCommand},
	}

	guidMap = map[string]string{
		"/":     "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
		"/home": "933AC7E1-2EB4-4F13-B844-0E14E2AEF915",
		"/srv":  "3B8F8425-20E0-4F3B-907F-1A25A76F98E8",
		"swap":  "0657FD6D-A4AB-43C4-84E5-0933C84B4F4F",
		"efi":   "C12A7328-F81F-11D2-BA4B-00A0C93EC93B",
	}
)

// MakeFs runs mkfs.* commands for a BlockDevice definition
func (bd *BlockDevice) MakeFs() error {
	if bd.Type == BlockDeviceTypeDisk {
		return errors.Errorf("Trying to run MakeFs() against a disk, partition required")
	}

	if op, ok := bdOps[bd.FsType]; ok {
		args := []string{
			"umount",
			"-f",
			"-A",
			fmt.Sprintf("/dev/%s", bd.Name),
		}

		if err := cmd.Run(nil, true, args...); err != nil {
			log.ErrorError(err)
		}

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
	var args []string

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		args = []string{
			"mkdir",
			"-p",
			"-v",
			targetPath,
		}

		err = cmd.RunAndLog(true, args...)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	args = []string{
		"mount",
		fmt.Sprintf("/dev/%s", bd.Name),
		targetPath,
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// UmountAll unmounts all mounted block devices on rootDir
func UmountAll(rootDir string) error {
	args := []string{
		"umount",
		"-f",
		"-R",
		rootDir,
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// WritePartitionTable writes the defined partitions to the actual block device
func (bd *BlockDevice) WritePartitionTable() error {
	if bd.Type != BlockDeviceTypeDisk {
		return errors.Errorf("Type is partition, disk required")
	}

	prg := progress.NewLoop(fmt.Sprintf("Writing partition table to: %s", bd.Name))
	args := []string{
		"parted",
		"-s",
		fmt.Sprintf("/dev/%s", bd.Name),
		"mklabel",
		"gpt",
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"parted",
		"-a",
		"optimal",
		fmt.Sprintf("/dev/%s", bd.Name),
		"--script",
	}

	var start uint64
	bootPartition := -1
	guids := map[int]string{}

	for idx, curr := range bd.Children {
		op, found := bdOps[curr.FsType]
		if !found {
			return errors.Errorf("No makePartCommand() implementation for: %s",
				curr.FsType)
		}

		end := start + (uint64(curr.Size) >> 20)
		cmd, err := op.makePartCommand(curr, start, end)
		if err != nil {
			return err
		}

		if curr.MountPoint == "/boot" {
			bootPartition = idx + 1
		}

		guid, err := curr.getGUID()
		if err != nil {
			return err
		}

		guids[idx+1] = guid
		args = append(args, cmd)
		start = end
	}

	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"parted",
		fmt.Sprintf("/dev/%s", bd.Name),
		fmt.Sprintf("set %d boot on", bootPartition),
	}

	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}
	prg.Done()

	prg = progress.MultiStep(len(guids), "Adjusting filesystem configurations")
	cnt := 1
	for idx, guid := range guids {
		args = []string{
			"sgdisk",
			fmt.Sprintf("/dev/%s", bd.Name),
			fmt.Sprintf("--typecode=%d:%s", idx, guid),
		}

		err = cmd.RunAndLog(true, args...)
		if err != nil {
			return errors.Wrap(err)
		}

		prg.Partial(cnt)
		cnt = cnt + 1
	}
	prg.Done()

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

func mountDevFs(rootDir string) error {
	mPointPath := filepath.Join(rootDir, "dev")
	args := []string{
		"mkdir",
		"-v",
		"-p",
		mPointPath,
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"mount",
		"--bind",
		"/dev",
		mPointPath,
	}

	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func mountSysFs(rootDir string) error {
	mPointPath := filepath.Join(rootDir, "sys")
	args := []string{
		"mkdir",
		"-v",
		"-p",
		mPointPath,
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"mount",
		"--bind",
		"/sys",
		mPointPath,
	}

	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func mountProcFs(rootDir string) error {
	mPointPath := filepath.Join(rootDir, "proc")
	args := []string{
		"mkdir",
		"-v",
		"-p",
		mPointPath,
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"mount",
		"-t",
		"proc",
		"proc",
		mPointPath,
	}

	err = cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func ext4MakePartCommand(bd *BlockDevice, start uint64, end uint64) (string, error) {
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
		fmt.Sprintf("/dev/%s", bd.Name),
	}

	err := cmd.RunAndLog(true, args...)
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
		fmt.Sprintf("/dev/%s", bd.Name),
	}

	err := cmd.RunAndLog(true, args...)
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
		fmt.Sprintf("/dev/%s", bd.Name),
	}

	err := cmd.RunAndLog(true, args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
