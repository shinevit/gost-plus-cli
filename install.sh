#!/bin/bash

# Detect actual user
REAL_USER="${SUDO_USER:-$(whoami)}"

# Check if running as root, otherwise use sudo for privileged commands
if [ "$EUID" -ne 0 ]; then
    SUDO="sudo"
else
    SUDO=""
fi

# Set the desired GitHub repository
repo="shinevit/gost-plus-cli"
base_url="https://api.github.com/repos/$repo/releases"

# Function to download and install gost
install_gost_tunnel() {
    version=$1
    # Detect the operating system
    if [[ "$(uname)" == "Linux" ]]; then
        os="linux"
    elif [[ "$(uname)" == "Darwin" ]]; then
        os="darwin"
    elif [[ "$(uname)" == "MINGW"* ]]; then
        os="windows"
    else
        echo "Unsupported operating system."
        exit 1
    fi

    # Detect the current CPU architecture
    arch=$(uname -m)
    case $arch in
    armv5*)
        cpu_arch="armv5"
        ;;
    armv6*)
        cpu_arch="armv6"
        ;;
    armv7*)
        cpu_arch="armv7"
        ;;
    aarch64 | arm64)
        cpu_arch="arm64"
        ;;
    i686)
        cpu_arch="386"
        ;;
    x86_64)
        cpu_arch="amd64"
        ;;
    *)
        echo "Unsupported CPU architecture."
        exit 1
        ;;
    esac
    get_download_url="$base_url/tags/$version"
    download_url=$(curl -s "$get_download_url" | grep "browser_download_url" | grep "${os}" | grep "${cpu_arch}" | head -n 1 | sed -E 's/.*"browser_download_url": "([^"]+)".*/\1/')

    # Download the binary
    echo "Downloading 'gost-tunnel' version $version..."
    
    # Use different archive format based on OS
    if [[ "$os" == "windows" ]]; then
        # Windows uses .zip archives
        curl -fsSL -o gost-tunnel.zip $download_url
        echo "Installing 'gost-tunnel'..."
        unzip -o gost-tunnel.zip
        
        INSTALL_DIR="/usr/bin"
        mkdir -p "$INSTALL_DIR"
        mv gost-tunnel.exe "$INSTALL_DIR/gost-tunnel.exe"
    else
        # Linux and macOS use .tar.gz archives
        curl -fsSL -o gost-tunnel.tar.gz $download_url
        echo "Installing 'gost-tunnel'..."
        tar -xzf gost-tunnel.tar.gz
        
        chmod +x gost-tunnel
        $SUDO mv gost-tunnel /usr/local/bin/gost-tunnel
    fi
    
    if [ -d /run/systemd/system ]; then
        # "System is running systemd"
        echo ""
        read -p "Would you like to install 'gost-tunnel' as Systemd service? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if [ "$REAL_USER" = "root" ]; then
                echo "Running as root user. It is recommended to run this script with sudo from a non-root user."
                while true; do
                    read -p "Enter the username to run the service as: " input_user
                    if id "$input_user" >/dev/null 2>&1; then
                        REAL_USER="$input_user"
                        break
                    else
                        echo "User '$input_user' does not exist. Please try again."
                    fi
                done
            fi

            # Get user details
            USER_HOME=$(eval echo "~$REAL_USER")
            USER_GROUP=$(id -gn "$REAL_USER")
            
            echo "Installing service for user: $REAL_USER (Group: $USER_GROUP, Home: $USER_HOME)"

            # Stop existing service if running
            $SUDO systemctl stop gost-plus 2>/dev/null || true

            # Install Config
            CONFIG_DIR="$USER_HOME/.config/gost.plus"
            CONFIG_FILE="$CONFIG_DIR/config.yml"
            SOURCE_CONFIG="./config.yml" 
            
            if [ ! -d "$CONFIG_DIR" ]; then
                echo "Creating config directory: $CONFIG_DIR"
                mkdir -p "$CONFIG_DIR"
                # Only chown if running as root
                if [ "$EUID" -eq 0 ]; then
                    chown "$REAL_USER:$USER_GROUP" "$CONFIG_DIR"
                fi
            fi

            if [ ! -f "$CONFIG_FILE" ]; then
                if [ -f "$SOURCE_CONFIG" ]; then
                    echo "Copying default config to $CONFIG_FILE"
                    cp "$SOURCE_CONFIG" "$CONFIG_FILE"
                    if [ "$EUID" -eq 0 ]; then
                        chown "$REAL_USER:$USER_GROUP" "$CONFIG_FILE"
                    fi
                    chmod 600 "$CONFIG_FILE"
                else
                    echo "Warning: Source config '$SOURCE_CONFIG' not found. Skipping config copy."
                fi
            else
                echo "Config file already exists at $CONFIG_FILE. Skipping."
            fi

            # Install Service
            SERVICE_TEMPLATE="./gost-plus.service"
            TARGET_SERVICE="/etc/systemd/system/gost-plus.service"

            if [ -f "$SERVICE_TEMPLATE" ]; then
                echo "Generating unit file for the service..."
                # We need to write to a temp file first because we might not have permission to write to /etc directly
                TEMP_SERVICE=$(mktemp)
                sed -e "s|{{USER}}|$REAL_USER|g" \
                    -e "s|{{HOME}}|$USER_HOME|g" \
                    "$SERVICE_TEMPLATE" > "$TEMP_SERVICE"
                
                $SUDO mv "$TEMP_SERVICE" "$TARGET_SERVICE"
                $SUDO chmod 644 "$TARGET_SERVICE"
                
                echo "Reloading systemd..."
                $SUDO systemctl daemon-reload
                $SUDO systemctl enable gost-plus
                $SUDO systemctl start gost-plus
                
                echo "Service installed and started!"
                echo -e "\nService status:"
                $SUDO systemctl status gost-plus --no-pager
                echo -e "\nTo view the service journal: \033[1;33mjournalctl -u gost-plus -f\033[0m"
                echo -e "To view application logs: \033[1;33mtail -f \"$USER_HOME/.config/gost.plus/logs/gost-plus.log\" | jq -C .\033[0m"
            else
                echo "Error: Service template '$SERVICE_TEMPLATE' not found."
            fi
        else
             echo "'gost-tunnel' installation completed (binary only)!"
        fi
    else 
        echo "'gost-tunnel' installation completed!"
    fi

    # Cleanup downloaded archive
    rm -f gost-tunnel.tar.gz gost-tunnel.zip
}

# Retrieve available versions from GitHub API
versions=$(curl -s "$base_url" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')

# Check if --latest option provided
if [[ "$1" == "--latest" ]]; then
    # Install the latest version automatically
    latest_version=$(echo "$versions" | head -n 1)
    install_gost_tunnel $latest_version
else
    # Display available versions to the user
    echo "Available gost-tunnel versions:"
    i=1
    for version in $versions; do
        echo "$i) $version"
        i=$((i + 1))
    done
    
    echo ""
    read -p "Select a version (enter number): " choice
    
    # Convert versions to array
    versions_array=($versions)
    
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#versions_array[@]}" ]; then
        selected_version="${versions_array[$((choice - 1))]}"
        install_gost_tunnel "$selected_version"
    else
        echo "Invalid choice! Please select a valid version number."
        exit 1
    fi
fi
