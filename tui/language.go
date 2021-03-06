// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package tui

import (
	"github.com/clearlinux/clr-installer/language"

	"github.com/VladimirMarkelov/clui"
)

// LanguagePage is the Page implementation for the language configuration page
type LanguagePage struct {
	BasePage
	avLanguages []*language.Language
	langListBox *clui.ListBox
}

// GetConfigDefinition returns if the config was interactively defined by the user,
// was loaded from a config file or if the config is not set.
func (page *LanguagePage) GetConfigDefinition() int {
	lang := page.getModel().Language

	if lang == nil {
		return ConfigNotDefined
	} else if lang.IsUserDefined() {
		return ConfigDefinedByUser
	}

	return ConfigDefinedByConfig
}

// SetDone sets the keyboard page flag done, and sets back the configuration to the data model
func (page *LanguagePage) SetDone(done bool) bool {
	page.done = done
	page.getModel().Language = page.avLanguages[page.langListBox.SelectedItem()]
	return true
}

// DeActivate will reset the selection case the user has pressed cancel
func (page *LanguagePage) DeActivate() {
	if page.action == ActionDoneButton {
		return
	}

	for idx, curr := range page.avLanguages {
		if !curr.Equals(page.getModel().Language) {
			continue
		}

		page.langListBox.SelectItem(idx)
		return
	}
}

func newLanguagePage(tui *Tui) (Page, error) {
	avLanguages, err := language.Load()
	if err != nil {
		return nil, err
	}

	page := &LanguagePage{
		avLanguages: avLanguages,
		BasePage: BasePage{
			// Tag this Page as required to be complete for the Install to proceed
			required: true,
		},
	}

	page.setupMenu(tui, TuiPageLanguage, "Choose language", DoneButton|CancelButton, TuiPageMenu)

	lbl := clui.CreateLabel(page.content, 2, 2, "Select System Language", Fixed)
	lbl.SetPaddings(0, 2)

	page.langListBox = clui.CreateListBox(page.content, AutoSize, 17, Fixed)

	defLanguage := 0
	for idx, curr := range page.avLanguages {
		page.langListBox.AddItem(curr.String())

		if curr.Equals(page.getModel().Language) {
			defLanguage = idx
		}
	}

	page.langListBox.SelectItem(defLanguage)
	page.activated = page.doneBtn

	return page, nil
}
