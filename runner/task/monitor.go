package task

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/runner"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/x/service"
)

type monitorTunnelsTask struct {
	logger logger.Logger
}

func NewMonitorTask() runner.Task {
	return &monitorTunnelsTask{
		logger: logger.Default().WithFields(map[string]any{
			"kind": "monitor",
		}),
	}
}

func NewMonitorTaskWith(logger logger.Logger) runner.Task {
	return &monitorTunnelsTask{
		logger: logger,
	}
}

func (t *monitorTunnelsTask) ID() runner.TaskID {
	return runner.TaskMonitorTunnels
}

// Create an HTTP httpClient that skips certificate verification
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func (t *monitorTunnelsTask) Run(ctx context.Context) error {
	log := t.logger

	for i := range tunnel.Count() {
		tun := tunnel.GetIndex(i)
		if tun == nil {
			continue
		}

		if tun.IsClosed() {
			log.Info("Skipping intentionally closed tunnel...")
			continue
		}

		isActive := false
		if tun.Type() == tunnel.HTTPTunnel || tun.Type() == tunnel.FileTunnel {
			entrypoint := tun.Entrypoint()
			log.Infof("Checking tunnel %s connection at %s", tun.Name(), entrypoint)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, entrypoint, nil)
			if err != nil {
				log.Warnf("Failed to create request for %s: %v", entrypoint, err)
				continue
			}
			req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

			// Trying to perform GET request to the entrypoint
			resp, err := httpClient.Do(req)
			if resp.StatusCode >= 500 || err != nil { // for test: resp.StatusCode == 401 ||
				log.Warnf("Tunnel '%s' does not exist. Http status: %s, reason: %v", tun.Name(), resp.Status, err)
			} else { // even 40x range is success
				log.Infof("Tunnel '%s' connectivity is succeeded: http status: %s", tun.Name(), resp.Status)
				defer resp.Body.Close()
				isActive = true
			}
		} else {
			status := tun.Status()
			isActive = status != nil && status.State() == service.StateRunning
		}

		if !isActive {
			log.Infof("Detected inactive tunnel '%s' (%s), attempting to reconnect...", tun.Name(), tun.ID())

			// Attempt to restart the existing tunnel
			if err := restartTunnel(tun, log); err != nil {
				log.Errorf("Failed to restart tunnel '%s': %v", tun.Name(), err)
			} else {
				log.Infof("Successfully restarted tunnel '%s'", tun.Name())
			}
		}
	}

	return nil
}

func restartTunnel(tun tunnel.Tunnel, log logger.Logger) error {
	log.Infof("Attempting to restart the tunnel ID: %s with the same options...", tun.ID())

	tun.Close() // Close it to release resources for just in case

	opts := tun.Options()
	tunnelID := tun.ID()
	tunnelType := tun.Type()
	newTunnel := tunnel.CreateTunnel(tunnelType, opts)

	if err := newTunnel.Run(); err != nil {
		log.Errorf("Failed to run recreated tunnel: %v", err)
		return err
	}

	tunnel.Set(newTunnel) // Replace the old tunnel with the new one
	log.Infof("Done. Tunnel %s has been recreated with same ID and options", tunnelID)

	return nil
}
