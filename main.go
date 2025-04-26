package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/runner"
	"github.com/go-gost/gost.plus/runner/task"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/go-gost/gost.plus/version"
)

// Command line flags
type CommandFlags struct {
	LocalEndpoint  string
	TunnelType     string
	RemoteName     string
	Username       string
	Password       string
	Hostname       string
	EnableTLS      bool
	ShowVersion    bool
	StatsInterval  time.Duration
	DeleteTunnelID string
	ListTunnels    bool
	ShowHelp       bool
}

func main() {
	// Parse command line arguments
	flags := parseFlags()

	// Process command line flags
	if flags.ShowHelp {
		printUsage()
		os.Exit(0)
	}

	if flags.ShowVersion {
		fmt.Print(GetVersion())
		os.Exit(0)
	}

	// Initialize configuration and load tunnels
	initializeSystem()

	// Handle tunnel management commands
	if flags.DeleteTunnelID != "" {
		deleteTunnel(flags.DeleteTunnelID)
		os.Exit(0)
	}

	if flags.ListTunnels {
		listAllTunnels()
		os.Exit(0)
	}

	// Create or start tunnels
	var newTunnel tunnel.Tunnel
	createNewTunnel := flags.LocalEndpoint != ""

	if createNewTunnel {
		newTunnel = createAndStartTunnel(flags)
	} else {
		startExistingTunnels()
	}

	// Start stats update and display
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startStatsRunner(ctx, flags.StatsInterval)
	doneChan := make(chan struct{})

	if tunnel.Count() > 0 {
		go displayStats(doneChan, flags.StatsInterval)
	}

	// Wait for shutdown signal
	waitForShutdown(doneChan)

	// Cleanup and save configuration
	cleanupAndExit(createNewTunnel, newTunnel)
}

// Parses command line arguments and returns a CommandFlags struct
func parseFlags() CommandFlags {
	flags := CommandFlags{}

	// Define command line arguments
	flag.StringVar(&flags.LocalEndpoint, "local", "", "Local endpoint to listen on")
	flag.StringVar(&flags.TunnelType, "tunnel_type", "http", "Tunnel type: http, tcp, udp, file")
	flag.StringVar(&flags.RemoteName, "name", "", "Name for the tunnel (optional)")
	flag.StringVar(&flags.Username, "username", "", "Username for authentication (optional)")
	flag.StringVar(&flags.Password, "password", "", "Password for authentication (optional)")
	flag.StringVar(&flags.Hostname, "hostname", "", "Hostname for the tunnel (optional)")
	flag.BoolVar(&flags.EnableTLS, "tls", false, "Enable TLS")
	flag.BoolVar(&flags.ShowVersion, "version", false, "Show version information")
	flag.DurationVar(&flags.StatsInterval, "stats-interval", time.Second, "Stats update interval")
	flag.StringVar(&flags.DeleteTunnelID, "delete", "", "Delete tunnel by ID")
	flag.BoolVar(&flags.ListTunnels, "list", false, "List all tunnels")
	flag.BoolVar(&flags.ShowHelp, "help", false, "Show help information")

	// Override default usage function
	flag.Usage = printUsage

	// Parse command line arguments
	flag.Parse()

	return flags
}

// Prints the usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", GetVersion())
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Start all configured tunnels\n")
	fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Create HTTP tunnel (default type)\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:8080 --name web-service\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Create HTTP tunnel (full syntax)\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:8080 --tunnel_type http --name web-service --username admin --password admin --stats-interval 5s\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Create TCP tunnel\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:22 --tunnel_type tcp --name ssh-service\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Create UDP tunnel\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:53 --tunnel_type udp --name dns-service\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Create file sharing tunnel from current directory\n")
	fmt.Fprintf(os.Stderr, "  %s --local . --tunnel_type file --name file-share\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # List all tunnels\n")
	fmt.Fprintf(os.Stderr, "  %s --list\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Delete a tunnel\n")
	fmt.Fprintf(os.Stderr, "  %s --delete \"tunnel-id\"\n\n", os.Args[0])
}

// Returns a formatted version string
func GetVersion() string {
	return fmt.Sprintf("GOST+ CLI Version %s\n", version.Version)
}

// Initializes the configuration and loads tunnels
func initializeSystem() {
	config.Init()
	tunnel.LoadConfig()
	entrypoint.LoadConfig()
}

// Deletes a tunnel by ID
func deleteTunnel(id string) {
	t := tunnel.Get(id)
	if t == nil {
		fmt.Printf("Tunnel with ID '%s' not found.\n", id)
		return
	}

	name := t.Name()
	tunnel.Delete(id)

	entrypoint.SaveConfig()
	err := tunnel.SaveConfig()
	if err != nil {
		logger.Default().Error(err)
		fmt.Printf("Error while saving config after deletion: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Tunnel '%s' (ID: %s) has been deleted.\n", name, id)
}

// Lists all configured tunnels
func listAllTunnels() {
	fmt.Println("Available tunnels:")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Printf("%-36s %-20s %-10s %-10s\n", "ID", "NAME", "TYPE", "STATUS")
	fmt.Println("-------------------------------------------------------------------------")

	for i := range tunnel.Count() {
		t := tunnel.GetIndex(i)
		if t == nil {
			continue
		}
		status := "CLOSED"
		if !t.IsClosed() {
			status = "ACTIVE"
		}
		fmt.Printf("%-36s %-20s %-10s %-10s\n", t.ID(), t.Name(), t.Type(), status)
	}
	fmt.Println("-------------------------------------------------------------------------")
}

// Creates and starts a new tunnel based on command line flags
func createAndStartTunnel(flags CommandFlags) tunnel.Tunnel {
	tunnelTypeStr := strings.ToLower(flags.TunnelType)

	validTypes := map[string]bool{
		tunnel.HTTPTunnel: true,
		tunnel.TCPTunnel:  true,
		tunnel.UDPTunnel:  true,
		tunnel.FileTunnel: true,
	}

	// Validate tunnel type
	if !validTypes[tunnelTypeStr] {
		fmt.Printf("Invalid tunnel type: %s. Valid types are: http, tcp, udp, file\n", tunnelTypeStr)
		os.Exit(1)
	}

	// Create tunnel options
	options := []tunnel.Option{
		tunnel.EndpointOption(flags.LocalEndpoint),
		tunnel.EnableTLSOption(flags.EnableTLS),
	}

	// Add optional parameters if provided
	if flags.RemoteName != "" {
		options = append(options, tunnel.NameOption(flags.RemoteName))
	}
	if flags.Username != "" {
		options = append(options, tunnel.UsernameOption(flags.Username))
	}
	if flags.Password != "" {
		options = append(options, tunnel.PasswordOption(flags.Password))
	}
	if flags.Hostname != "" {
		options = append(options, tunnel.HostnameOption(flags.Hostname))
	}

	// Create appropriate tunnel based on type
	var newTunnel = createTunnel(tunnelTypeStr, options)
	if newTunnel == nil {
		fmt.Printf("Failed to create %s tunnel\n", tunnelTypeStr)
		os.Exit(1)
	}

	// Run the tunnel
	err := newTunnel.Run()
	if err != nil {
		fmt.Printf("Failed to start tunnel: %v\n", err)
		os.Exit(1)
	}

	// Add tunnel to the global list for stats tracking
	tunnel.Add(newTunnel)

	// Print tunnel information
	fmt.Printf("\n%s Tunnel started:\n", strings.ToUpper(tunnelTypeStr))
	fmt.Printf("ID: %s\n", newTunnel.ID())
	fmt.Printf("Name: %s\n", newTunnel.Name())
	fmt.Printf("Local Endpoint: %s\n", newTunnel.Endpoint())
	fmt.Printf("Remote Entrypoint: %s\n", newTunnel.Entrypoint())
	fmt.Printf("\nPress Ctrl+C to stop the tunnel\n\n")

	return newTunnel
}

func createTunnel(tunnelTypeStr string, options []tunnel.Option) tunnel.Tunnel {
	var newTunnel tunnel.Tunnel
	switch tunnelTypeStr {
	case tunnel.HTTPTunnel:
		newTunnel = tunnel.NewHTTPTunnel(options...)
	case tunnel.TCPTunnel:
		newTunnel = tunnel.NewTCPTunnel(options...)
	case tunnel.UDPTunnel:
		newTunnel = tunnel.NewUDPTunnel(options...)
	case tunnel.FileTunnel:
		newTunnel = tunnel.NewFileTunnel(options...)
	}
	return newTunnel
}

// Starts all tunnels from the configuration
func startExistingTunnels() {
	if tunnel.Count() == 0 {
		fmt.Println("No tunnels configured. Use --local to create a new tunnel.")
		printUsage()
		os.Exit(0)
	}

	fmt.Println("Starting tunnels from configuration...")

	for i := range tunnel.Count() {
		t := tunnel.GetIndex(i)
		if t == nil || t.IsClosed() {
			continue
		}

		err := t.Run()
		if err != nil {
			logger.Default().Errorf("Failed to start tunnel %s: %v", t.ID(), err)
			continue
		}

		fmt.Printf("Started %s tunnel: %s (ID: %s)\n", strings.ToUpper(t.Type()), t.Name(), t.ID())
	}

	fmt.Println("\nPress Ctrl+C to stop all tunnels")
}

// Starts the stats update runner
func startStatsRunner(ctx context.Context, interval time.Duration) {
	err := runner.Exec(ctx, task.UpdateStats(),
		runner.WithAync(true),
		runner.WithInterval(interval),
		runner.WithCancel(true),
	)
	if err != nil {
		logger.Default().Error(err)
	}
}

// Waits for a shutdown signal
func waitForShutdown(doneChan chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	close(doneChan) // Signal stats display to stop
}

// Performs cleanup tasks and exits
func cleanupAndExit(createNewTunnel bool, newTunnel tunnel.Tunnel) {
	fmt.Println("\nShutting down tunnels...")

	// Save configuration before exit
	entrypoint.SaveConfig()
	err := tunnel.SaveConfig()
	if err != nil {
		logger.Default().Error(err)
	}

	// If we created a new tunnel, close it specifically
	if createNewTunnel && newTunnel != nil {
		err := newTunnel.Close()
		if err != nil {
			logger.Default().Errorf("Error closing tunnel %s: %v", newTunnel.ID(), err)
		}

		err = newTunnel.Err()
		if err != nil {
			logger.Default().Errorf("Tunnel %s errors: %v", newTunnel.ID(), err)
		}
	}
}

// Shows statistics about active tunnels
func displayStats(done chan struct{}, interval time.Duration) {
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
				clearLine()
				fmt.Printf("[%s-%s] Conn: %d/%d | Transfer: ↑ %.2f KB/s ↓ %.2f KB/s | Total: ↑ %.2f MB ↓ %.2f MB | Err: %d",
					t.Name(), strings.ToUpper(t.Type()),
					stats.CurrentConns, stats.TotalConns,
					float64(stats.OutputRateBytes)/1024, float64(stats.InputRateBytes)/1024,
					float64(stats.OutputBytes)/(1024*1024), float64(stats.InputBytes)/(1024*1024),
					stats.TotalErrs)
				fmt.Println()
				lineCount++
			}

			// Update line count for next iteration
			lastNumLines = lineCount

		case <-done:
			return
		}
	}
}

// Returns the number of active tunnels
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
