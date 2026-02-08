#!/bin/bash
# Pre-removal script for sysc-greet

set -e

echo "==> Disabling greetd service..."
systemctl disable greetd.service 2>/dev/null || true

echo "==> sysc-greet has been removed"
