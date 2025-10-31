package task_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-gost/gost.plus/runner/task"
	mocks "github.com/go-gost/gost.plus/tests/mocks"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/utils"
	"github.com/go-gost/x/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	xgoMock "github.com/xhd2015/xgo/runtime/mock"
)

func TestMonitorTunnelsTask_Run_ActiveHttpOrFileTunnel(t *testing.T) {
	tests := []struct {
		name           string
		tunnelType     string
		httpStatus     int
		expectedActive bool
	}{
		{
			name:           "Active http tunnel for success status",
			tunnelType:     tunnel.HTTPTunnel,
			httpStatus:     http.StatusOK,
			expectedActive: true,
		},
		{
			name:           "Active http tunnel if status is unauthorized",
			tunnelType:     tunnel.HTTPTunnel,
			httpStatus:     http.StatusUnauthorized,
			expectedActive: true,
		},
		{
			name:           "Active http tunnel if 301 status",
			tunnelType:     tunnel.HTTPTunnel,
			httpStatus:     http.StatusMovedPermanently,
			expectedActive: true,
		},
		{
			name:           "Active file tunnel for success status",
			tunnelType:     tunnel.FileTunnel,
			httpStatus:     http.StatusOK,
			expectedActive: true,
		},
		{
			name:           "Active file tunnel if status is unauthorized",
			tunnelType:     tunnel.FileTunnel,
			httpStatus:     http.StatusUnauthorized,
			expectedActive: true,
		},
		{
			name:           "Active file tunnel if 301 status",
			tunnelType:     tunnel.FileTunnel,
			httpStatus:     http.StatusMovedPermanently,
			expectedActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLogger := mocks.NewMockLogger(t)
			expectedHttpStatus := tt.httpStatus // 200, 301, 401
			mockServer := mocks.MockHTTPServer(expectedHttpStatus)
			defer mockServer.Close()

			mockLogger.On("Infof", mock.Anything, mock.Anything).Return()

			tunnelName := "test-tunnel-" + tt.tunnelType
			mockTunnel := tunnel.NewMockTunnel(t)
			mockTunnel.On("IsClosed").Return(false)
			mockTunnel.On("IsActive").Return(tt.expectedActive).Maybe()
			mockTunnel.On("Entrypoint").Return(mockServer.URL)
			mockTunnel.On("Type").Return(tt.tunnelType)
			mockTunnel.On("Name").Return(tunnelName)
			mockTunnel.On("ID").Return("test-tunnel-id").Maybe()
			mockTunnel.On("Close").Return(nil).Maybe()
			mockTunnel.On("Options").Return(nil).Maybe()
			// Don't mock Status() cause it has a non-exportable state

			xgoMock.Patch(tunnel.GetAll, func() []tunnel.Tunnel {
				return []tunnel.Tunnel{mockTunnel}
			})

			// Act
			monitorTask := task.NewMonitorTaskWith(mockLogger)
			err := monitorTask.Run(context.Background())

			// Assert
			assert.NoError(t, err, "Monitor task should run without error")
			assert.Equal(t, "service.tunnel.monitor", string(monitorTask.ID()), "Task ID should match its expected value")
			mockLogger.AssertExpectations(t)
			mockLogger.AssertCalled(t, "Infof", "Checking tunnel %s HTTP connection at %s", []any{tunnelName, mockServer.URL})
			mockLogger.AssertCalled(t, "Infof", "Tunnel %s HTTP connectivity check succeeded: %s", []any{tunnelName, fmt.Sprintf("%d %s", expectedHttpStatus, http.StatusText(expectedHttpStatus))})
			mockTunnel.AssertNotCalled(t, "Close", "Close should not be called")
			mockTunnel.AssertExpectations(t)
		})
	}
}

func TestMonitorTunnelsTask_Run_Should_Restart_FailedHTTPLikeTunnel(t *testing.T) {
	testCases := []struct {
		name       string
		tunnelType string
		httpStatus int
		message    string
	}{
		{
			name:       "Inactive http tunnel with 403 Forbidden status",
			tunnelType: tunnel.HTTPTunnel,
			httpStatus: http.StatusForbidden,
			message:    "Tunnel %s HTTP connectivity check failed: %s",
		},
		{
			name:       "Inactive http tunnel with 503 Service Unavailable status",
			tunnelType: tunnel.HTTPTunnel,
			httpStatus: http.StatusServiceUnavailable,
			message:    "Tunnel '%s' does not exist. Http status: %s",
		},
		{
			name:       "Inactive http tunnel with 502 Bad Gateway status",
			tunnelType: tunnel.HTTPTunnel,
			httpStatus: http.StatusBadGateway,
			message:    "Tunnel '%s' does not exist. Http status: %s",
		},
		{
			name:       "Inactive file tunnel with 403 Forbidden status",
			tunnelType: tunnel.FileTunnel,
			httpStatus: http.StatusForbidden,
			message:    "Tunnel %s HTTP connectivity check failed: %s",
		},
		{
			name:       "Inactive file tunnel with 503 Service Unavailable status",
			tunnelType: tunnel.FileTunnel,
			httpStatus: http.StatusServiceUnavailable,
			message:    "Tunnel '%s' does not exist. Http status: %s",
		},
		{
			name:       "Inactive file tunnel with 502 Bad Gateway status",
			tunnelType: tunnel.FileTunnel,
			httpStatus: http.StatusBadGateway,
			message:    "Tunnel '%s' does not exist. Http status: %s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockLogger := mocks.NewMockLogger(t)
			expectedHttpStatus := tc.httpStatus
			mockServer := mocks.MockHTTPServer(expectedHttpStatus)
			defer mockServer.Close()

			mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
			mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()

			tunnelName := "test-tunnel-" + tc.tunnelType
			mockTunnel := tunnel.NewMockTunnel(t)
			mockTunnel.On("ID").Return("test-tunnel-id")
			mockTunnel.On("Name").Return(tunnelName)
			mockTunnel.On("Type").Return(tc.tunnelType)
			mockTunnel.On("Options").Return(tunnel.Options{}).Maybe()
			mockTunnel.On("Entrypoint").Return(mockServer.URL)
			mockTunnel.On("IsClosed").Return(false).Maybe()
			mockTunnel.On("IsActive").Return(false).Maybe()
			mockTunnel.On("Status").Return(nil).Maybe()
			mockTunnel.On("Close").Return(nil).Maybe()
			mockTunnel.On("Run").Return(nil).Maybe()

			xgoMock.Patch(tunnel.GetAll, func() []tunnel.Tunnel {
				return []tunnel.Tunnel{mockTunnel}
			})
			xgoMock.Patch(tunnel.CreateTunnel, func(st string, opts tunnel.Options) tunnel.Tunnel {
				return mockTunnel
			})
			xgoMock.Patch(tunnel.Set, func(tunnel.Tunnel) {})

			// Act
			monitorTask := task.NewMonitorTaskWith(mockLogger)
			err := monitorTask.Run(context.Background())

			// Assert
			assert.NoError(t, err, "Monitor task should run without error")
			mockLogger.AssertExpectations(t)
			mockLogger.AssertCalled(t, "Infof", "Checking tunnel %s HTTP connection at %s", []any{tunnelName, mockServer.URL})
			mockLogger.AssertCalled(t, "Warnf", tc.message, []any{tunnelName, fmt.Sprintf("%d %s", expectedHttpStatus, http.StatusText(expectedHttpStatus))})
			mockLogger.AssertCalled(t, "Infof", "Detected inactive tunnel '%s' (%s), attempting to reconnect...", []any{tunnelName, "test-tunnel-id"})
			mockLogger.AssertCalled(t, "Infof", "Attempting to restart the tunnel ID: %s with the same options...", []any{"test-tunnel-id"})
			mockLogger.AssertCalled(t, "Infof", "Successfully closed old tunnel %s", []any{"test-tunnel-id"})
			mockLogger.AssertCalled(t, "Infof", "Successfully restarted tunnel %s with same ID and options", []any{"test-tunnel-id"})
			mockTunnel.AssertExpectations(t)
		})
	}
}

func TestMonitorTunnelsTask_Run_Should_Skip_Any_Closed_Tunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return().Once()

	tunnelName := "test-any-tunnel"
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("Name").Return(tunnelName).Once()
	mockTunnel.On("IsClosed").Return(true).Once()

	xgoMock.Patch(tunnel.GetAll, func() []tunnel.Tunnel {
		return []tunnel.Tunnel{mockTunnel}
	})

	// Act
	monitorTask := task.NewMonitorTaskWith(mockLogger)
	err := monitorTask.Run(context.Background())

	// Assert
	assert.NoError(t, err, "Monitor task should run without error")
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Infof", "Skipping closed tunnel '%s'", []any{tunnelName})
	mockTunnel.AssertExpectations(t)
}

func TestMonitorTunnelsTask_Run_Should_Restart_InactiveTCPTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return().Maybe()

	expectedTunnelID := "test-uid"
	expectedTunnelName := "test-tunnel"
	expectedState := service.StateFailed

	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("ID").Return(expectedTunnelID).Maybe()
	mockTunnel.On("Name").Return(expectedTunnelName).Maybe()
	mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
	mockTunnel.On("IsClosed").Return(false).Maybe()
	mockTunnel.On("Status").Return(nil).Maybe()
	mockTunnel.On("Entrypoint").Return("localhost:8080").Maybe()
	mockTunnel.On("IsActive").Return(false).Maybe()
	mockTunnel.On("Close").Return(nil).Maybe()
	mockTunnel.On("Options").Return(tunnel.Options{}).Maybe()
	mockTunnel.On("Run").Return(nil).Maybe()

	xgoMock.Patch(tunnel.GetAll, func() []tunnel.Tunnel {
		return []tunnel.Tunnel{mockTunnel}
	})

	xgoMock.Patch(utils.GetState, func(tun tunnel.Tunnel) service.State {
		return expectedState
	})
	xgoMock.Patch(tunnel.CreateTunnel, func(st string, opts tunnel.Options) tunnel.Tunnel {
		return mockTunnel
	})
	xgoMock.Patch(tunnel.Set, func(tunnel.Tunnel) {})

	// Act
	monitorTask := task.NewMonitorTaskWith(mockLogger)
	err := monitorTask.Run(context.Background())

	// Assert
	assert.NoError(t, err, "Monitor task should run without error")
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Infof", "%s Tunnel '%s' seems inactive (service %v)", []any{"TCP", expectedTunnelName, expectedState})
	mockLogger.AssertCalled(t, "Infof", "Detected inactive tunnel '%s' (%s), attempting to reconnect...", []any{expectedTunnelName, expectedTunnelID})
	mockLogger.AssertCalled(t, "Infof", "Attempting to restart the tunnel ID: %s with the same options...", []any{expectedTunnelID})
	mockLogger.AssertCalled(t, "Infof", "Successfully closed old tunnel %s", []any{expectedTunnelID})
	mockLogger.AssertCalled(t, "Infof", "Successfully restarted tunnel %s with same ID and options", []any{expectedTunnelID})
	mockTunnel.AssertExpectations(t)
}

func TestMonitorTunnelsTask_Run_Should_Not_Be_Restared_ActiveTCPTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return().Once()

	expectedTunnelID := "test-uid"
	expectedTunnelName := "test-tunnel"
	expectedState := service.StateRunning

	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("ID").Return(expectedTunnelID).Maybe()
	mockTunnel.On("Name").Return(expectedTunnelName).Maybe()
	mockTunnel.On("Type").Return(tunnel.TCPTunnel).Maybe()
	mockTunnel.On("IsClosed").Return(false).Maybe()
	mockTunnel.On("Status").Return(nil).Maybe()
	mockTunnel.On("Entrypoint").Return("localhost:8080").Maybe()
	mockTunnel.On("IsActive").Return(true).Maybe()
	mockTunnel.On("Close").Return(nil).Maybe()
	mockTunnel.On("Options").Return(tunnel.Options{}).Maybe()
	mockTunnel.On("Run").Return(nil).Maybe()

	xgoMock.Patch(tunnel.GetAll, func() []tunnel.Tunnel {
		return []tunnel.Tunnel{mockTunnel}
	})

	xgoMock.Patch(utils.GetState, func(tun tunnel.Tunnel) service.State {
		return expectedState
	})
	xgoMock.Patch(tunnel.CreateTunnel, func(st string, opts tunnel.Options) tunnel.Tunnel {
		return mockTunnel
	})
	xgoMock.Patch(tunnel.Set, func(tunnel.Tunnel) {})

	// Act
	monitorTask := task.NewMonitorTaskWith(mockLogger)
	err := monitorTask.Run(context.Background())

	// Assert
	assert.NoError(t, err, "Monitor task should run without error")
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Infof", "%s Tunnel '%s' is active (service %v)", []any{"TCP", expectedTunnelName, expectedState})
	mockTunnel.AssertExpectations(t)
}
