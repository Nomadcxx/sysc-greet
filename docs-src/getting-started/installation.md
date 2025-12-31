# Installation

This guide covers installation methods for sysc-greet across different Linux distributions.

## Quick Install Script

One-line install for most systems:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The interactive installer will prompt you to:
1. Choose your compositor (niri, hyprland, or sway)
2. Configure compositor settings
3. Install dependencies automatically

## Manual Build

### Prerequisites

- Go 1.25+
- greetd
- Wayland compositor (niri, hyprland, or sway)
- kitty (terminal emulator)
- gSlapper (wallpaper daemon)
- swww (legacy wallpaper daemon, optional fallback)

### Build Steps

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go build -o sysc-greet ./cmd/sysc-greet/
sudo install -Dm755 sysc-greet /usr/local/bin/sysc-greet
```

### Run Installer

The installer handles compositor configuration automatically:

```bash
go run ./cmd/installer/
```

## Arch Linux (AUR)

sysc-greet provides three AUR packages for different compositors:

```bash
# Recommended (niri)
yay -S sysc-greet

# Hyprland variant
yay -S sysc-greet-hyprland

# Sway variant
yay -S sysc-greet-sway
```

## NixOS (Flake)

### Add to flake.nix

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    sysc-greet = {
      url = "github:Nomadcxx/sysc-greet";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, sysc-greet, ... }: {
    nixosConfigurations.your-hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./configuration.nix
        sysc-greet.nixosModules.default
      ];
    };
  };
}
```

### Add to configuration.nix

```nix
{
  services.sysc-greet = {
    enable = true;
    compositor = "niri";  # or "hyprland" or "sway"
  };

  # Optional: Set initial session for auto-login
  services.sysc-greet.settings.initial_session = {
    command = "Hyprland";
    user = "your-username";
  };
}
```

### Rebuild System

```bash
sudo nixos-rebuild switch --flake .#your-hostname
```

## Post-Installation Setup

### Configure greetd

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
# Choose your compositor:
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
# command = "start-hyprland -- -c /etc/greetd/hyprland-greeter-config.conf"
# command = "sway --unsupported-gpu -c /etc/greetd/sway-greeter-config"
user = "greeter"
```

### Install Compositor Config

Copy the appropriate config to `/etc/greetd/`:

```bash
# niri
sudo cp config/niri-greeter-config.kdl /etc/greetd/

# hyprland
sudo cp config/hyprland-greeter-config.conf /etc/greetd/

# sway
sudo cp config/sway-greeter-config /etc/greetd/
```

### Create Greeter User

```bash
sudo useradd -M -G video -s /usr/bin/nologin greeter
sudo mkdir -p /var/cache/sysc-greet /var/lib/greeter/Pictures/wallpapers
sudo chown -R greeter:greeter /var/cache/sysc-greet /var/lib/greeter
sudo chmod 755 /var/lib/greeter
```

### Enable Service

```bash
sudo systemctl enable greetd.service
```

## Verification

After installation, test the greeter:

```bash
sysc-greet --test
```

For fullscreen testing:
```bash
kitty --start-as=fullscreen sysc-greet --test
```

## Troubleshooting

### IPC Client Error

If you see `FATAL: Failed to create IPC client`, check that:
1. You are not setting `GREETD_SOCK` environment variable manually
2. greetd is actually running and has created the socket
3. You are running sysc-greet through greetd (not directly in terminal)

### Compositor Not Starting

Check compositor logs:
```bash
journalctl -u greetd -n 50
```

### Greeter Not Appearing

Verify:
1. greetd service is enabled and running: `systemctl status greetd`
2. compositor config exists in `/etc/greetd/`
3. greeter user has proper permissions

For more help, see [Troubleshooting Guide](../getting-started/troubleshooting.md).
