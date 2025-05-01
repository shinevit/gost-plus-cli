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
	"github.com/go-gost/gost.plus/stats"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/version"
	"github.com/go-gost/x/service"
)

// Command line flags
type CommandFlags struct {
	LocalEndpoint   string
	TunnelType      string
	ServiceName     string
	Username        string
	Password        string
	Hostname        string
	EnableTLS       bool
	ShowVersion     bool
	StatsInterval   time.Duration
	MonitorInterval time.Duration
	DeleteTunnelID  string
	ListTunnels     bool
	ShowHelp        bool
	NoStats         bool
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

	// Initializes the configuration and loads tunnels
	config.Init()
	tunnel.InitFromConfig()

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
		newTunnel = createAndRunTunnel(flags)
	} else {
		startExistingTunnels()
	}

	// Start stats update and display
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	doneChan := make(chan struct{})

	startTunnelMonitor(ctx, flags.MonitorInterval)

	if !flags.NoStats {
		startStatsRunner(ctx, flags.StatsInterval)
		if tunnel.Count() > 0 {
			go stats.DisplayStats(doneChan, flags.StatsInterval)
		}
	}

	// Wait for shutdown signal
	waitForShutdown(doneChan)

	// Save configuration and cleanup
	cleanupAndExit(createNewTunnel, newTunnel)
}

// Parses command line arguments and returns a CommandFlags struct
func parseFlags() CommandFlags {
	flags := CommandFlags{}

	// Define command line arguments
	flag.StringVar(&flags.LocalEndpoint, "local", "", "Local endpoint to listen on")
	flag.StringVar(&flags.TunnelType, "tunnel_type", "http", "Tunnel type: http, file")
	flag.StringVar(&flags.ServiceName, "name", "", "Name for the tunnel (optional)")
	flag.StringVar(&flags.Username, "username", "", "Username for authentication (optional)")
	flag.StringVar(&flags.Password, "password", "", "Password for authentication (optional)")
	flag.StringVar(&flags.Hostname, "hostname", "", "Hostname for the tunnel (optional)")
	flag.BoolVar(&flags.EnableTLS, "tls", false, "Enable TLS")
	flag.BoolVar(&flags.ShowVersion, "version", false, "Show version information")
	flag.DurationVar(&flags.StatsInterval, "stats-interval", time.Second, "Stats update interval")
	flag.DurationVar(&flags.MonitorInterval, "monitor-interval", 60*time.Second, "Tunnel connection monitoring interval")
	flag.StringVar(&flags.DeleteTunnelID, "delete", "", "Delete tunnel by ID")
	flag.BoolVar(&flags.ListTunnels, "list", false, "List all tunnels")
	flag.BoolVar(&flags.ShowHelp, "help", false, "Show help information")
	flag.BoolVar(&flags.NoStats, "no-stats", false, "Disable statistics in daemon mode")

	// Override default usage function
	flag.Usage = printUsage

	// Parse command line arguments
	flag.Parse()

	return flags
}

// Get the App name
func appName() string {
	return strings.TrimPrefix(os.Args[0], "./")
}

// Prints the usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", GetVersion())
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", appName())
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Start all configured tunnels\n")
	fmt.Fprintf(os.Stderr, "  %s\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # Start all configured tunnels silently, i.e. without stats\n")
	fmt.Fprintf(os.Stderr, "  %s --no-stats\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # Monitor and auto-restart tunnels every minute\n")
	fmt.Fprintf(os.Stderr, "  %s --monitor-interval 2m\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # Create HTTP tunnel (default type)\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:8080 --name web-service\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # Create HTTP tunnel (full syntax)\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:8080 --tunnel_type http --name web-service --username admin --password admin --stats-interval 5s\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # Create file sharing tunnel from current directory\n")
	fmt.Fprintf(os.Stderr, "  %s --local . --tunnel_type file --name file-share\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # List all tunnels\n")
	fmt.Fprintf(os.Stderr, "  %s --list\n\n", appName())
	fmt.Fprintf(os.Stderr, "  # Delete a tunnel\n")
	fmt.Fprintf(os.Stderr, "  %s --delete \"tunnel-id\"\n\n", appName())
}

// Returns a formatted version string
func GetVersion() string {
	return fmt.Sprintf("GOST+ CLI Version %s\n", version.Version)
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

		status := strings.ToUpper(string(t.Status().State()))
		fmt.Printf("%-36s %-20s %-10s %-10s\n", t.ID(), t.Name(), t.Type(), status)
	}
	fmt.Println("-------------------------------------------------------------------------")
}

// Creates and runs a new tunnel based on command line flags
func createAndRunTunnel(flags CommandFlags) tunnel.Tunnel {
	tunnelType := strings.ToLower(flags.TunnelType)

	validTypes := map[string]bool{
		tunnel.HTTPTunnel: true,
		tunnel.FileTunnel: true,
		// tunnel.TCPTunnel:  true,
		// tunnel.UDPTunnel:  true,
	}

	// Validate a supported tunnel type
	if !validTypes[tunnelType] {
		fmt.Printf("Invalid tunnel type: %s. Supported types: http, file\n", tunnelType)
		os.Exit(1)
	}

	options := tunnel.Options{
		Name:      flags.ServiceName,
		Endpoint:  flags.LocalEndpoint,
		Hostname:  flags.Hostname,
		Username:  flags.Username,
		Password:  flags.Password,
		EnableTLS: flags.EnableTLS,
		Keepalive: true,
	}

	var newTunnel = tunnel.CreateTunnel(tunnelType, options)
	if newTunnel == nil {
		fmt.Printf("Failed to create %s tunnel\n", tunnelType)
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
	fmt.Printf("\n%s Tunnel started:\n", strings.ToUpper(tunnelType))
	fmt.Printf("ID: %s\n", newTunnel.ID())
	fmt.Printf("Name: %s\n", newTunnel.Name())

	if flags.TunnelType == tunnel.FileTunnel {
		fmt.Printf("Local folder: %s\n", newTunnel.Endpoint())
	} else {
		fmt.Printf("Local endpoint: %s\n", newTunnel.Endpoint())
	}

	fmt.Printf("Remote entrypoint: %s\n", newTunnel.Entrypoint())
	fmt.Printf("\nPress Ctrl+C to stop the tunnel\n\n")
	return newTunnel
}

// Starts all tunnels from the configuration
func startExistingTunnels() {
	if tunnel.Count() == 0 {
		fmt.Println("No tunnels configured. Use --local to create a new tunnel.")
		printUsage()
		os.Exit(0)
	}

	fmt.Println("\nStarting tunnels from configuration...")

	for i := range tunnel.Count() {
		t := tunnel.GetIndex(i)
		if t == nil || t.IsClosed() {
			continue
		}

		if t.Status().State() != service.StateRunning {
			err := t.Run()
			if err != nil {
				logger.Default().Errorf("Failed to start tunnel %s: %v", t.ID(), err)
				continue
			}
		}

		fmt.Printf("Started %s tunnel: %s [ID: %s, URL: %s]\n", strings.ToUpper(t.Type()), t.Name(), t.ID(), t.Entrypoint())
	}

	fmt.Println("\nPress Ctrl+C to stop all tunnels")
}

// Starts the stats updater
func startStatsRunner(ctx context.Context, interval time.Duration) {
	err := runner.Exec(ctx, task.UpdateStats(),
		runner.WithAsync(true),
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

func startTunnelMonitor(ctx context.Context, interval time.Duration) {
	err := runner.Exec(ctx, task.MonitorTunnels(),
		runner.WithAsync(true),
		runner.WithInterval(interval),
		runner.WithCancel(true),
	)
	if err != nil {
		logger.Default().Error(err)
	}
}
