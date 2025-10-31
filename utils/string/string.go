package strings

import (
	"unicode"
	"unicode/utf8"
)

// Returns string with a title case of first letter.
func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s // Return original string if invalid UTF-8
	}

	if unicode.IsUpper(r) || !unicode.IsLetter(r) {
		return s
	}

	// Combine the first rune to title with the rest of the string
	return string(unicode.ToTitle(r)) + s[size:]
}
