package tui

import (
	"fmt"

	"github.com/VladimirMarkelov/clui"
	"github.com/clearlinux/clr-installer/swupd"
)

// BundlePage is the Page implementation for the proxy configuration page
type BundlePage struct {
	BasePage
}

// BundleCheck maps a map name and description with the actual checkbox
type BundleCheck struct {
	bundle *swupd.Bundle
	check  *clui.CheckBox
}

var (
	bundles = []*BundleCheck{}
)

// Activate marks the checkbox selections based on the data model
func (bp *BundlePage) Activate() {
	model := bp.getModel()

	for _, curr := range bundles {
		state := 0

		if model.ContainsBundle(curr.bundle.Name) {
			state = 1
		}

		curr.check.SetState(state)
	}
}

func newBundlePage(mi *Tui) (Page, error) {
	bdls, err := swupd.LoadBundleList()
	if err != nil {
		return nil, err
	}

	for _, curr := range bdls {
		bundles = append(bundles, &BundleCheck{curr, nil})
	}

	page := &BundlePage{}
	page.setupMenu(mi, TuiPageBundle, "Bundle Selection", NoButtons)

	clui.CreateLabel(page.content, 2, 2, "Select additional bundles to be added to the system", Fixed)

	frm := clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	frm.SetPack(clui.Vertical)

	lblFrm := clui.CreateFrame(frm, AutoSize, AutoSize, BorderNone, Fixed)
	lblFrm.SetPack(clui.Vertical)
	lblFrm.SetPaddings(2, 0)

	for _, curr := range bundles {
		lbl := fmt.Sprintf("%s: %s", curr.bundle.Name, curr.bundle.Desc)
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
				page.getModel().AddBundle(curr.bundle.Name)
			} else {
				page.getModel().RemoveBundle(curr.bundle.Name)
			}
		}

		page.SetDone(true)
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	return page, nil
}
