package cli

import (
	"testing"
	"time"

	args "github.com/go-gost/gost.plus/utils/cli"
	"github.com/stretchr/testify/assert"
)

// mockArgs is a simple implementation of cli.Args for testing
type mockArgs struct {
	args []string
}

func (m *mockArgs) Get(n int) string {
	if n < 0 || n >= len(m.args) {
		return ""
	}
	return m.args[n]
}

func (m *mockArgs) First() string {
	if len(m.args) > 0 {
		return m.args[0]
	}
	return ""
}

func (m *mockArgs) Tail() []string {
	if len(m.args) < 2 {
		return []string{}
	}
	return m.args[1:]
}

func (m *mockArgs) Len() int {
	return len(m.args)
}

func (m *mockArgs) Present() bool {
	return len(m.args) > 0
}

func (m *mockArgs) Slice() []string {
	return m.args
}

func runWithArgs(t *testing.T, testArgs []string, testFunc func(*args.CLIArgsParser)) {
	// Create a new parser
	parser := args.NewCLIArgsParser()

	// Create mock arguments
	mockArgs := &mockArgs{args: testArgs}

	// Parse the arguments
	err := parser.ParseArgs(mockArgs)
	assert.NoError(t, err)

	// Call the test function with the parsed arguments
	testFunc(parser)
}

func TestCliArgsParser(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedParsed map[string]string
	}{
		{
			name: "prefixed_args",
			args: []string{"--name", "test", "--port", "8080"},
			expectedParsed: map[string]string{
				"name": "test",
				"port": "8080",
			},
		},
		{
			name: "single_dash_flags",
			args: []string{"-v", "-p", "8080", "-n=test"},
			expectedParsed: map[string]string{
				"v": "true",
				"p": "8080",
				"n": "test",
			},
		},
		{
			name: "shell_metachar_redirection",
			args: []string{"-no-stats", "-monitor", "3m", ">", "/dev/null", "2>&1"},
			expectedParsed: map[string]string{
				"no-stats": "true",
				"monitor":  "3m",
			},
		},
		{
			name: "shell_metachar_pipe",
			args: []string{"--name=test", "|", "grep", "something"},
			expectedParsed: map[string]string{
				"name": "test",
			},
		},
		{
			name: "combined_single_dash_flags",
			args: []string{"-abc", "value"},
			expectedParsed: map[string]string{
				"abc": "value",
			},
		},
		{
			name: "mixed_double_and_single_dash",
			args: []string{"--name", "test", "-p", "8080", "-v"},
			expectedParsed: map[string]string{
				"name": "test",
				"p":    "8080",
				"v":    "true",
			},
		},
		{
			name:           "non_prefixed_args",
			args:           []string{"name", "test"},
			expectedParsed: map[string]string{},
		},
		{
			name: "mixed_args",
			args: []string{"--name=test"},
			expectedParsed: map[string]string{
				"name": "test",
			},
		},
		{
			name: "boolean_flag",
			args: []string{"--verbose"},
			expectedParsed: map[string]string{
				"verbose": "true",
			},
		},
		{
			name: "boolean_flag_with_dash",
			args: []string{"--flag-bool"},
			expectedParsed: map[string]string{
				"flag-bool": "true",
			},
		},
		{
			name: "single_dash_boolean_with_value",
			args: []string{"-v", "true"},
			expectedParsed: map[string]string{
				"v": "true",
			},
		},
		{
			name: "single_dash_with_equals",
			args: []string{"-name=test"},
			expectedParsed: map[string]string{
				"name": "test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runWithArgs(t, tc.args, func(parser *args.CLIArgsParser) {
				assert.Equal(t, tc.expectedParsed, parser.GetParsed())
			})
		})
	}
}

func TestStringVar(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		keys     []string
		expected string
	}{
		{
			name:     "no args with default",
			args:     []string{},
			keys:     []string{"test"},
			expected: "default",
		},
		{
			name:     "with value",
			args:     []string{"--test", "value"},
			keys:     []string{"test"},
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var result string
				parser.StringVar(&result, "default", tt.keys...)
				assert.Equal(t, tt.expected, result)
			})
		})
	}
}

func TestIntVar(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		keys     []string
		expected int
	}{
		{
			name:     "no args with default",
			args:     []string{},
			keys:     []string{"port"},
			expected: 8080,
		},
		{
			name:     "with value",
			args:     []string{"--port", "3000"},
			keys:     []string{"port"},
			expected: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var result int
				parser.IntVar(&result, 8080, tt.keys...)
				assert.Equal(t, tt.expected, result)
			})
		})
	}
}

func TestBoolVar(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		keys     []string
		expected bool
	}{
		{
			name:     "no args with default false",
			args:     []string{},
			keys:     []string{"flag"},
			expected: false,
		},
		{
			name:     "boolean flag present",
			args:     []string{"--flag"},
			keys:     []string{"flag"},
			expected: true,
		},
		{
			name:     "boolean flag with explicit true",
			args:     []string{"--flag", "true"},
			keys:     []string{"flag"},
			expected: true,
		},
		{
			name:     "boolean flag with explicit false",
			args:     []string{"--flag", "false"},
			keys:     []string{"flag"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var result bool
				parser.BoolVar(&result, false, tt.keys...)
				assert.Equal(t, tt.expected, result, "Test case: %s", tt.name)
			})
		})
	}
}

func TestBoolEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		keys     []string
		expected bool
	}{
		{
			name:     "explicit true",
			args:     []string{"--flag", "true"},
			keys:     []string{"flag"},
			expected: true,
		},
		{
			name:     "explicit false",
			args:     []string{"--flag", "false"},
			keys:     []string{"flag"},
			expected: false,
		},
		{
			name:     "implicit true",
			args:     []string{"--flag"},
			keys:     []string{"flag"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var result bool
				parser.BoolVar(&result, false, tt.keys...)
				assert.Equal(t, tt.expected, result, "Test case: %s", tt.name)
			})
		})
	}
}

func TestDurationVar(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		keys     []string
		expected time.Duration
	}{
		{
			name:     "no args with default",
			args:     []string{},
			keys:     []string{"timeout"},
			expected: 5 * time.Second,
		},
		{
			name:     "with value",
			args:     []string{"--timeout", "10s"},
			keys:     []string{"timeout"},
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var result time.Duration
				parser.DurationVar(&result, 5*time.Second, tt.keys...)
				assert.Equal(t, tt.expected, result)
			})
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		runWithArgs(t, []string{}, func(parser *args.CLIArgsParser) {
			var s string
			parser.StringVar(&s, "default", "test")
			assert.Equal(t, "default", s)
		})
	})

	t.Run("duplicate keys", func(t *testing.T) {
		runWithArgs(t, []string{"--test", "first", "--test", "second"}, func(parser *args.CLIArgsParser) {
			var s string
			parser.StringVar(&s, "default", "test")
			assert.Equal(t, "second", s) // Last value is used
		})
	})

	t.Run("mixed case keys", func(t *testing.T) {
		runWithArgs(t, []string{"--test", "value"}, func(parser *args.CLIArgsParser) {
			var s string
			// The parser converts keys to lowercase
			parser.StringVar(&s, "default", "TEST") // Should match "test"
			assert.Equal(t, "value", s)
		})
	})
}

func TestIntEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		keys     []string
		expected int
	}{
		{
			name:     "valid int",
			args:     []string{"--number", "42"},
			keys:     []string{"number"},
			expected: 42,
		},
		{
			name:     "invalid int",
			args:     []string{"--number", "notanumber"},
			keys:     []string{"number"},
			expected: 0, // default value when parsing fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var result int
				parser.IntVar(&result, 0, tt.keys...)
				assert.Equal(t, tt.expected, result)
			})
		})
	}
}

func TestUnboundedArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "no unbounded args",
			args:     []string{"--flag", "value", "--another=test"},
			expected: []string{},
		},
		{
			name:     "single unbounded arg",
			args:     []string{"positional1", "--flag", "value"},
			expected: []string{"positional1"},
		},
		{
			name:     "multiple unbounded args",
			args:     []string{"pos1", "--flag", "value", "pos2", "pos3"},
			expected: []string{"pos1", "pos2", "pos3"},
		},
		{
			name:     "mixed with flags and values",
			args:     []string{"--flag", "value", "pos1", "pos2", "--another=test", "pos3"},
			expected: []string{"pos1", "pos2", "pos3"},
		},
		{
			name:     "only unbounded args",
			args:     []string{"pos1", "pos2", "pos3"},
			expected: []string{"pos1", "pos2", "pos3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				// Check unbounded args by index
				for i, expected := range tt.expected {
					actual := parser.UnboundedArg(i, "")
					assert.Equal(t, expected, actual, "Unbinded arg at index %d does not match", i)
				}

				// Test out of bounds access
				assert.Equal(t, "default", parser.UnboundedArg(len(tt.expected), "default"),
					"Out of bounds access should return default value")
			})
		})
	}
}

func TestFloat64Var(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected float64
	}{
		{
			name:     "valid float with decimal",
			args:     []string{"--ratio", "3.14"},
			expected: 3.14,
		},
		{
			name:     "valid float without decimal",
			args:     []string{"--ratio", "42"},
			expected: 42.0,
		},
		{
			name:     "scientific notation",
			args:     []string{"--ratio", "1.23e-4"},
			expected: 1.23e-4,
		},
		{
			name:     "invalid float",
			args:     []string{"--ratio", "notafloat"},
			expected: 2.71828, // default value when parsing fails
		},
		{
			name:     "no args with default",
			args:     []string{},
			expected: 2.71828, // default value when key is not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithArgs(t, tt.args, func(parser *args.CLIArgsParser) {
				var f float64
				parser.Float64Var(&f, 2.71828, "ratio")    // e as default value
				assert.InDelta(t, tt.expected, f, 0.00001) // Using InDelta to handle floating point precision
			})
		})
	}
}

func TestStringVarWithAlias(t *testing.T) {
	t.Run("with alias", func(t *testing.T) {
		runWithArgs(t, []string{"-b", "kubernetes"}, func(parser *args.CLIArgsParser) {
			var s string
			parser.StringVar(&s, "default", "b", "binding")
			assert.Equal(t, "kubernetes", s)
		})
	})

	t.Run("with primary name", func(t *testing.T) {
		runWithArgs(t, []string{"--binding", "kubernetes"}, func(parser *args.CLIArgsParser) {
			var s string
			parser.StringVar(&s, "default", "b", "binding")
			assert.Equal(t, "kubernetes", s)
		})
	})

	t.Run("with default", func(t *testing.T) {
		runWithArgs(t, []string{}, func(parser *args.CLIArgsParser) {
			var s string
			parser.StringVar(&s, "default", "b", "binding")
			assert.Equal(t, "default", s)
		})
	})
}
