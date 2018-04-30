package tui

import (
	"fmt"

	"github.com/VladimirMarkelov/clui"
)

// BundlePage is the Page implementation for the proxy configuration page
type BundlePage struct {
	BasePage
}

// Bundle maps a map name and description with the actual checkbox
type Bundle struct {
	name  string
	desc  string
	check *clui.CheckBox
}

var (
	bundles = []*Bundle{
		{"editors", "Popular text editors (terminal-based)", nil},
		{"user-basic", "Captures most console work flows", nil},
		{"desktop-autostart", "UI that automatically starts on boot", nil},
		{"dev-utils", "Utilities to assist application development", nil},
	}
)

// Activate marks the checkbox selections based on the data model
func (bp *BundlePage) Activate() {
	model := bp.getModel()

	for _, curr := range bundles {
		state := 0

		if model.ContainsBundle(curr.name) {
			state = 1
		}

		curr.check.SetState(state)
	}
}

func newBundlePage(mi *Tui) (Page, error) {
	page := &BundlePage{}
	page.setupMenu(mi, TuiPageBundle, "Bundle Selection", NoButtons)

	clui.CreateLabel(page.content, 2, 2, "Select additional bundles to be added to the system", Fixed)

	frm := clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	frm.SetPack(clui.Vertical)

	lblFrm := clui.CreateFrame(frm, AutoSize, AutoSize, BorderNone, Fixed)
	lblFrm.SetPack(clui.Vertical)
	lblFrm.SetPaddings(2, 0)

	for _, curr := range bundles {
		lbl := fmt.Sprintf("%s: %s", curr.name, curr.desc)
		curr.check = clui.CreateCheckBox(lblFrm, AutoSize, lbl, AutoSize)
		curr.check.SetPack(clui.Horizontal)
	}

	fldFrm := clui.CreateFrame(frm, 30, AutoSize, BorderNone, Fixed)
	fldFrm.SetPack(clui.Vertical)

	cancelBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Cancel", Fixed)
	cancelBtn.OnClick(func(ev clui.Event) {
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	confirmBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Confirm", Fixed)
	confirmBtn.OnClick(func(ev clui.Event) {
		for _, curr := range bundles {
			if curr.check.State() == 1 {
				page.getModel().AddBundle(curr.name)
			} else {
				page.getModel().RemoveBundle(curr.name)
			}
		}

		page.SetDone(true)
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	return page, nil
}
