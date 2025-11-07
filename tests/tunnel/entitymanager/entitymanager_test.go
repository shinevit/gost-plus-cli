package entitymanager

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	mockWrapper "github.com/go-gost/gost.plus/tests/tunnel"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entitymanager"
	opt "github.com/go-gost/gost.plus/utils/fp/option"
	"github.com/go-gost/x/service"
	"github.com/stretchr/testify/assert"
)

func TestFlow1_NetworkEntityManager(t *testing.T) {
	defaultId := "fallback"

	t.Run("non-nil NetworkEntityManager", func(t *testing.T) {
		// Create a mock tunnel
		mockTunnel := tunnel.NewMockTunnel(t)
		mockTunnel.On("IsClosed").Return(false).Maybe()
		mockTunnel.On("ID").Return("test-id").Maybe()
		mockTunnel.On("Name").Return("test-tunnel").Maybe()
		mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
		mockTunnel.On("IsActive").Return(true).Maybe()
		mockTunnel.On("Status").Return(&service.Status{}).Maybe()

		// Create a NetworkEntityManager
		tun := mockWrapper.NewEntity(mockTunnel)
		manager := &entitymanager.NetworkEntityManager{
			ID:     "test-manager-ID",
			Entity: &tun,
		}

		getIDFlow := opt.Flow1(
			opt.MapResult(func(m *entitymanager.NetworkEntityManager) string {
				return m.ID
			}),
			opt.GetOrElse(defaultId),
		)

		err, actualID := getIDFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, manager.ID, actualID)
	})

	t.Run("nil NetworkEntityManager", func(t *testing.T) {
		var manager *entitymanager.NetworkEntityManager = nil

		getIDFlow := opt.Flow1(
			opt.MapResult(func(m *entitymanager.NetworkEntityManager) string {
				return m.ID
			}),
			opt.GetOrElse(defaultId),
		)

		err, actualID := getIDFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, defaultId, actualID)
	})
}

func TestFlow2_NetworkEntityManager(t *testing.T) {
	defaultTunnelId := "fallback"

	t.Run("all non-nil values", func(t *testing.T) {
		// Create a mock tunnel
		mockTunnel := tunnel.NewMockTunnel(t) // *tunnel.MockTunnel
		mockTunnel.On("IsClosed").Return(false).Maybe()
		mockTunnel.On("ID").Return("test-tunnel-id").Maybe()
		mockTunnel.On("Name").Return("test-tunnel").Maybe()
		mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
		mockTunnel.On("IsActive").Return(true).Maybe()
		mockTunnel.On("Status").Return(&service.Status{}).Maybe()

		// Create a NetworkEntityManager
		tun := mockWrapper.NewEntity(mockTunnel)
		manager := &entitymanager.NetworkEntityManager{
			ID:     "test-manager",
			Entity: &tun,
		}

		getTunnelIDFlow := opt.Flow2(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.MapResult(func(tun *tunnel.Tunnel) string {
				return (*tun).ID()
			}),
			opt.GetOrElse(defaultTunnelId),
		)

		err, actualTunnelID := getTunnelIDFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, mockTunnel.ID(), actualTunnelID)
	})

	t.Run("entity nil", func(t *testing.T) {
		// Create a mock tunnel
		mockTunnel := tunnel.NewMockTunnel(t)
		mockTunnel.On("IsClosed").Return(false).Maybe()
		mockTunnel.On("ID").Return("test-tunnel-id").Maybe()
		mockTunnel.On("Name").Return("test-tunnel").Maybe()
		mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
		mockTunnel.On("IsActive").Return(true).Maybe()
		mockTunnel.On("Status").Return(&service.Status{}).Maybe()

		// Create a NetworkEntityManager
		manager := &entitymanager.NetworkEntityManager{
			ID:     "test-manager",
			Entity: nil,
		}

		getTunnelIDFlow := opt.Flow2(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.MapResult(func(tun *tunnel.Tunnel) string {
				return (*tun).ID()
			}),
			opt.GetOrElse(defaultTunnelId),
		)

		err, actualTunnelID := getTunnelIDFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, defaultTunnelId, actualTunnelID)
	})

	t.Run("nil manager", func(t *testing.T) {
		var manager *entitymanager.NetworkEntityManager = nil

		getTunnelIDFlow := opt.Flow2(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.MapResult(func(tun *tunnel.Tunnel) string {
				return (*tun).ID()
			}),
			opt.GetOrElse(defaultTunnelId),
		)

		err, actualTunnelID := getTunnelIDFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, defaultTunnelId, actualTunnelID)
	})
}

func TestFlow3_NetworkEntityManager(t *testing.T) {
	defaultState := service.State("")

	t.Run("all non-nil values", func(t *testing.T) {
		// Create a mock tunnel
		mockTunnel := tunnel.NewMockTunnel(t)
		mockTunnel.On("IsClosed").Return(false).Maybe()
		mockTunnel.On("ID").Return("test-id").Maybe()
		mockTunnel.On("Name").Return("test-tunnel").Maybe()
		mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
		mockTunnel.On("IsActive").Return(true).Maybe()
		mockTunnel.On("Status").Return(&service.Status{}).Maybe()

		// Create a NetworkEntityManager
		tun := mockWrapper.NewEntity(mockTunnel)
		manager := &entitymanager.NetworkEntityManager{
			ID:     "test-manager",
			Entity: &tun,
		}

		getStateFlow := opt.Flow3(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.Chain(func(tun *tunnel.Tunnel) *service.Status {
				return (*tun).Status()
			}),
			opt.MapResult(func(st *service.Status) service.State {
				return st.State()
			}),
			opt.GetOrElse(defaultState),
		)

		err, actualState := getStateFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, defaultState, actualState)
	})

	t.Run("nil manager", func(t *testing.T) {
		var manager *entitymanager.NetworkEntityManager = nil

		getStateFlow := opt.Flow3(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.Chain(func(tun *tunnel.Tunnel) *service.Status {
				return (*tun).Status()
			}),
			opt.MapResult(func(st *service.Status) service.State { // Map: A => B
				return st.State()
			}),
			opt.GetOrElse(defaultState),
		)

		err, actualState := getStateFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, defaultState, actualState)
	})
}

func TestFlow3_NetworkEntityManager_WithNilStatus(t *testing.T) {
	defaultStateStr := string(service.StateFailed)

	t.Run("all non-nil values", func(t *testing.T) {
		mockTunnel := tunnel.NewMockTunnel(t)
		mockTunnel.On("IsClosed").Return(false).Maybe()
		mockTunnel.On("ID").Return("test-id").Maybe()
		mockTunnel.On("Name").Return("test-tunnel").Maybe()
		mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
		mockTunnel.On("IsActive").Return(true).Maybe()
		mockTunnel.On("Status").Return(&service.Status{}).Maybe()

		// Create a NetworkEntityManager
		tun := mockWrapper.NewEntity(mockTunnel)
		manager := &entitymanager.NetworkEntityManager{
			ID:     "test-manager-id",
			Entity: &tun,
		}

		getStateFlow := opt.Flow3(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.Chain(func(tun *tunnel.Tunnel) *service.Status {
				return (*tun).Status()
			}),
			opt.MapResult(func(st *service.Status) string {
				return string(st.State())
			}),
			opt.GetOrElse(defaultStateStr),
		)

		err, actualStateStr := getStateFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, "", actualStateStr)
	})

	t.Run("nil value of State", func(t *testing.T) {
		mockTunnel := tunnel.NewMockTunnel(t)
		mockTunnel.On("IsClosed").Return(false).Maybe()
		mockTunnel.On("ID").Return("test-id").Maybe()
		mockTunnel.On("Name").Return("test-tunnel").Maybe()
		mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
		mockTunnel.On("IsActive").Return(true).Maybe()
		mockTunnel.On("Status").Return(nil).Maybe()

		// Create a NetworkEntityManager
		tun := mockWrapper.NewEntity(mockTunnel)
		manager := &entitymanager.NetworkEntityManager{
			ID:     "test-manager-id",
			Entity: &tun,
		}

		getStateFlow := opt.Flow3(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.Chain(func(tun *tunnel.Tunnel) *service.Status {
				return (*tun).Status()
			}),
			opt.MapResult(func(st *service.Status) string {
				return string(st.State())
			}),
			opt.GetOrElse(defaultStateStr),
		)

		err, actualStateStr := getStateFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, "failed", actualStateStr)
	})

	t.Run("nil manager", func(t *testing.T) {
		var manager *entitymanager.NetworkEntityManager = nil

		getStateFlow := opt.Flow3(
			opt.Chain(func(m *entitymanager.NetworkEntityManager) *tunnel.Tunnel {
				return m.Entity
			}),
			opt.Chain(func(tun *tunnel.Tunnel) *service.Status {
				return (*tun).Status()
			}),
			opt.MapResult(func(st *service.Status) string {
				return string(st.State())
			}),
			opt.GetOrElse(defaultStateStr),
		)

		err, actualStateStr := getStateFlow.Run(manager)

		assert.Nil(t, err)
		assert.Equal(t, "failed", actualStateStr)
	})
}

func TestDeleteEntity(t *testing.T) {
	tests := []struct {
		name           string
		entityManager  entitymanager.NetworkEntityManager
		label          string
		expectedOutput string
		expectedError  bool
		setupMocks     func(*tunnel.MockTunnel)
		setupManager   func(*entitymanager.NetworkEntityManager, *tunnel.MockTunnel)
	}{
		{
			name: "successful deletion",
			setupMocks: func(mt *tunnel.MockTunnel) {
				// Set up all required mock methods
				mt.On("Name").Return("test-tunnel").Maybe()
				mt.On("ID").Return("test-id").Maybe()
				mt.On("Type").Return("tcp").Maybe()
				mt.On("Endpoint").Return("localhost:8080").Maybe()
			},
			setupManager: func(em *entitymanager.NetworkEntityManager, mt *tunnel.MockTunnel) {
				if mt != nil {
					tun := mockWrapper.NewEntity(mt)
					em.Entity = &tun
				}
			},
			entityManager: entitymanager.NetworkEntityManager{
				ID: "test-id",
				Delete: func(id string) {
					assert.Equal(t, "test-id", id, "Delete should be called with correct ID")
				},
				SaveConfig: func() error { return nil },
			},
			label:          "Tunnel",
			expectedOutput: "Tunnel 'test-tunnel' (ID: test-id) has been deleted.\n",
			expectedError:  false,
		},
		{
			name:       "nil entity",
			setupMocks: nil,
			setupManager: func(em *entitymanager.NetworkEntityManager, mt *tunnel.MockTunnel) {
				em.Entity = nil
			},
			entityManager: entitymanager.NetworkEntityManager{
				ID: "test-id",
				Delete: func(id string) {
					t.Fatal("Delete should not be called when entity is nil")
				},
				SaveConfig: func() error {
					t.Fatal("SaveConfig should not be called when entity is nil")
					return nil
				},
			},
			label:          "Tunnel",
			expectedOutput: "",
			expectedError:  true,
		},
		{
			name: "save config error",
			setupMocks: func(mt *tunnel.MockTunnel) {
				// Set up all required mock methods
				mt.On("Name").Return("test-tunnel").Maybe()
				mt.On("ID").Return("test-id").Maybe()
				mt.On("Type").Return("tcp").Maybe()
				mt.On("Endpoint").Return("localhost:8080").Maybe()
			},
			setupManager: func(em *entitymanager.NetworkEntityManager, mt *tunnel.MockTunnel) {
				if mt != nil {
					tun := mockWrapper.NewEntity(mt)
					em.Entity = &tun
				}
			},
			entityManager: entitymanager.NetworkEntityManager{
				ID: "test-id",
				Delete: func(id string) {
					assert.Equal(t, "test-id", id, "Delete should be called with correct ID")
				},
				SaveConfig: func() error {
					return fmt.Errorf("config save error")
				},
			},
			label:          "Tunnel",
			expectedOutput: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock tunnel if needed
			var mockTunnel *tunnel.MockTunnel
			if tt.setupMocks != nil {
				mockTunnel = tunnel.NewMockTunnel(t)
				tt.setupMocks(mockTunnel)
			}

			// Create a copy of the entity manager for this test case
			em := tt.entityManager

			// Setup the manager if needed
			if tt.setupManager != nil {
				tt.setupManager(&em, mockTunnel)
			}

			// Redirect stdout to capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the method
			err := em.DeleteEntity(tt.label)

			// Restore stdout
			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Get the captured output
			output := buf.String()

			// Check the output
			if tt.name == "save config error" {
				// For save config error, we expect an error message in the output
				expectedErrorMsg := "Error while saving config after deletion: config save error"
				assert.Contains(t, output, expectedErrorMsg, "Error message should be in the output")
			} else if tt.expectedOutput != "" {
				assert.Contains(t, output, tt.expectedOutput, "Unexpected output")
			} else if output != "" {
				assert.Empty(t, output, "Expected no output but got: %s", output)
			}

			// Check the error
			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none")
				switch tt.name {
				case "nil entity":
					assert.Contains(t, err.Error(), "not found", "Error message should indicate entity not found")
				case "save config error":
					assert.Contains(t, err.Error(), "Error while saving config after deletion", "Error message should indicate config save error")
				}
			} else {
				assert.NoError(t, err, "Unexpected error")
			}
		})
	}
}
