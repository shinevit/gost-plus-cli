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


### Build the console application:
```bash
go build -o gost-tunnel main.go
```

### Usage examples on CLI:

```bash
# Show a help for the application
./gost-tunnel --help

# List all configured tunnels
./gost-tunnel --list

# Create a new tunnel quickly
./gost-tunnel --local localhost:8080 --name "my-tunnel"

# Create a new tunnel. Full Example
./gost-tunnel --local localhost:8080 --tunnel_type http --name mytunnel --username myuser --password mypass --hostname example.com --tls --stats-interval 5s

# Delete a tunnel by ID
./gost-tunnel --delete "12345678-1234-1234-1234-123456789abc"

# Start all configured tunnels from config
./gost-tunnel

# Show version
./gost-tunnel --version
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
