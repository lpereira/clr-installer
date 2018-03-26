package tui

import (
	"clr-installer/language"
	"github.com/VladimirMarkelov/clui"
)

// LanguagePage is the Page implementation for the language configuration page
type LanguagePage struct {
	BasePage
	avLanguages []*language.Language
	langListBox *clui.ListBox
}

// SetDone sets the keyboard page flag done, and sets back the configuration to the data model
func (page *LanguagePage) SetDone(done bool) bool {
	page.done = done
	page.mi.model.Language = page.avLanguages[page.langListBox.SelectedItem()].Code
	return true
}

func newLanguagePage(mi *Tui) (Page, error) {
	avLanguages, err := language.Load()
	if err != nil {
		return nil, err
	}

	page := &LanguagePage{
		avLanguages: avLanguages,
	}

	page.setupMenu(mi, TuiPageLanguage, "Choose language", DoneButton)

	lbl := clui.CreateLabel(page.content, 2, 2, "Select System Language", Fixed)
	lbl.SetPaddings(0, 2)

	page.langListBox = clui.CreateListBox(page.content, AutoSize, 17, Fixed)

	defLanguage := 0
	for idx, curr := range page.avLanguages {
		page.langListBox.AddItem(curr.String())

		if curr.IsDefault() {
			defLanguage = idx
		}
	}

	page.langListBox.SelectItem(defLanguage)
	page.activated = page.doneBtn

	return page, nil
}
