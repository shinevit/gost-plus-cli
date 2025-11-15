#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as sudo
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as a priviledged user with sudo"
    exit 1
fi

# Detect actual user
REAL_USER="${SUDO_USER:-$(whoami)}"
if [ "$REAL_USER" = "root" ]; then
    log_warn "Running as root user. It is recommended to run this script with sudo from a non-root user."
    while true; do
        read -p "Enter the username to run the service as: " input_user
        if id "$input_user" >/dev/null 2>&1; then
            REAL_USER="$input_user"
            break
        else
            log_error "User '$input_user' does not exist. Please try again."
        fi
    done
fi

# Get user details
USER_HOME=$(eval echo "~$REAL_USER")
USER_GROUP=$(id -gn "$REAL_USER")

log_info "Installing for user: $REAL_USER (Group: $USER_GROUP, Home: $USER_HOME)"

# Stop existing service if running
if systemctl is-active --quiet gost-plus; then
    log_info "Stopping existing service..."
    systemctl stop gost-plus
fi

# 1. Install Binary
BINARY_PATH="./gost-tunnel"
if [ ! -f "$BINARY_PATH" ]; then
    log_error "Binary 'gost-tunnel' not found in current directory."
    exit 1
fi

log_info "Installing binary from $BINARY_PATH..."
cp "$BINARY_PATH" /usr/local/bin/gost-tunnel
chmod +x /usr/local/bin/gost-tunnel

# 2. Install Config
SOURCE_CONFIG="./config.yml"
CONFIG_DIR="$USER_HOME/.config/gost.plus"
CONFIG_FILE="$CONFIG_DIR/config.yml"

if [ ! -d "$CONFIG_DIR" ]; then
    log_info "Creating config directory: $CONFIG_DIR"
    mkdir -p "$CONFIG_DIR"
    chown "$REAL_USER:$USER_GROUP" "$CONFIG_DIR"
fi

if [ ! -f "$CONFIG_FILE" ]; then
    if [ -f "$SOURCE_CONFIG" ]; then
        log_info "Copying default config to $CONFIG_FILE"
        cp "$SOURCE_CONFIG" "$CONFIG_FILE"
        chown "$REAL_USER:$USER_GROUP" "$CONFIG_FILE"
    else
        log_warn "Source config '$SOURCE_CONFIG' not found. Skipping config copy."
    fi
else
    log_info "Config file already exists at $CONFIG_FILE. Skipping."
fi

# 3. Install Service
SERVICE_TEMPLATE="./gost-plus.service"
TARGET_SERVICE="/etc/systemd/system/gost-plus.service"

if [ ! -f "$SERVICE_TEMPLATE" ]; then
    log_error "Service template '$SERVICE_TEMPLATE' not found."
    exit 1
fi

log_info "Generating unit file for the service..."
# Read template and substitute variables
sed -e "s|{{USER}}|$REAL_USER|g" \
    -e "s|{{HOME}}|$USER_HOME|g" \
    "$SERVICE_TEMPLATE" > "$TARGET_SERVICE"

log_info "Reloading systemd..."
systemctl daemon-reload

log_info "Enabling and starting service..."
systemctl enable gost-plus
systemctl start gost-plus

# 4. Status
echo ""
systemctl status gost-plus --no-pager

echo ""
log_info "Installation complete!"
echo ""
echo -e "To view the service journal: ${YELLOW}journalctl -u gost-plus -f${NC}"
echo -e "To view application logs: ${YELLOW}tail -f \"${CONFIG_DIR}/logs/gost-plus.log\" | jq -C .${NC}"
