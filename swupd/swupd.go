package swupd

import (
	"fmt"
	"path/filepath"

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/errors"
)

// SoftwareUpdater abstracts the swupd executable, environment and operations
type SoftwareUpdater struct {
	rootDir  string
	stateDir string
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
