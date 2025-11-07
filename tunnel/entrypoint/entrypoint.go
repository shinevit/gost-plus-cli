package entrypoint

import (
	"errors"
	"slices"
	"sync"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/tunnel"
	xservice "github.com/go-gost/x/service"
)

const (
	TCPEntryPoint = "tcp"
	UDPEntryPoint = "udp"
)

var (
	ErrEntryPointClosed = errors.New("entrypoint closed")
)

type EntryPoint = tunnel.Tunnel

type entryPointList struct {
	list []EntryPoint
	mux  sync.RWMutex
}

var (
	entryPoints entryPointList
)

func Count() int {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()
	return len(entryPoints.list)
}

func Add(s EntryPoint) {
	entryPoints.mux.Lock()
	defer entryPoints.mux.Unlock()
	entryPoints.list = append(entryPoints.list, s)
}

func Set(s EntryPoint) {
	if s == nil {
		return
	}

	old := Get(s.ID())
	if old == nil {
		return
	}
	s.Favorite(old.IsFavorite())

	entryPoints.mux.Lock()
	defer entryPoints.mux.Unlock()

	for i, ep := range entryPoints.list {
		if ep != nil && ep.ID() == s.ID() {
			entryPoints.list[i] = s
		}
	}
}

func GetIndex(index int) EntryPoint {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()
	if index < 0 || index >= len(entryPoints.list) {
		return nil
	}
	return entryPoints.list[index]
}

func Get(id string) EntryPoint {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()

	for _, s := range entryPoints.list {
		if s != nil && s.ID() == id {
			return s
		}
	}
	return nil
}

func Delete(id string) {
	entryPoints.mux.Lock()
	defer entryPoints.mux.Unlock()

	for i, s := range entryPoints.list {
		if s != nil && s.ID() == id {
			s.Close()
			entryPoints.list[i] = nil
			return
		}
	}
}

func LoadFromConfig() {
	for _, cfg := range config.Get().EntryPoints {
		if cfg == nil {
			continue
		}

		ep := createEntryPoint(cfg.Type, tunnel.Options{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Endpoint:  cfg.Endpoint,
			Hostname:  cfg.Hostname,
			Username:  cfg.Username,
			Password:  cfg.Password,
			EnableTLS: cfg.EnableTLS,
			Keepalive: cfg.Keepalive,
			TTL:       cfg.TTL,
			CreatedAt: cfg.CreatedAt,
			Stats:     cfg.Stats,
		})
		if ep == nil {
			continue
		}

		if cfg.Closed {
			ep.Close()
		}

		ep.Favorite(cfg.Favorite)
		Add(ep)
	}
}

func SaveConfig() error {
	cfg := config.Get()
	cfg.EntryPoints = nil

	for i := range Count() {
		ep := GetIndex(i)
		if ep == nil {
			continue
		}

		opts := ep.Options()

		cfg.EntryPoints = append(cfg.EntryPoints, &config.Tunnel{
			ID:        ep.ID(),
			Name:      ep.Name(),
			Type:      ep.Type(),
			Endpoint:  ep.Entrypoint(),
			Hostname:  opts.Hostname,
			Username:  opts.Username,
			Password:  opts.Password,
			EnableTLS: opts.EnableTLS,
			Favorite:  ep.IsFavorite(),
			Closed:    ep.IsClosed(),
			CreatedAt: opts.CreatedAt,
			Stats:     ep.Stats(),
		})
	}

	config.Set(cfg)

	if err := cfg.Write(); err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}

func GetAll() []EntryPoint {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()
	return slices.Clone(entryPoints.list)
}

func createEntryPoint(st string, opts tunnel.Options) (ep EntryPoint) {
	options := []tunnel.Option{
		tunnel.IDOption(opts.ID),
		tunnel.NameOption(opts.Name),
		tunnel.EndpointOption(opts.Endpoint),
		tunnel.HostnameOption(opts.Hostname),
		tunnel.UsernameOption(opts.Username),
		tunnel.EnableTLSOption(opts.EnableTLS),
		tunnel.CreatedAtOption(opts.CreatedAt),
	}
	if !opts.Password.IsEmpty() {
		options = append(options, tunnel.PasswordOption(opts.Password.Reveal()))
	}
	switch st {
	case TCPEntryPoint:
		ep = NewTCPEntryPoint(options...)
	case UDPEntryPoint:
		ep = NewUDPEntryPoint(options...)
	default:
		return nil
	}

	ep.SetStats(opts.Stats)
	return
}

func isActive(en EntryPoint) bool {
	if status := en.Status(); status != nil {
		return status.State() == xservice.StateRunning ||
			status.State() == xservice.StateReady
	}
	return false
}
