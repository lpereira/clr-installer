package tui

import (
	"clr-installer/keyboard"
	"github.com/VladimirMarkelov/clui"
)

// KeyboardPage is the Page implementation for the keyboard configuration page
type KeyboardPage struct {
	BasePage
	avKeymaps  []*keyboard.Keymap
	kbdListBox *clui.ListBox
}

// SetDone sets the keyboard page flag done, and sets back the configuration to the data model
func (page *KeyboardPage) SetDone(done bool) bool {
	page.done = done
	page.getModel().Keyboard = page.kbdListBox.SelectedItemText()
	return true
}

func newKeyboardPage(mi *Tui) (Page, error) {
	kmaps, err := keyboard.LoadKeymaps()
	if err != nil {
		return nil, err
	}

	page := &KeyboardPage{
		avKeymaps: kmaps,
	}

	page.setupMenu(mi, TuiPageKeyboard, "Configure the keyboard", DoneButton)

	lbl := clui.CreateLabel(page.content, 2, 2, "Select Keyboard", Fixed)
	lbl.SetPaddings(0, 2)

	page.kbdListBox = clui.CreateListBox(page.content, AutoSize, 17, Fixed)

	defKeyboard := 0
	for idx, curr := range page.avKeymaps {
		page.kbdListBox.AddItem(curr.Code)

		if curr.IsDefault() {
			defKeyboard = idx
		}
	}

	page.kbdListBox.SelectItem(defKeyboard)
	page.activated = page.doneBtn

	return page, nil
}
