#!/bin/bash
# SYSC-Greet Installation Script

set -e

echo "Building sysc-greet installer..."
cd cmd/installer
go build -o ../../install-sysc-greet
cd ../..

echo "Running installer..."
sudo ./install-sysc-greet

echo ""
echo "Installation complete!"
echo "Configure greetd to use sysc-greet:"
echo "  Edit /etc/greetd/config.toml"
echo "  Set: command = \"kitty --class=greeter -e sysc-greet\""
