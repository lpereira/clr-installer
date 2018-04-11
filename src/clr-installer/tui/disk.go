package tui

import (
	"github.com/VladimirMarkelov/clui"
)

// DiskMenuPage is the Page implementation for the disk partitioning menu page
type DiskMenuPage struct {
	BasePage
}

const (
	diskDesc = `The installer can use a guided standard	disk partitioning scheme or, you
can manually set your own.`
)

// GetConfigDefinition returns if the config was interactively defined by the user,
// was loaded from a config file or if the config is not set.
func (page *DiskMenuPage) GetConfigDefinition() int {
	tm := page.getModel().TargetMedias

	if tm == nil {
		return ConfigNotDefined
	}

	for _, bd := range tm {
		if !bd.IsUserDefined() {
			return ConfigDefinedByConfig
		}

		for _, ch := range bd.Children {
			if !ch.IsUserDefined() {
				return ConfigDefinedByConfig
			}
		}
	}

	return ConfigDefinedByUser
}

// The disk page gives the user the option so select how to set the storage device,
// if to manually configure it or a guided standard partition schema
func newDiskPage(mi *Tui) (Page, error) {
	page := &DiskMenuPage{}
	page.setupMenu(mi, TuiPageDiskMenu, "Partition disks", BackButton)

	lbl := clui.CreateLabel(page.content, 2, 2, "Partition Disks", clui.Fixed)
	lbl.SetPaddings(0, 2)

	lbl = clui.CreateLabel(page.content, 70, 3, diskDesc, Fixed)
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
