package model

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"clr-installer/errors"
	"clr-installer/network"
	"clr-installer/storage"
)

// Version of Clear Installer.
// Also used by the Makefile for releases.
const Version = "0.1.0"

// SystemInstall represents the system install "configuration", the target
// medias, bundles to install and whatever state a install may require
type SystemInstall struct {
	TargetMedias      []*storage.BlockDevice `yaml:"targetMedia"`
	NetworkInterfaces []*network.Interface   `yaml:"networkInterfaces"`
	Keyboard          string                 `yaml:"keyboard,omitempty,flow"`
	Language          string                 `yaml:"language,omitempty,flow"`
	Bundles           []string               `yaml:"bundles,omitempty,flow"`
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

// AddNetworkInterface adds an Interface instance to the list of NetworkInterfaces
func (si *SystemInstall) AddNetworkInterface(iface *network.Interface) {
	if si.NetworkInterfaces == nil {
		si.NetworkInterfaces = []*network.Interface{}
	}

	si.NetworkInterfaces = append(si.NetworkInterfaces, iface)
}

// LoadFile loads a model from a yaml file pointed by path
func LoadFile(path string) (*SystemInstall, error) {
	var result SystemInstall

	if _, err := os.Stat(path); err == nil {
		configStr, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = yaml.Unmarshal(configStr, &result)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}

	return &result, nil
}

// WriteFile writes a yaml formatted representation of si into the provided file path
func (si *SystemInstall) WriteFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	b, err := yaml.Marshal(si)
	if err != nil {
		return err
	}

	// Write our header
	_, err = f.WriteString("#clear-linux-config\n")
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}
