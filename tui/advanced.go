// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"github.com/VladimirMarkelov/clui"
)

// AdvancedSubMenuPage is the Page implementation for Advance/Optional configuration options
type AdvancedSubMenuPage struct {
	BasePage
}

const (
	AdvancedDesc = `Advanced/Optional configuration items which influence the Installer.`
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

	page.setupMenu(mi, TuiPageAdvancedMenu, "Advanced/Optional Menu", BackButton)

	lbl := clui.CreateLabel(page.content, 2, 2, "Advanced/Optional Menu", clui.Fixed)
	lbl.SetPaddings(0, 2)

	lbl = clui.CreateLabel(page.content, 70, 3, AdvancedDesc, Fixed)
	lbl.SetMultiline(true)

	clui.CreateLabel(page.content, AutoSize, 2, "Partitioning methods:", Fixed)

	gBtn := CreateSimpleButton(page.content, AutoSize, AutoSize, " Guided - use entire disk", Fixed)
	gBtn.SetAlign(AlignLeft)
	gBtn.OnClick(func(ev clui.Event) {
		mi.gotoPage(TuiPageGuidedPart, mi.currPage)
	})

	mBtn := CreateSimpleButton(page.content, AutoSize, AutoSize, " Manual configuration", Fixed)
	mBtn.SetAlign(AlignLeft)
	mBtn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(TuiPageManualPart, mi.currPage)
	})

	page.activated = gBtn
	return page, nil
}