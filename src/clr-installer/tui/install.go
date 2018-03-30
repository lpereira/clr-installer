package tui

import (
	"os/user"
	"path/filepath"
	"time"

	"clr-installer/controller"
	"clr-installer/progress"
	"github.com/VladimirMarkelov/clui"
)

// InstallPage is the Page implementation for installation progress page, it also implements
// the progress.Client interface
type InstallPage struct {
	BasePage
	rebootBtn *SimpleButton
	prgBar    *clui.ProgressBar
	prgLabel  *clui.Label
	prgMax    int
	descFile  string
}

var (
	loopWaitDuration = 2 * time.Second
)

// Done is part of the progress.Client implementation and sets the progress bar to "full"
func (page *InstallPage) Done() {
	page.prgBar.SetValue(page.prgMax)
	clui.RefreshScreen()
}

// Step is part of the progress.Client implementation and moves the progress bar one step
// case it becomes full it starts again
func (page *InstallPage) Step() {
	if page.prgBar.Value() == page.prgMax {
		page.prgBar.SetValue(0)
	} else {
		page.prgBar.Step()
	}
	clui.RefreshScreen()
}

// Desc is part of the progress.Client implementation and sets the progress bar label
func (page *InstallPage) Desc(desc string) {
	page.prgLabel.SetTitle(desc)
	clui.RefreshScreen()
}

// Partial is part of the progress.Client implementation and adjusts the progress bar to the
// current completion percentage
func (page *InstallPage) Partial(total int, step int) {
	perc := (step / total)
	value := page.prgMax * perc
	page.prgBar.SetValue(int(value))
}

// LoopWaitDuration is part of the progress.Client implementation and returns the time duration
// each step should wait until calling Step again
func (page *InstallPage) LoopWaitDuration() time.Duration {
	return loopWaitDuration
}

// Activate is called when the page is "shown"
func (page *InstallPage) Activate() {
	go func() {
		progress.Set(page)

		err := controller.Install(page.mi.rootDir, page.mi.model)
		if err != nil {
			page.Panic(err)
		}

		if err := page.mi.model.WriteFile(page.descFile); err != nil {
			page.Panic(err)
		}

		prg := progress.NewLoop("Cleaning up install environment")
		if err := controller.Cleanup(page.mi.rootDir, true); err != nil {
			page.Panic(err)
		}
		prg.Done()

		page.prgLabel.SetTitle("Installation complete")
		page.rebootBtn.SetEnabled(true)
		clui.ActivateControl(page.GetWindow(), page.rebootBtn)
		clui.RefreshScreen()

		page.mi.installed = true
	}()
}

func newInstallPage(mi *Tui) (Page, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	page := &InstallPage{}
	page.setup(mi, TuiPageInstall, NoButtons)
	page.descFile = filepath.Join(usr.HomeDir, "clr-installer.json")

	lbl := clui.CreateLabel(page.content, 2, 2, "Installing Clear Linux", Fixed)
	lbl.SetPaddings(0, 2)

	progressFrame := clui.CreateFrame(page.content, AutoSize, 3, BorderNone, clui.Fixed)
	progressFrame.SetPack(clui.Vertical)

	page.prgBar = clui.CreateProgressBar(progressFrame, AutoSize, AutoSize, clui.Fixed)

	page.prgMax, _ = page.prgBar.Size()
	page.prgBar.SetLimits(0, page.prgMax)

	page.prgLabel = clui.CreateLabel(progressFrame, 1, 1, "Installing", Fixed)
	page.prgLabel.SetPaddings(0, 3)

	page.rebootBtn = CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Reboot", Fixed)
	page.rebootBtn.OnClick(func(ev clui.Event) {
		go clui.Stop()
	})
	page.rebootBtn.SetEnabled(false)

	return page, nil
}
