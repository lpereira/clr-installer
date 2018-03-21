package tui

// NetworkPage is the Page implementation for the network configuration page
type NetworkPage struct {
	BasePage
}

func newNetworkPage(mi *Tui) (Page, error) {
	page := &NetworkPage{}
	page.setupMenu(mi, TuiPageNetwork, "Configure the network", BackButton)
	return page, nil
}
