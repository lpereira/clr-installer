package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"clr-installer/cmd"
	"clr-installer/errors"
)

// A BlockDevice describes a block device and its partitions
type BlockDevice struct {
	Name            string         // device name
	Model           string         // device model
	MajorMinor      string         // major:minor device number
	FsType          string         // filesystem type
	UUID            string         // filesystem uuid
	MountPoint      string         // where the device is mounted
	Size            float64        // size of the device
	Type            int            // device type
	State           int            // device state (running, live etc)
	ReadOnly        bool           // read-only device
	RemovableDevice bool           // removable device
	Children        []*BlockDevice // children devices/partitions
}

const (
	// BlockDeviceTypeDisk identifies a BlockDevice as a disk
	BlockDeviceTypeDisk = 1

	// BlockDeviceTypePart identifies a BlockDevice as a partition
	BlockDeviceTypePart = 2

	// BlockDeviceTypeRom identifies a BlockDevice as a rom
	BlockDeviceTypeRom = 3

	// BlockDeviceStateUnknown identifies a BlockDevice in a unknown state
	BlockDeviceStateUnknown = 0

	// BlockDeviceStateRunning identifies a BlockDevice as running
	BlockDeviceStateRunning = 1

	// BlockDeviceStateLive identifies a BlockDevice as live
	BlockDeviceStateLive = 2
)

var (
	lsblkBinary = "lsblk"
	storageExp  = regexp.MustCompile(`([0-9](\.)?[0-9]*)?([m,g,t,k,p])`)
)

// Validate checks if the minimal requirements for a installation is met
func (bd *BlockDevice) Validate() error {
	bootPartition := false
	rootPartition := false

	for _, ch := range bd.Children {
		if ch.FsType == "vfat" && ch.MountPoint == "/boot" {
			bootPartition = true
		}

		if ch.MountPoint == "/" {
			rootPartition = true
		}
	}

	if !bootPartition {
		return errors.Errorf("Could not find a suitable EFI partition")
	}

	if !rootPartition {
		return errors.Errorf("Could not find a root partition")
	}

	return nil
}

// RemoveChild removes a partition from disk block device
func (bd *BlockDevice) RemoveChild(child *BlockDevice) {
	nList := []*BlockDevice{}

	for _, curr := range bd.Children {
		if curr == child {
			continue
		}

		nList = append(nList, curr)
	}

	bd.Children = nList
}

// AddChild adds a partition to a disk block device
func (bd *BlockDevice) AddChild(child *BlockDevice) {
	if bd.Children == nil {
		bd.Children = []*BlockDevice{}
	}

	bd.Children = append(bd.Children, child)

	if child.Name == "" {
		child.Name = fmt.Sprintf("%s%d", bd.Name, len(bd.Children))
	}
}

// HumanReadableSize converts the size representation in bytes to the closest
// human readable format i.e 10M, 1G, 2T etc
func HumanReadableSize(size float64) (string, error) {
	sizes := []struct {
		unit string
		mask float64
	}{
		{"P", 1 << 50},
		{"T", 1 << 40},
		{"G", 1 << 30},
		{"M", 1 << 20},
		{"K", 1 << 10},
	}

	for _, curr := range sizes {
		csize := size / curr.mask
		if csize < 1 {
			continue
		}

		return fmt.Sprintf("%.2f%s", csize, curr.unit), nil
	}

	return "", fmt.Errorf("Could not format desk/partition size")
}

// FreeSpace returns the block device available/free space considering the currently
// configured partition table
func (bd *BlockDevice) FreeSpace() (float64, error) {
	if bd.Type != BlockDeviceTypeDisk {
		return 0, errors.Errorf("FreeSpace() must only be called with a disk block device")
	}

	var total float64
	for _, curr := range bd.Children {
		total = total + curr.Size
	}

	return bd.Size - total, nil
}

// HumanReadableSize converts the size representation in bytes to the closest
// human readable format i.e 10M, 1G, 2T etc
func (bd *BlockDevice) HumanReadableSize() (string, error) {
	return HumanReadableSize(bd.Size)
}

// ListBlockDevices Lists all available/attached block devices
func ListBlockDevices() ([]*BlockDevice, error) {
	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, false, lsblkBinary, "-J", "-b", "-O")
	if err != nil {
		return nil, fmt.Errorf("%s", w.String())
	}

	bd, err := parseBlockDevicesDescriptor(w.Bytes())
	if err != nil {
		return nil, err
	}

	return bd, nil
}

func parseBlockDevicesDescriptor(data []byte) ([]*BlockDevice, error) {
	root := struct {
		BlockDevices []*BlockDevice `json:"blockdevices"`
	}{}

	err := json.Unmarshal(data, &root)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return root.BlockDevices, nil
}

func getNextStrToken(dec *json.Decoder, name string) (string, error) {
	t, _ := dec.Token()
	if t == nil {
		return "", nil
	}

	str, valid := t.(string)
	if !valid {
		return "", errors.Errorf("\"%s\" token should have a string value", name)
	}

	return str, nil
}

func getNextBoolToken(dec *json.Decoder, name string) (bool, error) {
	str, err := getNextStrToken(dec, name)
	if err != nil {
		return false, err
	}

	if str == "0" {
		return false, nil
	} else if str == "1" {
		return true, nil
	} else if str == "" {
		return false, nil
	}

	return false, errors.Errorf("Unknown ro value: %s", str)
}

// ParseVolumeSize will parse a string formated (1M, 10G, 2T) size and return its representation
// in bytes
func ParseVolumeSize(str string) (float64, error) {
	str = strings.ToLower(str)

	if !storageExp.MatchString(str) {
		return strconv.ParseFloat(str, 64)
	}

	unit := storageExp.ReplaceAllString(str, `$3`)
	size, err := strconv.ParseFloat(storageExp.ReplaceAllString(str, `$1`), 64)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	switch unit {
	case "k":
		size = size * (1 << 10)
	case "m":
		size = size * (1 << 20)
	case "g":
		size = size * (1 << 30)
	case "t":
		size = size * (1 << 40)
	case "p":
		size = size * (1 << 50)
	}

	return size, nil
}

// UnmarshalJSON decodes a BlockDevice, targered to integrate with json
// decoding framework
func (bd *BlockDevice) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}

		str, valid := t.(string)
		if !valid {
			continue
		}

		switch str {
		case "name":
			var name string

			name, err = getNextStrToken(dec, "name")
			if err != nil {
				return err
			}

			bd.Name = name
		case "model":
			var model string

			model, err = getNextStrToken(dec, "model")
			if err != nil {
				return err
			}

			bd.Model = model
		case "maj:min":
			var majMin string

			majMin, err = getNextStrToken(dec, "maj:min")
			if err != nil {
				return err
			}

			bd.MajorMinor = majMin
		case "size":
			var size string

			size, err = getNextStrToken(dec, "size")
			if err != nil {
				return err
			}

			bd.Size, err = ParseVolumeSize(size)
			if err != nil {
				return err
			}
		case "fstype":
			var fstype string

			fstype, err = getNextStrToken(dec, "fstype")
			if err != nil {
				return err
			}

			bd.FsType = fstype
		case "uuid":
			var uuid string

			uuid, err = getNextStrToken(dec, "uuid")
			if err != nil {
				return err
			}

			bd.UUID = uuid
		case "type":
			var tp string

			tp, err = getNextStrToken(dec, "type")
			if err != nil {
				return err
			}

			if tp == "disk" {
				bd.Type = BlockDeviceTypeDisk
			} else if tp == "part" {
				bd.Type = BlockDeviceTypePart
			} else if tp == "rom" {
				bd.Type = BlockDeviceTypeRom
			} else {
				return errors.Errorf("Unknown block device type: %s", tp)
			}
		case "state":
			var state string

			state, err = getNextStrToken(dec, "state")
			if err != nil {
				return err
			}

			if state == "running" {
				bd.State = BlockDeviceStateRunning
			} else if state == "live" {
				bd.State = BlockDeviceStateLive
			} else if state == "" {
				bd.Type = BlockDeviceStateUnknown
			} else {
				return errors.Errorf("Unrecognized block device state: %s", state)
			}
		case "mountpoint":
			var mpoint string

			mpoint, err = getNextStrToken(dec, "mountpoint")
			if err != nil {
				return err
			}

			bd.MountPoint = mpoint
		case "ro":
			bd.ReadOnly, err = getNextBoolToken(dec, "ro")
			if err != nil {
				return err
			}
		case "rm":
			bd.RemovableDevice, err = getNextBoolToken(dec, "rm")
			if err != nil {
				return err
			}
		case "children":
			bd.Children = []*BlockDevice{}
			err := dec.Decode(&bd.Children)
			if err != nil {
				return errors.Errorf("Invalid \"children\" token")
			}
		}
	}

	return nil
}

// SupportedFileSystems exposes the currently supported file system
func SupportedFileSystems() []string {
	return []string{"ext4", "vfat", "swap"}
}

// NewStandardPartitions will return a list of BlockDevice representing a
// default set of partitions required for an installation
func NewStandardPartitions(disk *BlockDevice) []*BlockDevice {
	bootSize := float64(150 * (1 << 20))
	swapSize := float64(2 * (1 << 30))
	rootSize := float64(disk.Size - bootSize - swapSize)

	// TODO review this standard partition schema (maybe add a default configuration)
	return []*BlockDevice{
		{
			Name:       fmt.Sprintf("%s1", disk.Name),
			Size:       bootSize,
			Type:       BlockDeviceTypePart,
			FsType:     "vfat",
			MountPoint: "/boot",
		},
		{
			Name:   fmt.Sprintf("%s2", disk.Name),
			Size:   swapSize,
			Type:   BlockDeviceTypePart,
			FsType: "swap",
		},
		{
			Name:       fmt.Sprintf("%s3", disk.Name),
			Size:       rootSize,
			Type:       BlockDeviceTypePart,
			FsType:     "ext4",
			MountPoint: "/",
		},
	}
}
