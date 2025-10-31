package option

import (
	"strings"
	"testing"

	opt "github.com/go-gost/gost.plus/utils/fp/option"
	"github.com/stretchr/testify/assert"
)

// Test structures for option testing
type TestStruct struct {
	Value string
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestSome(t *testing.T) {
	t.Run("with non-nil value", func(t *testing.T) {
		val := "test"
		opt := opt.Some(&val)
		assert.NotNil(t, opt, "Option should not be nil")
		assert.Equal(t, val, opt.GetOrElse(""), "Should return the original value")
	})

	t.Run("with nil value", func(t *testing.T) {
		opt := opt.Some[TestStruct](nil)
		assert.True(t, opt.GetOrElse(TestStruct{Value: "default"}).Value == "default", "Should handle nil value")
	})
}

func TestNone(t *testing.T) {
	o := opt.None[string]()
	assert.True(t, o.GetOrElse("default") == "default", "None should return default value")
}

func TestGetOrElse(t *testing.T) {
	t.Run("with Some value", func(t *testing.T) {
		val := "test"
		o := opt.Some(&val)
		result := o.GetOrElse("default")
		assert.Equal(t, val, result, "Should return the original value")
	})

	t.Run("with None", func(t *testing.T) {
		o := opt.None[string]()
		result := o.GetOrElse("default")
		assert.Equal(t, "default", result, "Should return the default value")
	})
}

func TestMatch(t *testing.T) {
	t.Run("with Some value", func(t *testing.T) {
		val := "test"
		o := opt.Some(&val)

		o.Match(
			func(s *string) {
				assert.Equal(t, val, *s, "Should match the original value")
			},
			func() {
				t.Error("Should not call none function")
			},
		)
	})

	t.Run("with None", func(t *testing.T) {
		o := opt.None[string]()
		called := false

		o.Match(
			func(s *string) {
				t.Error("Should not call some function")
			},
			func() {
				called = true
			},
		)

		assert.True(t, called, "None function should be called")
	})
}

func TestFilter(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name     string
		value    *int
		expected int
	}{
		{"even number", intPtr(2), 2},
		{"odd number", intPtr(3), 0},
		{"nil value", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := opt.Some(tt.value).Filter(isEven)
			result := o.GetOrElse(0)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMap(t *testing.T) {
	toUpper := func(s string) string { return strings.ToUpper(s) }

	tests := []struct {
		name     string
		value    *string
		expected string
	}{
		{"non-empty string", stringPtr("test"), "TEST"},
		{"nil value", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := opt.Map(opt.Some(tt.value), toUpper)
			result := o.GetOrElse("")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlatMap(t *testing.T) {
	toLengthOpt := func(s string) opt.Option[int] {
		return opt.Some(intPtr(len(s)))
	}

	tests := []struct {
		name     string
		value    *string
		expected int
	}{
		{"non-empty string", stringPtr("test"), 4},
		{"nil value", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := opt.FlatMap(opt.Some(tt.value), toLengthOpt)
			result := o.GetOrElse(0)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrElse(t *testing.T) {
	tests := []struct {
		name     string
		opt      opt.Option[string]
		fallback opt.Option[string]
		expected string
	}{
		{"some with some", opt.Some(stringPtr("first")), opt.Some(stringPtr("second")), "first"},
		{"none with some", opt.None[string](), opt.Some(stringPtr("second")), "second"},
		{"some with none", opt.Some(stringPtr("first")), opt.None[string](), "first"},
		{"none with none", opt.None[string](), opt.None[string](), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opt.OrElse(tt.fallback).GetOrElse("")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFold(t *testing.T) {
	t.Run("with Some value", func(t *testing.T) {
		o := opt.Some(stringPtr("test"))
		result := opt.Fold(o, "default", func(s string) string {
			return strings.ToUpper(s)
		})
		assert.Equal(t, "TEST", result)
	})

	t.Run("with None", func(t *testing.T) {
		o := opt.None[string]()
		result := opt.Fold(o, "default", func(s string) string {
			return "should not be called"
		})
		assert.Equal(t, "default", result)
	})
}
