package tui

import (
	"github.com/clearlinux/clr-installer/network"

	"github.com/VladimirMarkelov/clui"
	term "github.com/nsf/termbox-go"
)

// NetworkInterfacePage is the Page implementation for the network configuration page
type NetworkInterfacePage struct {
	BasePage
	IPEdit      *clui.EditField
	NetMaskEdit *clui.EditField
	GatewayEdit *clui.EditField
	DNSEdit     *clui.EditField
	ifaceLbl    *clui.Label
	DHCPCheck   *clui.CheckBox

	defaultValues struct {
		IP      string
		NetMask string
		Gateway string
		DNS     string
		DHCP    bool
	}
}

func (page *NetworkInterfacePage) getSelectedInterface() *network.Interface {
	var iface *network.Interface
	var ok bool

	prevPage := page.mi.getPage(TuiPageNetwork)
	if iface, ok = prevPage.GetData().(*network.Interface); !ok {
		return nil
	}

	return iface
}

// Activate will set the fields with the selected interface info
func (page *NetworkInterfacePage) Activate() {
	sel := page.getSelectedInterface()

	page.ifaceLbl.SetTitle(sel.Name)
	page.IPEdit.SetTitle("")
	page.NetMaskEdit.SetTitle("")
	page.GatewayEdit.SetTitle(sel.Gateway)
	page.DNSEdit.SetTitle(sel.DNS)

	page.defaultValues.Gateway = sel.Gateway
	page.defaultValues.DNS = sel.DNS
	page.defaultValues.DHCP = sel.DHCP

	showIPv4 := sel.HasIPv4Addr()
	for _, addr := range sel.Addrs {
		if showIPv4 && addr.Version != network.IPv4 {
			continue
		}

		page.IPEdit.SetTitle(addr.IP)
		page.NetMaskEdit.SetTitle(addr.NetMask)

		page.defaultValues.IP = addr.IP
		page.defaultValues.NetMask = addr.NetMask
		break
	}

	page.setDHCP(sel.DHCP)
}

func (page *NetworkInterfacePage) getDHCP() bool {
	state := page.DHCPCheck.State()
	if state == 1 {
		return true
	}

	return false
}

func (page *NetworkInterfacePage) setDHCP(DHCP bool) {
	state := 0

	if DHCP {
		state = 1
	}

	page.DHCPCheck.SetState(state)
}

func newFieldLabel(frame *clui.Frame, text string) *clui.Label {
	lbl := clui.CreateLabel(frame, AutoSize, 2, text, Fixed)
	lbl.SetAlign(AlignRight)
	return lbl
}

func validateIPEdit(k term.Key, ch rune) bool {
	validKeys := []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.'}

	if k == term.KeyBackspace || k == term.KeyBackspace2 {
		return false
	}

	for _, curr := range validKeys {
		if curr == ch {
			return false
		}
	}

	return true
}

func newNetworkInterfacePage(mi *Tui) (Page, error) {
	page := &NetworkInterfacePage{}
	page.setup(mi, TuiPageInterface, NoButtons)

	frm := clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	frm.SetPack(clui.Horizontal)

	lblFrm := clui.CreateFrame(frm, 20, AutoSize, BorderNone, Fixed)
	lblFrm.SetPack(clui.Vertical)
	lblFrm.SetPaddings(1, 0)

	newFieldLabel(lblFrm, "Interface:")
	newFieldLabel(lblFrm, "Ip address:")
	newFieldLabel(lblFrm, "Subnet mask:")
	newFieldLabel(lblFrm, "Gateway:")
	newFieldLabel(lblFrm, "DNS:")

	fldFrm := clui.CreateFrame(frm, 30, AutoSize, BorderNone, Fixed)
	fldFrm.SetPack(clui.Vertical)

	ifaceFrm := clui.CreateFrame(fldFrm, 5, 2, BorderNone, Fixed)
	ifaceFrm.SetPack(clui.Vertical)

	page.ifaceLbl = clui.CreateLabel(ifaceFrm, AutoSize, 2, "", Fixed)
	page.ifaceLbl.SetAlign(AlignLeft)

	page.IPEdit = newEditField(fldFrm, validateIPEdit)
	page.NetMaskEdit = newEditField(fldFrm, validateIPEdit)
	page.GatewayEdit = newEditField(fldFrm, validateIPEdit)
	page.DNSEdit = newEditField(fldFrm, validateIPEdit)

	dhcpFrm := clui.CreateFrame(fldFrm, 5, 2, BorderNone, Fixed)
	dhcpFrm.SetPack(clui.Vertical)

	page.DHCPCheck = clui.CreateCheckBox(dhcpFrm, 1, "Automatic/dhcp", Fixed)

	page.DHCPCheck.OnChange(func(ev int) {
		enable := true

		if ev == 1 {
			enable = false
		}

		page.IPEdit.SetEnabled(enable)
		page.NetMaskEdit.SetEnabled(enable)
		page.GatewayEdit.SetEnabled(enable)
		page.DNSEdit.SetEnabled(enable)
	})

	btnFrm := clui.CreateFrame(fldFrm, 30, 1, BorderNone, Fixed)
	btnFrm.SetPack(clui.Horizontal)
	btnFrm.SetGaps(1, 1)
	btnFrm.SetPaddings(2, 0)

	cancelBtn := CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Cancel", Fixed)
	cancelBtn.OnClick(func(ev clui.Event) {
		mi.gotoPage(TuiPageNetwork, mi.currPage)
	})

	confirmBtn := CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Confirm", Fixed)
	confirmBtn.OnClick(func(ev clui.Event) {

		IP := page.IPEdit.Title()
		NetMask := page.NetMaskEdit.Title()
		DHCP := page.getDHCP()
		Gateway := page.GatewayEdit.Title()
		DNS := page.DNSEdit.Title()
		changed := false

		if IP != page.defaultValues.IP {
			changed = true
		}

		if NetMask != page.defaultValues.NetMask {
			changed = true
		}

		if DHCP != page.defaultValues.DHCP {
			changed = true
		}

		if Gateway != page.defaultValues.Gateway {
			changed = true
		}

		if DNS != page.defaultValues.DNS {
			changed = true
		}

		if changed {
			sel := page.getSelectedInterface()
			if !sel.HasIPv4Addr() {
				sel.AddAddr(IP, NetMask, network.IPv4)
			} else {
				for _, addr := range sel.Addrs {
					if addr.Version != network.IPv4 {
						continue
					}

					addr.IP = IP
					addr.NetMask = NetMask
					break
				}
			}

			sel.DHCP = DHCP
			sel.Gateway = Gateway
			sel.DNS = DNS
			page.getModel().AddNetworkInterface(sel)
		}

		mi.gotoPage(TuiPageNetwork, mi.currPage)
		page.mi.getPage(TuiPageNetwork).SetDone(true)
	})

	return page, nil
}
