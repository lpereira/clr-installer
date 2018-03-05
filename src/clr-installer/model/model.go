package model

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"clr-installer/errors"
	"clr-installer/storage"
)

// SystemInstall represents the system install "configuration", the target
// medias, bundles to install and whatever state a install may require
type SystemInstall struct {
	TargetMedias []*storage.BlockDevice
	Bundles      []string
}

// Validate checks the model for possible inconsistencies or "minimum required"
// information
func (si *SystemInstall) Validate() error {
	if si == nil {
		return errors.Errorf("model is nil")
	}

	if si.TargetMedias == nil || len(si.TargetMedias) == 0 {
		return errors.Errorf("System Installation must provide a target media")
	}

	bootPartition := false
	rootPartition := false

	for _, curr := range si.TargetMedias {
		for _, ch := range curr.Children {
			if ch.FsType == "vfat" && ch.MountPoint == "/boot" {
				bootPartition = true
			}

			if ch.MountPoint == "/" {
				rootPartition = true
			}
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

// LoadFile loads a model from a json file pointed by path
func LoadFile(path string) (*SystemInstall, error) {
	var result SystemInstall

	if _, err := os.Stat(path); err == nil {
		configStr, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = json.Unmarshal(configStr, &result)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}

	return &result, nil
}
