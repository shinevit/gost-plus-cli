#!/bin/bash

# Stop the service if it's running
sudo systemctl stop gost-plus 2>/dev/null || true

# Copy the binary
sudo cp -v ./gost-tunnel /usr/local/bin/
sudo chmod +x /usr/local/bin/gost-tunnel

# Copy the service file
sudo cp -v ./gost-plus.service /etc/systemd/system/gost-plus.service

# Reload systemd
sudo systemctl daemon-reload

# Enable and start the service
sudo systemctl enable gost-plus
sudo systemctl start gost-plus

# Show the service status
echo -e "\nService status:"
sudo systemctl status gost-plus --no-pager

echo -e "\nTo view logs: journalctl -u gost-plus -f"
