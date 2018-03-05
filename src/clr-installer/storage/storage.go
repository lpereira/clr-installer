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
	MajorMinor      string         // major:minor device number
	FsType          string         // filesystem type
	UUID            string         // filesystem uuid
	MountPoint      string         // where the device is mounted
	Size            float64        // size of the device
	Type            int            // device type
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
)

var (
	lsblkBinary = "lsblk"
	storageExp  = regexp.MustCompile(`([0-9](\.)?[0-9]*)?([m,g,t,k,p])`)
)

// ListBlockDevices Lists all available/attached block devices
func ListBlockDevices() ([]*BlockDevice, error) {
	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, false, lsblkBinary, "-J", "-b")
	if err != nil {
		return nil, fmt.Errorf("%s", w.String())
	}

	bd, err := parseBlockDevicesDescriptor(w.Bytes())
	if err != nil {
		return nil, err
	}

	w = bytes.NewBuffer(nil)
	err = cmd.Run(w, false, lsblkBinary, "-J", "-f")
	if err != nil {
		return nil, fmt.Errorf("%s", w.String())
	}

	fs, err := parseBlockDevicesDescriptor(w.Bytes())
	if err != nil {
		return nil, err
	}

	baseMp := map[string]*BlockDevice{}
	fsMp := map[string]*BlockDevice{}

	mapDescriptors(baseMp, bd)
	mapDescriptors(fsMp, fs)

	for k, v := range baseMp {
		v.FsType = fsMp[k].FsType
	}

	return bd, nil
}

func mapDescriptors(mp map[string]*BlockDevice, bds []*BlockDevice) {
	for _, curr := range bds {
		mp[curr.Name] = curr

		for _, ch := range curr.Children {
			mapDescriptors(mp, []*BlockDevice{ch})
		}
	}
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

func parseVolumeSize(str string) (float64, error) {
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

			bd.Size, err = parseVolumeSize(size)
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
