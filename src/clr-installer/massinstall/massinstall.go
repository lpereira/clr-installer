package massinstall

import (
	"fmt"
	"time"

	"clr-installer/controller"
	"clr-installer/log"
	"clr-installer/model"
	"clr-installer/progress"
)

// MassInstall is the frontend implementation for the "mass installer" it also
// implements the progress interface: progress.Client
type MassInstall struct {
	configFile string
	prgDesc    string
	prgIndex   int
}

// New creates a new instance of MassInstall frontend implementation
func New(config string) *MassInstall {
	return &MassInstall{configFile: config}
}

// Step is the progress step implementation for progress.Client interface
func (mi *MassInstall) Step() {
	elms := []string{"|", "-", "\\", "|", "/", "-", "\\"}

	fmt.Printf("%s [%s]\r", mi.prgDesc, elms[mi.prgIndex])

	if mi.prgIndex+1 == len(elms) {
		mi.prgIndex = 0
	} else {
		mi.prgIndex = mi.prgIndex + 1
	}
}

// LoopWaitDuration is part of the progress.Client implementation and returns the
// duration each loop progress step should wait
func (mi *MassInstall) LoopWaitDuration() time.Duration {
	return 50 * time.Millisecond
}

// Desc is part of the implementation for ProgresIface and is used to adjust the progress bar
// label content
func (mi *MassInstall) Desc(desc string) {
	mi.prgDesc = desc
}

// Partial is part of the progress.Client implementation and sets the progress bar based
// on actuall progression
func (mi *MassInstall) Partial(total int, step int) {
	line := fmt.Sprintf("%s %.0f%%\r", mi.prgDesc, (float64(step)/float64(total))*100)
	fmt.Printf("%s", line)
}

// Done is part of the progress.Client implementation and represents the progress task "done"
// notification
func (mi *MassInstall) Done() {
	mi.prgIndex = 0
	fmt.Printf("%s [done]\n", mi.prgDesc)
}

// MustRun is part of the Frontend implementation and tells the core implementation that this
// frontend wants or should be executed
func (mi *MassInstall) MustRun() bool {
	return mi.configFile != ""
}

// Run is part of the Frontend implementation and is the actual entry point for the
// "mass installer" frontend
func (mi *MassInstall) Run(rootDir string) error {
	progress.Set(mi)

	log.Debug("Loading config file: %s", mi.configFile)
	md, err := model.LoadFile(mi.configFile)
	if err != nil {
		return err
	}

	log.Debug("Starting install")
	err = controller.Install(rootDir, md)
	if err != nil {
		log.ErrorError(err)
	}

	prg := progress.NewLoop("Cleaning up install environment")
	if controller.Cleanup(rootDir, true) != nil {
		log.ErrorError(err)
	}
	prg.Done()

	return nil
}