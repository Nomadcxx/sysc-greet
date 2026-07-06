#!/bin/bash
# Post-installation script for sysc-greet

set -e

echo "==> Setting up sysc-greet..."

# Create greeter user if it doesn't exist
if ! id greeter &>/dev/null; then
    echo "==> Creating greeter user..."
    useradd -M -G video,render,input -s /usr/bin/nologin greeter
else
    echo "==> Updating greeter user groups..."
    usermod -aG video,render,input greeter
fi

# Set permissions
echo "==> Setting permissions..."
chown -R greeter:greeter /var/cache/sysc-greet 2>/dev/null || true
chown -R greeter:greeter /var/lib/greeter 2>/dev/null || true
chmod 755 /var/lib/greeter

# Detect installed Wayland backend and configure greetd
echo "==> Detecting greeter backend..."

COMPOSITOR=""
if command -v cagebreak &>/dev/null; then
    COMPOSITOR="cagebreak"
    GREETD_COMMAND="cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
elif command -v niri &>/dev/null; then
    COMPOSITOR="niri"
    GREETD_COMMAND="niri -c /etc/greetd/niri-greeter-config.kdl"
elif command -v sway &>/dev/null; then
    COMPOSITOR="sway"
    GREETD_COMMAND="sway -c /etc/greetd/sway-greeter-config"
elif command -v Hyprland &>/dev/null || command -v hyprland &>/dev/null; then
    COMPOSITOR="hyprland"
    # Use legacy hyprland command (not start-hyprland) for compatibility
    GREETD_COMMAND="Hyprland -c /etc/greetd/hyprland-greeter-config.conf"
fi

if [ -z "$COMPOSITOR" ]; then
    echo "WARNING: No supported greeter backend detected (cagebreak, niri, sway, hyprland)"
    echo "Please install cagebreak (recommended) and manually configure /etc/greetd/config.toml"
else
    echo "Detected backend: $COMPOSITOR"

    if [ "$COMPOSITOR" = "hyprland" ]; then
        echo ""
        echo "WARNING: Hyprland greeter support is deprecated and will be removed in ~3 months."
        echo "         Migrate to cagebreak: SYSC_COMPOSITOR=cagebreak sudo ./install.sh"
        echo "         See https://nomadcxx.github.io/sysc-greet/compositors/cagebreak/"
        echo ""
    fi

    # Only create greetd config if it doesn't exist or is empty
    if [ ! -s /etc/greetd/config.toml ]; then
        echo "==> Configuring greetd for $COMPOSITOR..."
        cat > /etc/greetd/config.toml <<EOF2
[terminal]
vt = 1

[default_session]
command = "$GREETD_COMMAND"
user = "greeter"

[initial_session]
command = "$GREETD_COMMAND"
user = "greeter"
EOF2
        echo "Created /etc/greetd/config.toml"
    else
        echo "Existing /etc/greetd/config.toml found, not modifying"
        echo "If you want to use sysc-greet, update the command to: $GREETD_COMMAND"
        # Remove source_profile=false if present (was added in v1.1.3, causes issues)
        if grep -q 'source_profile' /etc/greetd/config.toml; then
            echo "==> Removing source_profile setting from greetd config..."
            sed -i '/^\[general\]/d; /^source_profile/d' /etc/greetd/config.toml
            # Clean up empty lines left behind
            sed -i '/^$/N;/^\n$/d' /etc/greetd/config.toml
        fi
    fi
fi

# Enable greetd service
echo "==> Enabling greetd service..."
systemctl enable greetd.service 2>/dev/null || true

echo ""
echo "==> sysc-greet installed successfully!"
echo ""
echo "Reboot to see sysc-greet"
echo ""
