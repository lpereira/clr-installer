// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/clearlinux/clr-installer/args"
	"github.com/clearlinux/clr-installer/errors"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/model"

	"github.com/VladimirMarkelov/clui"
	"github.com/nsf/termbox-go"
)

// Tui is the main tui data struct and holds data about the higher level data for this
// front end, it also implements the Frontend interface
type Tui struct {
	pages         []Page
	currPage      Page
	prevPage      Page
	model         *model.SystemInstall
	rootDir       string
	paniced       chan error
	installReboot bool
}

var (
	// errorLabelBg is a custom theme element, it has the error label background color definition
	errorLabelBg termbox.Attribute

	// errorLabelFg is a custom theme element, it has the error label foreground color definition
	errorLabelFg termbox.Attribute
)

// New creates a new Tui frontend instance
func New() *Tui {
	return &Tui{pages: []Page{}}
}

// MustRun is part of the Frontend interface implementation and tells the core that this
// frontend wants/must run.
func (mi *Tui) MustRun(args *args.Args) bool {
	return true
}

func lookupThemeDir() (string, error) {
	var result string

	themeDirs := []string{
		os.Getenv("CLR_INSTALLER_THEME_DIR"),
	}

	src, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	if strings.Contains(src, "/.gopath/bin") {
		themeDirs = append(themeDirs, strings.Replace(src, "bin", "../themes", 1))
	}

	themeDirs = append(themeDirs, "/usr/share/clr-installer/themes/")

	for _, curr := range themeDirs {
		if _, err := os.Stat(curr); os.IsNotExist(err) {
			continue
		}

		result = curr
		break
	}

	if result == "" {
		panic(errors.Errorf("Could not find a theme dir"))
	}

	return result, nil
}

// Run is part of the Frontend interface implementation and is the tui frontend main entry point
func (mi *Tui) Run(md *model.SystemInstall, rootDir string) (bool, error) {
	clui.InitLibrary()
	defer clui.DeinitLibrary()

	mi.model = md
	themeDir, err := lookupThemeDir()
	if err != nil {
		return false, err
	}

	clui.SetThemePath(themeDir)

	if !clui.SetCurrentTheme("clr-installer") {
		panic("Could not change theme")
	}

	errorLabelBg = clui.RealColor(clui.ColorDefault, "ErrorLabelBack")
	errorLabelFg = clui.RealColor(clui.ColorDefault, "ErrorLabelText")

	mi.rootDir = rootDir
	mi.paniced = make(chan error, 1)

	menus := []struct {
		desc string
		fc   func(*Tui) (Page, error)
	}{
		{"timezone", newTimezonePage},
		{"language", newLanguagePage},
		{"keyboard", newKeyboardPage},
		{"disk menu", newDiskPage},
		{"network", newNetworkPage},
		{"proxy", newProxyPage},
		{"network validate", newNetworkValidatePage},
		{"network interface", newNetworkInterfacePage},
		{"main menu", newMenuPage},
		{"guided partitioning", newGuidedPartitionPage},
		{"manual partitioning", newManualPartitionPage},
		{"disk partition", newDiskPartitionPage},
		{"bundle selection", newBundlePage},
		{"add user", newUseraddPage},
		{"telemetry enabling", newTelemetryPage},
		{"install", newInstallPage},
	}

	for _, menu := range menus {
		var page Page

		if page, err = menu.fc(mi); err != nil {
			return false, err
		}

		mi.pages = append(mi.pages, page)
	}

	mi.gotoPage(TuiPageMenu, mi.currPage)

	var paniced error

	go func() {
		if paniced = <-mi.paniced; paniced != nil {
			clui.Stop()
			log.ErrorError(paniced)
		}
	}()

	clui.MainLoop()

	if paniced != nil {
		panic(paniced)
	}

	return mi.installReboot, nil
}

func (mi *Tui) gotoPage(id int, currPage Page) {
	if mi.currPage != nil {
		mi.currPage.GetWindow().SetVisible(false)
		mi.currPage.DeActivate()

		// TODO clui is not hiding cursor when we hide/destroy an edit widget
		termbox.HideCursor()
	}

	mi.currPage = mi.getPage(id)
	mi.prevPage = currPage

	mi.currPage.Activate()
	mi.currPage.GetWindow().SetVisible(true)

	clui.ActivateControl(mi.currPage.GetWindow(), mi.currPage.GetActivated())
}

func (mi *Tui) getPage(page int) Page {
	for _, curr := range mi.pages {
		if curr.GetID() == page {
			return curr
		}
	}

	return nil
}
