#!/bin/bash

set -e

SERVICE_NAME="go-serial-tty-client"
BINARY_NAME="go-serial-tty-client"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/default"
SYSTEMD_DIR="/etc/systemd/system"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root"
  exit 1
fi

echo "Installing $SERVICE_NAME..."

# Build the project (ensure we have the latest binary)
# Assuming make is available, otherwise we expect the binary to exist
if [ -f "Makefile" ]; then
    echo "Building project..."
    make build
fi

if [ ! -f "$BINARY_NAME" ]; then
    echo "Error: Binary $BINARY_NAME not found. Please build it first."
    exit 1
fi

# Install binary
echo "Copying binary to $INSTALL_DIR..."
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Install config
if [ ! -f "$CONFIG_DIR/$SERVICE_NAME" ]; then
    echo "Copying default config to $CONFIG_DIR/$SERVICE_NAME..."
    cp "$SERVICE_NAME.default" "$CONFIG_DIR/$SERVICE_NAME"
else
    echo "Config file already exists at $CONFIG_DIR/$SERVICE_NAME, skipping overwrite."
fi

# Install systemd service
echo "Copying service file to $SYSTEMD_DIR..."
cp "$SERVICE_NAME.service" "$SYSTEMD_DIR/"

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable and start service
echo "Enabling and starting service..."
systemctl enable "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

echo "Installation complete!"
echo "Check status with: systemctl status $SERVICE_NAME"
echo "Edit config at: $CONFIG_DIR/$SERVICE_NAME"
