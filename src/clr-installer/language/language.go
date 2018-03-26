package language

import (
	"bytes"
	"strings"

	"clr-installer/cmd"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Language represents a system language, containing the locale code and lang tag representation
type Language struct {
	Code string
	Tag  language.Tag
}

// String converts a Language to string, namely it returns the tag's name - or the language desc
func (l *Language) String() string {
	return display.English.Tags().Name(l.Tag)
}

// IsDefault returns true if a given language is the default one
func (l *Language) IsDefault() bool {
	return l.Tag == language.AmericanEnglish
}

// Load uses localectl to load the currently available locales/Languages
func Load() ([]*Language, error) {
	result := []*Language{}

	w := bytes.NewBuffer(nil)
	err := cmd.Run(w, "localectl", "list-locales", "--no-pager")
	if err != nil {
		return nil, err
	}

	tks := strings.Split(w.String(), "\n")
	for _, curr := range tks {
		if curr == "" {
			continue
		}

		code := strings.Replace(curr, ".UTF-8", "", 1)

		lang := &Language{
			Code: curr,
			Tag:  language.MustParse(code),
		}

		result = append(result, lang)
	}

	return result, nil
}
