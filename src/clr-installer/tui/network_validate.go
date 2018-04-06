package tui

import (
	"time"

	"clr-installer/network"
	"clr-installer/progress"
	"github.com/VladimirMarkelov/clui"
)

// NetworkValidatePage is the Page implementation for network config validation, it also
// implements the progress.Client interface
type NetworkValidatePage struct {
	BasePage
	prgLabel *clui.Label
	prgBar   *clui.ProgressBar
	prgMax   int
	doneBtn  *SimpleButton
}

const (
	networkTestDesc = `By pressing the "Test" button we'll attempt to apply the specified
network settings and try to reach the required network servers.`
)

// Done is part of the progress.Client implementation and sets the progress bar to "full"
func (page *NetworkValidatePage) Done() {
	page.prgBar.SetValue(page.prgMax)
	clui.RefreshScreen()
}

// Step is part of the progress.Client implementation and moves the progress bar one step
// case it becomes full it starts again
func (page *NetworkValidatePage) Step() {
	if page.prgBar.Value() == page.prgMax {
		page.prgBar.SetValue(0)
	} else {
		page.prgBar.Step()
	}
	clui.RefreshScreen()
}

// Desc is part of the progress.Client implementation and sets the progress bar label
func (page *NetworkValidatePage) Desc(desc string) {
	page.prgLabel.SetTitle(desc)
	clui.RefreshScreen()
}

// Partial is part of the progress.Client implementation and adjusts the progress bar to the
// current completion percentage
func (page *NetworkValidatePage) Partial(total int, step int) {
}

// LoopWaitDuration is part of the progress.Client implementation and returns the time duration
// each step should wait until calling Step again
func (page *NetworkValidatePage) LoopWaitDuration() time.Duration {
	return 1 * time.Second
}

// Activate resets the page state
func (page *NetworkValidatePage) Activate() {
	page.doneBtn.SetEnabled(false)
	page.prgLabel.SetTitle("")
	page.prgBar.SetValue(0)
}

func newNetworkValidatePage(mi *Tui) (Page, error) {
	page := &NetworkValidatePage{}
	page.setupMenu(mi, TuiPageNetworkValidate, "Test Network Settings", NoButtons)

	lbl := clui.CreateLabel(page.content, 2, 2, "Test Network Settings", clui.Fixed)
	lbl.SetPaddings(0, 2)

	lbl = clui.CreateLabel(page.content, 70, 3, networkTestDesc, Fixed)
	lbl.SetMultiline(true)

	progressFrame := clui.CreateFrame(page.content, AutoSize, 3, BorderNone, clui.Fixed)
	progressFrame.SetPack(clui.Vertical)

	page.prgBar = clui.CreateProgressBar(progressFrame, AutoSize, AutoSize, clui.Fixed)

	page.prgLabel = clui.CreateLabel(progressFrame, 1, 1, "", Fixed)
	page.prgLabel.SetPaddings(0, 3)

	page.prgMax, _ = page.prgBar.Size()
	page.prgBar.SetLimits(0, page.prgMax)

	page.doneBtn = CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Done", Fixed)
	page.doneBtn.SetEnabled(false)

	page.doneBtn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(TuiPageMenu, page.mi.currPage)
	})

	btn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Test", Fixed)
	btn.OnClick(func(ev clui.Event) {
		go func() {
			progress.Set(page)

			prg := progress.NewLoop("Applying network settings")
			if err := network.Apply("/", page.getModel().NetworkInterfaces); err != nil {
				page.Panic(err)
			}
			prg.Done()

			prg = progress.NewLoop("Restarting network interfaces")
			if err := network.Restart(); err != nil {
				page.Panic(err)
			}
			prg.Done()

			prg = progress.NewLoop("Testing connectivity")
			if err := network.Test(); err != nil {
				page.prgLabel.SetTitle("Failed, network is not working.")
			} else {
				page.prgLabel.SetTitle("Success.")
			}
			prg.Done()

			page.doneBtn.SetEnabled(true)
			clui.ActivateControl(mi.currPage.GetWindow(), page.doneBtn)
			clui.RefreshScreen()
		}()
	})

	page.activated = btn

	return page, nil
}
