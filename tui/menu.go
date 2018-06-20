// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"fmt"
	//	"strings"

	"github.com/clearlinux/clr-installer/log"

	"github.com/VladimirMarkelov/clui"
)

// MenuPage is the Page implementation for the main menu page
type MenuPage struct {
	BasePage
	btns       []*SimpleButton
	installBtn *SimpleButton
}

func (page *MenuPage) addMenuItem(item Page) bool {

	buttonPrefix := page.GetButtonPrefix(item)
	title := fmt.Sprintf(" %s %s", buttonPrefix, item.GetMenuTitle())
	btn := CreateSimpleButton(page.content, AutoSize, AutoSize, title, Fixed)
	btn.SetStyle("Menu")
	btn.SetAlign(AlignLeft)

	btn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(item.GetID(), page.mi.currPage)
	})

	page.btns = append(page.btns, btn)

	log.Debug("buttonPrefix: %q", buttonPrefix)
	//return strings.Compare(buttonPrefix, MenuButtonPrefixUncompleted) == 0
	return buttonPrefix != MenuButtonPrefixUncompleted
}

// Activate is called when the page is "shown" and it repaints the main menu based on the
// available menu pages and their done/undone status
func (page *MenuPage) Activate() {
	for _, curr := range page.btns {
		curr.Destroy()
	}
	page.btns = []*SimpleButton{}

	previous := false
	activeSet := false
	for _, curr := range page.mi.pages {
		// Skip Menu Pages that are not required
		if !curr.IsRequired() {
			continue
		}

		if page.mi.prevPage != nil {
			// Is this menu option match the previous page?
			previous = page.mi.prevPage.GetID() == curr.GetID()
		}

		// Does the menu item added have the data set completed?
		completed := page.addMenuItem(curr)
		log.Debug("Completed = %s", completed)

		// If we haven't found the first active choice, set it
		if !activeSet && !completed {
			// Make last button added Active
			page.activated = page.btns[len(page.btns)-1]
			activeSet = true
		}

		// Special case if the previous page and the data set is not completed
		// we want THIS to be the active choice for easy return
		if previous && !completed {
			// Make last button added Active
			page.activated = page.btns[len(page.btns)-1]
			activeSet = true
		}
	}

	if page.getModel() != nil && page.getModel().Validate() == nil {
		page.installBtn.SetEnabled(true)
		page.activated = page.installBtn
	}
}

const (
	menuHelp = `Choose the next steps. Use <Tab> or arrow keys (up and down) to navigate
between the elements.
`
)

func newMenuPage(mi *Tui) (Page, error) {
	page := &MenuPage{}
	page.setup(mi, TuiPageMenu, NoButtons)

	lbl := clui.CreateLabel(page.content, 2, 3, menuHelp, Fixed)
	lbl.SetMultiline(true)
	lbl.SetPaddings(0, 2)

	cancelBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Cancel", Fixed)
	cancelBtn.OnClick(func(ev clui.Event) {
		go clui.Stop()
	})

	page.installBtn = CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Install", Fixed)
	page.installBtn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(TuiPageInstall, page.mi.currPage)
	})

	page.installBtn.SetEnabled(false)

	return page, nil
}
