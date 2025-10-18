#!/bin/bash
# sysc-greet one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash

set -e

echo "sysc-greet installer"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash"
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install Go first: https://go.dev/doc/install"
    exit 1
fi

# Ask user for compositor choice
echo "Select compositor:"
echo "1) niri"
echo "2) hyprland"
echo "3) sway"
read -p "Choice [1-3]: " choice < /dev/tty

case $choice in
    1) COMPOSITOR="niri" ;;
    2) COMPOSITOR="hyprland" ;;
    3) COMPOSITOR="sway" ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

echo "Selected: $COMPOSITOR"
echo ""

# Check for compositor
if ! command -v $COMPOSITOR &> /dev/null; then
    echo "Error: $COMPOSITOR is not installed"
    echo "Please install $COMPOSITOR first"
    exit 1
fi

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Cloning sysc-greet..."
git clone https://github.com/Nomadcxx/sysc-greet.git
cd sysc-greet

echo "Building installer..."
go build -o install-sysc-greet ./cmd/installer/

echo "Running installer with compositor: $COMPOSITOR"
# Pass compositor to installer via environment variable
SYSC_COMPOSITOR=$COMPOSITOR ./install-sysc-greet

# Cleanup
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "Installation complete."
echo "Reboot to see sysc-greet."
