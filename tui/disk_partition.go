// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package tui

import (
	"fmt"

	"github.com/clearlinux/clr-installer/storage"

	"github.com/VladimirMarkelov/clui"

	term "github.com/nsf/termbox-go"
)

// DiskPartitionPage is the Page implementation for partition configuration page
type DiskPartitionPage struct {
	BasePage
	fsList        *clui.ListBox
	mPointEdit    *clui.EditField
	mPointWarning *clui.Label
	sizeEdit      *clui.EditField
	confirmBtn    *SimpleButton
	deleteBtn     *SimpleButton
	cancelBtn     *SimpleButton
	sizeWarning   *clui.Label
	sizeInfo      *clui.Label
}

const (
	nPartitionHelp = "Set the new partition's file system, mount point and size."

	// partConfirmBtn mask defines a partition configuration page will have a confirm button
	partConfirmBtn = 1 << 1

	// partDeleteBtn mask defines a partition configuration page will have a delete button
	partDeleteBtn = 1 << 2

	// partCancelBtn mask defines a partition configuration page will have a cancel button
	partCancelBtn = 1 << 3

	// partAllBtns mask defines a partition configuration page will have show both:
	// delete, add and confirm buttons
	partAllBtns = partConfirmBtn | partDeleteBtn | partCancelBtn
)

func (page *DiskPartitionPage) setPartitionButtonsVisible(visible bool, mask int) {
	if mask&partConfirmBtn == partConfirmBtn {
		page.confirmBtn.SetVisible(visible)
		page.setConfirmButton()
	}

	if mask&partDeleteBtn == partDeleteBtn {
		page.deleteBtn.SetVisible(visible)
	}

	if mask&partCancelBtn == partCancelBtn {
		page.cancelBtn.SetVisible(visible)
	}
}

func (page *DiskPartitionPage) setPartitionForm(part *storage.BlockDevice) {
	idx := page.fsList.FindItem(part.FsType, true)
	page.fsList.SelectItem(idx)

	page.mPointEdit.SetEnabled(true)
	if part.FsType == "swap" {
		page.mPointEdit.SetEnabled(false)
	} else {
		page.mPointEdit.SetTitle(part.MountPoint)
		page.validateMountPoint()
	}

	size, err := part.HumanReadableSize()
	if err != nil {
		page.Panic(err)
	}

	page.sizeEdit.SetTitle(size)

	page.setPartitionButtonsVisible(true, partAllBtns)
}

func (page *DiskPartitionPage) getSelectedBlockDevice() *SelectedBlockDevice {
	var sel *SelectedBlockDevice
	var ok bool

	prevPage := page.tui.getPage(TuiPageManualPart)
	if sel, ok = prevPage.GetData().(*SelectedBlockDevice); !ok {
		return nil
	}

	return sel
}

// Activate is called when the window is "shown", this implementation adjusts
// the currently displayed data
func (page *DiskPartitionPage) Activate() {
	sel := page.getSelectedBlockDevice()

	if sel == nil {
		return
	}

	page.mPointEdit.SetTitle("")
	page.mPointWarning.SetTitle("")
	page.sizeEdit.SetTitle("")
	page.sizeInfo.SetTitle("'+/=' to force Maximum size")
	page.sizeWarning.SetTitle("")

	page.setPartitionForm(sel.part)

	if sel.addMode {
		page.setPartitionButtonsVisible(false, partCancelBtn)
		// In Add partition mode, the Delete button is really
		// our "Cancel" as the new partition was already added.
		page.deleteBtn.SetTitle("Cancel")
		// and the Confirm button is really our "Add" button
		page.confirmBtn.SetTitle("Add")
	} else {
		page.setPartitionButtonsVisible(true, partCancelBtn)
		page.deleteBtn.SetTitle("Delete")
		page.confirmBtn.SetTitle("Confirm")
	}
}

func (page *DiskPartitionPage) setConfirmButton() {
	if page.mPointWarning.Title() == "" && page.sizeWarning.Title() == "" {
		page.confirmBtn.SetEnabled(true)
	} else {
		page.confirmBtn.SetEnabled(false)
	}
}

func (page *DiskPartitionPage) validateMountPoint() {
	page.mPointWarning.SetTitle(storage.IsValidMount(page.mPointEdit.Title()))

	page.setConfirmButton()
}

func newDiskPartitionPage(tui *Tui) (Page, error) {
	page := &DiskPartitionPage{}

	page.setup(tui, TuiPageDiskPart, NoButtons, TuiPageMenu)

	lbl := clui.CreateLabel(page.content, 2, 2, "Partition Setup", Fixed)
	lbl.SetPaddings(0, 2)

	clui.CreateLabel(page.content, 2, 2, nPartitionHelp, Fixed)

	frm := clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	frm.SetPack(clui.Horizontal)

	lblFrm := clui.CreateFrame(frm, 20, AutoSize, BorderNone, Fixed)
	lblFrm.SetPack(clui.Vertical)
	lblFrm.SetPaddings(1, 0)

	lbl = clui.CreateLabel(lblFrm, AutoSize, 3, "File System:", Fixed)
	lbl.SetAlign(AlignRight)

	lbl = clui.CreateLabel(lblFrm, AutoSize, 3, "Mount Point:", Fixed)
	lbl.SetAlign(AlignRight)

	lbl = clui.CreateLabel(lblFrm, AutoSize, 2, "Size:", Fixed)
	lbl.SetAlign(AlignRight)

	fldFrm := clui.CreateFrame(frm, 30, AutoSize, BorderNone, Fixed)
	fldFrm.SetPack(clui.Vertical)

	page.fsList = clui.CreateListBox(fldFrm, 1, 2, Fixed)
	page.fsList.SetAlign(AlignLeft)

	for _, fs := range storage.SupportedFileSystems() {
		page.fsList.AddItem(fs)
	}
	page.fsList.SelectItem(0)

	mPointFrm := clui.CreateFrame(fldFrm, 4, AutoSize, BorderNone, Fixed)
	mPointFrm.SetPack(clui.Vertical)
	mPointFrm.SetPaddings(0, 1)

	page.mPointEdit = clui.CreateEditField(mPointFrm, 3, "", Fixed)
	page.mPointEdit.OnChange(func(ev clui.Event) {
		page.validateMountPoint()
	})

	page.mPointWarning = clui.CreateLabel(mPointFrm, 1, 1, "", Fixed)
	page.mPointWarning.SetMultiline(true)
	page.mPointWarning.SetBackColor(errorLabelBg)
	page.mPointWarning.SetTextColor(errorLabelFg)

	page.fsList.OnSelectItem(func(evt clui.Event) {
		page.mPointEdit.SetEnabled(true)

		if page.fsList.SelectedItemText() == "swap" {
			page.mPointEdit.SetEnabled(false)
			page.mPointEdit.SetTitle("")
			page.mPointWarning.SetTitle("")
		}
	})

	sizeFrm := clui.CreateFrame(fldFrm, 5, AutoSize, BorderNone, Fixed)
	sizeFrm.SetPack(clui.Vertical)

	page.sizeEdit = clui.CreateEditField(sizeFrm, 3, "", Fixed)
	page.sizeEdit.OnChange(func(ev clui.Event) {
		sel := page.getSelectedBlockDevice()
		page.sizeWarning.SetTitle(sel.part.IsValidSize(page.sizeEdit.Title()))

		page.setConfirmButton()
	})
	page.sizeEdit.OnKeyPress(func(k term.Key, ch rune) bool {
		maxSizeKeys := []rune{'=', '+'}
		for _, curr := range maxSizeKeys {
			if curr == ch {
				sel := page.getSelectedBlockDevice()
				page.sizeEdit.SetTitle(fmt.Sprintf("%v", sel.part.MaxParitionSize()))
				return true
			}
		}

		return false
	})

	page.sizeInfo = clui.CreateLabel(sizeFrm, 1, 1, "", Fixed)
	page.sizeInfo.SetMultiline(false)
	page.sizeWarning = clui.CreateLabel(sizeFrm, 1, 3, "", Fixed)
	page.sizeWarning.SetMultiline(true)
	page.sizeWarning.SetBackColor(errorLabelBg)
	page.sizeWarning.SetTextColor(errorLabelFg)

	btnFrm := clui.CreateFrame(fldFrm, 30, 1, BorderNone, Fixed)
	btnFrm.SetPack(clui.Horizontal)
	btnFrm.SetGaps(1, 1)
	btnFrm.SetPaddings(2, 0)

	page.confirmBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Confirm", Fixed)
	page.confirmBtn.OnClick(func(ev clui.Event) {
		sel := page.getSelectedBlockDevice()

		if sel.part != nil {
			sel.part.FsType = page.fsList.SelectedItemText()
			sel.part.MountPoint = page.mPointEdit.Title()
			size, err := storage.ParseVolumeSize(page.sizeEdit.Title())
			if err == nil {
				sel.part.Size = size
			}
		}

		page.GotoPage(TuiPageManualPart)
	})

	page.deleteBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Delete", Fixed)
	page.deleteBtn.OnClick(func(ev clui.Event) {
		sel := page.getSelectedBlockDevice()
		sel.bd.RemoveChild(sel.part)

		page.GotoPage(TuiPageManualPart)
	})

	page.cancelBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Cancel", Fixed)
	page.cancelBtn.OnClick(func(ev clui.Event) {
		page.GotoPage(TuiPageManualPart)
	})

	page.activated = page.fsList
	return page, nil
}
