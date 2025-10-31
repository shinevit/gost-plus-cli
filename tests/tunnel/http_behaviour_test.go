package tunnel

import (
	"sync"
	"testing"
	"time"

	"github.com/go-gost/gost.plus/tunnel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestTunnel(t *testing.T, id string) *tunnel.MockTunnel {
	t.Helper()
	mockTunnel := tunnel.NewMockTunnel(t)
	mockTunnel.On("ID").Return(id).Maybe()
	mockTunnel.On("Type").Return("http").Maybe()
	mockTunnel.On("Name").Return("test-tunnel-" + id).Maybe()
	mockTunnel.On("Endpoint").Return("localhost:0").Maybe()
	mockTunnel.On("IsClosed").Return(false).Maybe()
	mockTunnel.On("IsActive").Return(false).Maybe()
	return mockTunnel
}

func TestHTTPTunnel_Run_AlreadyRunning(t *testing.T) {
	// Create a mock tunnel and wrap it
	mockTunnel := createTestTunnel(t, "test-id")

	// Set up expectations for the first Run() call
	mockTunnel.On("Run").Return(nil).Once()

	// First run should succeed
	err := mockTunnel.Run()
	assert.NoError(t, err, "First Run() should succeed")

	// Set up expectations for the second Run() call
	mockTunnel.On("Run").Return(tunnel.ErrTunnelDuplicateInstance).Once()

	// Second run should fail with duplicate instance error
	err = mockTunnel.Run()
	assert.ErrorIs(t, err, tunnel.ErrTunnelDuplicateInstance, "Second Run() should fail with duplicate instance error")

	// Set up expectations for Close()
	mockTunnel.On("Close").Return(nil).Once()

	// Cleanup
	err = mockTunnel.Close()
	assert.NoError(t, err, "Close() should succeed")

	// Verify all expectations were met
	mockTunnel.AssertExpectations(t)
}

func TestHTTPTunnel_Run_Concurrent(t *testing.T) {
	// Create a mock tunnel and wrap it
	mockTunnel := createTestTunnel(t, "test-concurrent")

	// Use a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Channel to collect results from goroutines
	results := make(chan error, 5)

	// Use a sync.Once to ensure only one Run() call succeeds
	var once sync.Once

	// Set up expectations for Run() - only one call should succeed
	mockTunnel.On("Run").Run(func(args mock.Arguments) {
		once.Do(func() {})
	}).Return(nil).Once()

	// Set up expectations for subsequent Run() calls - they should fail
	mockTunnel.On("Run").Return(tunnel.ErrTunnelDuplicateInstance).Times(4)

	// Start multiple goroutines trying to run the tunnel
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := mockTunnel.Run()
			results <- err
		}()
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	successCount := 0
	var firstErr error
	for err := range results {
		if err == nil {
			successCount++
		} else if firstErr == nil {
			firstErr = err
		}
	}

	// Only one goroutine should have succeeded
	assert.Equal(t, 1, successCount, "Only one Run() should succeed")
	if successCount == 0 {
		t.Fatalf("No goroutine was able to run the tunnel: %v", firstErr)
	}

	// Set up expectations for Close()
	mockTunnel.On("Close").Return(nil).Once()

	// Cleanup
	err := mockTunnel.Close()
	assert.NoError(t, err, "Close() should succeed")

	// Verify all expectations were met
	mockTunnel.AssertExpectations(t)
}

func TestHTTPTunnel_Run_AfterClose(t *testing.T) {
	// Create a mock tunnel and wrap it
	mockTunnel := createTestTunnel(t, "test-after-close")

	// Set up expectations for the first Run() call
	mockTunnel.On("Run").Return(nil).Once()

	// First run should succeed
	err := mockTunnel.Run()
	assert.NoError(t, err, "First Run() should succeed")

	// Set up expectations for Close()
	mockTunnel.On("Close").Return(nil).Once()

	// Close the tunnel
	err = mockTunnel.Close()
	assert.NoError(t, err, "Close() should succeed")

	// Set up expectations for IsClosed() to return true after close
	mockTunnel.On("IsClosed").Return(true).Maybe()

	// Set up expectations for Run() after close - should fail with ErrTunnelClosed
	mockTunnel.On("Run").Return(tunnel.ErrTunnelClosed).Once()

	// Running after close should fail
	err = mockTunnel.Run()
	assert.ErrorIs(t, err, tunnel.ErrTunnelClosed, "Run() after Close() should fail with ErrTunnelClosed")

	// Verify all expectations were met
	mockTunnel.AssertExpectations(t)
}

func TestHTTPTunnel_Run_ConcurrentCalls(t *testing.T) {
	// Create a mock tunnel and wrap it
	mockTunnel := createTestTunnel(t, "test-concurrent")

	// Channel to coordinate the test
	started := make(chan struct{})
	done := make(chan struct{})

	// Set up expectations for the first Run() call
	mockTunnel.On("Run").Run(func(args mock.Arguments) {
		// Signal that the first goroutine has started
		close(started)
		// Wait for the test to complete
		<-done
	}).Return(nil).Once()

	// Set up expectations for the second Run() call - should fail
	mockTunnel.On("Run").Return(tunnel.ErrTunnelDuplicateInstance).Once()

	// Channel to collect errors from goroutines
	errCh := make(chan error, 2)

	// First goroutine - will block until we signal it to complete
	go func() {
		errCh <- mockTunnel.Run()
	}()

	// Wait for the first goroutine to start
	<-started

	// Start the second goroutine - this should fail immediately
	go func() {
		errCh <- mockTunnel.Run()
	}()

	// Wait a moment for the second goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Signal the first goroutine to complete
	close(done)

	// Wait for both goroutines to complete
	err1 := <-errCh
	err2 := <-errCh

	// One should succeed, the other should fail
	if err1 == nil && err2 == nil {
		t.Fatal("Expected one of the Run() calls to fail with duplicate instance error")
	}

	// Set up expectations for Close()
	mockTunnel.On("Close").Return(nil).Once()

	// Cleanup
	err := mockTunnel.Close()
	assert.NoError(t, err, "Close() should succeed")

	// Verify all expectations were met
	mockTunnel.AssertExpectations(t)
}
