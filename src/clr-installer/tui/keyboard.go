package tui

import (
	"bytes"
	"strings"

	"clr-installer/cmd"
	"github.com/VladimirMarkelov/clui"
)

// KeyboardPage is the Page implementation for the keyboard configuration page
type KeyboardPage struct {
	BasePage
	avKeymaps  []string
	kbdListBox *clui.ListBox
}

func (page *KeyboardPage) initKeymaps() error {
	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, false, "localectl", "list-keymaps", "--no-pager")
	if err != nil {
		return err
	}

	tks := strings.Split(w.String(), "\n")
	for _, curr := range tks {
		if curr == "" {
			continue
		}
		page.avKeymaps = append(page.avKeymaps, curr)
	}

	return nil
}

// SetDone sets the keyboard page flag done, and sets back the configuration to the data model
func (page *KeyboardPage) SetDone(done bool) bool {
	page.done = done
	page.mi.model.Keyboard = page.kbdListBox.SelectedItemText()
	return true
}

func newKeyboardPage(mi *Tui) (Page, error) {
	page := &KeyboardPage{
		avKeymaps: []string{},
	}

	if err := page.initKeymaps(); err != nil {
		return nil, err
	}

	page.setupMenu(mi, TuiPageKeyboard, "Configure the keyboard", DoneButton)

	lbl := clui.CreateLabel(page.content, 2, 2, "Select Keyboard", Fixed)
	lbl.SetPaddings(0, 2)

	page.kbdListBox = clui.CreateListBox(page.content, AutoSize, 17, Fixed)

	defKeyboard := 0
	for idx, curr := range page.avKeymaps {
		page.kbdListBox.AddItem(curr)

		if curr == "us" {
			defKeyboard = idx
		}
	}

	page.kbdListBox.SelectItem(defKeyboard)
	page.activated = page.doneBtn

	return page, nil
}
