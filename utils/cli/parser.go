package cli

import (
	"strconv"
	"strings"
	"time"

	"unicode"

	"github.com/urfave/cli/v2"
)

func isShellMetaChar(s string) bool {
	if s == "" {
		return false
	}
	// Check for common shell metacharacters: > >> < | & ; \n ( ) { } $
	metaChars := ">|&;(){}$"
	return strings.ContainsAny(s, metaChars) || unicode.IsSpace(rune(s[0]))
}

type CLIArgsParser struct {
	parsed    map[string]string
	unbounded []string
}

func NewCLIArgsParser() *CLIArgsParser {
	return &CLIArgsParser{
		parsed: make(map[string]string),
	}
}

func (p *CLIArgsParser) ParseArgs(args cli.Args) error {
	p.parsed = make(map[string]string) // Reset parsed flags
	p.unbounded = make([]string, 0)    // Reset slice

	rawArgs := args.Slice()
	for i := 0; i < len(rawArgs); {
		arg := rawArgs[i]

		if isShellMetaChar(arg) {
			// stop processing further arguments
			return nil
		}

		// Handle --key=value and -key=value
		if strings.HasPrefix(arg, "-") {
			// Handle --flag or -f
			key := strings.TrimLeft(arg, "-")

			// Check if it's a key-value pair with equals
			if strings.Contains(key, "=") {
				parts := strings.SplitN(key, "=", 2)
				p.parsed[strings.ToLower(parts[0])] = parts[1]
				i++
				continue
			}

			// Check if next arg is a value (not a flag)
			if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				p.parsed[strings.ToLower(key)] = rawArgs[i+1]
				i += 2
			} else {
				// boolean flag
				p.parsed[strings.ToLower(key)] = "true"
				i++
			}
		} else {
			p.addUnbounded(arg)
			i++
		}
	}
	return nil
}

// StringVar sets the value of the string pointer `p` to the value associated with the first key found.
func (p *CLIArgsParser) StringVar(ptr *string, defaultValue string, keys ...string) {
	for _, key := range keys {
		if val, ok := p.parsed[strings.ToLower(key)]; ok {
			*ptr = val
			return
		}
	}
	*ptr = defaultValue
}

// IntVar sets the value of the int pointer `p` to the integer value associated with the first key found.
func (p *CLIArgsParser) IntVar(ptr *int, defaultValue int, keys ...string) {
	for _, key := range keys {
		if val, ok := p.parsed[strings.ToLower(key)]; ok {
			if i, err := strconv.Atoi(val); err == nil {
				*ptr = i
				return
			}
		}
	}
	*ptr = defaultValue
}

// BoolVar sets the value of the bool pointer `p` to the boolean value associated with the first key found.
// This method now handles flags like `--verbose` (presence means true)
// and `--verbose=true` or `--verbose=false`.
func (p *CLIArgsParser) BoolVar(ptr *bool, defaultValue bool, keys ...string) {
	for _, key := range keys {
		keyLower := strings.ToLower(key)
		if val, ok := p.parsed[keyLower]; ok {
			if b, err := strconv.ParseBool(val); err == nil {
				*ptr = b
				return
			}
			// If value is not explicitly true/false, but the flag was present (e.g., just "--verbose"),
			// treat it as true.
			*ptr = true
			return
		}
	}
	*ptr = defaultValue
}

// Float64Var sets the value of the float64 pointer `p` to the floating-point value associated with the first key found.
func (p *CLIArgsParser) Float64Var(ptr *float64, defaultValue float64, keys ...string) {
	for _, key := range keys {
		if val, ok := p.parsed[strings.ToLower(key)]; ok {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				*ptr = f
				return
			}
		}
	}
	*ptr = defaultValue
}

// DurationVar sets the value of the time.Duration pointer `p` to the duration value associated with the first key found.
func (p *CLIArgsParser) DurationVar(ptr *time.Duration, defaultValue time.Duration, keys ...string) {
	for _, key := range keys {
		if val, ok := p.parsed[strings.ToLower(key)]; ok {
			if d, err := time.ParseDuration(val); err == nil {
				*ptr = d
				return
			}
		}
	}
	*ptr = defaultValue
}

// GetParsed returns the parsed map for testing purposes.
func (p *CLIArgsParser) GetParsed() map[string]string {
	return p.parsed
}

// Returns the positional argument at the specified index, or the default value if the index is out of bounds.
func (p *CLIArgsParser) UnboundedArg(index int, defaultValue string) string {
	if index < 0 || index >= len(p.unbounded) {
		return defaultValue
	}
	return p.unbounded[index]
}

func (p *CLIArgsParser) addUnbounded(value string) {
	p.unbounded = append(p.unbounded, value)
}
