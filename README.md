# GOST-PLUS CLI Tunnel

A cross-platform command-line client designed to tunnel local network services securely to the public internet using the [GOST.PLUS](https://gost.plus) infrastructure.<br>

It enables developers _ngrok_-like functionality exposing local HTTP resources through encrypted tunnels with Basic Auth and minimal configuration.

## Supported tunnels
* File Tunnel - exposes local files to the public network
* HTTP Tunnel - exposes local HTTP service to the public network

## Key Features:
- 🔒 Secure Tunnels: TLS-encrypted connections with Basic Auth using GOST.PLUS relay nodes
- 💻 Cross-Platform Support: Runs seamlessly on Linux, macOS, and Windows
- 🚀 Zero Configuration Mode: Start tunneling with a single command
- 🔧 Configurable & Scriptable: YAML config support and CLI flags
- 💊 Tunnel monitoring & Channel Self-recovering
- 🛠️ Integrated Logging & Stats: Real-time connection metrics and JSON logging for diagnostics
- ⚙️ Systemd service integration for Linux.

## Use Cases:
- IoT device connectivity on edge devices such as Raspberry Pi
- Ideal for rapid prototyping
- Webhook testing and remote access
- Private API exposure for collaborators

## Screenshot

### Terminal

Running with stats

<img src="assets/tunnel-run-stats.png" width="512" />

Logging output

<img src="assets/tunnel-logging.png" width="512" />


### Build Console App:
```bash
go build -ldflags="-s -w" -o gost-tunnel main.go
# or
make
```

### Get started
1. Generate a secure 12-character password:
```bash
openssl rand -base64 12
```

2. Create and start a tunnel (feel free to create multiple ones):
```bash
# for the first time
gost-tunnel --local 127.0.0.1:8080 \
      --tunnel_type http \
      --name rpi-service \
      --username user \
      --password my_password \
      --stats-interval 1s

# next time just run
gost-tunnel
```

**Note:** It's tested on `Raspberry Pi Zero 2W` and macOS hosts. Use it on your own risk.

3. Watch logs with `jq`:
```bash
# Linux
tail -f /home/[user]/.config/gost.plus/logs/gost-plus.log | jq -C .
```

4. Run it as a systemd service in the background on Linux (optional):
```bash
chmod +x ./install.sh
./install.sh
```

### Commands:

```bash
# Show a help for the application
gost-tunnel --help

# List all configured tunnels
gost-tunnel --list

# Create a new tunnel quickly
gost-tunnel --local localhost:8080 --name "my-tunnel"

# Create a new tunnel. Full Example
gost-tunnel --local localhost:8080 --tunnel_type http --name mytunnel --username myuser --password mypass --hostname example.com --tls --stats-interval 2s

# Delete a tunnel by ID
gost-tunnel --delete "12345678-1234-1234-1234-123456789abc"

# Start all configured tunnels using config
gost-tunnel

# Start all configured tunnels without statistics updates, e.g. for daemon mode
gost-tunnel --no-stats

# Show version
gost-tunnel --version
```

### App related folders and configuration

**MacOS**
- app dir: /Users/[user]/Library/Application\ Support/gost.plus
- app configuration file: /Users/[user]/Library/Application\ Support/gost.plus/config.yml
- app log file: /Users/[user]/Library/Application\ Support/gost.plus/logs/gost-plus.log

**Linux**
- app dir: /home/[user]/.config/gost.plus
- app configuration file: /home/[user]/.config/gost.plus/config.yml
- app log file: /home/[user]/.config/gost.plus/logs/gost-plus.log

### Log monitoring
- Keep track pretty formatted logs:
```bash
tail -f /Users/[user]/Library/Application\ Support/gost.plus/logs/gost-plus.log | jq -C .
```

- Keep track logs without formatting
```bash
tail -f /Users/[user]/Library/Application\ Support/gost.plus/logs/gost-plus.log | jq -c .
```

## Unit Testing

### Running Tests

```bash
# To run all Unit tests:
xgo test -v ./tests/

# To run specific Unit tests on a file:
xgo test -v ./tests/stats_test.go
```

Start tests explorer on web browser:
```bash
xgo e
```

### Test Coverage

Check test coverage with the following commands:

```bash
# Generate coverage profile
xgo test -cover -coverpkg=./... -coverprofile=coverage.out ./...
# or
xgo test -cover -coverpkg=./runner/task/... -coverprofile=coverage.out ./...

# View coverage on Web browser
go tool cover -html=coverage.out
# or view the coverage on terminal
go tool cover -func=coverage.out

# Save coverage report to HTML file
go tool cover -html=coverage.out -o coverage.html
```
