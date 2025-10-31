package string

import (
	"testing"
	"unicode"
	"unicode/utf8"

	strings "github.com/go-gost/gost.plus/utils/string"
	"github.com/stretchr/testify/assert"
)

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single lowercase letter",
			input:    "a",
			expected: "A",
		},
		{
			name:     "single uppercase letter",
			input:    "A",
			expected: "A",
		},
		{
			name:     "all lowercase word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "already capitalized word",
			input:    "Hello",
			expected: "Hello",
		},
		{
			name:     "all uppercase word",
			input:    "HELLO",
			expected: "HELLO",
		},
		{
			name:     "numbers and symbols",
			input:    "123abc",
			expected: "123abc",
		},
		{
			name:     "starts with space",
			input:    " hello",
			expected: " hello",
		},
		{
			name:     "starts with number",
			input:    "1hello",
			expected: "1hello",
		},
		{
			name:     "starts with symbol",
			input:    "@hello",
			expected: "@hello",
		},
		{
			name:     "unicode character (cyrillic)",
			input:    "привет",
			expected: "Привет",
		},
		{
			name:     "unicode character (chinese)",
			input:    "你好",
			expected: "你好", // Chinese characters don't have case, should remain unchanged
		},
		{
			name:     "starts with combining character",
			input:    "́hello", // Combining acute accent
			expected: "́hello", // Should remain unchanged as the first rune is not a letter
		},
		{
			name:     "starts with multi-byte character",
			input:    "éclair",
			expected: "Éclair",
		},
		{
			name:     "starts with non-printable character",
			input:    "\x00hello",
			expected: "\x00hello",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := strings.CapitalizeFirst(tc.input)
			assert.Equal(t, tc.expected, result, "CapitalizeFirst(%q) = %q; want %q", tc.input, result, tc.expected)
		})
	}
}

// Fuzz test to find edge cases
func FuzzCapitalizeFirst(f *testing.F) {
	// Seed corpus
	for _, tc := range []string{
		"", "a", "A", "hello", "Hello", "123", " hello", "@test", "привет", "你好", "éclair",
	} {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8 strings
		if !utf8.ValidString(input) {
			t.Skip()
		}

		result := strings.CapitalizeFirst(input)

		// Basic properties to check for any input
		if input == "" {
			if result != "" {
				t.Errorf("CapitalizeFirst(\"\") = %q; want \"\"", result)
			}
			return
		}

		// Check if the first rune is a letter
		firstRune, _ := utf8.DecodeRuneInString(input)
		if unicode.IsLetter(firstRune) {
			// If the first rune is lowercase, it should be uppercase in the result
			if unicode.IsLower(firstRune) {
				expectedFirstRune := unicode.ToUpper(firstRune)
				actualFirstRune, _ := utf8.DecodeRuneInString(result)
				if actualFirstRune != expectedFirstRune {
					t.Errorf("First rune not capitalized correctly: got %U, want %U", actualFirstRune, expectedFirstRune)
				}
			}

			// The rest of the string should remain unchanged
			_, size := utf8.DecodeRuneInString(input)
			if len(input) > size {
				if result[size:] != input[size:] {
					t.Errorf("Suffix changed: got %q, want %q", result[size:], input[size:])
				}
			}
		} else {
			// For non-letter first rune, the string should be unchanged
			if result != input {
				t.Errorf("String with non-letter first rune was modified: got %q, want %q", result, input)
			}
		}

		// The result should be valid UTF-8
		if !utf8.ValidString(result) {
			t.Errorf("Result is not valid UTF-8: %q", result)
		}
	})
}

// Helper function to compare strings while ignoring differences in case of the first rune
func equalIgnoreFirstRuneCase(s1, s2 string) bool {
	if s1 == s2 {
		return true
	}

	r1, size1 := utf8.DecodeRuneInString(s1)
	r2, size2 := utf8.DecodeRuneInString(s2)

	if r1 == utf8.RuneError || r2 == utf8.RuneError {
		return false
	}

	// Compare the first runes case-insensitively
	if unicode.ToUpper(r1) != unicode.ToUpper(r2) {
		return false
	}

	// Compare the rest of the strings exactly
	return s1[size1:] == s2[size2:]
}
