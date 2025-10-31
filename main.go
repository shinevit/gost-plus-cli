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
	"github.com/go-gost/gost.plus/tunnel/entitymanager"
	"github.com/go-gost/gost.plus/tunnel/entrypoint"
	"github.com/go-gost/gost.plus/utils"
	"github.com/go-gost/gost.plus/version"
	_ "github.com/go-gost/gost.plus/winres"
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
	TunnelID        string
	IsEntrypoint    bool
	TCPFlag         bool
	UDPFlag         bool
	TTLSec          int
	ListAll         bool
	ListTunnels     bool
	ListEntryPoints bool
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
	entrypoint.InitFromConfig()
	fmt.Println("\nStarting tunnels from configuration...")
	fmt.Println("Starting entrypoints from configuration...")

	// Handle tunnel management commands
	if flags.DeleteTunnelID != "" {
		err1 := deleteTunnel(flags.DeleteTunnelID)
		err2 := deleteEntryPoint(flags.DeleteTunnelID)
		if err1 != nil && err2 != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if flags.ListTunnels {
		listAllTunnels()
		os.Exit(0)
	}
	if flags.ListEntryPoints {
		listAllEntryPoints()
		os.Exit(0)
	}
	if flags.ListAll {
		listAllTunnels()
		listAllEntryPoints()
		os.Exit(0)
	}

	if flags.IsEntrypoint {
		if flags.TunnelID == "" {
			fmt.Println("--tunnel_id is required to add a new entrypoint")
			os.Exit(1)
		}
		if flags.LocalEndpoint == "" {
			fmt.Println("--local [endpoint] is required for entrypoint")
			os.Exit(1)
		}
		if !flags.TCPFlag && !flags.UDPFlag {
			fmt.Println("Type of entrypoint (--tcp, --udp) is required.")
			os.Exit(1)
		}
		createEntryPoint(flags)

	} else if flags.LocalEndpoint != "" && flags.TunnelType != "" {
		createTunnel(flags)
	}

	if tunnel.Count() == 0 && entrypoint.Count() == 0 {
		fmt.Println("No tunnels or entrypoints are configured. Go for creating any of them.")
		printUsage()
		os.Exit(0)
	} else if tunnel.Count() > 0 && entrypoint.Count() == 0 {
		listAllTunnels()
	} else if entrypoint.Count() > 0 && tunnel.Count() == 0 {
		listAllEntryPoints()
	} else { // tunnel.Count() > 0 && entrypoint.Count() > 0
		listAllTunnels()
		listAllEntryPoints()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Give tunnels some time to establish initial connections before starting the monitoring
	// This prevents the monitoring task from immediately trying to restart tunnels that are still connecting
	fmt.Printf("Starting tunnel monitoring with %v interval...\n", flags.MonitorInterval)
	startTunnelMonitorWithDelay(ctx, flags.MonitorInterval, flags.MonitorInterval)

	doneChan := make(chan struct{})
	if !flags.NoStats {
		// Start statistics updating and displaying
		startStatsRunner(ctx, flags.StatsInterval)
		if tunnel.Count() > 0 || entrypoint.Count() > 0 {
			go stats.DisplayStats(doneChan, flags.StatsInterval)
		}
		fmt.Printf("\nPress Ctrl+C to exit\n\n")
	}

	waitForShutdown(doneChan)
}

// Parses command line arguments and returns a CommandFlags struct
func parseFlags() CommandFlags {
	flags := CommandFlags{}

	// Define command line arguments
	flag.StringVar(&flags.LocalEndpoint, "local", "", "Local endpoint to listen on")
	flag.StringVar(&flags.TunnelType, "tunnel_type", "http", "Tunnel type: http, file, tcp, udp")
	flag.StringVar(&flags.ServiceName, "name", "", "Name for the tunnel (optional)")
	flag.StringVar(&flags.Username, "username", "", "Username for authentication (optional)")
	flag.StringVar(&flags.Password, "password", "", "Password for authentication (optional)")
	flag.StringVar(&flags.Hostname, "hostname", "", "Rewritten hostname on headers (optional)")
	flag.BoolVar(&flags.EnableTLS, "tls", false, "Enable TLS (optional)")
	flag.BoolVar(&flags.ShowVersion, "version", false, "Show the version")
	flag.DurationVar(&flags.StatsInterval, "stats-interval", time.Second, "Stats update interval")
	flag.DurationVar(&flags.MonitorInterval, "monitor-interval", 60*time.Second, "Tunnel connection monitoring interval")
	flag.StringVar(&flags.DeleteTunnelID, "delete", "", "Delete tunnel or entrypoint by ID")

	flag.BoolVar(&flags.IsEntrypoint, "entrypoint", false, "Create an entrypoint")
	flag.StringVar(&flags.TunnelID, "tunnel_id", "", "An existing tunnel ID to connect from entrypoint")
	flag.IntVar(&flags.TTLSec, "ttl", 0, "Time to live for UDP entrypoint in sec (optional)")
	flag.BoolVar(&flags.TCPFlag, "tcp", false, "Indicates tcp protocol of entrypoint/tunnel")
	flag.BoolVar(&flags.UDPFlag, "udp", false, "Indicates udp protocol of entrypoint/tunnel")

	flag.BoolVar(&flags.ListAll, "list", false, "List all tunnels and entrypoints")
	flag.BoolVar(&flags.ListTunnels, "tunnels", false, "List all tunnels")
	flag.BoolVar(&flags.ListEntryPoints, "entrypoints", false, "List all entrypoints")

	flag.BoolVar(&flags.ShowHelp, "help", false, "Show help information")
	flag.BoolVar(&flags.NoStats, "no-stats", false, "Disable statistics for a daemon mode")

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

// printUsage prints the usage information for the CLI application
func printUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", GetVersion())
	fmt.Fprintf(os.Stderr, "A secure tunneling solution for exposing local services to the internet\n\n")

	// Global options
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", appName())

	// Tunnel types
	fmt.Fprintf(os.Stderr, "Tunnel Types (--tunnel_type):\n")
	fmt.Fprintf(os.Stderr, "  http\tHTTP/HTTPS tunnel (default)\n")
	fmt.Fprintf(os.Stderr, "  file\tFile sharing tunnel\n")
	fmt.Fprintf(os.Stderr, "  tcp\tTCP port forwarding\n")
	fmt.Fprintf(os.Stderr, "  udp\tUDP port forwarding\n\n")

	// TCP/UDP Tunnel options
	fmt.Fprintf(os.Stderr, "TCP/UDP Tunnel Options (--tunnel_type tcp/udp):\n")
	fmt.Fprintf(os.Stderr, "  --local <endpoint>\t[REQUIRED] Local endpoint to forward (e.g., 127.0.0.1:22)\n")
	fmt.Fprintf(os.Stderr, "  --name <name>\t\tName for the tunnel (optional)\n")

	// HTTP Tunnel options
	fmt.Fprintf(os.Stderr, "HTTP Tunnel Options (--tunnel_type http):\n")
	fmt.Fprintf(os.Stderr, "  --local <endpoint>\t[REQUIRED] Local HTTP server to forward (e.g., 127.0.0.1:8080)\n")
	fmt.Fprintf(os.Stderr, "  --name <name>\t\tName for the tunnel (optional)\n")
	fmt.Fprintf(os.Stderr, "  --username <user>\tUsername for authentication (optional)\n")
	fmt.Fprintf(os.Stderr, "  --password <pass>\tPassword for authentication (optional)\n")
	fmt.Fprintf(os.Stderr, "  --hostname <host>\tRewritten hostname on headers (optional)\n")
	fmt.Fprintf(os.Stderr, "  --tls\t\t\tEnable TLS (optional)\n")

	// File Tunnel options
	fmt.Fprintf(os.Stderr, "File Tunnel Options (--tunnel_type file):\n")
	fmt.Fprintf(os.Stderr, "  --local <path>\t[REQUIRED] Local directory to share (e.g., ./files)\n")
	fmt.Fprintf(os.Stderr, "  --name <name>\t\tName for the tunnel (optional)\n")
	fmt.Fprintf(os.Stderr, "  --username <user>\tUsername for authentication (optional)\n")
	fmt.Fprintf(os.Stderr, "  --password <pass>\tPassword for authentication (optional)\n\n")

	// TCP Entrypoint options
	fmt.Fprintf(os.Stderr, "TCP Entrypoint Options (--entrypoint --tcp):\n")
	fmt.Fprintf(os.Stderr, "  --entrypoint\t\tCreate an entrypoint\n")
	fmt.Fprintf(os.Stderr, "  --tcp\t\t\tUse TCP protocol\n")
	fmt.Fprintf(os.Stderr, "  --local <endpoint>\t[REQUIRED] Local endpoint to forward to (e.g., 127.0.0.1:22)\n")
	fmt.Fprintf(os.Stderr, "  --name <name>\t\tName for the entrypoint (optional)\n")
	fmt.Fprintf(os.Stderr, "  --tunnel_id <id>\t[REQUIRED] Tunnel ID to connect to\n")

	// UDP Entrypoint options
	fmt.Fprintf(os.Stderr, "UDP Entrypoint Options (--entrypoint --udp):\n")
	fmt.Fprintf(os.Stderr, "  --entrypoint\t\tCreate an entrypoint\n")
	fmt.Fprintf(os.Stderr, "  --udp\t\t\tUse UDP protocol\n")
	fmt.Fprintf(os.Stderr, "  --local <endpoint>\t[REQUIRED] Local endpoint to forward to (e.g., 127.0.0.1:553)\n")
	fmt.Fprintf(os.Stderr, "  --name <name>\t\tName for the entrypoint (optional)\n")
	fmt.Fprintf(os.Stderr, "  --tunnel_id <id>\t[REQUIRED] Tunnel ID to connect to\n")
	fmt.Fprintf(os.Stderr, "  --ttl <seconds>\tTime to live for UDP packets with keep-alive (optional)\n\n")

	// Global options
	fmt.Fprintf(os.Stderr, "Global Options:\n")
	fmt.Fprintf(os.Stderr, "  --help, -h\t\tShow this help message\n")
	fmt.Fprintf(os.Stderr, "  --version, -v\t\tShow version information\n")
	fmt.Fprintf(os.Stderr, "  --no-stats\t\tDisable statistics for daemon mode\n")
	fmt.Fprintf(os.Stderr, "  --stats-interval\tStats update interval (default: 1s)\n")
	fmt.Fprintf(os.Stderr, "  --monitor-interval\tTunnel connection monitoring interval (default: 60s)\n\n")

	// Management commands
	fmt.Fprintf(os.Stderr, "Management Commands:\n")
	fmt.Fprintf(os.Stderr, "  --list\t\tList all tunnels and entrypoints\n")
	fmt.Fprintf(os.Stderr, "  --tunnels\t\tList all tunnels\n")
	fmt.Fprintf(os.Stderr, "  --entrypoints\t\tList all entrypoints\n")
	fmt.Fprintf(os.Stderr, "  --delete <id>\t\tDelete tunnel or entrypoint by ID\n\n")

	// Examples
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  # Start all configured tunnels and entrypoints\n")
	fmt.Fprintf(os.Stderr, "  %s\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # Create a simple HTTP tunnel\n")
	fmt.Fprintf(os.Stderr, "  %s --local localhost:8080 --name web-service\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # Create a TCP tunnel\n")
	fmt.Fprintf(os.Stderr, "  %s --local 192.168.1.100:22 --tunnel_type tcp --name \"SSH Access\"\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # Create a TCP entrypoint\n")
	fmt.Fprintf(os.Stderr, "  %s --entrypoint --tcp --local localhost:2222 --tunnel_id \"tunnel-id-here\" --name \"SSH Access\"\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # Create a UDP entrypoint with TTL\n")
	fmt.Fprintf(os.Stderr, "  %s --entrypoint --udp --local localhost:553 --tunnel_id \"tunnel-id-here\" --name \"DNS Server\" --ttl 120\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # Run as a daemon without statistics\n")
	fmt.Fprintf(os.Stderr, "  %s --no-stats\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # List tunnels and entrypoints\n")
	fmt.Fprintf(os.Stderr, "  %s --list\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # List tunnels\n")
	fmt.Fprintf(os.Stderr, "  %s --tunnels\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # List entrypoints\n")
	fmt.Fprintf(os.Stderr, "  %s --entrypoints\n\n", appName())

	fmt.Fprintf(os.Stderr, "  # Delete tunnel or entrypoint\n")
	fmt.Fprintf(os.Stderr, "  %s --delete \"tunnel-or-entrypoint-id\"\n\n", appName())
}

// Returns a formatted version string
func GetVersion() string {
	return fmt.Sprintf("GOST+ CLI Version %s\n", version.Version)
}

// Deletes a tunnel by ID
func deleteTunnel(id string) error {
	if tunnel.Count() == 0 {
		return nil
	}

	tun := tunnel.Get(id) // nilable
	if tun == nil {
		return nil
	}

	em := entitymanager.NetworkEntityManager{
		ID:         id,
		Entity:     &tun,
		Delete:     tunnel.Delete,
		SaveConfig: tunnel.SaveConfig,
	}
	if err := em.DeleteEntity("Tunnel"); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting a tunnel: %v\n", err)
		return err
	}
	return nil
}

// Deletes an entrypoint by ID
func deleteEntryPoint(id string) error {
	if entrypoint.Count() == 0 {
		return nil
	}

	ep := entrypoint.Get(id)
	if ep == nil {
		return nil
	}

	em := entitymanager.NetworkEntityManager{
		ID:         id,
		Entity:     &ep,
		Delete:     entrypoint.Delete,
		SaveConfig: entrypoint.SaveConfig,
	}
	if err := em.DeleteEntity("EntryPoint"); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting an entrypoint: %v", err)
		return err
	}
	return nil
}

// Lists all configured tunnels
func listAllTunnels() {
	count := tunnel.Count()
	listEntities("tunnels", entitymanager.MakeEntityIterator(count, tunnel.GetIndex))
}
func listAllEntryPoints() {
	count := entrypoint.Count()
	listEntities("entrypoints", entitymanager.MakeEntityIterator(count, entrypoint.GetIndex))
}

func listEntities(entitiesType string, next entitymanager.EntityIteratorFn) {
	fmt.Printf("\nAvailable %s:\n", entitiesType)
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Printf("%-38s %-20s %-10s %-10s\n", "ID", "NAME", "TYPE", "STATUS")
	fmt.Println("------------------------------------------------------------------------------")

	for {
		manager, ok := next()
		if !ok {
			break
		}
		entityPtr := manager.Entity
		if entityPtr == nil {
			continue
		}
		entity := *entityPtr

		fmt.Printf("%-38s %-20s %-10s %-10s\n",
			entity.ID(),
			entity.Name(),
			entity.Type(),
			utils.GetDisplayState(entity))
		fmt.Printf("Entrypoint: %s\n", entity.Entrypoint())
		fmt.Println("------------------------------------------------------------------------------")
	}
}

/**
 * Creates and runs a new tunnel
 */
func createTunnel(flags CommandFlags) tunnel.Tunnel {
	tunnelType := strings.ToLower(flags.TunnelType)

	validTypes := map[string]bool{
		tunnel.HTTPTunnel: true,
		tunnel.FileTunnel: true,
		tunnel.TCPTunnel:  true,
		tunnel.UDPTunnel:  true,
	}

	// Validate a supported tunnel type
	if !validTypes[tunnelType] {
		fmt.Printf("Unsupported tunnel type: %s. Valid: http, file, tcp, udp\n", tunnelType)
		os.Exit(1)
	}

	// Check if the tunnel exists
	for i := range tunnel.Count() {
		tun := tunnel.GetIndex(i)
		if tun == nil {
			continue
		}

		localEndpoint := tun.Endpoint()
		if localEndpoint == flags.LocalEndpoint && tun.Type() == tunnelType {
			fmt.Printf("%s tunnel '%s' is found in state: %v, ID: %s\n",
				strings.ToUpper(tun.Type()),
				localEndpoint,
				utils.GetState(tun),
				tun.ID())
			fmt.Println("Remove it first to update or create a new one")
			return nil
		}
	}

	defer tunnel.SaveConfig()

	options := tunnel.Options{
		Name:     flags.ServiceName,
		Endpoint: flags.LocalEndpoint,
		Username: flags.Username,
	}
	if flags.Password != "" {
		options.Password.Set(flags.Password)
	}
	if tunnelType == tunnel.HTTPTunnel {
		options.EnableTLS = flags.EnableTLS
		options.Hostname = flags.Hostname
	}

	var newTunnel = tunnel.CreateTunnel(tunnelType, options)
	if newTunnel == nil {
		fmt.Printf("Failed to create %s tunnel\n", tunnelType)
		os.Exit(1)
	}

	// Add tunnel to the global list for stats tracking
	tunnel.Add(newTunnel)

	// Run the tunnel
	if err := newTunnel.Run(); err != nil {
		newTunnel.Close()
		fmt.Printf("Failed to start tunnel: %v\n", err)
		os.Exit(1)
	}

	// Print out the tunnel information
	fmt.Printf("\n%s tunnel is started:\n", strings.ToUpper(tunnelType))
	fmt.Printf("ID: %s\n", newTunnel.ID())
	fmt.Printf("Name: %s\n", newTunnel.Name())

	if flags.TunnelType == tunnel.FileTunnel {
		fmt.Printf("Local folder: %s\n", newTunnel.Endpoint())
		fmt.Printf("Remote entrypoint: %s\n", newTunnel.Entrypoint())
	} else if flags.TunnelType == tunnel.HTTPTunnel {
		fmt.Printf("Local endpoint: %s\n", newTunnel.Endpoint())
		fmt.Printf("Remote entrypoint: %s\n", newTunnel.Entrypoint())
	} else if flags.TunnelType == tunnel.TCPTunnel {
		fmt.Printf("TCP tunnel, endpoint: %s\n", newTunnel.Endpoint())
	} else {
		fmt.Printf("UDP tunnel, endpoint: %s\n", newTunnel.Endpoint())
	}

	return newTunnel
}

/**
 * Creates a new entrypoint and connect it to a remote tunnel
 */
func createEntryPoint(flags CommandFlags) error {
	defer entrypoint.SaveConfig()

	var ep entrypoint.EntryPoint
	tunnelId := strings.TrimSpace(flags.TunnelID)
	idOpt := tunnel.IDOption(tunnelId)
	nameOpt := tunnel.NameOption(strings.TrimSpace(flags.ServiceName))
	endpointOpt := tunnel.EndpointOption(strings.TrimSpace(flags.LocalEndpoint))

	// Check if the entrypoint exists
	ep = entrypoint.Get(tunnelId)
	if ep != nil {
		fmt.Printf("%s entrypoint '%s' is found in state: %v, ID: %s\n",
			strings.ToUpper(ep.Type()),
			ep.Endpoint(),
			utils.GetState(ep),
			ep.ID())
		fmt.Println("Remove it first to update or create a new one")
		return nil
	}

	if flags.TCPFlag {
		ep = entrypoint.NewTCPEntryPoint(idOpt, nameOpt, endpointOpt)
	} else { // UDP
		var keepaliveOpt tunnel.Option
		var ttlOpt tunnel.Option
		if flags.TTLSec > 0 {
			keepaliveOpt = tunnel.KeepaliveOption(true)
			ttlOpt = tunnel.TTLOption(flags.TTLSec)
		}
		ep = entrypoint.NewUDPEntryPoint(idOpt, nameOpt, endpointOpt, keepaliveOpt, ttlOpt)
	}

	entrypoint.Add(ep)

	if err := ep.Run(); err != nil {
		ep.Close()
		return err
	}

	// Print out the entrypoint information
	fmt.Printf("\n%s entrypoint is connected to %s:\n", strings.ToUpper(ep.Type()), ep.Endpoint())
	fmt.Printf("ID: %s\n", ep.ID())
	fmt.Printf("Name: %s\n", ep.Name())
	fmt.Printf("Listening to %s locally\n", ep.Entrypoint())

	return nil
}

// Start statistics updater
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

// Waits for a signal shutdown
func waitForShutdown(doneChan chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	close(doneChan) // Signal stats display to stop

	fmt.Println("\nDone")
}

// Starts monitoring of tunnels and entrypoints connections
func startTunnelMonitor(ctx context.Context, interval time.Duration) error {
	return startTunnelMonitorWithDelay(ctx, interval, 0)
}

// Starts monitoring of tunnels and entrypoints connections with initial delay
func startTunnelMonitorWithDelay(ctx context.Context, interval time.Duration, delay time.Duration) error {
	err := runner.Exec(ctx, task.NewMonitorTask(),
		runner.WithAsync(true),
		runner.WithDelay(delay),
		runner.WithInterval(interval),
		runner.WithCancel(true),
	)
	if err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}
