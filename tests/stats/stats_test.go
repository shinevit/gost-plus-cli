package stats

import (
	"bytes"
	"testing"
	"time"

	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/stats"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/stretchr/testify/assert"
	xgoMock "github.com/xhd2015/xgo/runtime/mock"
)

func setupTestEnvironment(tunnels []tunnel.Tunnel, entrypoints []entrypoint.EntryPoint) func() {
	// Setup tunnels
	xgoMock.Patch(tunnel.GetAll, func() []tunnel.Tunnel {
		// Return a copy of the tunnels slice to avoid race conditions
		result := make([]tunnel.Tunnel, len(tunnels))
		copy(result, tunnels)
		return result
	})

	// Setup entrypoints
	xgoMock.Patch(entrypoint.GetAll, func() []entrypoint.EntryPoint {
		// Return a copy of the entrypoints slice to avoid race conditions
		result := make([]entrypoint.EntryPoint, len(entrypoints))
		copy(result, entrypoints)
		return result
	})

	// Capture output
	originalWriter := stats.StdOutWriter
	stats.StdOutWriter = &bytes.Buffer{}
	return func() {
		stats.StdOutWriter = originalWriter
	}
}

func createActiveMockTunnel(t *testing.T, id, name, tunnelType string) *tunnel.MockTunnel {
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("IsClosed").Return(false).Maybe()
	mockTunnel.On("ID").Return(id).Maybe()
	mockTunnel.On("Name").Return(name).Maybe()
	mockTunnel.On("Type").Return(tunnelType).Maybe()
	mockTunnel.On("IsActive").Return(true).Maybe()
	return mockTunnel
}

func createActiveMockEntrypoint(t *testing.T, id, name, entrypointType string) *tunnel.MockTunnel {
	mockEntrypoint := tunnel.NewMockTunnel(t)
	mockEntrypoint.On("IsClosed").Return(false).Maybe()
	mockEntrypoint.On("ID").Return(id).Maybe()
	mockEntrypoint.On("Name").Return(name).Maybe()
	mockEntrypoint.On("Type").Return(entrypointType).Maybe()
	mockEntrypoint.On("IsActive").Return(true).Maybe()
	return mockEntrypoint
}

func TestDisplayStats_Shows_LatestStats_ForTunnels(t *testing.T) {
	// Arrange
	mockTunnel := createActiveMockTunnel(t, "test-tunnel-id", "test-tunnel", "http")

	// Setup mock expectations for the tunnel with updated stats
	mockTunnel.On("Stats").Return(config.ServiceStats{
		CurrentConns:    5,
		TotalConns:      15,
		OutputRateBytes: 5120,    // 5 KB/s
		InputRateBytes:  7168,    // 7 KB/s
		OutputBytes:     3145728, // 3 MB
		InputBytes:      5242880, // 5 MB
		TotalErrs:       2,
	}).Maybe() // Use Maybe() since we don't know exactly how many times it will be called

	tunnels := []tunnel.Tunnel{mockTunnel}
	cleanup := setupTestEnvironment(tunnels, nil)
	defer cleanup()

	// Act
	done := make(chan struct{})
	outputChan := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		old := stats.StdOutWriter
		stats.StdOutWriter = &buf
		defer func() { stats.StdOutWriter = old }()

		stats.DisplayStats(done, 20*time.Millisecond)
		outputChan <- buf.String()
	}()

	// Wait for stats to be displayed
	time.Sleep(100 * time.Millisecond)

	// Close the done channel to stop the stats display
	close(done)

	// Get the output
	output := <-outputChan

	// Assert
	assert.Contains(t, output, "Monitoring 1 tunnels.", "Should show correct monitoring title")
	assert.Contains(t, output, "[test-tunnel-HTTP]", "Should display tunnel name and type")

	// Check for stats in the output
	assert.Contains(t, output, "Conn: 5/15", "Should show current and total connections")
	assert.Contains(t, output, "↑ 5.00 KB/s", "Should show upload speed")
	assert.Contains(t, output, "↓ 7.00 KB/s", "Should show download speed")
	assert.Contains(t, output, "Total: ↑ 3.00 MB ↓ 5.00 MB", "Should show total upload/download")
	assert.Contains(t, output, "Err: 2", "Should show error count")
}

func TestDisplayStats_Shows_Last_Indicators(t *testing.T) {
	// Arrange
	mockTunnel := createActiveMockTunnel(t, "test-tunnel-id", "test-tunnel", "http")

	// Setup mock expectations for the tunnel
	mockTunnel.On("Stats").Return(config.ServiceStats{
		CurrentConns:    1,
		TotalConns:      1,
		OutputRateBytes: 1024,    // 1 KB/s
		InputRateBytes:  2048,    // 2 KB/s
		OutputBytes:     1048576, // 1 MB
		InputBytes:      2097152, // 2 MB
		TotalErrs:       0,
	}).Once()

	// Second call: zero rates (should show last non-zero rates)
	mockTunnel.On("Stats").Return(config.ServiceStats{
		CurrentConns:    1,
		TotalConns:      1,
		OutputRateBytes: 0,
		InputRateBytes:  0,
		OutputBytes:     1048576, // 1 MB
		InputBytes:      2097152, // 2 MB
		TotalErrs:       0,
	}).Once()

	// Allow additional calls to Stats() with zero values
	mockTunnel.On("Stats").Return(config.ServiceStats{
		CurrentConns:    1,
		TotalConns:      1,
		OutputRateBytes: 0,
		InputRateBytes:  0,
		OutputBytes:     1048576,
		InputBytes:      2097152,
		TotalErrs:       0,
	}).Maybe()

	tunnels := []tunnel.Tunnel{mockTunnel}
	cleanup := setupTestEnvironment(tunnels, nil)
	defer cleanup()

	// Act
	done := make(chan struct{})
	outputChan := make(chan string, 1)
	completed := make(chan bool, 1)

	// Create a custom writer that captures output
	var buf bytes.Buffer
	oldWriter := stats.StdOutWriter
	stats.StdOutWriter = &buf
	defer func() { stats.StdOutWriter = oldWriter }()

	// Start the display stats in a goroutine
	go func() {
		stats.DisplayStats(done, 10*time.Millisecond) // Shorter interval for faster test
		outputChan <- buf.String()
		close(completed)
	}()

	// Wait for at least one update
	time.Sleep(50 * time.Millisecond)

	// Stop the display stats
	close(done)

	// Wait for completion with timeout
	select {
	case <-completed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Test timed out waiting for DisplayStats to complete")
	}

	// Get the output
	output := ""
	select {
	case output = <-outputChan:
	default:
		output = buf.String()
	}

	// Should show the last non-zero rates with "(last)" indicator
	assert.Contains(t, output, "(last)", "Should show (last) indicator for non-zero rates")

	mockTunnel.AssertExpectations(t)
}

func TestDisplayStats_ActiveTunnelAndEntrypointCount(t *testing.T) {
	// Arrange
	// Active tunnel - use Maybe() for methods that might be called multiple times
	mockTunnel1 := createActiveMockTunnel(t, "tunnel-1", "test-tunnel", "http")
	mockTunnel1.On("IsActive").Return(true).Maybe()
	mockTunnel1.On("IsClosed").Return(false).Maybe()
	mockTunnel1.On("ID").Return("tunnel-1").Maybe()
	mockTunnel1.On("Name").Return("test-tunnel").Maybe()
	mockTunnel1.On("Type").Return("http").Maybe()
	mockTunnel1.On("Stats").Return(config.ServiceStats{
		CurrentConns:    5,
		TotalConns:      15,
		OutputRateBytes: 5120,    // 5 KB/s
		InputRateBytes:  7168,    // 7 KB/s
		OutputBytes:     3145728, // 3 MB
		InputBytes:      5242880, // 5 MB
		TotalErrs:       2,
	}).Maybe()

	// Closed tunnel
	mockTunnel2 := tunnel.NewMockTunnel(t)
	mockTunnel2.On("IsActive").Return(false).Maybe()
	mockTunnel2.On("IsClosed").Return(true).Maybe()

	// Active entrypoint
	mockEntrypoint1 := createActiveMockEntrypoint(t, "ep-1", "test-entrypoint", "tcp")
	mockEntrypoint1.On("IsActive").Return(true).Maybe()
	mockEntrypoint1.On("IsClosed").Return(false).Maybe()
	mockEntrypoint1.On("ID").Return("ep-1").Maybe()
	mockEntrypoint1.On("Name").Return("test-entrypoint").Maybe()
	mockEntrypoint1.On("Type").Return("tcp").Maybe()
	mockEntrypoint1.On("Stats").Return(config.ServiceStats{
		CurrentConns:    3,
		TotalConns:      10,
		OutputRateBytes: 2048,    // 2 KB/s
		InputRateBytes:  3072,    // 3 KB/s
		OutputBytes:     1048576, // 1 MB
		InputBytes:      2097152, // 2 MB
		TotalErrs:       1,
	}).Maybe()

	// Closed entrypoint
	mockEntrypoint2 := tunnel.NewMockTunnel(t)
	mockEntrypoint2.On("IsActive").Return(false).Maybe()
	mockEntrypoint2.On("IsClosed").Return(true).Maybe()

	tunnels := []tunnel.Tunnel{mockTunnel1, mockTunnel2}
	entrypoints := []entrypoint.EntryPoint{mockEntrypoint1, mockEntrypoint2}

	cleanup := setupTestEnvironment(tunnels, entrypoints)
	defer cleanup()

	// Create a custom writer that captures output
	var buf bytes.Buffer
	oldWriter := stats.StdOutWriter
	stats.StdOutWriter = &buf
	defer func() { stats.StdOutWriter = oldWriter }()

	// Act
	done := make(chan struct{})
	completed := make(chan struct{})

	// Start the display stats in a goroutine
	go func() {
		stats.DisplayStats(done, 10*time.Millisecond) // Shorter interval for faster test
		close(completed)
	}()

	// Wait for at least one update (reduced time to speed up test)
	time.Sleep(30 * time.Millisecond)

	// Stop the display stats
	close(done)

	// Wait for completion with timeout
	select {
	case <-completed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Test timed out waiting for DisplayStats to complete")
	}

	// Get the output
	output := buf.String()

	// Assert
	assert.Contains(t, output, "Monitoring 1 tunnels and 1 entrypoints.",
		"Should correctly count active tunnels and entrypoints")
}

func TestDisplayStats_Shows_LatestStats_ForEntrypoints(t *testing.T) {
	// Arrange
	mockEntrypoint := createActiveMockEntrypoint(t, "test-ep-id", "test-entrypoint", "tcp")

	// Setup stats
	serviceStats := config.ServiceStats{
		CurrentConns:    3,
		TotalConns:      20,
		OutputRateBytes: 2048,    // 2 KB/s
		InputRateBytes:  1024,    // 1 KB/s
		OutputBytes:     2097152, // 2 MB
		InputBytes:      1048576, // 1 MB
		TotalErrs:       5,
	}

	// Setup expectations
	mockEntrypoint.On("IsClosed").Return(false)
	mockEntrypoint.On("ID").Return("test-ep-id")
	mockEntrypoint.On("Name").Return("test-entrypoint")
	mockEntrypoint.On("Type").Return("tcp")
	mockEntrypoint.On("IsActive").Return(true)
	mockEntrypoint.On("Stats").Return(serviceStats)

	entrypoints := []entrypoint.EntryPoint{mockEntrypoint}
	cleanup := setupTestEnvironment(nil, entrypoints)
	defer cleanup()

	// Act
	done := make(chan struct{})
	go func() {
		// Run display stats with a very short interval
		stats.DisplayStats(done, 10*time.Millisecond)
	}()

	// Give it some time to process
	time.Sleep(50 * time.Millisecond)
	close(done)

	// Small delay to allow final output to be captured
	time.Sleep(10 * time.Millisecond)

	// Assert
	output := stats.StdOutWriter.(*bytes.Buffer).String()

	// Check monitoring title
	assert.Contains(t, output, "Monitoring 1 entrypoints.", "Should show correct monitoring title")

	// Check entrypoint stats
	assert.Contains(t, output, "[test-entrypoint-TCP]", "Should display entrypoint name and type")
	assert.Contains(t, output, "Conn: 3/20", "Should show current and total connections")
	assert.Contains(t, output, "↑ 2.00 KB/s", "Should show upload speed")
	assert.Contains(t, output, "↓ 1.00 KB/s", "Should show download speed")
	assert.Contains(t, output, "Total: ↑ 2.00 MB ↓ 1.00 MB", "Should show total upload/download")
	assert.Contains(t, output, "Err: 5", "Should show error count")

	// Verify all expected calls were made
	mockEntrypoint.AssertExpectations(t)
}
