package tui

import (
	"clr-installer/keyboard"

	"github.com/VladimirMarkelov/clui"
	"github.com/nsf/termbox-go"
)

// KeyboardPage is the Page implementation for the keyboard configuration page
type KeyboardPage struct {
	BasePage
	avKeymaps  []*keyboard.Keymap
	kbdListBox *clui.ListBox
}

// DeActivate is called when the window is "hidden" and hides the cursor
func (page *KeyboardPage) DeActivate() {
	// TODO clui is not hiding cursor for this case - fix it on clui side
	termbox.HideCursor()
}

// GetConfigDefinition returns if the config was interactively defined by the user,
// was loaded from a config file or if the config is not set.
func (page *KeyboardPage) GetConfigDefinition() int {
	kbd := page.getModel().Keyboard

	if kbd == nil {
		return ConfigNotDefined
	} else if kbd.IsUserDefined() {
		return ConfigDefinedByUser
	}

	return ConfigDefinedByConfig
}

// SetDone sets the keyboard page flag done, and sets back the configuration to the data model
func (page *KeyboardPage) SetDone(done bool) bool {
	page.done = done
	page.getModel().Keyboard = page.avKeymaps[page.kbdListBox.SelectedItem()]
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

	page.kbdListBox = clui.CreateListBox(page.content, AutoSize, 10, Fixed)

	defKeyboard := 0
	for idx, curr := range page.avKeymaps {
		page.kbdListBox.AddItem(curr.Code)

		if curr.Equals(page.getModel().Keyboard) {
			defKeyboard = idx
		}
	}
	page.kbdListBox.SelectItem(defKeyboard)
	page.kbdListBox.OnActive(func(active bool) {
		if active {
			return
		}

		idx := page.kbdListBox.SelectedItem()
		selected := page.avKeymaps[idx]

		if page.getModel().Language.Code == selected.Code {
			return
		}

		if err := keyboard.Apply(selected); err != nil {
			page.Panic(err)
		}
	})

	frame := clui.CreateFrame(page.content, AutoSize, 7, BorderNone, Fixed)
	frame.SetPack(clui.Vertical)
	frame.SetPaddings(0, 1)

	clui.CreateLabel(frame, AutoSize, 1, "Test our keyboard", Fixed)
	newEditField(frame, nil)

	page.activated = page.doneBtn

	return page, nil
}
