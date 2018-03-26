package keyboard

import (
	"bytes"
	"strings"

	"clr-installer/cmd"
)

// Keymap represents a system' keymap
type Keymap struct {
	Code string
}

// IsDefault returns true if a given keymap is the default one
func (l *Keymap) IsDefault() bool {
	return l.Code == "us"
}

// LoadKeymaps loads the system's available keymaps
func LoadKeymaps() ([]*Keymap, error) {
	result := []*Keymap{}

	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, "localectl", "list-keymaps", "--no-pager")
	if err != nil {
		return nil, err
	}

	tks := strings.Split(w.String(), "\n")
	for _, curr := range tks {
		if curr == "" {
			continue
		}

		result = append(result, &Keymap{Code: curr})
	}

	return result, nil
}
