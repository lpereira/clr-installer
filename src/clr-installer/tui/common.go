package tui

import (
	"os/user"
	"path/filepath"

	"clr-installer/model"
	"github.com/VladimirMarkelov/clui"
)

// BasePage is the common implementation for the TUI frontend
// other pages will inherit this base page behaviours
type BasePage struct {
	mi        *Tui          // the Tui frontend reference
	window    *clui.Window  // the page window
	mFrame    *clui.Frame   // main frame
	content   *clui.Frame   // the main content frame
	cFrame    *clui.Frame   // control frame
	backBtn   *SimpleButton // back button
	doneBtn   *SimpleButton // done button
	activated clui.Control  // activated control
	menuTitle string        // the title used to show on main menu
	done      bool          // marks if an item is completed
	id        int           // the page id
	data      interface{}   // arbitrary page context data
}

// Page defines the methods a Page must implement
type Page interface {
	GetID() int
	GetWindow() *clui.Window
	GetActivated() clui.Control
	GetMenuTitle() string
	SetData(data interface{})
	GetData() interface{}
	SetDone(done bool) bool
	GetDone() bool
	Activate()
	DeActivate()
}

var (
	descFile string
)

const (
	// AutoSize is shortcut for clui.AutoSize flag
	AutoSize = clui.AutoSize

	// Fixed is shortcut for clui.Fixed flag
	Fixed = clui.Fixed

	// BorderNone is shortcut for clui.BorderNone flag
	BorderNone = clui.BorderNone

	// AlignLeft is shortcut for clui.AlignLeft flag
	AlignLeft = clui.AlignLeft

	// AlignRight is shortcut for clui.AlignRight flag
	AlignRight = clui.AlignRight

	// NoButtons mask defines a common Page will not set any control button
	NoButtons = 0

	// BackButton mask defines a common Page will have a back button
	BackButton = 1 << 1

	// DoneButton mask defines a common Page will have  done button
	DoneButton = 1 << 2

	// AllButtons mask defines a common Page will have both Back and Done buttons
	AllButtons = BackButton | DoneButton

	// TuiPageMenu is the id for menu page
	TuiPageMenu = iota

	// TuiPageInstall is the id for install page
	TuiPageInstall

	// TuiPageLanguage is the id for language page
	TuiPageLanguage

	// TuiPageKeyboard is the id for keyboard page
	TuiPageKeyboard

	// TuiPageDiskMenu is the id for disk configuration menu
	TuiPageDiskMenu

	// TuiPageDiskPart is the id for disk partition configuration page
	TuiPageDiskPart

	// TuiPageGuidedPart is the id for disk guided partitioning page
	TuiPageGuidedPart

	// TuiPageManualPart is the id for disk manual partitioning page
	TuiPageManualPart

	// TuiPageNetwork is the id for network configuration page
	TuiPageNetwork

	// TuiPageNetworkValidate is the id for the network validation page
	TuiPageNetworkValidate

	// TuiPageInterface is the id for the network interface configuration page
	TuiPageInterface
)

// DeActivate is a stub implementation for the DeActivate method of Page interface
func (page *BasePage) DeActivate() {}

// Activate is a stub implementation for the Activate method of Page interface
func (page *BasePage) Activate() {}

// SetDone sets the page's done flag
func (page *BasePage) SetDone(done bool) bool {
	page.done = done
	return true
}

// Panic write an error to the tui paniced channel - we'll deal the error, stop clui
// mainloop and nicely panic() the application
func (page *BasePage) Panic(err error) {
	page.mi.paniced <- err
}

// GetDone returns the current value of a page's done flag
func (page *BasePage) GetDone() bool {
	return page.done
}

// GetData returns the current value of a page's data member
func (page *BasePage) GetData() interface{} {
	return page.data
}

// SetData set the current value for the page's data member
func (page *BasePage) SetData(data interface{}) {
	page.data = data
}

// GetMenuTitle returns the current page's title string
func (page *BasePage) GetMenuTitle() string {
	return page.menuTitle
}

// GetActivated returns the control set as activated for a page
func (page *BasePage) GetActivated() clui.Control {
	return page.activated
}

// GetWindow returns the current page's window control
func (page *BasePage) GetWindow() *clui.Window {
	return page.window
}

// GetID returns the current page's identifier
func (page *BasePage) GetID() int {
	return page.id
}

func (page *BasePage) setupMenu(mi *Tui, id int, menuTitle string, btns int) {
	page.setup(mi, id, btns)
	page.menuTitle = menuTitle
}

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	descFile = filepath.Join(usr.HomeDir, "clr-installer.yaml")
}

func (page *BasePage) setup(mi *Tui, id int, btns int) {
	page.id = id
	page.mi = mi
	page.newWindow()

	page.mFrame = clui.CreateFrame(page.window, 78, 22, BorderNone, clui.Fixed)
	page.mFrame.SetPack(clui.Vertical)

	page.content = clui.CreateFrame(page.mFrame, 8, 21, BorderNone, clui.Fixed)
	page.content.SetPack(clui.Vertical)
	page.content.SetPaddings(2, 1)

	page.cFrame = clui.CreateFrame(page.mFrame, AutoSize, 1, BorderNone, Fixed)
	page.cFrame.SetPack(clui.Horizontal)
	page.cFrame.SetGaps(1, 1)
	page.cFrame.SetPaddings(2, 0)

	if btns&BackButton == BackButton {
		page.newBackButton()
	}

	if btns&DoneButton == DoneButton {
		page.newDoneButton(mi)
	}

	page.window.SetVisible(false)
}

func (page *BasePage) newWindow() {
	sw, sh := clui.ScreenSize()

	x := (sw - 80) / 2
	y := (sh - 24) / 2

	page.window = clui.AddWindow(x, y, 80, 24, " [Clear Linux Installer ("+model.Version+")] ")
	page.window.SetTitleButtons(0)

	page.window.OnScreenResize(func(evt clui.Event) {
		ww, wh := page.window.Size()

		x := (evt.Width - ww) / 2
		y := (evt.Height - wh) / 2

		page.window.SetPos(x, y)
		page.window.PlaceChildren()
	})
}

func (page *BasePage) newBackButton() {
	btn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "< Main Menu", Fixed)

	btn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(TuiPageMenu, page.mi.currPage)
	})

	page.backBtn = btn
}

func (page *BasePage) newDoneButton(mi *Tui) {
	btn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Done", Fixed)

	btn.OnClick(func(ev clui.Event) {
		if mi.currPage.SetDone(true) {
			mi.gotoPage(TuiPageMenu, page.mi.currPage)
		}
	})
	page.doneBtn = btn
}
