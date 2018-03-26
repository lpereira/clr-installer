package tui

import (
	"bytes"
	"strings"

	"clr-installer/cmd"
	"github.com/VladimirMarkelov/clui"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// LanguagePage is the Page implementation for the language configuration page
type LanguagePage struct {
	BasePage
	avLanguages []language.Tag
	langListBox *clui.ListBox
}

func (page *LanguagePage) initLanguages() error {
	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, "localectl", "list-locales", "--no-pager")
	if err != nil {
		return err
	}

	tks := strings.Split(w.String(), "\n")
	for _, curr := range tks {
		if curr == "" {
			continue
		}
		code := strings.Replace(curr, ".UTF-8", "", 1)
		page.avLanguages = append(page.avLanguages, language.MustParse(code))
	}

	return nil
}

// SetDone sets the keyboard page flag done, and sets back the configuration to the data model
func (page *LanguagePage) SetDone(done bool) bool {
	page.done = done
	page.mi.model.Language = page.langListBox.SelectedItemText()
	return true
}

func newLanguagePage(mi *Tui) (Page, error) {
	page := &LanguagePage{
		avLanguages: []language.Tag{},
	}

	if err := page.initLanguages(); err != nil {
		return nil, err
	}

	page.setupMenu(mi, TuiPageLanguage, "Choose language", DoneButton)

	lbl := clui.CreateLabel(page.content, 2, 2, "Select System Language", Fixed)
	lbl.SetPaddings(0, 2)

	page.langListBox = clui.CreateListBox(page.content, AutoSize, 17, Fixed)

	defLanguage := 0
	for idx, curr := range page.avLanguages {
		page.langListBox.AddItem(display.English.Tags().Name(curr))

		if curr == language.AmericanEnglish {
			defLanguage = idx
		}
	}

	page.langListBox.SelectItem(defLanguage)
	page.activated = page.doneBtn

	return page, nil
}
