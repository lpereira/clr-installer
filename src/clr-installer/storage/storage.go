package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"clr-installer/cmd"
	"clr-installer/errors"
)

// A BlockDevice describes a block device and its partitions
type BlockDevice struct {
	Name            string           // device name
	Model           string           // device model
	MajorMinor      string           // major:minor device number
	FsType          string           // filesystem type
	UUID            string           // filesystem uuid
	MountPoint      string           // where the device is mounted
	Size            uint64           // size of the device
	Type            BlockDeviceType  // device type
	State           BlockDeviceState // device state (running, live etc)
	ReadOnly        bool             // read-only device
	RemovableDevice bool             // removable device
	Children        []*BlockDevice   // children devices/partitions
	userDefined     bool             // was this value set by user?
}

// Version used for reading and writing YAML
type blockDeviceYAMLMarshal struct {
	Name            string         `yaml:"name,omitempty"`
	Model           string         `yaml:"model,omitempty"`
	MajorMinor      string         `yaml:"majMin,omitempty"`
	FsType          string         `yaml:"fstype,omitempty"`
	UUID            string         `yaml:"uuid,omitempty"`
	MountPoint      string         `yaml:"mountpoint,omitempty"`
	Size            string         `yaml:"size,omitempty"`
	ReadOnly        string         `yaml:"ro,omitempty"`
	RemovableDevice string         `yaml:"rm,omitempty"`
	Type            string         `yaml:"type,omitempty"`
	State           string         `yaml:"state,omitempty"`
	Children        []*BlockDevice `yaml:"children,omitempty"`
}

// BlockDeviceState is the representation of a block device state (live, running, etc)
type BlockDeviceState int

// BlockDeviceType is the representation of a block device type (disk, part, rom, etc)
type BlockDeviceType int

const (
	// BlockDeviceTypeDisk identifies a BlockDevice as a disk
	BlockDeviceTypeDisk = iota

	// BlockDeviceTypePart identifies a BlockDevice as a partition
	BlockDeviceTypePart

	// BlockDeviceTypeRom identifies a BlockDevice as a rom
	BlockDeviceTypeRom

	// BlockDeviceTypeLVM2Group indeitifies a BlockDevice as a lvm2 group
	BlockDeviceTypeLVM2Group

	// BlockDeviceTypeLVM2Volume identifies a BlockDevice as a lvm2 volume
	BlockDeviceTypeLVM2Volume

	// BlockDeviceTypeUnknown identifies a BlockDevice as unknown
	BlockDeviceTypeUnknown

	// BlockDeviceStateUnknown identifies a BlockDevice in a unknown state
	BlockDeviceStateUnknown = iota

	// BlockDeviceStateRunning identifies a BlockDevice as running
	BlockDeviceStateRunning

	// BlockDeviceStateLive identifies a BlockDevice as live
	BlockDeviceStateLive
)

var (
	lsblkBinary         = "lsblk"
	storageExp          = regexp.MustCompile(`([0-9](\.)?[0-9]*)?([m,g,t,k,p])`)
	blockDeviceStateMap = map[BlockDeviceState]string{
		BlockDeviceStateRunning: "running",
		BlockDeviceStateLive:    "live",
		BlockDeviceStateUnknown: "",
	}
	blockDeviceTypeMap = map[BlockDeviceType]string{
		BlockDeviceTypeDisk:       "disk",
		BlockDeviceTypePart:       "part",
		BlockDeviceTypeRom:        "rom",
		BlockDeviceTypeLVM2Group:  "LVM2_member",
		BlockDeviceTypeLVM2Volume: "lvm",
		BlockDeviceTypeUnknown:    "",
	}
)

func (bt BlockDeviceType) String() string {
	return blockDeviceTypeMap[bt]
}

func parseBlockDeviceType(bdt string) (BlockDeviceType, error) {
	for k, v := range blockDeviceTypeMap {
		if v == bdt {
			return k, nil
		}
	}

	return BlockDeviceTypeUnknown, errors.Errorf("Unknown block device type: %s", bdt)
}

func (bs BlockDeviceState) String() string {
	return blockDeviceStateMap[bs]
}

func parseBlockDeviceState(bds string) (BlockDeviceState, error) {
	for k, v := range blockDeviceStateMap {
		if v == bds {
			return k, nil
		}
	}

	return BlockDeviceStateUnknown, errors.Errorf("Unrecognized block device state: %s", bds)
}

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
func HumanReadableSize(size uint64) (string, error) {

	if size == 0 {
		return fmt.Sprintf("0"), nil
	}

	sizes := []struct {
		unit string
		mask uint64
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

		return fmt.Sprintf("%v%s", csize, curr.unit), nil
	}

	return "", errors.Errorf("Could not format desk/partition size")
}

// FreeSpace returns the block device available/free space considering the currently
// configured partition table
func (bd *BlockDevice) FreeSpace() (uint64, error) {
	if bd.Type != BlockDeviceTypeDisk {
		return 0, errors.Errorf("FreeSpace() must only be called with a disk block device")
	}

	var total uint64
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
// userDefined will be inserted in the resulting list reather the loaded ones
func ListBlockDevices(userDefined []*BlockDevice) ([]*BlockDevice, error) {
	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, lsblkBinary, "-J", "-b", "-O")
	if err != nil {
		return nil, fmt.Errorf("%s", w.String())
	}

	bds, err := parseBlockDevicesDescriptor(w.Bytes())
	if err != nil {
		return nil, err
	}

	if userDefined == nil || len(userDefined) == 0 {
		return bds, nil
	}

	merged := []*BlockDevice{}
	for _, loaded := range bds {
		added := false

		for _, udef := range userDefined {
			if !loaded.Equals(udef) {
				continue
			}

			merged = append(merged, udef)
			added = true
			break
		}

		if !added {
			merged = append(merged, loaded)
		}
	}

	return merged, nil
}

// Equals compares two BlockDevice instances
func (bd *BlockDevice) Equals(cmp *BlockDevice) bool {
	if cmp == nil {
		return false
	}

	return bd.Name == cmp.Name && bd.Model == cmp.Model && bd.MajorMinor == cmp.MajorMinor
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
func ParseVolumeSize(str string) (uint64, error) {
	var size uint64

	str = strings.ToLower(str)

	if !storageExp.MatchString(str) {
		return strconv.ParseUint(str, 0, 64)
	}

	unit := storageExp.ReplaceAllString(str, `$3`)
	fsize, err := strconv.ParseFloat(storageExp.ReplaceAllString(str, `$1`), 64)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	switch unit {
	case "k":
		fsize = fsize * (1 << 10)
	case "m":
		fsize = fsize * (1 << 20)
	case "g":
		fsize = fsize * (1 << 30)
	case "t":
		fsize = fsize * (1 << 40)
	case "p":
		fsize = fsize * (1 << 50)
	}

	size = uint64(math.Round(fsize))

	return size, nil
}

// UnmarshalJSON decodes a BlockDevice, targeted to integrate with json
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

			bd.Type, err = parseBlockDeviceType(tp)
			if err != nil {
				return err
			}
		case "state":
			var state string

			state, err = getNextStrToken(dec, "state")
			if err != nil {
				return err
			}

			bd.State, err = parseBlockDeviceState(state)
			if err != nil {
				return err
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
				return errors.Errorf("Invalid \"children\" token: %s", err)
			}
		}
	}

	return nil
}

// MarshalYAML is the yaml Marshaller implementation
func (bd *BlockDevice) MarshalYAML() (interface{}, error) {

	var bdm blockDeviceYAMLMarshal

	bdm.Name = bd.Name
	bdm.Model = bd.Model
	bdm.MajorMinor = bd.MajorMinor
	bdm.FsType = bd.FsType
	bdm.UUID = bd.UUID
	bdm.MountPoint = bd.MountPoint
	bdm.Size = strconv.FormatUint(bd.Size, 10)
	bdm.ReadOnly = strconv.FormatBool(bd.ReadOnly)
	bdm.RemovableDevice = strconv.FormatBool(bd.RemovableDevice)
	bdm.Type = bd.Type.String()
	bdm.State = bd.State.String()
	bdm.Children = bd.Children

	return bdm, nil
}

// UnmarshalYAML is the yaml Unmarshaller implementation
func (bd *BlockDevice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var unmarshBlockDevice blockDeviceYAMLMarshal

	if err := unmarshal(&unmarshBlockDevice); err != nil {
		return err
	}

	// Copy the unmarshaled data
	bd.userDefined = false
	bd.Name = unmarshBlockDevice.Name
	bd.Model = unmarshBlockDevice.Model
	bd.MajorMinor = unmarshBlockDevice.MajorMinor
	bd.FsType = unmarshBlockDevice.FsType
	bd.UUID = unmarshBlockDevice.UUID
	bd.MountPoint = unmarshBlockDevice.MountPoint
	bd.Children = unmarshBlockDevice.Children
	// Convert String to Uint64
	if unmarshBlockDevice.Size != "" {
		uSize, err := ParseVolumeSize(unmarshBlockDevice.Size)
		if err != nil {
			return err
		}
		bd.Size = uSize
	}

	// Map the BlockDeviceType
	if unmarshBlockDevice.Type != "" {
		iType, err := parseBlockDeviceType(unmarshBlockDevice.Type)
		if err != nil {
			return errors.Errorf("Device: %s: %v", unmarshBlockDevice.Name, err)
		}
		if iType < 0 || iType > BlockDeviceTypeUnknown {
		}
		bd.Type = iType
	}

	// Map the BlockDeviceState
	if unmarshBlockDevice.State != "" {
		iState, err := parseBlockDeviceState(unmarshBlockDevice.State)
		if err != nil {
			return errors.Errorf("Device: %s: %v", unmarshBlockDevice.Name, err)
		}
		bd.State = iState
	}

	// Map the ReanOnly bool
	if unmarshBlockDevice.ReadOnly != "" {
		bReadOnly, err := strconv.ParseBool(unmarshBlockDevice.ReadOnly)
		if err != nil {
			return err
		}
		bd.ReadOnly = bReadOnly
	}

	// Map the RemovableDevice bool
	if unmarshBlockDevice.RemovableDevice != "" {
		bRemovableDevice, err := strconv.ParseBool(unmarshBlockDevice.RemovableDevice)
		if err != nil {
			return err
		}
		bd.RemovableDevice = bRemovableDevice
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
	bootSize := uint64(150 * (1 << 20))
	swapSize := uint64(2 * (1 << 30))
	rootSize := uint64(disk.Size - bootSize - swapSize)

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
