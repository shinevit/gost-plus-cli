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

func TestMakeEntityIterator(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		getItem     entitymanager.EntityGetterFn
		expectedIDs []string
		expectEmpty bool
	}{
		{
			name:        "empty iterator",
			count:       0,
			getItem:     func(i int) tunnel.Tunnel { return nil },
			expectedIDs: []string{},
			expectEmpty: true,
		},
		{
			name:  "single item",
			count: 1,
			getItem: func(i int) tunnel.Tunnel {
				if i == 0 {
					mockTunnel := tunnel.NewMockTunnel(t)
					mockTunnel.On("ID").Return("tunnel-1").Once()
					return mockTunnel
				}
				return nil
			},
			expectedIDs: []string{"tunnel-1"},
			expectEmpty: false,
		},
		{
			name:  "multiple items with nil",
			count: 5,
			getItem: func(i int) tunnel.Tunnel {
				// Return nil for even indices
				if i%2 == 0 {
					return nil
				}
				mockTunnel := tunnel.NewMockTunnel(t)
				mockTunnel.On("ID").Return(fmt.Sprintf("tunnel-%d", i)).Once()
				return mockTunnel
			},
			expectedIDs: []string{"tunnel-1", "tunnel-3"},
			expectEmpty: false,
		},
		{
			name:        "nil expected IDs",
			count:       2,
			getItem:     func(i int) tunnel.Tunnel { return nil },
			expectedIDs: nil,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the iterator
			iterator := entitymanager.MakeEntityIterator(tt.count, tt.getItem)

			// Track the entities we get from the iterator
			var entities []entitymanager.NetworkEntityManager
			var ids []string

			// Iterate through all items
			for {
				entity, hasMore := iterator()
				if !hasMore {
					break
				}
				entities = append(entities, entity)
				ids = append(ids, entity.ID)
			}

			// Check if we got the expected number of items
			if tt.expectEmpty {
				assert.True(t, len(ids) == 0, "Expected no items but got %d", len(ids))
			} else {
				if tt.expectedIDs == nil {
					tt.expectedIDs = []string{}
				}
				assert.Equal(t, len(tt.expectedIDs), len(ids), "Unexpected number of items")
				for i, id := range tt.expectedIDs {
					assert.Equal(t, id, ids[i], "Unexpected ID at index %d", i)
				}
			}

			// Verify that the iterator returns false when done
			_, hasMore := iterator()
			assert.False(t, hasMore, "iterator should be exhausted")
		})
	}
}

func TestMakeEntityIteratorFrom(t *testing.T) {
	t.Run("create iterator from source", func(t *testing.T) {
		// Setup test data
		tunnel1 := tunnel.NewMockTunnel(t)
		tunnel1.On("ID").Return("tunnel-1")

		tunnel2 := tunnel.NewMockTunnel(t)
		tunnel2.On("ID").Return("tunnel-2")

		tunnels := []tunnel.Tunnel{tunnel1, tunnel2}

		// Create source
		source := entitymanager.EntitySource{
			Count: len(tunnels),
			GetEntity: func(i int) tunnel.Tunnel {
				if i < len(tunnels) {
					return tunnels[i]
				}
				return nil
			},
		}

		// Create iterator from source
		iterator := entitymanager.MakeEntityIteratorFrom(source)

		// Test iteration
		entity1, hasMore1 := iterator()
		assert.True(t, hasMore1)
		assert.Equal(t, "tunnel-1", entity1.ID)

		entity2, hasMore2 := iterator()
		assert.True(t, hasMore2)
		assert.Equal(t, "tunnel-2", entity2.ID)

		// Should be no more items
		_, hasMore3 := iterator()
		assert.False(t, hasMore3)

		// Verify all expected calls were made
		tunnel1.AssertExpectations(t)
		tunnel2.AssertExpectations(t)
	})
}
