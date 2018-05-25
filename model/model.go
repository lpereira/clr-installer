package model

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/clearlinux/clr-installer/errors"
	"github.com/clearlinux/clr-installer/keyboard"
	"github.com/clearlinux/clr-installer/language"
	"github.com/clearlinux/clr-installer/network"
	"github.com/clearlinux/clr-installer/storage"
	"github.com/clearlinux/clr-installer/telemetry"
)

// Version of Clear Installer.
// Also used by the Makefile for releases.
const Version = "0.1.0"

// SystemInstall represents the system install "configuration", the target
// medias, bundles to install and whatever state a install may require
type SystemInstall struct {
	TargetMedias      []*storage.BlockDevice `yaml:"targetMedia"`
	NetworkInterfaces []*network.Interface   `yaml:"networkInterfaces"`
	Keyboard          *keyboard.Keymap       `yaml:"keyboard,omitempty,flow"`
	Language          *language.Language     `yaml:"language,omitempty,flow"`
	Bundles           []string               `yaml:"bundles,omitempty,flow"`
	HTTPSProxy        string                 `yaml:"httpsProxy,omitempty,flow"`
	Telemetry         *telemetry.Telemetry   `yaml:"telemetry,omitempty,flow"`
}

// ContainsBundle returns true if the data model has a bundle and false otherwise
func (si *SystemInstall) ContainsBundle(bundle string) bool {
	for _, curr := range si.Bundles {
		if curr == bundle {
			return true
		}
	}

	return false
}

// RemoveBundle removes a bundle from the data model
func (si *SystemInstall) RemoveBundle(bundle string) {
	bundles := []string{}

	for _, curr := range si.Bundles {
		if curr != bundle {
			bundles = append(bundles, curr)
		}
	}

	si.Bundles = bundles
}

// AddBundle adds a new bundle to the data model, we make sure to not duplicate entries
func (si *SystemInstall) AddBundle(bundle string) {
	for _, curr := range si.Bundles {
		if curr == bundle {
			return
		}
	}

	si.Bundles = append(si.Bundles, bundle)
}

// Validate checks the model for possible inconsistencies or "minimum required"
// information
func (si *SystemInstall) Validate() error {
	// si will be nil if we fail to unmarshal (coverage tests has a case for that)
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

	if si.Keyboard == nil {
		return errors.Errorf("Keyboard not set")
	}

	if si.Language == nil {
		return errors.Errorf("System Language not set")
	}

	if si.Telemetry == nil {
		return errors.Errorf("Telemetry not acknowledged")
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

// EnableTelemetry operates on the telemetry flag and enables or disables the target
// systems telemetry support based in enable argument
func (si *SystemInstall) EnableTelemetry(enable bool) {
	if si.Telemetry == nil {
		si.Telemetry = &telemetry.Telemetry{}
	}

	si.Telemetry.SetEnable(enable)
}

// IsTelemetryEnabled returns true if telemetry is enabled, false otherwise
func (si *SystemInstall) IsTelemetryEnabled() bool {
	if si.Telemetry == nil {
		return false
	}

	return si.Telemetry.Enabled
}

// WriteFile writes a yaml formatted representation of si into the provided file path
func (si *SystemInstall) WriteFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
