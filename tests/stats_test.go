package tests

import (
	"bytes"
	"testing"
	"time"

	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/stats"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/stretchr/testify/assert"
	xgoMock "github.com/xhd2015/xgo/runtime/mock"
)

func setupTestEnvironment(tunnels []tunnel.Tunnel) func() {
	xgoMock.Patch(tunnel.Count, func() int {
		return len(tunnels)
	})
	xgoMock.Patch(tunnel.GetIndex, func(idx int) tunnel.Tunnel {
		if idx >= 0 && idx < len(tunnels) {
			return tunnels[idx]
		}
		return nil
	})

	// Capture output
	originalWriter := stats.StdOutWriter
	stats.StdOutWriter = &bytes.Buffer{}
	return func() {
		stats.StdOutWriter = originalWriter
	}
}

func runDisplayStats(interval time.Duration, sleepTime time.Duration) *bytes.Buffer {
	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(sleepTime)
	}()

	stats.DisplayStats(done, interval)
	return stats.StdOutWriter.(*bytes.Buffer)
}

func createActiveMockTunnel(t *testing.T, id, name, tunnelType string) *tunnel.MockTunnel {
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("IsClosed").Return(false)
	mockTunnel.On("ID").Return(id)
	mockTunnel.On("Name").Return(name)
	mockTunnel.On("Type").Return(tunnelType)
	return mockTunnel
}

func TestDisplayStats_ActiveTunnelCount(t *testing.T) {
	// Arrange
	mockTunnel1 := createActiveMockTunnel(t, "tunnel-1", "test-tunnel", "type")
	mockTunnel1.On("Stats").Return(config.ServiceStats{})

	mockTunnel2 := tunnel.NewMockTunnel(t)
	mockTunnel2.On("IsClosed").Return(true) // closed tunnel

	tunnels := []tunnel.Tunnel{mockTunnel1, mockTunnel2}
	cleanup := setupTestEnvironment(tunnels)
	defer cleanup()

	// Act
	outputBuf := runDisplayStats(50*time.Millisecond, 100*time.Millisecond)

	// Assert
	output := outputBuf.String()
	assert.Contains(t, output, "Monitoring 1 active tunnels", "Should correctly count only the active tunnel")
	mockTunnel1.AssertExpectations(t)
	mockTunnel2.AssertExpectations(t)
}

func TestDisplayStats_Shows_LatestStats(t *testing.T) {
	// Arrange
	mockTunnel := createActiveMockTunnel(t, "test-tunnel-id", "test-tunnel", "http")

	var (
		currentConns  uint64 = 2
		totalConns    uint64 = 10
		uploadRate    uint64 = 1024    // 1 KB/s
		downloadRate  uint64 = 2048    // 2 KB/s
		totalUpload   uint64 = 1048576 // 1 MB
		totalDownload uint64 = 2097152 // 2 MB
		totalErrors   uint64 = 0
	)

	mockTunnel.On("Stats").Return(func() config.ServiceStats {
		return config.ServiceStats{
			CurrentConns:    currentConns,
			TotalConns:      totalConns,
			OutputRateBytes: uploadRate,
			InputRateBytes:  downloadRate,
			OutputBytes:     totalUpload,
			InputBytes:      totalDownload,
			TotalErrs:       totalErrors,
		}
	})

	tunnels := []tunnel.Tunnel{mockTunnel}
	cleanup := setupTestEnvironment(tunnels)
	defer cleanup()

	// Act
	interval := 20 * time.Millisecond
	done := make(chan struct{})
	go func() {
		stats.DisplayStats(done, interval)
	}()

	time.Sleep(30 * time.Millisecond) // wait for initial display and then modify stats

	currentConns = 3
	totalConns = 12
	uploadRate = 3072       // 3 KB/s
	downloadRate = 4096     // 4 KB/s
	totalUpload = 2097152   // 2 MB
	totalDownload = 3145728 // 3 MB
	totalErrors = 1

	time.Sleep(30 * time.Millisecond)

	currentConns = 5
	totalConns = 15
	uploadRate = 5120       // 5 KB/s
	downloadRate = 7168     // 7 KB/s
	totalUpload = 3145728   // 3 MB
	totalDownload = 5242880 // 5 MB
	totalErrors = 2

	time.Sleep(30 * time.Millisecond)

	close(done)

	// Assert
	output := stats.StdOutWriter.(*bytes.Buffer).String()

	assert.Contains(t, output, "Monitoring 1 active tunnels")
	assert.Contains(t, output, "[test-tunnel-HTTP]")

	// Initial stats
	assert.Contains(t, output, "Conn: 2/10")
	assert.Contains(t, output, "↑ 1.00 KB/s")
	assert.Contains(t, output, "↓ 2.00 KB/s")
	assert.Contains(t, output, "Total: ↑ 1.00 MB ↓ 2.00 MB")
	assert.Contains(t, output, "Err: 0")

	// Second update
	assert.Contains(t, output, "Conn: 3/12")
	assert.Contains(t, output, "↑ 3.00 KB/s")
	assert.Contains(t, output, "↓ 4.00 KB/s")
	assert.Contains(t, output, "Total: ↑ 2.00 MB ↓ 3.00 MB")
	assert.Contains(t, output, "Err: 1")

	// Final update
	assert.Contains(t, output, "Conn: 5/15")
	assert.Contains(t, output, "↑ 5.00 KB/s")
	assert.Contains(t, output, "↓ 7.00 KB/s")
	assert.Contains(t, output, "Total: ↑ 3.00 MB ↓ 5.00 MB")
	assert.Contains(t, output, "Err: 2")

	mockTunnel.AssertExpectations(t)
}

func TestDisplayStats_Shows_Last_Indicators(t *testing.T) {
	// Arrange
	mockTunnel := createActiveMockTunnel(t, "test-tunnel-id", "test-tunnel", "http")

	statsCalls := 0
	mockTunnel.On("Stats").Return(func() config.ServiceStats {
		statsCalls++
		switch {
		case statsCalls <= 2: // Normal rates
			return config.ServiceStats{
				CurrentConns:    2,
				TotalConns:      10,
				OutputRateBytes: 2048,    // 2 KB/s
				InputRateBytes:  3072,    // 3 KB/s
				OutputBytes:     1048576, // 1 MB
				InputBytes:      2097152, // 2 MB
			}
		case statsCalls <= 4: // Upload very low, Download normal
			return config.ServiceStats{
				CurrentConns:    2,
				TotalConns:      10,
				OutputRateBytes: 9,       // 0.009 KB/s (just below threshold)
				InputRateBytes:  3072,    // 3 KB/s (normal)
				OutputBytes:     1048576, // 1 MB
				InputBytes:      2097152, // 2 MB
			}
		case statsCalls <= 6: // Upload normal, Download very low
			return config.ServiceStats{
				CurrentConns:    2,
				TotalConns:      10,
				OutputRateBytes: 2048,    // 2 KB/s (normal)
				InputRateBytes:  9,       // 0.009 KB/s (just below threshold)
				OutputBytes:     1048576, // 1 MB
				InputBytes:      2097152, // 2 MB
			}
		case statsCalls <= 8: // Both Upload and Download very low
			return config.ServiceStats{
				CurrentConns:    2,
				TotalConns:      10,
				OutputRateBytes: 9,       // 0.009 KB/s (just below threshold)
				InputRateBytes:  9,       // 0.009 KB/s (just below threshold)
				OutputBytes:     1048576, // 1 MB
				InputBytes:      2097152, // 2 MB
			}
		default: // Back to normal rates
			return config.ServiceStats{
				CurrentConns:    2,
				TotalConns:      10,
				OutputRateBytes: 4096,    // 4 KB/s (normal)
				InputRateBytes:  5120,    // 5 KB/s (normal)
				OutputBytes:     1048576, // 1 MB
				InputBytes:      2097152, // 2 MB
			}
		}
	})

	tunnels := []tunnel.Tunnel{mockTunnel}
	cleanup := setupTestEnvironment(tunnels)
	defer cleanup()

	// Act
	interval := 20 * time.Millisecond
	done := make(chan struct{})
	go func() {
		stats.DisplayStats(done, interval)
	}()

	time.Sleep(200 * time.Millisecond) // wait for all stats changes to be displayed
	close(done)

	// Assert
	output := stats.StdOutWriter.(*bytes.Buffer).String()

	assert.Contains(t, output, "Transfer: ↑ 2.00 KB/s ↓ 3.00 KB/s") // normal rates - no indicators
	assert.Contains(t, output, "↑ 2.00 KB/s ↓ 3.00 KB/s |")
	assert.Contains(t, output, "↑ 2.00 KB/s (last) ↓ 3.00 KB/s")        // upload indicator only
	assert.Contains(t, output, "↑ 2.00 KB/s ↓ 3.00 KB/s (last)")        // download indicator only
	assert.Contains(t, output, "↑ 2.00 KB/s (last) ↓ 3.00 KB/s (last)") // both indicators
	assert.Contains(t, output, "↑ 4.00 KB/s ↓ 5.00 KB/s")               // back to normal rates

	mockTunnel.AssertExpectations(t)
}
