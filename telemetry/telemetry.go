// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package telemetry

// Telemetry represents the target system telemetry enabling flag
type Telemetry struct {
	Enabled     bool
	userDefined bool
}

// IsUserDefined returns true if the configuration was interactively
// defined by the user
func (tl *Telemetry) IsUserDefined() bool {
	return tl.userDefined
}

// MarshalYAML marshals Telemetry into YAML format
func (tl *Telemetry) MarshalYAML() (interface{}, error) {
	return tl.Enabled, nil
}

// UnmarshalYAML unmarshals Telemetry from YAML format
func (tl *Telemetry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var enabled bool

	if err := unmarshal(&enabled); err != nil {
		return err
	}

	tl.Enabled = enabled
	tl.userDefined = false
	return nil
}

// SetEnable sets the enabled flag and sets this is an user defined configuration
func (tl *Telemetry) SetEnable(enable bool) {
	tl.Enabled = enable
	tl.userDefined = true
}
