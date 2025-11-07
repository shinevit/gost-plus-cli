package utils

import (
	"testing"

	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/utils"
	"github.com/go-gost/x/service"
	xservice "github.com/go-gost/x/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stateValidator func(t *testing.T, result xservice.State, expectedStatus *xservice.Status)
type displayStateValidator func(t *testing.T, result string, expectedStatus *xservice.Status)

func TestGetDisplayState(t *testing.T) {
	validateExactDisplayState := func(expected string) displayStateValidator {
		return func(t *testing.T, result string, _ *xservice.Status) {
			assert.Equal(t, expected, result, "Unexpected display state returned")
		}
	}

	tests := []struct {
		name       string
		setupMocks func(*tunnel.MockTunnel) *xservice.Status
		validate   displayStateValidator
	}{
		{
			name: "WhenClosed_ReturnsUppercaseClosed",
			setupMocks: func(mt *tunnel.MockTunnel) *xservice.Status {
				mt.On("IsClosed").Return(true).Once()
				status := (*xservice.Status)(nil)
				mt.On("Status").Return(status).Once()
				return status
			},
			validate: validateExactDisplayState("CLOSED"),
		},
		{
			name: "WhenNotClosedAndNoStatus_ReturnsUppercaseReady",
			setupMocks: func(mt *tunnel.MockTunnel) *xservice.Status {
				mt.On("IsClosed").Return(false).Once()
				status := (*xservice.Status)(nil)
				mt.On("Status").Return(status).Once()
				return status
			},
			validate: validateExactDisplayState("READY"),
		},
		{
			name: "WhenNotClosedAndHasStatus_ReturnsUppercaseStatusState",
			setupMocks: func(mt *tunnel.MockTunnel) *xservice.Status {
				mt.On("IsClosed").Return(false).Once()
				status := &xservice.Status{}
				mt.On("Status").Return(status).Once()
				return status
			},
			validate: func(t *testing.T, result string, _ *xservice.Status) {
				// For an empty status, we expect an empty string
				assert.Equal(t, "", result, "Display state should be empty for empty status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTunnel := tunnel.NewMockTunnel(t)

			// Setup common mocks
			tunnelID := uuid.New().String()
			mockTunnel.On("ID").Return(tunnelID).Maybe()
			mockTunnel.On("Name").Return("mock-tunnel").Maybe()
			mockTunnel.On("Type").Return(tunnel.HTTPTunnel).Maybe()
			mockTunnel.On("Close").Return(nil).Maybe()
			mockTunnel.On("Options").Return(tunnel.Options{}).Maybe()

			expectedStatus := tt.setupMocks(mockTunnel)
			tunnel.Add(mockTunnel)

			result := utils.GetDisplayState(mockTunnel)
			tt.validate(t, result, expectedStatus)

			mockTunnel.AssertExpectations(t)
		})
	}
}

func TestGetState(t *testing.T) {
	validateExactState := func(expected xservice.State) stateValidator {
		return func(t *testing.T, result xservice.State, _ *xservice.Status) {
			assert.Equal(t, expected, result, "Unexpected state returned")
		}
	}

	validateNonErrorState := func(t *testing.T, result xservice.State, _ *xservice.Status) {
		assert.NotEqual(t, service.StateFailed, result, "Should not return failed state when status exists")
		assert.NotEqual(t, service.StateClosed, result, "Should not return closed state when not closed")
	}

	tests := []struct {
		name       string
		setupMocks func(*tunnel.MockTunnel) *xservice.Status
		validate   stateValidator
	}{
		{
			name: "WhenClosed_ReturnsClosedState",
			setupMocks: func(mt *tunnel.MockTunnel) *xservice.Status {
				mt.On("IsClosed").Return(true).Once()
				status := (*xservice.Status)(nil)
				mt.On("Status").Return(status).Once()
				return status
			},
			validate: validateExactState(service.StateClosed),
		},
		{
			name: "WhenNotClosedAndNilStatus_ReturnsReadyState",
			setupMocks: func(mt *tunnel.MockTunnel) *xservice.Status {
				mt.On("IsClosed").Return(false).Once()
				status := (*xservice.Status)(nil)
				mt.On("Status").Return(status).Once()
				return status
			},
			validate: validateExactState(service.StateReady),
		},
		{
			name: "WhenNotClosedAndHasEmptyStatus_ReturnsStatusState",
			setupMocks: func(mt *tunnel.MockTunnel) *xservice.Status {
				mt.On("IsClosed").Return(false).Once()
				status := &xservice.Status{}
				mt.On("Status").Return(status).Once()
				return status
			},
			validate: validateNonErrorState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTunnel := tunnel.NewMockTunnel(t)

			// Setup common mocks
			tunnelID := uuid.New().String()
			mockTunnel.On("ID").Return(tunnelID).Maybe()
			mockTunnel.On("Name").Return("mock-tunnel").Maybe()
			mockTunnel.On("Type").Return(tunnel.HTTPTunnel).Maybe()
			mockTunnel.On("Close").Return(nil).Maybe()
			mockTunnel.On("Options").Return(tunnel.Options{}).Maybe()

			expectedStatus := tt.setupMocks(mockTunnel)
			tunnel.Add(mockTunnel)

			result := utils.GetState(mockTunnel)
			tt.validate(t, result, expectedStatus)

			mockTunnel.AssertExpectations(t)
		})
	}
}
