package frontend

import (
	"clr-installer/model"
)

// Frontend is the common interface for the frontend entry point
type Frontend interface {
	// MustRun is the method where the frontend implementation tells the
	// core code that this frontend wants to run
	MustRun() bool

	// Run is the actual entry point
	Run(md *model.SystemInstall, rootDir string) (bool, error)
}
