package slice

import (
	"fmt"
	"testing"

	opt "github.com/go-gost/gost.plus/utils/fp/option"
	fp "github.com/go-gost/gost.plus/utils/fp/slice"
	"github.com/stretchr/testify/assert"
)

func TestFlatMap_WithIntsAndStringTransformation_ReturnsFlattenedStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) []string
		expected []string
	}{
		{
			name:  "numbers to words",
			input: []int{1, 2, 3},
			fn: func(n int) []string {
				if n == 1 {
					return []string{"one"}
				} else if n == 2 {
					return []string{"two", "even"}
				} else if n == 3 {
					return []string{"three", "odd"}
				}
				return []string{}
			},
			expected: []string{"one", "two", "even", "three", "odd"},
		},
		{
			name:  "empty input",
			input: []int{},
			fn: func(n int) []string {
				return []string{}
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.FlatMap(tt.input, tt.fn)
			assert.NotNil(t, result, "FlatMap should never return nil, expected empty slice")
			assert.IsType(t, tt.expected, result, "Unexpected return type from FlatMap")
			assert.Equal(t, tt.expected, result, "FlatMap returned unexpected elements")
		})
	}
}

func TestFlatMap_WithIntSlices_ReturnsFlattenedInts(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]int
		fn       func([]int) []int
		expected []int
	}{
		{
			name:  "nil input",
			input: [][]int(nil),
			fn: func(arr []int) []int {
				return arr
			},
			expected: []int{},
		},
		{
			name:  "flatten arrays",
			input: [][]int{{1, 2}, {3, 4}, {5, 6}},
			fn: func(arr []int) []int {
				return arr
			},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:  "mixed array sizes",
			input: [][]int{{1}, {2, 3}, {}, {4, 5, 6}},
			fn: func(arr []int) []int {
				return arr
			},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:  "empty input",
			input: [][]int{},
			fn: func(arr []int) []int {
				return arr
			},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.FlatMap(tt.input, tt.fn)
			assert.NotNil(t, result, "FlatMap should never return nil, expected empty slice")
			assert.IsType(t, tt.expected, result, "Unexpected return type from FlatMap")
			assert.Equal(t, tt.expected, result, "FlatMap returned unexpected elements")
		})
	}
}

func TestFilter_WithInts_ReturnsFilteredValues(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) bool
		expected []int
	}{
		{
			name:  "filter even numbers",
			input: []int{1, 2, 3, 4, 5, 6},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{2, 4, 6},
		},
		{
			name:  "empty input",
			input: []int{},
			fn: func(n int) bool {
				return false
			},
			expected: []int{},
		},
		{
			name:  "nil input",
			input: []int(nil),
			fn: func(n int) bool {
				return false
			},
			expected: []int{},
		},
		{
			name:  "all items match",
			input: []int{1, 2, 3, 4},
			fn: func(n int) bool {
				return n > 0
			},
			expected: []int{1, 2, 3, 4},
		},
		{
			name:  "no items match",
			input: []int{1, 3, 5, 7},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{},
		},
		{
			name:  "filter out odd numbers",
			input: []int{1, 2, 3, 4, 5, 6},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{2, 4, 6},
		},
		{
			name:  "filter positive numbers",
			input: []int{-3, -2, 0, 1, 2, 3},
			fn: func(n int) bool {
				return n > 0
			},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Filter(tt.input, tt.fn)
			assert.NotNil(t, result, "Filter should never return nil, expected empty slice")
			assert.IsType(t, tt.expected, result, "Unexpected return type from Filter")
			assert.Equal(t, tt.expected, result, "Filter returned unexpected elements")
		})
	}
}

func TestFilterNot_WithInts_ReturnsInverseFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) bool
		expected []int
	}{
		{
			name:  "nil input",
			input: []int(nil),
			fn: func(n int) bool {
				return false
			},
			expected: []int{},
		},
		{
			name:  "filter out even numbers",
			input: []int{1, 2, 3, 4, 5, 6},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{1, 3, 5},
		},
		{
			name:  "all items match predicate",
			input: []int{2, 4, 6},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{},
		},
		{
			name:  "no items match predicate",
			input: []int{1, 3, 5},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{1, 3, 5},
		},
		{
			name:  "empty input",
			input: []int{},
			fn: func(n int) bool {
				return true
			},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.FilterNot(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterNot_WithStrings_ReturnsInverseFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		fn       func(string) bool
		expected []string
	}{
		{
			name:  "filter out short strings",
			input: []string{"a", "abc", "hello", "world", "x"},
			fn: func(s string) bool {
				return len(s) <= 2
			},
			expected: []string{"abc", "hello", "world"},
		},
		{
			name:  "filter out empty strings",
			input: []string{"", "a", "", "b", ""},
			fn: func(s string) bool {
				return s == ""
			},
			expected: []string{"a", "b"},
		},
		{
			name:  "filter out non-error messages",
			input: []string{"error: foo", "warning: bar", "info: test", "error: baz"},
			fn: func(s string) bool {
				return len(s) < 6 || s[:6] != "error:"
			},
			expected: []string{"error: foo", "error: baz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.FilterNot(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result, "FilterNot returned unexpected elements")
		})
	}
}

func TestFilterAndFilterNot_WithInts_AreComplementary(t *testing.T) {
	tests := []struct {
		name                string
		input               []int
		predicate           func(int) bool
		expectedFiltered    []int
		expectedFilteredNot []int
	}{
		{
			name:  "even/odd numbers",
			input: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			expectedFiltered:    []int{2, 4, 6, 8, 10},
			expectedFilteredNot: []int{1, 3, 5, 7, 9},
		},
		{
			name:  "all items match predicate",
			input: []int{2, 4, 6, 8},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			expectedFiltered:    []int{2, 4, 6, 8},
			expectedFilteredNot: nil, // Can be nil or empty slice
		},
		{
			name:  "no items match predicate",
			input: []int{1, 3, 5, 7},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			expectedFiltered:    nil, // Can be nil or empty slice
			expectedFilteredNot: []int{1, 3, 5, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := fp.Filter(tt.input, tt.predicate)
			filteredNot := fp.FilterNot(tt.input, tt.predicate)

			// Use ElementsMatch for more flexible slice comparison (handles both nil and empty slices)
			assert.ElementsMatch(t, tt.expectedFiltered, filtered, "Filter result mismatch")
			assert.ElementsMatch(t, tt.expectedFilteredNot, filteredNot, "FilterNot result mismatch")

			// Combined should equal original
			var combined []int
			if filtered != nil || filteredNot != nil {
				combined = append(filtered, filteredNot...)
			}
			assert.ElementsMatch(t, tt.input, combined,
				"Filter and FilterNot results should combine to original")
		})
	}
}

func TestFilterAndFilterNot_WithStrings_AreComplementary(t *testing.T) {
	tests := []struct {
		name                string
		input               []string
		predicate           func(string) bool
		expectedFiltered    []string
		expectedFilteredNot []string
	}{
		{
			name:  "long/short strings",
			input: []string{"hi", "hello", "ok", "world", "a", "testing"},
			predicate: func(s string) bool {
				return len(s) >= 5
			},
			expectedFiltered:    []string{"hello", "world", "testing"},
			expectedFilteredNot: []string{"hi", "ok", "a"},
		},
		{
			name:  "empty strings",
			input: []string{"", "a", "", "b", ""},
			predicate: func(s string) bool {
				return s == ""
			},
			expectedFiltered:    []string{"", "", ""},
			expectedFilteredNot: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := fp.Filter(tt.input, tt.predicate)
			filteredNot := fp.FilterNot(tt.input, tt.predicate)

			// Use ElementsMatch for more flexible slice comparison (handles both nil and empty slices)
			assert.ElementsMatch(t, tt.expectedFiltered, filtered, "Filter result mismatch")
			assert.ElementsMatch(t, tt.expectedFilteredNot, filteredNot, "FilterNot result mismatch")

			// Combined should equal original
			var combined []string
			if filtered != nil || filteredNot != nil {
				combined = append(filtered, filteredNot...)
			}
			assert.ElementsMatch(t, tt.input, combined,
				"Filter and FilterNot results should combine to original")
		})
	}
}

func TestCount_WithInts_ReturnsCorrectCount(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) bool
		expected uint64
	}{
		{
			name:  "count even numbers",
			input: []int{1, 2, 3, 4, 5, 6},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: 3,
		},
		{
			name:  "empty input",
			input: []int{},
			fn: func(n int) bool {
				return true
			},
			expected: 0,
		},
		{
			name:  "all items match",
			input: []int{2, 4, 6, 8},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: 4,
		},
		{
			name:  "no items match",
			input: []int{1, 3, 5, 7},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: 0,
		},
		{
			name:  "count positive numbers",
			input: []int{-2, -1, 0, 1, 2, 3},
			fn: func(n int) bool {
				return n > 0
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Count(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCount_WithStrings_ReturnsCorrectCount(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		fn       func(string) bool
		expected uint64
	}{
		{
			name:  "count long strings",
			input: []string{"a", "ab", "abc", "abcd"},
			fn: func(s string) bool {
				return len(s) > 2
			},
			expected: 2,
		},
		{
			name:  "count empty strings",
			input: []string{"", "a", "", "b", ""},
			fn: func(s string) bool {
				return s == ""
			},
			expected: 3,
		},
		{
			name:  "count strings with prefix",
			input: []string{"error: foo", "warning: bar", "info: test", "error: baz"},
			fn: func(s string) bool {
				return len(s) >= 6 && s[:6] == "error:"
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Count(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSum_WithInts_ReturnsCorrectSum(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) int
		expected int
	}{
		{
			name:  "sum of integers",
			input: []int{1, 2, 3, 4, 5},
			fn: func(n int) int {
				return n
			},
			expected: 15,
		},
		{
			name:  "empty slice",
			input: []int{},
			fn: func(n int) int {
				return n
			},
			expected: 0,
		},
		{
			name:  "negative numbers",
			input: []int{-1, -2, -3, -4, -5},
			fn: func(n int) int {
				return n
			},
			expected: -15,
		},
		{
			name:  "mixed positive and negative",
			input: []int{-1, 2, -3, 4, -5},
			fn: func(n int) int {
				return n
			},
			expected: -3,
		},
		{
			name:  "with transformation",
			input: []int{1, 2, 3, 4, 5},
			fn: func(n int) int {
				return n * 2
			},
			expected: 30, // (1+2+3+4+5)*2 = 30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Sum(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSum_WithFloat64_ReturnsCorrectSum(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		fn       func(float64) float64
		expected float64
	}{
		{
			name:  "sum of float64",
			input: []float64{1.1, 2.2, 3.3, 4.4},
			fn: func(n float64) float64 {
				return n
			},
			expected: 11.0,
		},
		{
			name:  "empty slice",
			input: []float64{},
			fn: func(n float64) float64 {
				return n
			},
			expected: 0.0,
		},
		{
			name:  "negative floats",
			input: []float64{-1.1, -2.2, -3.3, -4.4},
			fn: func(n float64) float64 {
				return n
			},
			expected: -11.0,
		},
		{
			name:  "with transformation",
			input: []float64{1.0, 2.0, 3.0},
			fn: func(n float64) float64 {
				return n * 1.5
			},
			expected: 9.0, // (1+2+3)*1.5 = 9.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.Sum(tt.input, tt.fn)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestMap_WithIntToString_ReturnsMappedStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []string{},
		},
		{
			name:     "Single element",
			input:    []int{42},
			expected: []string{"42"},
		},
		{
			name:     "Multiple elements",
			input:    []int{1, 2, 3, 4, 5},
			expected: []string{"1", "2", "3", "4", "5"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fp.Map(tc.input, func(i int) string {
				return fmt.Sprint(i)
			})
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMap_WithStructTransformation_ReturnsMappedValues(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	tests := []struct {
		name     string
		input    []person
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []person{},
			expected: []string{},
		},
		{
			name: "Single element",
			input: []person{
				{name: "Alice", age: 30},
			},
			expected: []string{"Alice:30"},
		},
		{
			name: "Multiple elements",
			input: []person{
				{name: "Alice", age: 30},
				{name: "Bob", age: 25},
			},
			expected: []string{"Alice:30", "Bob:25"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fp.Map(tc.input, func(p person) string {
				return fmt.Sprintf("%s:%d", p.name, p.age)
			})
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAny_WithDifferentInputs_ReturnsCorrectBool(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected bool
	}{
		{
			name:     "Nil slice",
			input:    nil,
			expected: false,
		},
		{
			name:     "Empty slice",
			input:    []int{},
			expected: false,
		},
		{
			name:     "Single element",
			input:    []int{42},
			expected: true,
		},
		{
			name:     "Multiple elements",
			input:    []int{1, 2, 3},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fp.Any(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExists_WithDifferentInputs_ReturnsCorrectBool(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		value    int
		expected bool
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			value:    42,
			expected: false,
		},
		{
			name:     "Element exists",
			input:    []int{1, 2, 3, 4, 5},
			value:    3,
			expected: true,
		},
		{
			name:     "Element doesn't exist",
			input:    []int{1, 2, 3, 4, 5},
			value:    42,
			expected: false,
		},
		{
			name:     "First element",
			input:    []int{1, 2, 3, 4, 5},
			value:    1,
			expected: true,
		},
		{
			name:     "Last element",
			input:    []int{1, 2, 3, 4, 5},
			value:    5,
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fp.Exists(tc.input, func(x int) bool {
				return x == tc.value
			})
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSum_WithOtherNumericTypes_ReturnsCorrectSums(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		input := []uint{1, 2, 3, 4, 5}
		fn := func(n uint) uint { return n }
		expected := uint(15)
		result := fp.Sum(input, fn)
		assert.Equal(t, expected, result)
	})

	t.Run("int8", func(t *testing.T) {
		input := []int8{1, 2, 3, 4, 5}
		fn := func(n int8) int8 { return n }
		expected := int8(15)
		result := fp.Sum(input, fn)
		assert.Equal(t, expected, result)
	})

	t.Run("float32", func(t *testing.T) {
		input := []float32{1.1, 2.2, 3.3, 4.4}
		fn := func(n float32) float32 { return n }
		expected := float32(11.0)
		result := fp.Sum(input, fn)
		assert.InDelta(t, expected, result, 0.0001)
	})
}

func TestFirst_WithInts_ReturnsFirstMatchingElement(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		pred     func(int) bool
		expected int
		found    bool
	}{
		{
			name:  "find first even number",
			input: []int{1, 3, 4, 6, 7},
			pred: func(n int) bool {
				return n%2 == 0
			},
			expected: 4,
			found:    true,
		},
		{
			name:  "no match found",
			input: []int{1, 3, 5, 7, 9},
			pred: func(n int) bool {
				return n%2 == 0
			},
			expected: 0,
			found:    false,
		},
		{
			name:  "empty input",
			input: []int{},
			pred: func(n int) bool {
				return true
			},
			expected: 0,
			found:    false,
		},
		{
			name:  "first element matches",
			input: []int{2, 4, 6, 8},
			pred: func(n int) bool {
				return n%2 == 0
			},
			expected: 2,
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.First(tt.input, tt.pred)
			actual := opt.Fold(result, 0, func(v int) int { return v })
			expected := tt.expected
			if !tt.found {
				expected = 0
			}
			assert.Equal(t, expected, actual, "Unexpected value from First")
		})
	}
}

func TestFirst_WithStrings_ReturnsFirstMatchingElement(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		pred     func(string) bool
		expected string
		found    bool
	}{
		{
			name:  "find first string with length > 3",
			input: []string{"a", "bb", "ccc", "dddd", "eeeee"},
			pred: func(s string) bool {
				return len(s) > 3
			},
			expected: "dddd",
			found:    true,
		},
		{
			name:  "no match found in strings",
			input: []string{"a", "b", "c"},
			pred: func(s string) bool {
				return len(s) > 5
			},
			expected: "",
			found:    false,
		},
		{
			name:  "first element matches in strings",
			input: []string{"apple", "banana", "cherry"},
			pred: func(s string) bool {
				return len(s) > 0
			},
			expected: "apple",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.First(tt.input, tt.pred)
			actual := opt.Fold(result, "", func(v string) string { return v })
			expected := tt.expected
			if !tt.found {
				expected = ""
			}
			assert.Equal(t, expected, actual, "Unexpected value from First")
		})
	}
}

func TestForAll_WithInts_ReturnsCorrectResult(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) bool
		expected bool
	}{
		{
			name:  "all even numbers",
			input: []int{2, 4, 6, 8, 10},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: true,
		},
		{
			name:  "not all even numbers",
			input: []int{2, 4, 5, 8, 10},
			fn: func(n int) bool {
				return n%2 == 0
			},
			expected: false,
		},
		{
			name:  "empty slice",
			input: []int{},
			fn: func(n int) bool {
				return false
			},
			expected: true, // Vacuous truth - all elements (none) satisfy the condition
		},
		{
			name:  "nil slice",
			input: nil,
			fn: func(n int) bool {
				return false
			},
			expected: true, // Vacuous truth - all elements (none) satisfy the condition
		},
		{
			name:  "all elements satisfy complex condition",
			input: []int{10, 20, 30, 40, 50},
			fn: func(n int) bool {
				return n > 0 && n < 100 && n%10 == 0
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.ForAll(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result, "ForAll returned unexpected result for test case: %s", tt.name)
		})
	}
}

func TestForAll_WithStrings_ReturnsCorrectResult(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		fn       func(string) bool
		expected bool
	}{
		{
			name:  "all strings with length > 2",
			input: []string{"abc", "def", "ghi"},
			fn: func(s string) bool {
				return len(s) > 2
			},
			expected: true,
		},
		{
			name:  "not all strings with length > 2",
			input: []string{"a", "bb", "ccc"},
			fn: func(s string) bool {
				return len(s) > 2
			},
			expected: false,
		},
		{
			name:  "empty string slice",
			input: []string{},
			fn: func(s string) bool {
				return false
			},
			expected: true, // Vacuous truth
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.ForAll(tt.input, tt.fn)
			assert.Equal(t, tt.expected, result, "ForAll returned unexpected result for test case: %s", tt.name)
		})
	}
}
