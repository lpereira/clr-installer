package tui

import (
	"fmt"

	"github.com/VladimirMarkelov/clui"
)

// MenuPage is the Page implementation for the main menu page
type MenuPage struct {
	BasePage
	btns       []*SimpleButton
	installBtn *SimpleButton
}

func (page *MenuPage) addMenuItem(item Page, activated bool) {
	done := "[ ]"

	if item.GetDone() {
		done = "[+]"
	} else if item.GetConfigDefinition() == ConfigDefinedByConfig {
		done = "[*]"
	}

	title := fmt.Sprintf(" %s %s", done, item.GetMenuTitle())
	btn := CreateSimpleButton(page.content, AutoSize, AutoSize, title, Fixed)
	btn.SetStyle("Menu")
	btn.SetAlign(AlignLeft)

	btn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(item.GetID(), page.mi.currPage)
	})

	page.btns = append(page.btns, btn)

	if activated {
		page.activated = btn
	}
}

// Activate is called when the page is "shown" and it repaints the main menu based on the
// available menu pages and their done/undone status
func (page *MenuPage) Activate() {
	for _, curr := range page.btns {
		curr.Destroy()
	}
	page.btns = []*SimpleButton{}

	previous := false
	for idx, curr := range page.mi.pages {
		activated := false

		if curr.GetMenuTitle() == "" {
			continue
		}

		if page.mi.prevPage == nil {
			if idx == 0 {
				activated = true
			}
		} else {
			if page.mi.prevPage.GetID() == curr.GetID() {
				previous = true
			} else if previous {
				activated = true
			}
		}

		page.addMenuItem(curr, activated)

		if previous && activated {
			previous = false
		}
	}

	if page.getModel() != nil && page.getModel().Validate() == nil {
		page.installBtn.SetEnabled(true)
	}
}

const (
	menuHelp = `Choose the next steps. Use <Tab> or arrow keys (up and down) to navigate
between the elements.
`
)

func newMenuPage(mi *Tui) (Page, error) {
	page := &MenuPage{}
	page.setup(mi, TuiPageMenu, NoButtons)

	lbl := clui.CreateLabel(page.content, 2, 3, menuHelp, Fixed)
	lbl.SetMultiline(true)
	lbl.SetPaddings(0, 2)

	cancelBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Cancel", Fixed)
	cancelBtn.OnClick(func(ev clui.Event) {
		go clui.Stop()
	})

	saveBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Save & Exit", Fixed)
	saveBtn.OnClick(func(ev clui.Event) {
		if err := page.getModel().WriteFile(descFile); err != nil {
			page.Panic(err)
		}
		go clui.Stop()
	})

	page.installBtn = CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Install", Fixed)
	page.installBtn.OnClick(func(ev clui.Event) {
		page.mi.gotoPage(TuiPageInstall, page.mi.currPage)
	})

	page.installBtn.SetEnabled(false)

	return page, nil
}
