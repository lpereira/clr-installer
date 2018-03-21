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
	Keyboard     string
	Language     string
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

	for _, curr := range si.TargetMedias {
		if err := curr.Validate(); err != nil {
			return err
		}
	}

	if si.Keyboard == "" {
		return errors.Errorf("Keyboard not set")
	}

	if si.Language == "" {
		return errors.Errorf("System Language not set")
	}

	return nil
}

// AddTargetMedia adds a BlockDevice instance to the list of TargetMedias
func (si *SystemInstall) AddTargetMedia(bd *storage.BlockDevice) {
	if si.TargetMedias == nil {
		si.TargetMedias = []*storage.BlockDevice{}
	}

	si.TargetMedias = append(si.TargetMedias, bd)
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
