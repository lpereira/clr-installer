package frontend

// Frontend is the common interface for the frontend entry point
type Frontend interface {
	// MustRun is the method where the frontend implementation tells the
	// core code that this frontend wants to run
	MustRun() bool

	// Run is the actual entry point
	Run(rootDir string) error
}
