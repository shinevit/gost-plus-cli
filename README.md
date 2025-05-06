# GOST-PLUS App in UI and console mode

A cross-platform GUI client for [GOST.PLUS](https://gost.plus) built with [gioui](https://gioui.org).

## Features

### File Tunnel

Expose local files to the public network.

<img src="assets/file-tunnel.gif">

### HTTP Tunnel

Expose local HTTP service to the public network.

<img src="assets/http-tunnel.gif">

### TCP Tunnel

Expose local TCP service to the public network.

<img src="assets/tcp-tunnel.gif">

### UDP Tunnel

Expose local UDP service to the public network.

<img src="assets/udp-tunnel.gif">

## Screenshot

### Desktop

<img src="assets/list.png" width="512" />
<img src="assets/menu.png" width="512" />
<img src="assets/add.png" width="512" />
<img src="assets/edit.png" width="512" />

### Mobile

<img src="assets/list-android.png" width="512" />
<img src="assets/add-android.png" width="512" />
<img src="assets/edit-android.png" width="512" />


### Build Console App:
```bash
go build -ldflags="-s -w" -o gost-tunnel main.go
# or
make
```

### Usage examples on CLI:

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
- app dir:
/Users/vitalii/Library/Application\ Support/gost.plus

- app configuration file:
/Users/vitalii/Library/Application\ Support/gost.plus/config.yml

- app log file:
/Users/vitalii/Library/Application\ Support/gost.plus/logs/gost-plus.log


### Log monitoring
- Keep track logs with pretty formatting:
```bash
tail -f /Users/vitalii/Library/Application\ Support/gost.plus/logs/gost-plus.log | jq -C .
```

- Keep track logs without formatting
```bash
tail -f /Users/vitalii/Library/Application\ Support/gost.plus/logs/gost-plus.log | jq -c .
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
