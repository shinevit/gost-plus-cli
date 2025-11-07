package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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
	args "github.com/go-gost/gost.plus/utils/cli"
	"github.com/go-gost/gost.plus/utils/fp/lazy"
	opt "github.com/go-gost/gost.plus/utils/fp/option"
	"github.com/go-gost/gost.plus/utils/fp/slice"
	"github.com/go-gost/gost.plus/version"
	_ "github.com/go-gost/gost.plus/winres"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type Nothing struct{}

var (
	_lazyInit = lazy.NewLazy(func() Nothing {
		// This initialization block is called once
		config.Init()
		tunnel.LoadFromConfig()
		entrypoint.LoadFromConfig()
		return Nothing{}
	})

	initOnce = func() {
		_lazyInit.Get()
	}
)

// Global flags values
var (
	statsInterval   time.Duration
	monitorInterval time.Duration
	statsDisabled   bool
)

const (
	DefaultStatsInterval   = time.Second
	DefaultMonitorInterval = time.Minute
)

func main() {
	app := &cli.App{
		Name:    appName(),
		Version: version.Version,
		Description: "The GOST+ CLI tool helps tunnel local and internal services over the internet,\n" +
			"inspects traffic, and automatically recovers connectivity.",
		Usage:          "secure tunneling tool for sharing local services over the internet",
		UsageText:      getMainUsageText(),
		DefaultCommand: "start",
		CommandNotFound: func(c *cli.Context, command string) {
			// Check if the command is a valid command (starts with a letter, followed by word characters)
			isCommand := regexp.MustCompile(`^[a-zA-Z]\w*$`).MatchString(command)
			if !isCommand {
				// it's not a valid command (could be a redirection, pipe, etc.)
				return
			}

			fmt.Fprintf(c.App.ErrWriter, "Unknown command %q\n\n", command)
			cli.ShowAppHelp(c)
			fmt.Fprintf(c.App.ErrWriter, "\nUnknown command %q\n", command)
			os.Exit(1)
		},
		Flags: []cli.Flag{
			getStatsIntervalFlag(),
			getMonitorIntervalFlag(),
			getDisableStatsFlag(),
		},
		Before: func(c *cli.Context) error {
			initOnce() // it's called before each command
			parser := args.NewCLIArgsParser()
			err := parser.ParseArgs(c.Args())
			if err == nil {
				parser.DurationVar(&statsInterval, c.Duration("stats"), "stats", "stats-interval")
				parser.DurationVar(&monitorInterval, c.Duration("monitor"), "monitor", "monitor-interval")
				parser.BoolVar(&statsDisabled, c.Bool("no-stats"), "no-stats")
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:      "http",
				Usage:     "Create HTTP tunnel",
				UsageText: getHttpTunnelUsageText(),
				Flags:     getHttpTunnelCommandFlags(),
				Action:    createTunnelAction("http"),
			},
			{
				Name:        "tls",
				Usage:       "Create HTTPS-based secured tunnel",
				UsageText:   getTLSTunnelUsageText(),
				Description: `Export https local services to public access`,
				Flags:       getHttpTunnelCommandFlags(),
				Action:      createTunnelAction("tls"),
			},
			{
				Name:      "file",
				Usage:     "Create File sharing tunnel",
				UsageText: getFileTunnelUsageText(),
				Flags:     getFileTunnelCommandFlags(),
				Action:    createTunnelAction("file"),
			},
			{
				Name:      "tcp",
				Usage:     "Create TCP tunnel",
				UsageText: getTCPTunnelUsageText(),
				Flags: slices.Concat(getTunnelNameCommandFlags(),
					[]cli.Flag{
						getStatsIntervalFlag(),
						getMonitorIntervalFlag(),
						getDisableStatsFlag(),
					}),
				Action: createTunnelAction("tcp"),
			},
			{
				Name:      "udp",
				Usage:     "Create UDP tunnel",
				UsageText: getUDPTunnelUsageText(),
				Flags: slices.Concat(getTunnelNameCommandFlags(),
					[]cli.Flag{
						getStatsIntervalFlag(),
						getMonitorIntervalFlag(),
						getDisableStatsFlag(),
					}),
				Action: createTunnelAction("udp"),
			},
			{
				Name:      "bind",
				Usage:     "Binds a tunnel to a local TCP or UDP port, creates an entrypoint",
				UsageText: getEntrypointsUsageText(),
				Flags:     getEntrypointsCommandFlags(),
				Action:    createEntrypointAction(),
			},
			{
				Name:      "start",
				Usage:     "Start an existing tunnel and/or entrypoint from the config",
				UsageText: getStartCommandUsageText(),
				Flags: slices.Concat(getTunnelNameCommandFlags(),
					[]cli.Flag{
						getStatsIntervalFlag(),
						getMonitorIntervalFlag(),
						getDisableStatsFlag(),
					}),
				Action: handleStartAction(),
			},
			{
				Name:      "list",
				Aliases:   []string{"ls"},
				Usage:     "List all tunnels and entrypoints",
				UsageText: fmt.Sprintf("%s list\n", appName()) + fmt.Sprintf("%s l\n", appName()) + fmt.Sprintf("%s l -h", appName()),
				Action: func(c *cli.Context) error {
					listAllTunnels()
					listAllEntryPoints()
					return nil
				},
			},
			{
				Name:      "tunnels",
				Aliases:   []string{"tl"},
				Usage:     "List tunnels",
				UsageText: fmt.Sprintf("%s tunnels\n", appName()) + fmt.Sprintf("%s tl\n", appName()) + fmt.Sprintf("%s tl -h", appName()),
				Action: func(c *cli.Context) error {
					listAllTunnels()
					return nil
				},
			},
			{
				Name:      "entrypoints",
				Aliases:   []string{"el"},
				Usage:     "List entrypoints",
				UsageText: fmt.Sprintf("%s entrypoints\n", appName()) + fmt.Sprintf("%s el\n", appName()) + fmt.Sprintf("%s el -h", appName()),
				Action: func(c *cli.Context) error {
					listAllEntryPoints()
					return nil
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"d"},
				Usage:     "Delete a tunnel or entrypoint by ID",
				UsageText: fmt.Sprintf("%s delete <ID>\n", appName()) + fmt.Sprintf("%s d <ID>\n", appName()) + fmt.Sprintf("%s d -h", appName()),
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return cli.Exit("Error: missing ID for delete command", 1)
					}
					id := c.Args().First()
					err1 := deleteTunnel(id)
					err2 := deleteEntryPoint(id)
					if err1 != nil && err2 != nil {
						return cli.Exit(fmt.Sprintf("Error: No tunnel or entrypoint found with ID: %s", id), 1)
					}
					return nil
				},
			},
			{
				Name:      "showConfig",
				Aliases:   []string{"sc"},
				Usage:     "Show the application configuration safely",
				UsageText: fmt.Sprintf("%s showConfig\n", appName()) + fmt.Sprintf("%s sc\n", appName()) + fmt.Sprintf("%s sc -h", appName()),
				Action:    showConfigSafeAction(),
			},
			{
				Name:      "env",
				Usage:     "Show app environment variables",
				UsageText: fmt.Sprintf("%s env\n", appName()) + fmt.Sprintf("%s env -h", appName()),
				Action: func(c *cli.Context) error {
					fmt.Println("\nEnvironment variables:")
					fmt.Println("   GOST_CONFIG_DIR", "   sets root app folder for config")
					fmt.Println("   AUTH_PASSWORD  ", "   sets an authentication password for HTTP/File tunnels")
					fmt.Println()
					return nil
				},
			},
			{
				Name:      "paths",
				Usage:     "Show app paths",
				UsageText: fmt.Sprintf("%s paths\n", appName()) + fmt.Sprintf("%s paths -h", appName()),
				Action: func(c *cli.Context) error {
					fmt.Println("\nConfig path: ", config.ConfigFilePath())
					fmt.Println("Log path:    ", config.LogFilePath())
					fmt.Println()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "\nIncorrect Usage: %v\n", err)
	}
}

func appName() string {
	return filepath.Base(os.Args[0])
}

func getStatsIntervalFlag() cli.Flag {
	return &cli.DurationFlag{
		Name:    "stats-interval",
		Aliases: []string{"stats"},
		Value:   DefaultStatsInterval,
		Usage:   "stats update interval",
	}
}

func getMonitorIntervalFlag() cli.Flag {
	return &cli.DurationFlag{
		Name:    "monitor-interval",
		Aliases: []string{"monitor"},
		Value:   DefaultMonitorInterval,
		Usage:   "tunnel connection monitoring interval",
	}
}

func getDisableStatsFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:  "no-stats",
		Usage: "disable statistics for a daemon mode",
	}
}

// http/tls + file tunnel
func getAuthCommandFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "user name for authentication",
		},
		&cli.StringFlag{
			Name:    "password",
			EnvVars: []string{"AUTH_PASSWORD"},
			Usage:   "password for authentication",
		},
	}
}

func getTunnelNameCommandFlags() []cli.Flag {
	return []cli.Flag{
		getNameCommandFlag("tunnel"),
	}
}

func getNameCommandFlag(itemName string) cli.Flag {
	return &cli.StringFlag{
		Name:     "name",
		Aliases:  []string{"n"},
		Usage:    itemName + " name",
		Required: false,
	}
}

func getMainUsageText() string {
	usages := []string{
		fmt.Sprintf("%s [global options] <command> [command options]", appName()),
		"or",
		fmt.Sprintf("%s <command> [command options] [global options]", appName()),
	}
	return strings.Join(usages, "\n")
}

func getHttpTunnelUsageText() string {
	usages := []string{
		fmt.Sprintf("%s http [host:]port [options]\n", appName()),

		fmt.Sprintf("AUTH_PASSWORD=secret %s http 80 -n \"Name\" -u test_user -hostname iot-host", appName()),
		fmt.Sprintf("%s http localhost:80 -n \"Name\"", appName()),
		fmt.Sprintf("%s http 192.168.0.100:80 -n \"Name\"", appName()),
	}
	return strings.Join(usages, "\n")
}

func getHttpTunnelCommandFlags() []cli.Flag {
	flags := []cli.Flag{
		getNameCommandFlag("tunnel"),
	}

	flags = slices.Concat(flags, getAuthCommandFlags())

	flags = append(flags, &cli.StringFlag{
		Name:  "hostname",
		Usage: "rewritten hostname on headers",
	})

	flags = append(flags, getStatsIntervalFlag())
	flags = append(flags, getMonitorIntervalFlag())
	flags = append(flags, getDisableStatsFlag())

	return flags
}

func getFileTunnelCommandFlags() []cli.Flag {
	flags := slices.Concat(getTunnelNameCommandFlags(), getAuthCommandFlags())

	flags = append(flags, getStatsIntervalFlag())
	flags = append(flags, getMonitorIntervalFlag())
	flags = append(flags, getDisableStatsFlag())

	return flags
}

func getTLSTunnelUsageText() string {
	usages := []string{
		fmt.Sprintf("%s tls [host:]port [options]\n", appName()),

		fmt.Sprintf("AUTH_PASSWORD=secret %s tls 443 -n \"Name\" -u test_user -hostname iot-host", appName()),
		fmt.Sprintf("%s tls localhost:443 -n \"Name\"", appName()),
		fmt.Sprintf("%s tls 192.168.0.100:443 -n \"Name\"", appName()),
	}
	return strings.Join(usages, "\n")
}

func getFileTunnelUsageText() string {
	usages := []string{
		fmt.Sprintf("%s file <files_dir> [options]\n", appName()),

		fmt.Sprintf("AUTH_PASSWORD=secret %s file \"./home/user/work docs\" -n \"Name\" -u test_user", appName()),
		fmt.Sprintf("%s file . -n \"Name\"", appName()),
	}
	return strings.Join(usages, "\n")
}

func getTCPTunnelUsageText() string {
	usages := []string{
		fmt.Sprintf("%s tcp [host:]port [options]\n", appName()),

		fmt.Sprintf("%s tcp 22 -n \"Name\"", appName()),
		fmt.Sprintf("%s tcp localhost:22 -n \"Name\"", appName()),
		fmt.Sprintf("%s tcp 192.168.0.100:22 -n \"Name\"", appName()),
	}
	return strings.Join(usages, "\n")
}

func getUDPTunnelUsageText() string {
	usages := []string{
		fmt.Sprintf("%s udp [host:]port [options]\n", appName()),

		fmt.Sprintf("%s udp 53 -n \"Name\"", appName()),
		fmt.Sprintf("%s udp localhost:53 -n \"Name\"", appName()),
		fmt.Sprintf("%s udp 192.168.0.100:53 -n \"Name\"", appName()),
	}
	return strings.Join(usages, "\n")
}

func getEntrypointsUsageText() string {
	usages := []string{
		fmt.Sprintf("%s bind <ID> -tcp [host:]port -n \"Name\"", appName()),
		fmt.Sprintf("%s bind <ID> -tcp port -n \"Name\"", appName()),
		fmt.Sprintf("%s bind <ID> -tcp localhost:port -n \"Name\"\n", appName()),

		fmt.Sprintf("%s bind <ID> -udp [host:]port -n \"Name\" -ttl 120s", appName()),
		fmt.Sprintf("%s bind <ID> -udp port -n \"Name\" -ttl 120s", appName()),
		fmt.Sprintf("%s bind <ID> -udp localhost:port -n \"Name\" -ttl 120s", appName()),
	}
	return strings.Join(usages, "\n")
}

func getStartCommandUsageText() string {
	usages := []string{
		fmt.Sprintf("%s start [options]\nor", appName()),
		fmt.Sprintf("%s start <ID> [options]\n", appName()),

		fmt.Sprintf("%s start -name Name", appName()),
		fmt.Sprintf("%s start -n Name -no-stats", appName()),
		fmt.Sprintf("%s start a7a1c126-970b-4886-8ff5-455902abcfd3", appName()),
		fmt.Sprintf("%s start -h", appName()),
	}
	return strings.Join(usages, "\n")
}

func getEntrypointsCommandFlags() []cli.Flag {
	tcpAddressFlag := &cli.StringFlag{
		Name:  "tcp",
		Usage: "local TCP endpoint is listening to",
	}
	udpAddressFlag := &cli.StringFlag{
		Name:  "udp",
		Usage: "local UDP endpoint is listening to",
	}
	ttlFlag := &cli.DurationFlag{
		Name:  "ttl",
		Usage: "time to live for UDP, sec",
	}

	return []cli.Flag{
		getNameCommandFlag("entrypoint"),
		tcpAddressFlag,
		udpAddressFlag,
		ttlFlag,
		getStatsIntervalFlag(),
		getDisableStatsFlag(),
	}
}

func createTunnelAction(cmd string) cli.ActionFunc {
	return func(c *cli.Context) error {
		if c.NArg() == 0 {
			return cli.Exit(fmt.Sprintf("Error: missing address for %s command", cmd), 1)
		}

		var (
			target    string
			name      string
			username  string
			password  string
			hostname  string
			enableTLS bool
		)

		parser := args.NewCLIArgsParser()
		err := parser.ParseArgs(c.Args())
		if err != nil {
			return cli.Exit(fmt.Sprintf("parsing arguments error: %v", err), 1)
		}

		parser.StringVar(&name, c.String("n"), "name", "n")
		parser.StringVar(&username, c.String("u"), "username", "u")
		password = c.String("password")
		parser.StringVar(&hostname, c.String("hostname"), "hostname")
		target = parser.UnboundedArg(0, "")

		if len(username) > 0 && len(password) == 0 {
			return cli.Exit("Incorrect Usage: both 'username' and $AUTH_PASSWORD are required", 1)
		}

		var tunnelType string
		switch cmd {
		case tunnel.HTTPTunnel:
			tunnelType = tunnel.HTTPTunnel
		case "tls":
			tunnelType = tunnel.HTTPTunnel
			enableTLS = true
		case tunnel.FileTunnel:
			tunnelType = tunnel.FileTunnel
		case tunnel.TCPTunnel:
			tunnelType = tunnel.TCPTunnel
		case tunnel.UDPTunnel:
			tunnelType = tunnel.UDPTunnel
		default:
			// Should not happen with the current commands setup
			return cli.Exit(fmt.Sprintf("unknown command: %s", cmd), 1)
		}

		// Parse endpoint
		if strings.HasPrefix(target, ":") { // :553
			target = "localhost" + target
		} else if _, _, err := net.SplitHostPort(target); err != nil { // 192.168.0.100:553
			_, err := strconv.Atoi(target)
			if err != nil { // 553
				return cli.Exit(fmt.Sprintf("Invalid port number '%v'", target), 1)
			} else {
				// for commands like 'gost-tunnel http 8080', default to localhost
				target = "localhost:" + target
			}
		}

		options := tunnel.Options{
			Name:      name,
			Endpoint:  target,
			Username:  username,
			EnableTLS: enableTLS,
			Hostname:  hostname,
		}
		if len(password) > 0 {
			options.Password.Set(password)
		}

		newTun := tunnel.CreateTunnel(tunnelType, options)
		if newTun == nil {
			return cli.Exit(fmt.Sprintf("Error of creating %s tunnel", tunnelType), 1)
		}

		// Check if a tunnel with the same name or a service endpoint exists
		exists := slice.Exists(tunnel.GetAll(), func(tun tunnel.Tunnel) bool {
			return tun.Name() == newTun.Name() || tun.Endpoint() == newTun.Endpoint()
		})
		if exists {
			msg := fmt.Sprintf("Duplicate tunnel detected by name or endpoint. Name: '%s', endpoint: %s", newTun.Name(), newTun.Endpoint())
			return cli.Exit(msg, 1)
		}

		tunnel.Add(newTun)
		tunnel.SaveConfig() // persist the tunnel

		fmt.Printf("\n%s tunnel created successfully:\n", strings.ToUpper(tunnelType))
		fmt.Printf("   ID: %s\n", newTun.ID())
		fmt.Printf("   Name: %s\n", newTun.Name())
		fmt.Printf("   Endpoint: %s\n", newTun.Entrypoint())
		fmt.Printf("   Forwarding to: %s\n", newTun.Endpoint())

		run(c) // Run all after creating this one
		return nil
	}
}

func createEntrypointAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if c.NArg() == 0 {
			return cli.Exit(fmt.Sprintf("Error: missing tunnel ID"), 1)
		}

		var (
			name          string
			tunnelId      string
			tcpEndpoint   string
			udpEndpoint   string
			localEndpoint string
			isTCP         bool
			ttl           time.Duration
		)

		parser := args.NewCLIArgsParser()
		err := parser.ParseArgs(c.Args())
		if err != nil {
			return cli.Exit(fmt.Sprintf("parsing arguments error: %v", err), 1)
		}

		parser.StringVar(&name, c.String("n"), "name", "n")
		parser.StringVar(&tcpEndpoint, c.String("tcp"), "tcp")
		parser.StringVar(&udpEndpoint, c.String("udp"), "udp")
		parser.DurationVar(&ttl, c.Duration("ttl"), "ttl")

		ttlSec := ttl.Milliseconds() / 1000
		tunnelId = parser.UnboundedArg(0, "")

		if _, err := uuid.Parse(tunnelId); err != nil {
			return cli.Exit(fmt.Sprintf("Error: invalid tunnel ID format '%s'", tunnelId), 1)
		}

		if len(tcpEndpoint) == 0 && len(udpEndpoint) == 0 {
			return cli.Exit("Incorrect Usage: Required flag \"tcp or udp\" is not set", 1)
		} else if len(tcpEndpoint) > 0 && len(udpEndpoint) > 0 {
			return cli.Exit("Incorrect Usage: Only one of flags \"tcp or udp\" is required", 1)
		}

		if len(tcpEndpoint) > 0 {
			localEndpoint = tcpEndpoint
			isTCP = true
		} else if len(udpEndpoint) > 0 {
			localEndpoint = udpEndpoint
		}

		// parse endpoint
		if strings.HasPrefix(localEndpoint, ":") { // :553
			localEndpoint = "localhost" + localEndpoint
		} else if _, _, err := net.SplitHostPort(localEndpoint); err != nil { // 192.168.0.100:553
			_, err := strconv.Atoi(localEndpoint)
			if err != nil { // 553
				return cli.Exit(fmt.Sprintf("Invalid port number '%v'", localEndpoint), 1)
			} else {
				localEndpoint = "localhost:" + localEndpoint
			}
		}

		idOpt := tunnel.IDOption(strings.ToLower(tunnelId))
		nameOpt := tunnel.NameOption(name)
		endpointOpt := tunnel.EndpointOption(localEndpoint)

		var newEp entrypoint.EntryPoint
		if isTCP {
			newEp = entrypoint.NewTCPEntryPoint(idOpt, nameOpt, endpointOpt)
		} else if ttlSec > 0 { // udp
			var ttlOpt, keepaliveOpt tunnel.Option
			ttlOpt = tunnel.TTLOption(int(ttlSec))
			keepaliveOpt = tunnel.KeepaliveOption(true)
			newEp = entrypoint.NewUDPEntryPoint(idOpt, nameOpt, endpointOpt, ttlOpt, keepaliveOpt)
		} else { // udp
			newEp = entrypoint.NewUDPEntryPoint(idOpt, nameOpt, endpointOpt)
		}

		exists := slice.Exists(entrypoint.GetAll(), func(ep entrypoint.EntryPoint) bool {
			return newEp.ID() == ep.ID() ||
				newEp.Name() == ep.Name() ||
				newEp.Entrypoint() == ep.Entrypoint() && newEp.Type() == ep.Type()
		})
		if exists {
			msg := fmt.Sprintf("Duplicate entrypoint detected by id, name or local endpoint. Name: '%s', endpoint: %s", newEp.Name(), newEp.Entrypoint())
			return cli.Exit(msg, 1)
		}

		entrypoint.Add(newEp)
		entrypoint.SaveConfig()

		fmt.Printf("\n%s entrypoint created successfully:\n", strings.ToUpper(newEp.Type()))
		fmt.Printf("   ID: %s\n", newEp.ID())
		fmt.Printf("   Name: %s\n", newEp.Name())
		fmt.Printf("   Listening on %s\n", newEp.Entrypoint())
		fmt.Printf("   Forwarding to remote endpoint: %s\n\n", newEp.Endpoint())

		run(c) // Run all after creating this one
		return nil
	}
}

// starts a specific tunnel or entrypoint by ID or -name or all if no any arguments provided
func handleStartAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		name := c.String("name")
		parser := args.NewCLIArgsParser()
		parser.ParseArgs(c.Args())

		if id := parser.UnboundedArg(0, ""); id != "" && name == "" {
			if _, err := uuid.Parse(id); err != nil {
				return cli.Exit(fmt.Sprintf("Invalid ID format: %s", id), 1)
			}

			id = strings.ToLower(id)
			tun := tunnel.Get(id)
			ep := entrypoint.Get(id)
			if tun != nil && ep != nil {
				fmt.Println("Starting specific tunnel and entrypoint by ID...")
				return runEntities([]tunnel.Tunnel{tun}, []entrypoint.EntryPoint{ep})
			} else if tun != nil {
				fmt.Println("Starting the tunnel with ID...")
				return runEntities([]tunnel.Tunnel{tun}, []entrypoint.EntryPoint{})
			} else if ep != nil {
				fmt.Println("Starting the entrypoint with ID...")
				return runEntities([]tunnel.Tunnel{}, []entrypoint.EntryPoint{ep})
			}

			return cli.Exit(fmt.Sprintf("No tunnel or entrypoint found with ID: %s", id), 0)
		}

		// If no name or ID is provided, start all
		if name == "" {
			fmt.Println("Starting all configured tunnels and entrypoints...")
			run(c)
			return nil
		}

		// Try to find by name
		var tunnels []tunnel.Tunnel
		var entrypoints []entrypoint.EntryPoint

		if tun := getTunnelFromConfig(name); tun != nil {
			tunnels = append(tunnels, tun)
		}
		if ep := getEntrypointFromConfig(name); ep != nil {
			entrypoints = append(entrypoints, ep)
		}

		if !slice.Any(slices.Concat(tunnels, entrypoints)) {
			return cli.Exit(fmt.Sprintf("No tunnel or entrypoint found with name: %s", name), 0)
		}

		fmt.Println("Starting specific tunnel/entrypoint...")
		return runEntities(tunnels, entrypoints)
	}
}

func showConfigSafeAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		cfg := config.Get()
		safeCfg := *cfg // create a deep copy
		safeCfg.Tunnels = make([]*config.Tunnel, len(cfg.Tunnels))
		for i, tun := range cfg.Tunnels {
			if tun == nil {
				continue
			}
			safeTun := *tun
			if !safeTun.Password.IsEmpty() {
				// hide an actual password if it's defined
				safeTun.Password = config.NewPassword(safeTun.Password.String())
			}
			safeCfg.Tunnels[i] = &safeTun
		}
		data, err := yaml.Marshal(safeCfg) // serialize the config to YAML
		if err != nil {
			return fmt.Errorf("failed to serialize config: %w", err)
		}
		fmt.Println()
		fmt.Println(string(data))
		return nil
	}
}

// finds a tunnel by its name
func getTunnelFromConfig(name string) tunnel.Tunnel {
	tunnelOpt := slice.First(tunnel.GetAll(), func(tun tunnel.Tunnel) bool {
		return tun != nil && strings.EqualFold(tun.Name(), name)
	})
	return tunnelOpt.GetOrElse(nil)
}

// finds an entrypoint by its name
func getEntrypointFromConfig(name string) entrypoint.EntryPoint {
	entrypointOpt := slice.First(entrypoint.GetAll(), func(ep entrypoint.EntryPoint) bool {
		return ep != nil && strings.EqualFold(ep.Name(), name)
	})
	return entrypointOpt.GetOrElse(nil)
}

// runs specific tunnels/entrypoints
func runEntities(tunnels []tunnel.Tunnel, entrypoints []entrypoint.EntryPoint) error {
	if !slice.Any(tunnels) && !slice.Any(entrypoints) {
		return nil
	}

	// Initialize context and shutdown handler
	ctx, cancel := waitForShutdown()
	defer cancel()

	runAll := func(items []tunnel.Tunnel) []tunnel.Tunnel {
		var started []tunnel.Tunnel
		slice.ForEach(items, func(item tunnel.Tunnel) {
			if err := item.Run(); err != nil {
				item.Close()
				fmt.Fprintf(os.Stderr, "%s %s failed: %v", strings.ToUpper(item.Type()), item.Name(), err)
				return
			}
			started = append(started, item)
		})
		return started
	}

	if slice.Any(tunnels) {
		startedTunnels := runAll(tunnels)
		listEntities("Started tunnels", startedTunnels)

		monitorInterval = opt.Cond(monitorInterval.Seconds() > 0, monitorInterval).GetOrElse(DefaultMonitorInterval)
		if err := startTunnelsMonitoring(ctx, monitorInterval, tunnels); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start tunnels monitoring: %v", err)
		}
	}
	if slice.Any(entrypoints) {
		startedEntrypoints := runAll(entrypoints)
		listEntities("Started entrypoints", startedEntrypoints)
	}

	if !statsDisabled {
		statsInterval = opt.Cond(statsInterval.Seconds() > 0, statsInterval).GetOrElse(DefaultStatsInterval)
		startStatsRunner(ctx, statsInterval)
		go stats.DisplayStats(ctx, statsInterval)
		fmt.Printf("\nPress Ctrl+C to exit\n\n")
	}

	// Wait for context cancellation triggered by the signal of interruption
	<-ctx.Done()
	fmt.Println("Done")
	return nil
}

// runs all tunnels and entrypoints from the config
func run(c *cli.Context) {
	// get all non-closed tunnels and entrypoints
	tunnels := slice.Filter(tunnel.GetAll(), func(tun tunnel.Tunnel) bool {
		return tun != nil && !tun.IsClosed()
	})
	entrypoints := slice.Filter(entrypoint.GetAll(), func(tun entrypoint.EntryPoint) bool {
		return tun != nil && !tun.IsClosed()
	})

	if !slice.Any(slices.Concat(tunnels, entrypoints)) {
		fmt.Fprintln(os.Stderr, "No tunnels or entrypoints are configured")
		cli.ShowAppHelpAndExit(c, 0)
	}

	if err := runEntities(tunnels, entrypoints); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Deletes a tunnel by ID
func deleteTunnel(id string) error {
	tun := tunnel.Get(id)
	if tun == nil {
		return fmt.Errorf("tunnel with id %s not found", id)
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
	ep := entrypoint.Get(id)
	if ep == nil {
		return fmt.Errorf("entrypoint with id %s not found", id)
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
	if count == 0 {
		fmt.Println("No tunnels are configured.")
		return
	}
	listEntities("Available tunnels", tunnel.GetAll())
}

func listAllEntryPoints() {
	count := entrypoint.Count()
	if count == 0 {
		fmt.Println("No entrypoints are configured.")
		return
	}
	listEntities("Available entrypoints", entrypoint.GetAll())
}

func listEntities(title string, items []tunnel.Tunnel) {
	fmt.Printf("\n%s:\n", title)
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Printf("%-38s %-20s %-10s %-10s\n", "ID", "NAME", "TYPE", "STATUS")
	fmt.Println("------------------------------------------------------------------------------")

	for _, entity := range items {
		if entity == nil {
			continue
		}
		fmt.Printf("%-38s %-20s %-10s %-10s\n",
			entity.ID(),
			entity.Name(),
			entity.Type(),
			utils.GetDisplayState(entity))
		fmt.Printf("Entrypoint: %s\n", entity.Entrypoint())
		fmt.Println("------------------------------------------------------------------------------")
	}
}

// Start statistics updater
func startStatsRunner(ctx context.Context, interval time.Duration) error {
	err := runner.Exec(ctx, task.UpdateStats(),
		runner.WithAsync(true),
		runner.WithInterval(interval),
		runner.WithCancel(true),
	)
	if err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}

// Starts monitoring of tunnels and entrypoints connections with an initial delay
func startTunnelsMonitoring(ctx context.Context, interval time.Duration, tunnels []tunnel.Tunnel) error {
	fmt.Printf("Starting tunnels monitoring with the interval %v...\n", interval)
	tunnelsID := slice.Map(tunnels, func(tun tunnel.Tunnel) string { return tun.ID() })
	err := runner.Exec(ctx, task.NewMonitorTaskFor(tunnelsID),
		runner.WithAsync(true),
		runner.WithDelay(interval),
		runner.WithInterval(interval),
		runner.WithCancel(true),
	)
	if err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}

// Waits for a shutdown signal
func waitForShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	return ctx, cancel
}
