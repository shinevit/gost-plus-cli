package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/go-gost/core/logger"
	xconfig "github.com/go-gost/x/config"
	logger_parser "github.com/go-gost/x/config/parsing/logger"
	"gopkg.in/yaml.v3"
)

const (
	configFile = "config.yml"
	logFile    = "gost-plus.log"
)

var (
	configDir string
)

func init() {
	config.Store(&Config{})
}

func Init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	configDir = ConfigDir()
	os.MkdirAll(configDir, 0755)
	slog.Debug(fmt.Sprintf("appDir: %s", configDir))

	cfg := Get()
	if err := cfg.load(); err != nil {
		slog.Error(fmt.Sprintf("load config: %v", err))
		if _, ok := err.(*os.PathError); ok {
			cfg.Write()
		}
	}
	Set(cfg)

	initLog()
}

func initLog() {
	cfg := Get().Log
	if cfg == nil {
		logDir := LogDir()
		os.MkdirAll(logDir, 0755)
		slog.Debug(fmt.Sprintf("log dir: %s", logDir))

		cfg = &xconfig.LogConfig{
			Output: LogFilePath(),
			Level:  string(logger.InfoLevel),
			Format: string(logger.JSONFormat),
			Rotation: &xconfig.LogRotationConfig{
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 10,
				LocalTime:  true,
			},
		}
	}

	logger.SetDefault(logger_parser.ParseLogger(&xconfig.LoggerConfig{Log: cfg}))
}

var (
	config atomic.Value
)

func Get() *Config {
	c := config.Load().(*Config)
	cfg := &Config{}
	*cfg = *c
	return cfg
}

func Set(c *Config) {
	if c == nil {
		c = &Config{}
	}
	config.Store(c)
}

type Settings struct {
	// Server address.
	// default value is tunnel.gost.plus
	Server string
	// Public entrypoint address.
	// default value is gost.plus
	Entrypoint string
	Lang       string
	Theme      string
}

type Tunnel struct {
	ID        string
	Name      string
	Type      string
	Endpoint  string
	Hostname  string   `yaml:",omitempty"`
	Username  string   `yaml:",omitempty"`
	Password  Password `yaml:",omitempty"`
	EnableTLS bool     `yaml:"enableTLS,omitempty"`
	Keepalive bool     `yaml:",omitempty"`
	TTL       int      `yaml:"ttl,omitempty"`

	Stats     ServiceStats
	Favorite  bool
	Closed    bool
	CreatedAt time.Time
}

type Config struct {
	Settings    *Settings `yaml:",omitempty"`
	Tunnels     []*Tunnel
	EntryPoints []*Tunnel
	Log         *xconfig.LogConfig
}

func (c *Config) load() error {
	f, err := os.Open(ConfigFilePath())
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(c)
}

func (c *Config) Write() error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()

	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(configDir, configFile), buf.Bytes(), 0600)
}

func ConfigDir() string {
	if len(configDir) > 0 {
		return configDir
	}

	dir := os.Getenv("GOST_CONFIG_DIR")
	var err error
	if dir == "" {
		dir, err = os.UserConfigDir()
		if err != nil {
			dir, err = os.UserHomeDir()
		}
	}
	if err != nil {
		slog.Error(fmt.Sprintf("appDir: %v", err))
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	return filepath.Join(dir, "gost.plus")
}

func LogDir() string {
	return filepath.Join(ConfigDir(), "logs")
}

func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), configFile)
}

func LogFilePath() string {
	return filepath.Join(LogDir(), logFile)
}

type ServiceStats struct {
	Time            time.Time
	TotalConns      uint64
	RequestRate     float64
	CurrentConns    uint64
	TotalErrs       uint64
	InputBytes      uint64
	InputRateBytes  uint64
	OutputBytes     uint64
	OutputRateBytes uint64
}

// Encodes a string to base64
func encodePassword(plainSecret string) string {
	if plainSecret == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(plainSecret))
}

// Decodes a base64 encoded string
func decodePassword(encodedSecret string) (string, error) {
	if encodedSecret == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encodedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode password: %w", err)
	}
	return string(decoded), nil
}

// Password represents an encoded password that handles base64 encoding/decoding
type Password string

func NewPassword(value string) Password {
	return Password(value)
}

// Hides from fmt.Println(s) logging
func (p Password) String() string {
	return strings.Repeat("*", p.Length())
}

// Prevents from logging it: fmt.Printf("%+v\n", s). Output: *****
func (p Password) GoString() string {
	return p.String()
}

// Returns the decoded/plain password string
func (p Password) Reveal() string {
	if p.IsEmpty() {
		return ""
	}

	decoded, err := decodePassword(string(p))
	if err != nil {
		// If decoding fails, return the original string (for backward compatibility)
		return string(p)
	}
	return decoded
}

// Encodes and sets the password
func (p *Password) Set(password string) {
	*p = Password(encodePassword(password))
}

// Returns true if the password is empty
func (p Password) IsEmpty() bool {
	return string(p) == ""
}

// Return a length of plain password
func (p Password) Length() int {
	return utf8.RuneCountInString(p.Reveal())
}
