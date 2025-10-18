#!/bin/bash
# sysc-greet installer
# Usage:
#   Download and run: curl -O https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh && sudo bash install.sh
#   Or with compositor: sudo COMPOSITOR=niri bash install.sh

set -e

echo "sysc-greet installer"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install Go first: https://go.dev/doc/install"
    exit 1
fi

# Auto-detect compositor if not specified
if [ -z "$COMPOSITOR" ]; then
    echo "Auto-detecting compositor..."
    if command -v niri &> /dev/null; then
        COMPOSITOR="niri"
        echo "Detected: niri"
    elif command -v hyprland &> /dev/null || command -v Hyprland &> /dev/null; then
        COMPOSITOR="hyprland"
        echo "Detected: hyprland"
    elif command -v sway &> /dev/null; then
        COMPOSITOR="sway"
        echo "Detected: sway"
    else
        echo "Error: No supported compositor found (niri, hyprland, or sway)"
        echo "Install one of them first, or specify: sudo COMPOSITOR=niri bash install.sh"
        exit 1
    fi
else
    echo "Using specified compositor: $COMPOSITOR"
fi

echo ""

# Validate compositor choice
case $COMPOSITOR in
    niri|hyprland|sway)
        if ! command -v $COMPOSITOR &> /dev/null && ! command -v ${COMPOSITOR^} &> /dev/null; then
            echo "Error: $COMPOSITOR is not installed"
            exit 1
        fi
        ;;
    *)
        echo "Error: Invalid compositor '$COMPOSITOR'"
        echo "Supported: niri, hyprland, sway"
        exit 1
        ;;
esac

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
