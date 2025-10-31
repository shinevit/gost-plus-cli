package fp

import (
	"testing"

	fp "github.com/go-gost/gost.plus/utils/fp"
	"github.com/stretchr/testify/assert"
)

func TestMap_WithInts_ReturnsMappedValues(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		fn       func(int) int
		expected int
	}{
		{
			name:  "double numbers",
			input: 5,
			fn: func(x int) int {
				return x * 2
			},
			expected: 10,
		},
		{
			name:  "zero value",
			input: 0,
			fn: func(x int) int {
				return x + 1
			},
			expected: 1,
		},
		{
			name:  "negative numbers",
			input: -3,
			fn: func(x int) int {
				return x * -1
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Map(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMap_WithStrings_ReturnsMappedValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fn       func(string) string
		expected string
	}{
		{
			name:  "capitalize first letter",
			input: "hello",
			fn: func(s string) string {
				if len(s) == 0 {
					return s
				}
				return string(s[0]-32) + s[1:] // Simple ASCII upper case
			},
			expected: "Hello",
		},
		{
			name:  "empty string",
			input: "",
			fn: func(s string) string {
				return "default"
			},
			expected: "default",
		},
		{
			name:  "uppercase all",
			input: "test",
			fn: func(s string) string {
				return "PREFIX_" + s
			},
			expected: "PREFIX_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Map(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result)
		})
	}
}
