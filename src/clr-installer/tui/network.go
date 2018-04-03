package tui

import (
	"fmt"

	"clr-installer/network"
	"github.com/VladimirMarkelov/clui"
)

// NetworkPage is the Page implementation for the network configuration page
type NetworkPage struct {
	BasePage
	frm        *clui.Frame
	btns       []*SimpleButton
	labels     []*clui.Label
	interfaces []*network.Interface
}

func (page *NetworkPage) showLabel(frm *clui.Frame, txt string) {
	label := clui.CreateLabel(frm, AutoSize, 1, txt, Fixed)
	page.labels = append(page.labels, label)
}

func (page *NetworkPage) showInterface(frm *clui.Frame, iface *network.Interface) {
	lbl := fmt.Sprintf(" interface: %s", iface.Name)

	btn := CreateSimpleButton(frm, AutoSize, 1, lbl, Fixed)
	btn.SetAlign(AlignLeft)

	btn.OnClick(func(ev clui.Event) {
		page.data = iface
		page.mi.gotoPage(TuiPageInterface, page.mi.currPage)
	})

	page.btns = append(page.btns, btn)

	for _, addr := range iface.Addrs {
		ipLabel := addr.VersionString()

		page.showLabel(frm, fmt.Sprintf("  %s:    %s", ipLabel, addr.IP))
		page.showLabel(frm, fmt.Sprintf("  netmask: %s", addr.NetMask))
	}

	if len(iface.Addrs) == 0 {
		page.showLabel(frm, fmt.Sprintf("  ipv4:    0.0.0.0"))
		page.showLabel(frm, fmt.Sprintf("  netmask: 0.0.0.0"))
	}
}

// Activate will recreate the network listing elements
func (page *NetworkPage) Activate() {
	var err error

	if page.interfaces == nil {
		page.interfaces, err = network.Interfaces()
		if err != nil {
			page.Panic(err)
		}
	}

	for _, curr := range page.btns {
		curr.Destroy()
	}
	page.btns = []*SimpleButton{}

	for _, curr := range page.labels {
		curr.Destroy()
	}
	page.labels = []*clui.Label{}

	for _, curr := range page.interfaces {
		page.showInterface(page.frm, curr)
	}
}

func newNetworkPage(mi *Tui) (Page, error) {
	page := &NetworkPage{}
	page.setupMenu(mi, TuiPageNetwork, "Configure network interfaces", BackButton)

	page.frm = clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	page.frm.SetPack(clui.Vertical)

	return page, nil
}
