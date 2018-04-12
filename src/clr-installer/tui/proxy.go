package tui

import (
	"github.com/VladimirMarkelov/clui"
	"github.com/nsf/termbox-go"
)

// ProxyPage is the Page implementation for the proxy configuration page
type ProxyPage struct {
	BasePage
}

// DeActivate is called when the window is "hidden" and hides the cursor
func (page *ProxyPage) DeActivate() {
	// TODO clui is not hiding cursor for this case - fix it on clui side
	termbox.HideCursor()
}

func newProxyPage(mi *Tui) (Page, error) {
	page := &ProxyPage{}
	page.setupMenu(mi, TuiPageProxy, "Proxy", NoButtons)

	clui.CreateLabel(page.content, 2, 2, "Configure the network proxy", Fixed)

	frm := clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	frm.SetPack(clui.Horizontal)

	lblFrm := clui.CreateFrame(frm, 20, AutoSize, BorderNone, Fixed)
	lblFrm.SetPack(clui.Vertical)
	lblFrm.SetPaddings(1, 0)

	newFieldLabel(lblFrm, "HTTPS Proxy:")

	fldFrm := clui.CreateFrame(frm, 30, AutoSize, BorderNone, Fixed)
	fldFrm.SetPack(clui.Vertical)

	iframe := clui.CreateFrame(fldFrm, 5, 2, BorderNone, Fixed)
	iframe.SetPack(clui.Vertical)

	httpsProxyEdit := clui.CreateEditField(iframe, 1, "", Fixed)

	btnFrm := clui.CreateFrame(fldFrm, 30, 1, BorderNone, Fixed)
	btnFrm.SetPack(clui.Horizontal)
	btnFrm.SetGaps(1, 1)
	btnFrm.SetPaddings(2, 0)

	cancelBtn := CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Cancel", Fixed)
	cancelBtn.OnClick(func(ev clui.Event) {
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	confirmBtn := CreateSimpleButton(btnFrm, AutoSize, AutoSize, "Confirm", Fixed)
	confirmBtn.OnClick(func(ev clui.Event) {
		page.getModel().HTTPSProxy = httpsProxyEdit.Title()
		page.SetDone(true)
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	return page, nil
}
