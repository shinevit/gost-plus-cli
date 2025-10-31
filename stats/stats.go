package stats

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	fp "github.com/go-gost/gost.plus/utils/fp/slice"
)

var StdOutWriter io.Writer = os.Stdout

type TunnelStats struct {
	UploadSpeed, DownloadSpeed float64
	TotalUpload, TotalDownload float64
}

func getActiveTunnels() uint64 {
	return fp.Count(tunnel.GetAll(), func(t tunnel.Tunnel) bool {
		return t != nil && t.IsActive()
	})
}

func getActiveEntrypoints() uint64 {
	return fp.Count(entrypoint.GetAll(), func(t entrypoint.EntryPoint) bool {
		return t != nil && t.IsActive()
	})
}

// Shows statistics about active tunnels
func DisplayStats(done chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Clear the terminal line and move cursor to beginning
	clearLine := func() {
		fmt.Fprintf(StdOutWriter, "\033[2K\r")
	}

	showMonitoringTitle()

	lastNumLines := 0
	// Store last non-zero transfer rates for each tunnel
	lastRates := make(map[string]TunnelStats)

	for {
		select {
		case <-ticker.C:
			// Move cursor up for each line we printed previously
			if lastNumLines > 0 {
				fmt.Fprintf(StdOutWriter, "\033[%dA", lastNumLines)
			}

			// Count active tunnels and print stats
			lineCount := 0

			observableItems := slices.Concat(tunnel.GetAll(), entrypoint.GetAll())
			for _, t := range observableItems {
				if t == nil || t.IsClosed() {
					continue
				}

				stats := t.Stats()
				currentStats := getCurrentStats(stats)

				// Get or initialize last memoStats
				tunnelID := t.ID()
				memoStats, exists := lastRates[tunnelID]
				if !exists {
					memoStats = TunnelStats{0, 0, 0, 0}
				}

				retainStats(currentStats, memoStats, lastRates, tunnelID)
				displayStats := getDisplayStats(currentStats, memoStats)
				uploadIndicator, downloadIndicator := getDisplaySpeedsIndicators(currentStats, displayStats)

				clearLine()

				fmt.Fprintf(StdOutWriter, "[%s-%s] Conn: %d/%d | Transfer: ↑ %.2f KB/s%s ↓ %.2f KB/s%s | Total: ↑ %.2f MB ↓ %.2f MB | Err: %d\n",
					t.Name(), strings.ToUpper(t.Type()),
					stats.CurrentConns, stats.TotalConns,
					displayStats.UploadSpeed, uploadIndicator,
					displayStats.DownloadSpeed, downloadIndicator,
					displayStats.TotalUpload,
					displayStats.TotalDownload,
					stats.TotalErrs)

				lineCount++
			}

			lastNumLines = lineCount

		case <-done:
			return
		}
	}
}

func showMonitoringTitle() {
	var message string
	staticMessage := "Statistics is updating:"
	activeTunnels := getActiveTunnels()
	activeEntrypoints := getActiveEntrypoints()
	if activeTunnels > 0 && activeEntrypoints > 0 {
		message = fmt.Sprintf("Monitoring %d tunnels and %d entrypoints. %s\n",
			activeTunnels,
			activeEntrypoints,
			staticMessage,
		)
	} else if activeTunnels > 0 && activeEntrypoints == 0 {
		message = fmt.Sprintf("Monitoring %d tunnels. %s\n",
			activeTunnels,
			staticMessage,
		)
	} else if activeTunnels == 0 && activeEntrypoints > 0 {
		message = fmt.Sprintf("Monitoring %d entrypoints. %s\n",
			activeEntrypoints,
			staticMessage,
		)
	} else {
		message = "No items for monitoring"
	}
	fmt.Fprint(StdOutWriter, message)
}

// Calculate current rates, cumulative Tx/Rx sizes
func getCurrentStats(stats config.ServiceStats) TunnelStats {
	return TunnelStats{
		float64(stats.OutputRateBytes) / 1024,
		float64(stats.InputRateBytes) / 1024,
		float64(stats.OutputBytes) / (1024 * 1024),
		float64(stats.InputBytes) / (1024 * 1024),
	}
}

func getDisplaySpeedsIndicators(currentStats, displayStats TunnelStats) (string, string) {
	uploadIndicator := ""
	downloadIndicator := ""
	if currentStats.UploadSpeed < 0.01 && displayStats.UploadSpeed > 0 {
		uploadIndicator = " (last)"
	}
	if currentStats.DownloadSpeed < 0.01 && displayStats.DownloadSpeed > 0 {
		downloadIndicator = " (last)"
	}

	return uploadIndicator, downloadIndicator
}

// Choose which rates to display - always prefer a non-zero value
func getDisplayStats(currentStats TunnelStats, memoRates TunnelStats) TunnelStats {
	displayUpload := currentStats.UploadSpeed
	displayDownload := currentStats.DownloadSpeed

	// If current rates are very low but we have saved rates, show saved rates
	if currentStats.UploadSpeed < 0.01 && memoRates.UploadSpeed > 0 {
		displayUpload = memoRates.UploadSpeed
	}
	if currentStats.DownloadSpeed < 0.01 && memoRates.DownloadSpeed > 0 {
		displayDownload = memoRates.DownloadSpeed
	}

	return TunnelStats{
		displayUpload,
		displayDownload,
		currentStats.TotalUpload,
		currentStats.TotalDownload,
	}
}

// Update last non-zero rates
func retainStats(
	currentStats TunnelStats,
	memoRates TunnelStats,
	lastRates map[string]TunnelStats,
	tunnelID string,
) {
	if currentStats.UploadSpeed > 0.01 {
		memoRates.UploadSpeed = currentStats.UploadSpeed
	}
	if currentStats.DownloadSpeed > 0.01 {
		memoRates.DownloadSpeed = currentStats.DownloadSpeed
	}

	// Always update totals when they increase
	if currentStats.TotalUpload > memoRates.TotalUpload {
		memoRates.TotalUpload = currentStats.TotalUpload
	}
	if currentStats.TotalDownload > memoRates.TotalDownload {
		memoRates.TotalDownload = currentStats.TotalDownload
	}

	lastRates[tunnelID] = memoRates
}
