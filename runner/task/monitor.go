package task

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/runner"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/utils"
)

type monitorTunnelsTask struct {
	logger       logger.Logger
	restartTimes map[string]time.Time // Track when each tunnel was last restarted
	restartMutex sync.RWMutex         // Protect restartTimes map
}

func NewMonitorTask() runner.Task {
	return NewMonitorTaskWith(logger.Default().WithFields(map[string]any{
		"kind": "monitor",
	}))
}

func NewMonitorTaskWith(logger logger.Logger) runner.Task {
	return &monitorTunnelsTask{
		logger:       logger,
		restartTimes: make(map[string]time.Time),
	}
}

func (t *monitorTunnelsTask) ID() runner.TaskID {
	return runner.TaskMonitorTunnels
}

// Create an HTTP httpClient that skips certificate verification
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func (t *monitorTunnelsTask) isTunnelActive(tun tunnel.Tunnel) bool {
	if tun.Type() == tunnel.HTTPTunnel || tun.Type() == tunnel.FileTunnel {
		return t.isHTTPTunnelActive(tun)
	}

	// For TCP and UDP tunnels, rely on service state only
	displayType := strings.ToUpper(tun.Type())
	state := utils.GetState(tun)
	isActive := tun.IsActive()
	if isActive {
		t.logger.Infof("%s Tunnel '%s' is active (service %v)", displayType, tun.Name(), state)
	} else {
		// For other states (failed, closed, etc.), the tunnel is not active
		t.logger.Infof("%s Tunnel '%s' seems inactive (service %v)", displayType, tun.Name(), state)
	}
	return isActive
}

// Checks connectivity for HTTP and FILE tunnels
// Status check is not enough for HTTP tunnels
func (t *monitorTunnelsTask) isHTTPTunnelActive(tun tunnel.Tunnel) bool {
	entrypoint := tun.Entrypoint()
	t.logger.Infof("Checking tunnel %s HTTP connection at %s", tun.Name(), entrypoint)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, entrypoint, nil)
	if err != nil {
		t.logger.Warnf("Failed to create request for %s: %v", entrypoint, err)
		return false
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		t.logger.Warnf("HTTP request failed for tunnel %s: %v", tun.Name(), err)
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx, 3xx, and 401 status codes as success
	// 401 means the endpoint is available but needs authorization, which means the service is operating
	if (resp.StatusCode >= 200 && resp.StatusCode < 400) || resp.StatusCode == 401 {
		t.logger.Infof("Tunnel %s HTTP connectivity check succeeded: %s", tun.Name(), resp.Status)
		return true
	}

	if resp.StatusCode == 503 || resp.StatusCode == 502 {
		// Check for server errors (5xx) or other network issues
		t.logger.Warnf("Tunnel '%s' does not exist. Http status: %s", tun.Name(), resp.Status)
		return false
	}

	// Consider 4xx (except 401) and other status codes as failure
	t.logger.Warnf("Tunnel %s HTTP connectivity check failed: %s", tun.Name(), resp.Status)
	return false
}

// shouldRestartTunnel checks if we should restart a tunnel, considering recent restart attempts
func (t *monitorTunnelsTask) shouldRestartTunnel(tunnelID string) bool {
	t.restartMutex.RLock()
	lastRestart, exists := t.restartTimes[tunnelID]
	t.restartMutex.RUnlock()

	if !exists {
		return true
	}

	// Don't restart if we restarted this tunnel within the last 1 minutes
	minRestartInterval := 1 * time.Minute
	if time.Since(lastRestart) < minRestartInterval {
		return false
	}
	return true
}

func (t *monitorTunnelsTask) Run(ctx context.Context) error {
	log := t.logger
	for _, tun := range tunnel.GetAll() {
		if tun == nil {
			continue
		}

		if tun.IsClosed() {
			log.Infof("Skipping closed tunnel '%s'", tun.Name())
			continue
		}
		// Check if tunnel is active (running or ready)
		if t.isTunnelActive(tun) {
			continue
		}
		// Check if we should restart this tunnel (prevent rapid successive restarts)
		tunnelID := tun.ID()
		if !t.shouldRestartTunnel(tunnelID) {
			log.Infof("Skipping restart of tunnel '%s' - recently restarted", tun.Name())
			continue
		}

		log.Infof("Detected inactive tunnel '%s' (%s), attempting to reconnect...", tun.Name(), tunnelID)
		// Record the restart attempt
		t.restartMutex.Lock()
		t.restartTimes[tunnelID] = time.Now()
		t.restartMutex.Unlock()

		// Attempt to restart an existing tunnel
		if err := restartTunnel(tun, log); err != nil {
			log.Errorf("Failed to restart tunnel '%s': %v", tun.Name(), err)
		}
	}

	return nil
}

func restartTunnel(tun tunnel.Tunnel, log logger.Logger) error {
	log.Infof("Attempting to restart the tunnel ID: %s with the same options...", tun.ID())

	tunnelID := tun.ID()
	opts := tun.Options()

	// Create new tunnel with same options
	newTunnel := tunnel.CreateTunnel(tun.Type(), opts)
	if newTunnel == nil {
		return fmt.Errorf("Failed to create a new tunnel of type %s", strings.ToUpper(tun.Type()))
	}

	// Start the new tunnel
	if err := newTunnel.Run(); err != nil {
		log.Errorf("Failed to recreate the tunnel %s: %v", tunnelID, err)
		// Try to close the failed tunnel
		newTunnel.Close()
		return err
	}

	// Check if tunnel is already closed
	if tun.IsClosed() {
		log.Warnf("Tunnel %s is already closed", tunnelID)
	} else {
		// Close the old tunnel to release resources
		if err := tun.Close(); err != nil {
			log.Errorf("Error closing old tunnel %s: %v", tunnelID, err)
		} else {
			log.Infof("Successfully closed old tunnel %s", tunnelID)
		}
	}

	// Replace the old tunnel with the new one in the tunnel registry
	tunnel.Set(newTunnel)
	log.Infof("Successfully restarted tunnel %s with same ID and options", tunnelID)

	return nil
}
