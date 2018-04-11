package frontend

import (
	"clr-installer/model"
)

// Args represents the user provided arguments
type Args struct {
	Version    bool
	Reboot     bool
	LogFile    string
	ConfigFile string
	LogLevel   int
	ForceTUI   bool
}

// Frontend is the common interface for the frontend entry point
type Frontend interface {
	// MustRun is the method where the frontend implementation tells the
	// core code that this frontend wants to run
	MustRun(args *Args) bool

	// Run is the actual entry point
	Run(md *model.SystemInstall, rootDir string) (bool, error)
}
