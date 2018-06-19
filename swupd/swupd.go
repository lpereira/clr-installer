// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package swupd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/conf"
	"github.com/clearlinux/clr-installer/errors"
)

var (
	// CoreBundles represents the core bundles installed in the Verify() operation
	CoreBundles = []string{
		"os-core",
		"os-core-update",
	}
)

// SoftwareUpdater abstracts the swupd executable, environment and operations
type SoftwareUpdater struct {
	rootDir  string
	stateDir string
}

// Bundle maps a map name and description with the actual checkbox
type Bundle struct {
	Name string // Name the bundle name or id
	Desc string // Desc is the bundle long description
}

// IsCoreBundle checks if bundle is in the list of core bundles
func IsCoreBundle(bundle string) bool {
	for _, curr := range CoreBundles {
		if curr == bundle {
			return true
		}
	}
	return false
}

// New creates a new instance of SoftwareUpdater with the rootDir properly adjusted
func New(rootDir string) *SoftwareUpdater {
	return &SoftwareUpdater{rootDir, filepath.Join(rootDir, "/var/lib/swupd")}
}

// Verify runs "swupd verify" operation
func (s *SoftwareUpdater) Verify(version string) error {
	args := []string{
		"swupd",
		"verify",
		fmt.Sprintf("--path=%s", s.rootDir),
		fmt.Sprintf("--statedir=%s", s.stateDir),
		"--install",
		"-m",
		version,
		"--force",
		"--no-scripts",
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	args = []string{
		"swupd",
		"bundle-add",
		fmt.Sprintf("--path=%s", s.rootDir),
		fmt.Sprintf("--statedir=%s", s.stateDir),
		"os-core-update",
	}

	err = cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// Update executes the "swupd update" operation
func (s *SoftwareUpdater) Update() error {
	args := []string{
		filepath.Join(s.rootDir, "/usr/bin/swupd"),
		"update",
		fmt.Sprintf("--path=%s", s.rootDir),
		fmt.Sprintf("--statedir=%s", s.stateDir),
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// BundleAdd executes the "swupd bundle-add" operation for a single bundle
func (s *SoftwareUpdater) BundleAdd(bundle string) error {
	args := []string{
		filepath.Join(s.rootDir, "/usr/bin/swupd"),
		"bundle-add",
		fmt.Sprintf("--path=%s", s.rootDir),
		fmt.Sprintf("--statedir=%s", s.stateDir),
		bundle,
	}

	err := cmd.RunAndLog(args...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// LoadBundleList loads the bundle definitions
func LoadBundleList() ([]*Bundle, error) {
	path, err := conf.LookupBundleListFile()
	if err != nil {
		return nil, err
	}

	root := struct {
		Bundles []*Bundle `json:"bundles"`
	}{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if err = json.Unmarshal(data, &root); err != nil {
		return nil, errors.Wrap(err)
	}

	return root.Bundles, nil
}
