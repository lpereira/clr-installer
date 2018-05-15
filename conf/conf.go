package conf

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// BundleListFile the file file for the containing the bundle list definition
	BundleListFile = "bundles.json"

	// ConfigFile is the install descriptor
	ConfigFile = "clr-installer.yaml"

	// DefaultConfigDir is the system wide default configuration directory
	DefaultConfigDir = "/usr/share/defaults/clr-installer"

	// SourcePath is the source path (within the .gopath)
	SourcePath = "src/github.com/clearlinux/clr-installer"
)

func isRunningFromSourceTree() (bool, string, error) {
	src, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return false, src, err
	}

	return !strings.HasPrefix(src, "/usr/bin"), src, nil
}

func lookupDefaultFile(file string) (string, error) {
	isSourceTree, sourcePath, err := isRunningFromSourceTree()
	if err != nil {
		return "", err
	}

	// use the config from source code's etc dir if not installed binary
	if isSourceTree {
		sourceRoot := strings.Replace(sourcePath, "bin", filepath.Join(SourcePath, "etc"), 1)
		return filepath.Join(sourceRoot, file), nil
	}

	return filepath.Join(DefaultConfigDir, file), nil
}

// LookupBundleListFile looks up the bundle list definition
// Guesses if we're running from source code our from system, if we're running from
// source code directory then we loads the source default file, otherwise tried to load
// the system installed file
func LookupBundleListFile() (string, error) {
	return lookupDefaultFile(BundleListFile)
}

// LookupDefaultConfig looks up the install descriptor
// Guesses if we're running from source code our from system, if we're running from
// source code directory then we loads the source default file, otherwise tried to load
// the system installed file
func LookupDefaultConfig() (string, error) {
	return lookupDefaultFile(ConfigFile)
}
