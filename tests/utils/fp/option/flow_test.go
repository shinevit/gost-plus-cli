package option

import (
	"strings"
	"testing"

	opt "github.com/go-gost/gost.plus/utils/fp/option"
	"github.com/stretchr/testify/assert"
)

// Test structures
type TunnelManager struct {
	ActiveTunnel *Tunnel
}
type Tunnel struct {
	Status *TunnelStatus
}
type TunnelStatus struct {
	State string
}

// Test Flow1 with basic transformation
func TestFlow1_Basic(t *testing.T) {
	t.Run("with valid input", func(t *testing.T) {

		input := &TunnelStatus{State: "running"}
		transform := opt.Flow1(
			opt.MapResult(func(s *TunnelStatus) string {
				return strings.ToUpper(s.State)
			}),
			opt.GetOrElse("default"),
		)

		_, result := transform.Run(input)
		assert.Equal(t, "RUNNING", result)
	})

	t.Run("with nil input", func(t *testing.T) {
		var input *string = nil
		defaultValue := "default"
		transform := opt.Flow1(
			opt.MapResult(func(s *string) string {
				if s == nil {
					return "nil"
				}
				return *s
			}),
			opt.GetOrElse(defaultValue),
		)

		_, result := transform.Run(input)
		assert.Equal(t, defaultValue, result)
	})
}

// Test Flow2 with two chained operations
func TestFlow2_Chaining(t *testing.T) {
	t.Run("with valid tunnel manager", func(t *testing.T) {
		manager := &TunnelManager{
			ActiveTunnel: &Tunnel{
				Status: &TunnelStatus{State: "active"},
			},
		}

		transform := opt.Flow2(
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.MapResult(func(t *Tunnel) string {
				if t == nil || t.Status == nil {
					return "inactive"
				}
				return t.Status.State
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(manager)
		assert.Equal(t, "active", result)
	})

	t.Run("with nil tunnel", func(t *testing.T) {
		manager := &TunnelManager{ActiveTunnel: nil}

		transform := opt.Flow2(
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.MapResult(func(t *Tunnel) string {
				return "should not reach here"
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(manager)
		assert.Equal(t, "disconnected", result)
	})
}

// Test Flow3 with three chained operations
func TestFlow3_DeepChaining(t *testing.T) {
	t.Run("with full chain", func(t *testing.T) {
		manager := &TunnelManager{
			ActiveTunnel: &Tunnel{
				Status: &TunnelStatus{State: "active"},
			},
		}

		transform := opt.Flow3(
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.Chain(func(t *Tunnel) *TunnelStatus {
				return t.Status
			}),
			opt.MapResult(func(s *TunnelStatus) string {
				return s.State
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(manager)
		assert.Equal(t, "active", result)
	})

	t.Run("with broken chain", func(t *testing.T) {
		manager := &TunnelManager{
			ActiveTunnel: &Tunnel{Status: nil},
		}

		transform := opt.Flow3(
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.Chain(func(t *Tunnel) *TunnelStatus {
				return t.Status
			}),
			opt.MapResult(func(s *TunnelStatus) string {
				return s.State
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(manager)
		assert.Equal(t, "disconnected", result)
	})
}

// Test Flow4 with four chained operations
func TestFlow4_ComplexChaining(t *testing.T) {
	type Config struct {
		Manager *TunnelManager
	}

	t.Run("with full chain", func(t *testing.T) {
		config := &Config{
			Manager: &TunnelManager{
				ActiveTunnel: &Tunnel{
					Status: &TunnelStatus{State: "active"},
				},
			},
		}

		transform := opt.Flow4(
			opt.Chain(func(c *Config) *TunnelManager {
				return c.Manager
			}),
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.Chain(func(t *Tunnel) *TunnelStatus {
				return t.Status
			}),
			opt.MapResult(func(s *TunnelStatus) string {
				return s.State
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(config)
		assert.Equal(t, "active", result)
	})
}

// Test Flow5 with five chained operations
func TestFlow5_DeepNesting(t *testing.T) {
	type System struct {
		Config struct {
			Manager *TunnelManager
		}
	}

	t.Run("with full chain", func(t *testing.T) {
		system := System{}
		system.Config.Manager = &TunnelManager{
			ActiveTunnel: &Tunnel{
				Status: &TunnelStatus{State: "active"},
			},
		}

		transform := opt.Flow5(
			opt.Chain(func(s *System) *TunnelManager {
				return s.Config.Manager
			}),
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.Chain(func(t *Tunnel) *TunnelStatus {
				return t.Status
			}),
			opt.Chain(func(t *TunnelStatus) *TunnelStatus {
				return t
			}),
			opt.MapResult(func(s *TunnelStatus) string {
				return s.State
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(&system)
		assert.Equal(t, "active", result)
	})

	t.Run("with early nil in chain", func(t *testing.T) {
		system := System{} // system.Config.Manager is nil

		transform := opt.Flow5(
			opt.Chain(func(s *System) *TunnelManager {
				return s.Config.Manager
			}),
			opt.Chain(func(m *TunnelManager) *Tunnel {
				return m.ActiveTunnel
			}),
			opt.Chain(func(t *Tunnel) *TunnelStatus {
				return t.Status
			}),
			opt.Chain(func(t *TunnelStatus) *TunnelStatus {
				return t
			}),
			opt.MapResult(func(s *TunnelStatus) string {
				return "should not reach here"
			}),
			opt.GetOrElse("disconnected"),
		)

		_, result := transform.Run(&system)
		assert.Equal(t, "disconnected", result)
	})
}

// Test error handling in Flow
func TestFlow_ErrorHandling(t *testing.T) {
	// This test ensures that a panic in any step is properly recovered
	// and the fallback value is returned
	input := &TunnelStatus{State: "running"}
	transform := opt.Flow1(
		opt.MapResult(func(s *TunnelStatus) string {
			panic("simulated error")
		}),
		opt.GetOrElse("recovered"),
	)

	err, result := transform.Run(input)
	assert.Error(t, err) // Error is handled by the Flow's recovery
	assert.Equal(t, "recovered", result)
}
