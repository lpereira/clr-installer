// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"fmt"

	"github.com/VladimirMarkelov/clui"
)

// AdvancedSubMenuPage is the Page implementation for Advance/Optional configuration options
type AdvancedSubMenuPage struct {
	BasePage
	btns []*SimpleButton
}

func (page *AdvancedSubMenuPage) addMenuItem(item Page) bool {

	buttonPrefix := page.GetButtonPrefix(item)
	title := fmt.Sprintf(" %s %s", buttonPrefix, item.GetMenuTitle())
	btn := CreateSimpleButton(page.content, AutoSize, AutoSize, title, Fixed)
	btn.SetStyle("Menu")
	btn.SetAlign(AlignLeft)

	btn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(item.GetID(), page.mi.currPage)
	})

	page.btns = append(page.btns, btn)

	return buttonPrefix != MenuButtonPrefixUncompleted
}

// Activate is called when the page is "shown" and it repaints the main menu based on the
// available menu pages and their done/undone status
func (page *AdvancedSubMenuPage) Activate() {
	for _, curr := range page.btns {
		curr.Destroy()
	}
	page.btns = []*SimpleButton{}

	previous := false
	activeSet := false
	for _, curr := range page.mi.pages {
		// Skip Menu Pages that are not required
		if curr.IsRequired() || curr.GetMenuTitle() == "" {
			continue
		}

		if page.mi.prevPage != nil {
			// Is this menu option match the previous page?
			previous = page.mi.prevPage.GetID() == curr.GetID()
		}

		// Does the menu item added have the data set completed?
		completed := page.addMenuItem(curr)

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
}

const (
	advancedDesc = `Advanced/Optional configuration items which influence the Installer.`
)

// The Advanced page gives the user the option so select how to set the storage device,
// if to manually configure it or a guided standard partition schema
func newAdvancedPage(mi *Tui) (Page, error) {
	page := &AdvancedSubMenuPage{
		BasePage: BasePage{
			// Tag this Page as required to be complete for the Install to proceed
			required: true,
		},
	}

	page.setupMenu(mi, TuiPageAdvancedMenu, "Advanced/Optional Menu", BackButton, TuiPageMenu)

	lbl := clui.CreateLabel(page.content, 2, 2, "Advanced/Optional Menu", clui.Fixed)
	lbl.SetPaddings(0, 2)

	lbl = clui.CreateLabel(page.content, 70, 3, advancedDesc, Fixed)
	lbl.SetMultiline(true)
	lbl.SetPaddings(0, 2)

	return page, nil
}
