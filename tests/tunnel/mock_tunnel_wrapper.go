package tunnel

import (
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/x/service"
)

type MockTunnelWrapper = tunnel.Tunnel

type mockTestEntity struct {
	mock *tunnel.MockTunnel
}

func NewEntity(mockTunnel *tunnel.MockTunnel) MockTunnelWrapper {
	s := &mockTestEntity{
		mock: mockTunnel,
	}
	return s
}

func (s mockTestEntity) ID() string {
	return s.mock.ID()
}

func (s mockTestEntity) Type() string {
	return s.mock.Type()
}

func (s mockTestEntity) Name() string {
	return s.mock.Name()
}

func (s mockTestEntity) Endpoint() string {
	return s.mock.Endpoint()
}

func (s mockTestEntity) Entrypoint() string {
	return s.mock.Entrypoint()
}

func (s mockTestEntity) Options() tunnel.Options {
	return s.mock.Options()
}

func (s *mockTestEntity) Favorite(b bool) {
	s.mock.Favorite(b)
}

func (s mockTestEntity) IsFavorite() bool {
	return s.mock.IsFavorite()
}

func (s mockTestEntity) Close() error {
	return s.mock.Close()
}

func (s mockTestEntity) Err() error {
	return s.mock.Err()
}

func (s mockTestEntity) IsActive() bool {
	return s.mock.IsActive()
}

func (s mockTestEntity) IsClosed() bool {
	return s.mock.IsClosed()
}

func (s mockTestEntity) Run() error {
	return s.mock.Run()
}

func (s *mockTestEntity) SetStats(stats config.ServiceStats) {
	s.mock.SetStats(stats)
}

func (s mockTestEntity) Stats() config.ServiceStats {
	return s.mock.Stats()
}

func (s mockTestEntity) Status() *service.Status {
	return s.mock.Status()
}
