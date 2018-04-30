package tui

import (
	"github.com/clearlinux/clr-installer/storage"

	"github.com/VladimirMarkelov/clui"
)

// DiskPartitionPage is the Page implementation for partition configuration page
type DiskPartitionPage struct {
	BasePage
	fsList      *clui.ListBox
	mPointEdit  *clui.EditField
	sizeEdit    *clui.EditField
	cancelBtn   *SimpleButton
	deleteBtn   *SimpleButton
	addBtn      *SimpleButton
	confirmBtn  *SimpleButton
	sizeWarning *clui.Label
}

const (
	nPartitionHelp = "Set the new partition's file system, mount point and size."

	// partDeleteBtn mask defines a partition configuration page will have a delete button
	partDeleteBtn = 1 << 1

	// partAddBtn mask defines a partition configuration page will have an add button
	partAddBtn = 1 << 2

	// partConfirmBtn mask defines a partition configuration page will have a confirm button
	partConfirmBtn = 1 << 3

	// partAllBtns mask defines a partition configuration page will have show both:
	// delete, add and confirm buttons
	partAllBtns = partDeleteBtn | partAddBtn | partConfirmBtn
)

func (page *DiskPartitionPage) setPartitionButtonsVisible(visible bool, mask int) {
	if mask&partDeleteBtn == partDeleteBtn {
		page.deleteBtn.SetVisible(visible)
	}

	if mask&partAddBtn == partAddBtn {
		page.addBtn.SetVisible(visible)
	}

	if mask&partConfirmBtn == partConfirmBtn {
		page.confirmBtn.SetVisible(visible)
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
	}

	size, err := part.HumanReadableSize()
	if err != nil {
		page.Panic(err)
	}

	page.sizeEdit.SetTitle(size)

	page.setPartitionButtonsVisible(false, partAllBtns)
	page.setPartitionButtonsVisible(true, partDeleteBtn|partConfirmBtn)
}

func (page *DiskPartitionPage) getSelectedBlockDevice() *SelectedBlockDevice {
	var sel *SelectedBlockDevice
	var ok bool

	prevPage := page.mi.getPage(TuiPageManualPart)
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
	page.sizeEdit.SetTitle("")
	page.sizeWarning.SetTitle("")

	if sel.part != nil {
		page.setPartitionForm(sel.part)
	} else if sel.bd != nil && sel.freeSpace != 0 {
		page.setPartitionButtonsVisible(false, partAllBtns)
		page.setPartitionButtonsVisible(true, partAddBtn)
	}
}

func newDiskPartitionPage(mi *Tui) (Page, error) {
	page := &DiskPartitionPage{}

	page.setup(mi, TuiPageDiskPart, NoButtons)

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

	lbl = clui.CreateLabel(lblFrm, AutoSize, 2, "Mount Point:", Fixed)
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

	page.fsList.OnSelectItem(func(evt clui.Event) {
		page.mPointEdit.SetEnabled(true)

		if page.fsList.SelectedItemText() == "swap" {
			page.mPointEdit.SetEnabled(false)
		}
	})

	sizeFrm := clui.CreateFrame(fldFrm, 5, AutoSize, BorderNone, Fixed)
	sizeFrm.SetPack(clui.Vertical)

	page.sizeEdit = clui.CreateEditField(sizeFrm, 3, "", Fixed)

	page.sizeWarning = clui.CreateLabel(sizeFrm, 1, 1, "", Fixed)
	page.sizeWarning.SetBackColor(errorLabelBg)
	page.sizeWarning.SetTextColor(errorLabelFg)

	btnFrm := clui.CreateFrame(fldFrm, 30, 1, BorderNone, Fixed)
	btnFrm.SetPack(clui.Horizontal)
	btnFrm.SetGaps(1, 1)
	btnFrm.SetPaddings(2, 0)

	page.cancelBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Cancel", Fixed)
	page.cancelBtn.OnClick(func(ev clui.Event) {
		mi.gotoPage(TuiPageManualPart, mi.currPage)
	})

	page.deleteBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Delete", Fixed)
	page.deleteBtn.OnClick(func(ev clui.Event) {
		sel := page.getSelectedBlockDevice()
		sel.bd.RemoveChild(sel.part)
		mi.gotoPage(TuiPageManualPart, mi.currPage)
	})

	page.addBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Add", Fixed)

	page.addBtn.OnClick(func(ev clui.Event) {
		sel := page.getSelectedBlockDevice()

		size, err := storage.ParseVolumeSize(page.sizeEdit.Title())
		if err != nil {
			page.sizeWarning.SetTitle("Invalid size")
			return
		}

		part := &storage.BlockDevice{
			FsType:     page.fsList.SelectedItemText(),
			MountPoint: page.mPointEdit.Title(),
			Size:       size,
		}

		sel.bd.AddChild(part)
		mi.gotoPage(TuiPageManualPart, mi.currPage)
	})

	page.confirmBtn = CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Confirm", Fixed)
	page.confirmBtn.OnClick(func(ev clui.Event) {
		sel := page.getSelectedBlockDevice()

		if sel.part != nil {
			sel.part.FsType = page.fsList.SelectedItemText()
			sel.part.MountPoint = page.mPointEdit.Title()
		}

		mi.gotoPage(TuiPageManualPart, mi.currPage)
	})

	page.activated = page.fsList
	return page, nil
}
