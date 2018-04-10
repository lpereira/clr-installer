package tui

import (
	"fmt"

	"clr-installer/storage"
	"github.com/VladimirMarkelov/clui"
	"github.com/nsf/termbox-go"
)

// ManualPartPage is the Page implementation for manual partitioning page
type ManualPartPage struct {
	BasePage
	bds  []*storage.BlockDevice
	btns []*SimpleButton
}

// SelectedBlockDevice holds the shared date between the manual partitioning page and
// the partition configuration page
type SelectedBlockDevice struct {
	bd        *storage.BlockDevice
	part      *storage.BlockDevice
	freeSpace uint64
}

const (
	manualDesc = `Select a partition to modify its configuration and to define it as the
target instalattion disk.`
)

var (
	partBtnBg termbox.Attribute
)

func (page *ManualPartPage) showManualDisk(bd *storage.BlockDevice, frame *clui.Frame) error {
	size, err := bd.HumanReadableSize()
	if err != nil {
		return err
	}

	mm := fmt.Sprintf("(%s)", bd.MajorMinor)
	lbl := fmt.Sprintf("%s %s %s %s", bd.Model, bd.Name, mm, size)

	btn := CreateSimpleButton(frame, AutoSize, AutoSize, lbl, Fixed)
	btn.SetAlign(AlignLeft)

	page.btns = append(page.btns, btn)

	for _, part := range bd.Children {
		sel := &SelectedBlockDevice{bd: bd, part: part}

		size, err = sel.part.HumanReadableSize()
		if err != nil {
			return err
		}

		txt := fmt.Sprintf("%10s %10s %s %s", sel.part.Name, size, sel.part.FsType,
			sel.part.MountPoint)

		btn = page.newPartBtn(frame, txt)
		btn.OnClick(func(ev clui.Event) {
			page.data = sel
			page.mi.gotoPage(TuiPageDiskPart, page.mi.currPage)
		})
	}

	freeSpace, err := bd.FreeSpace()
	if err != nil {
		return err
	}

	freeSpaceLbl, err := storage.HumanReadableSize(freeSpace)
	if err != nil {
		return err
	}

	btn = page.newPartBtn(frame, fmt.Sprintf("%16s: %s", "Free space", freeSpaceLbl))
	btn.OnClick(func(ev clui.Event) {
		page.data = &SelectedBlockDevice{bd: bd, freeSpace: freeSpace}
		page.mi.gotoPage(TuiPageDiskPart, page.mi.currPage)
	})

	return nil
}

func (page *ManualPartPage) newPartBtn(frame *clui.Frame, label string) *SimpleButton {
	btn := CreateSimpleButton(frame, AutoSize, AutoSize, label, Fixed)
	btn.SetAlign(AlignLeft)
	btn.SetBackColor(partBtnBg)

	page.btns = append(page.btns, btn)
	return btn
}

func (page *ManualPartPage) showManualStorageList() error {
	for _, bd := range page.bds {
		if bd.State == storage.BlockDeviceStateLive {
			continue
		}

		if err := page.showManualDisk(bd, page.content); err != nil {
			return err
		}
	}

	return nil
}

// Activate is called when the manual disk partitioning page is activated and resets the
// page's displayed data
func (page *ManualPartPage) Activate() {
	if page.data == nil {
		return
	}

	for _, curr := range page.btns {
		curr.Destroy()
	}

	if err := page.showManualStorageList(); err != nil {
		page.Panic(err)
	}
	page.data = nil

	for _, bd := range page.bds {
		if err := bd.Validate(); err == nil {
			page.doneBtn.SetEnabled(true)
		}
	}
}

// SetDone set's the configured disk into the model and sets the previous page
// as done
func (page *ManualPartPage) SetDone(done bool) bool {
	if sel, ok := page.data.(*SelectedBlockDevice); ok {
		page.getModel().AddTargetMedia(sel.bd)
	}

	diskPage := page.mi.getPage(TuiPageDiskMenu)
	diskPage.SetDone(done)
	page.mi.gotoPage(TuiPageMenu, diskPage)
	return false
}

func newManualPartitionPage(mi *Tui) (Page, error) {
	var err error

	partBtnBg = clui.RealColor(clui.ColorDefault, "ManualPartitionBack")

	page := &ManualPartPage{}
	page.setup(mi, TuiPageManualPart, AllButtons)

	lbl := clui.CreateLabel(page.content, 2, 2, "Manual Partition", Fixed)
	lbl.SetPaddings(0, 2)

	lbl = clui.CreateLabel(page.content, 70, 3, manualDesc, Fixed)
	lbl.SetMultiline(true)

	page.bds, err = storage.ListBlockDevices(page.getModel().TargetMedias)
	if err != nil {
		return nil, err
	}

	if err = page.showManualStorageList(); err != nil {
		return nil, err
	}

	page.doneBtn.SetEnabled(false)
	return page, nil
}
