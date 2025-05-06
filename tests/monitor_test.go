package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-gost/gost.plus/runner/task"
	"github.com/go-gost/gost.plus/tests/mocks"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	xgoMock "github.com/xhd2015/xgo/runtime/mock"
)

func TestMonitorTunnelsTask_Run_ActiveHTTPTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	expectedHttpStatus := http.StatusForbidden
	mockServer := mockHTTPServer(expectedHttpStatus) // or http.StatusOK
	defer mockServer.Close()

	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()

	// Create an active mock tunnel
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("IsClosed").Return(false)
	mockTunnel.On("Entrypoint").Return(mockServer.URL)
	mockTunnel.On("Type").Return(tunnel.HTTPTunnel)
	mockTunnel.On("Name").Return("test-tunnel")

	xgoMock.Patch(tunnel.Count, func() int {
		return 1
	})
	xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
		return mockTunnel
	})

	// Act
	monitorTask := task.NewMonitorTaskWith(mockLogger)
	err := monitorTask.Run(context.Background())

	// Assert
	assert.NoError(t, err, "Monitor task should run without error")
	assert.Equal(t, "service.tunnel.monitor", string(monitorTask.ID()), "Task ID should match its expected value")
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Infof", "Checking tunnel %s connection at %s", []any{"test-tunnel", mockServer.URL})
	mockLogger.AssertCalled(t, "Infof", "Tunnel '%s' connectivity is succeeded: http status: %s", []any{"test-tunnel", fmt.Sprintf("%d Forbidden", expectedHttpStatus)})
	mockTunnel.AssertNotCalled(t, "Close", "Close should not be called")
	mockTunnel.AssertExpectations(t)
}

// table-driven approach
func TestMonitorTunnelsTask_Run_ActiveFileTunnel(t *testing.T) {
	testCases := []struct {
		name       string
		httpStatus int
	}{
		{
			name:       "File tunnel with 403 Forbidden status",
			httpStatus: http.StatusForbidden,
		},
		{
			name:       "File tunnel with 200 OK status",
			httpStatus: http.StatusOK,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockLogger := mocks.NewMockLogger(t)
			expectedHttpStatus := tc.httpStatus
			mockServer := mockHTTPServer(expectedHttpStatus)
			defer mockServer.Close()

			mockLogger.On("Infof", mock.Anything, mock.Anything).Return()

			// Create an active mock tunnel
			mockTunnel := tunnel.NewMockTunnel(t)
			mockTunnel.On("IsClosed").Return(false)
			mockTunnel.On("Entrypoint").Return(mockServer.URL)
			mockTunnel.On("Type").Return(tunnel.FileTunnel)
			mockTunnel.On("Name").Return("test-tunnel")

			xgoMock.Patch(tunnel.Count, func() int {
				return 1
			})
			xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
				return mockTunnel
			})

			// Act
			monitorTask := task.NewMonitorTaskWith(mockLogger)
			err := monitorTask.Run(context.Background())

			// Assert
			assert.NoError(t, err, "Monitor task should run without error")
			mockLogger.AssertExpectations(t)
			mockLogger.AssertCalled(t, "Infof", "Checking tunnel %s connection at %s", []any{"test-tunnel", mockServer.URL})
			mockLogger.AssertCalled(t, "Infof", "Tunnel '%s' connectivity is succeeded: http status: %s", []any{
				"test-tunnel",
				fmt.Sprintf("%d %s", expectedHttpStatus, http.StatusText(expectedHttpStatus)),
			})
			mockTunnel.AssertNotCalled(t, "Close", "Close should not be called")
			mockTunnel.AssertExpectations(t)
		})
	}
}

func TestMonitorTunnelsTask_Run_InactiveHTTPTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	expectedHttpStatus := http.StatusBadGateway
	mockServer := mockHTTPServer(expectedHttpStatus)
	defer mockServer.Close()

	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return()

	expectedTunnelID := "test-uid"
	expectedTunnelName := "test-tunnel"
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("IsClosed").Return(false)
	mockTunnel.On("Entrypoint").Return(mockServer.URL)
	mockTunnel.On("Type").Return(tunnel.HTTPTunnel)
	mockTunnel.On("Name").Return(expectedTunnelName)
	mockTunnel.On("ID").Return(expectedTunnelID)
	mockTunnel.On("Close").Return(nil).Once()
	mockTunnel.On("Options").Return(tunnel.Options{})
	mockTunnel.On("Run").Return(nil)

	xgoMock.Patch(tunnel.Count, func() int {
		return 1
	})
	xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
		return mockTunnel
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
	mockLogger.AssertCalled(t, "Infof", "Checking tunnel %s connection at %s", []any{"test-tunnel", mockServer.URL})
	mockLogger.AssertCalled(t, "Warnf", "Tunnel '%s' does not exist. Http status: %s, reason: %v", []any{"test-tunnel", fmt.Sprintf("%d Bad Gateway", expectedHttpStatus), nil})
	mockLogger.AssertCalled(t, "Infof", "Done. Tunnel %s has been recreated with same ID and options", []any{expectedTunnelID})
	mockLogger.AssertCalled(t, "Infof", "Successfully restarted tunnel '%s'", []any{expectedTunnelName})
	mockTunnel.AssertExpectations(t)
}

func TestMonitorTunnelsTask_Run_On_ClosedTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("Info", mock.Anything).Return()

	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("IsClosed").Return(true)

	xgoMock.Patch(tunnel.Count, func() int {
		return 1
	})
	xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
		return mockTunnel
	})

	// Act
	monitorTask := task.NewMonitorTaskWith(mockLogger)
	err := monitorTask.Run(context.Background())

	// Assert
	assert.NoError(t, err, "Monitor task should run without error")
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Info", []any{"Skipping intentionally closed tunnel..."})
	mockTunnel.AssertExpectations(t)
}

func TestMonitorTunnelsTask_Run_InactiveTCPTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()

	expectedTunnelID := "test-uid"
	expectedTunnelName := "test-tunnel"
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("ID").Return(expectedTunnelID)
	mockTunnel.On("Name").Return(expectedTunnelName)
	mockTunnel.On("IsClosed").Return(false)
	mockTunnel.On("Type").Return(tunnel.TCPTunnel)
	mockTunnel.On("Status").Return(nil)
	mockTunnel.On("Close").Return(nil).Once()
	mockTunnel.On("Options").Return(tunnel.Options{})
	mockTunnel.On("Run").Return(nil)

	xgoMock.Patch(tunnel.Count, func() int {
		return 1
	})
	xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
		return mockTunnel
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
	mockLogger.AssertCalled(t, "Infof", "Detected inactive tunnel '%s' (%s), attempting to reconnect...", []any{expectedTunnelName, expectedTunnelID})
	mockLogger.AssertCalled(t, "Infof", "Done. Tunnel %s has been recreated with same ID and options", []any{expectedTunnelID})
	mockTunnel.AssertExpectations(t)
}

func TestMonitorTunnelsTask_Run_Error_InactiveTCPTunnel(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return().Once()
	mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return().Once()

	expectedTunnelID := "test-uid"
	expectedTunnelName := "test-tunnel"
	expectedErr := errors.New("failed to start")
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("ID").Return(expectedTunnelID)
	mockTunnel.On("Name").Return(expectedTunnelName)
	mockTunnel.On("IsClosed").Return(false)
	mockTunnel.On("Type").Return(tunnel.TCPTunnel)
	mockTunnel.On("Status").Return(nil)
	mockTunnel.On("Close").Return(nil).Once()
	mockTunnel.On("Options").Return(tunnel.Options{})
	mockTunnel.On("Run").Return(expectedErr)

	xgoMock.Patch(tunnel.Count, func() int {
		return 1
	})
	xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
		return mockTunnel
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
	mockLogger.AssertCalled(t, "Infof", "Detected inactive tunnel '%s' (%s), attempting to reconnect...", []any{expectedTunnelName, expectedTunnelID})
	mockLogger.AssertNumberOfCalls(t, "Errorf", 2)
	mockTunnel.AssertExpectations(t)
}
