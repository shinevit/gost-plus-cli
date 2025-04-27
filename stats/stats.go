package stats

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/tunnel"
)

type TunnelStats struct {
	UploadSpeed, DownloadSpeed float64
	TotalUpload, TotalDownload float64
}

func getActiveTunnelCount() int {
	count := 0
	for i := range tunnel.Count() {
		t := tunnel.GetIndex(i)
		if t != nil && !t.IsClosed() {
			count++
		}
	}
	return count
}

// Shows statistics about active tunnels
func DisplayStats(done chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Clear the terminal line and move cursor to beginning
	clearLine := func() {
		fmt.Printf("\033[2K\r")
	}

	// Get active tunnel count
	activeTunnels := getActiveTunnelCount()

	// Print initial header
	fmt.Printf("Monitoring %d active tunnels. Stats will appear below:\n", activeTunnels)
	lastNumLines := 0

	// Store last non-zero transfer rates for each tunnel
	lastRates := make(map[string]TunnelStats)

	for {
		select {
		case <-ticker.C:
			// Move cursor up for each line we printed previously
			if lastNumLines > 0 {
				fmt.Printf("\033[%dA", lastNumLines)
			}

			// Count active tunnels and print stats
			lineCount := 0

			for i := range tunnel.Count() {
				t := tunnel.GetIndex(i)
				if t == nil || t.IsClosed() {
					continue
				}

				stats := t.Stats()
				currentStats := getCurentStats(stats)

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

				fmt.Printf("[%s-%s] Conn: %d/%d | Transfer: ↑ %.2f KB/s%s ↓ %.2f KB/s%s | Total: ↑ %.2f MB ↓ %.2f MB | Err: %d\n",
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

// Calculate current rates, cumulative Tx/Rx sizes
func getCurentStats(stats config.ServiceStats) TunnelStats {
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
