package tunnel

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/config"
	xconfig "github.com/go-gost/x/config"
	_ "github.com/go-gost/x/connector/tunnel"
	_ "github.com/go-gost/x/dialer/ws"
	xservice "github.com/go-gost/x/service"
)

const (
	EndpointAddr = "gost.plus"
	ServerName   = "tunnel.gost.plus"
	ServerAddr   = ServerName + ":443"
)

const (
	FileTunnel = "file"
	HTTPTunnel = "http"
	TCPTunnel  = "tcp"
	UDPTunnel  = "udp"
)

var (
	ErrTunnelClosed            = errors.New("tunnel closed")
	ErrTunnelDuplicateInstance = errors.New("duplicate of tunnel instance")
)

type Options struct {
	ID        string
	Name      string
	Endpoint  string
	Hostname  string
	Username  string
	Password  config.Password
	EnableTLS bool
	Keepalive bool
	TTL       int
	CreatedAt time.Time
	Stats     config.ServiceStats
}

type Option func(opts *Options)

func IDOption(id string) Option {
	return func(opts *Options) {
		opts.ID = id
	}
}

func NameOption(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func EndpointOption(endpoint string) Option {
	return func(opts *Options) {
		opts.Endpoint = endpoint
	}
}

func HostnameOption(hostname string) Option {
	return func(opts *Options) {
		opts.Hostname = hostname
	}
}

func UsernameOption(username string) Option {
	return func(opts *Options) {
		opts.Username = username
	}
}

func PasswordOption(password string) Option {
	return func(opts *Options) {
		opts.Password.Set(password)
	}
}

func EnableTLSOption(b bool) Option {
	return func(opts *Options) {
		opts.EnableTLS = b
	}
}

func KeepaliveOption(b bool) Option {
	return func(opts *Options) {
		opts.Keepalive = b
	}
}

func TTLOption(ttl int) Option {
	return func(opts *Options) {
		opts.TTL = ttl
	}
}

func CreatedAtOption(createdAt time.Time) Option {
	return func(opts *Options) {
		opts.CreatedAt = createdAt
	}
}

type ServiceStatus interface {
	Status() *xservice.Status
}

type Tunnel interface {
	ID() string
	Type() string
	Name() string
	Endpoint() string
	Entrypoint() string
	Options() Options
	Run() error
	Status() *xservice.Status
	Stats() config.ServiceStats
	SetStats(stats config.ServiceStats)
	Favorite(b bool)
	IsFavorite() bool
	Close() error
	IsClosed() bool
	IsActive() bool
	Err() error
}

// manages the tunnels
type tunnelManager struct {
	tunnels map[string]Tunnel
	mux     sync.RWMutex
}

var (
	tm = &tunnelManager{
		tunnels: make(map[string]Tunnel),
	}
)

func Count() int {
	tm.mux.RLock()
	defer tm.mux.RUnlock()
	return len(tm.tunnels)
}

func Add(s Tunnel) {
	if s == nil {
		return
	}
	tm.mux.Lock()
	defer tm.mux.Unlock()

	if _, exists := tm.tunnels[s.ID()]; !exists {
		tm.tunnels[s.ID()] = s
	}
}

func Set(s Tunnel) {
	if s == nil {
		return
	}
	tm.mux.Lock()
	defer tm.mux.Unlock()

	if existingTunnel, exists := tm.tunnels[s.ID()]; exists {
		s.Favorite(existingTunnel.IsFavorite())
		tm.tunnels[s.ID()] = s
	}
}

func Get(id string) Tunnel {
	tm.mux.RLock()
	defer tm.mux.RUnlock()

	tun, exists := tm.tunnels[id]
	if exists {
		return tun
	}
	return nil
}

func Delete(id string) {
	tm.mux.Lock()
	defer tm.mux.Unlock()

	if tun, exists := tm.tunnels[id]; exists {
		tun.Close()
		delete(tm.tunnels, id)
	}
}

func ChainConfig(id string, name string) *xconfig.ChainConfig {
	return &xconfig.ChainConfig{
		Name: name,
		Hops: []*xconfig.HopConfig{
			{
				Name: name,
				Nodes: []*xconfig.NodeConfig{
					{
						Name: name,
						Addr: ServerAddr,
						Connector: &xconfig.ConnectorConfig{
							Type:     "tunnel",
							Metadata: map[string]any{"tunnel.id": id},
						},
						Dialer: &xconfig.DialerConfig{
							Type: "wss",
							TLS: &xconfig.TLSConfig{
								Secure:     true,
								ServerName: ServerName,
							},
						},
					},
				},
			},
		},
	}
}

func SaveConfig() error {
	cfg := config.Get()
	cfg.Tunnels = nil

	for _, tun := range tm.tunnels {
		opts := tun.Options()
		cfg.Tunnels = append(cfg.Tunnels, &config.Tunnel{
			ID:        tun.ID(),
			Name:      tun.Name(),
			Type:      tun.Type(),
			Endpoint:  tun.Endpoint(),
			Hostname:  opts.Hostname,
			Username:  opts.Username,
			Password:  opts.Password,
			EnableTLS: opts.EnableTLS,
			Favorite:  tun.IsFavorite(),
			Closed:    tun.IsClosed(),
			CreatedAt: opts.CreatedAt,
			Stats:     tun.Stats(),
			Keepalive: opts.Keepalive,
			TTL:       opts.TTL,
		})
	}

	config.Set(cfg)

	if err := cfg.Write(); err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}

// Initializes tunnels from the config file keeping them in memory
// It does not run tunnels
func LoadFromConfig() {
	for _, cfg := range config.Get().Tunnels {
		if cfg == nil {
			continue
		}

		tun := CreateTunnel(cfg.Type, Options{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Endpoint:  cfg.Endpoint,
			Hostname:  cfg.Hostname,
			Username:  cfg.Username,
			Password:  cfg.Password,
			EnableTLS: cfg.EnableTLS,
			CreatedAt: cfg.CreatedAt,
			Stats:     cfg.Stats,
		})
		if tun == nil {
			continue
		}

		if cfg.Closed {
			tun.Close()
		}

		tun.Favorite(cfg.Favorite)
		Add(tun)
	}
}

func CreateTunnel(st string, opts Options) (t Tunnel) {
	options := []Option{
		IDOption(opts.ID),
		NameOption(opts.Name),
		EndpointOption(opts.Endpoint),
		HostnameOption(opts.Hostname),
		UsernameOption(opts.Username),
		EnableTLSOption(opts.EnableTLS),
		CreatedAtOption(opts.CreatedAt),
	}
	if !opts.Password.IsEmpty() {
		options = append(options, PasswordOption(opts.Password.Reveal()))
	}

	switch st {
	case FileTunnel:
		t = NewFileTunnel(options...)
	case HTTPTunnel:
		t = NewHTTPTunnel(options...)
	case TCPTunnel:
		t = NewTCPTunnel(options...)
	case UDPTunnel:
		t = NewUDPTunnel(options...)
	default:
		return nil
	}

	t.SetStats(opts.Stats)
	return
}

// returns sorted tunnels by CreatedAt
func GetAll() []Tunnel {
	tm.mux.RLock()
	defer tm.mux.RUnlock()

	tunnels := make([]Tunnel, 0, len(tm.tunnels))
	for _, tun := range tm.tunnels {
		tunnels = append(tunnels, tun)
	}

	slices.SortFunc(tunnels, func(a, b Tunnel) int {
		if a != nil && b != nil && a.Options().CreatedAt.Before(b.Options().CreatedAt) {
			return -1
		}
		if a != nil && b != nil && a.Options().CreatedAt.After(b.Options().CreatedAt) {
			return 1
		}
		return 0
	})

	return tunnels
}

func getState(id string) xservice.State {
	if existing := Get(id); existing != nil {
		if status := existing.Status(); status != nil {
			return status.State()
		}
	}
	return ""
}

func isTunnelExisting(tun Tunnel) bool {
	if tun == nil {
		return false
	}
	existing := Get(tun.ID())
	return existing == tun
}
