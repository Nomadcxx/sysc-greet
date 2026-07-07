# Installation

This guide covers installation methods for sysc-greet across different Linux distributions.

## Quick Install Script

One-line install for most systems:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The interactive installer will prompt you to:
1. Choose your greeter backend (niri default, cagebreak, sway, or hyprland deprecated)
2. Configure backend settings
3. Install dependencies automatically

> Hyprland greeter support ends in ~3 months. Migrate to [cagebreak](../compositors/cagebreak.md) or [niri](../compositors/niri.md).

## Manual Build

### Prerequisites

- Go 1.25+
- greetd
- Wayland backend: niri (default), cagebreak, sway, or hyprland (deprecated)
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

sysc-greet provides AUR packages for different greeter backends:

```bash
# niri (default)
yay -S sysc-greet

# Hyprland variant (deprecated, replaced by cagebreak)
yay -S sysc-greet-hyprland

# Sway variant
yay -S sysc-greet-sway

# Cagebreak: install from AUR, then run the installer
paru -S cagebreak && sudo pacman -S socat
SYSC_COMPOSITOR=cagebreak curl -fsSL .../install.sh | sudo bash
```

## Pre-built Packages

Download pre-built packages from [GitHub Releases](https://github.com/Nomadcxx/sysc-greet/releases):

### Debian/Ubuntu (.deb)

```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.1.8/sysc-greet_1.1.8_amd64.deb
sudo apt install ./sysc-greet_1.1.8_amd64.deb
```

The package will:
1. Install sysc-greet to `/usr/local/bin/`
2. Install configs to `/usr/share/sysc-greet/`
3. Detect your compositor and configure greetd
4. Enable the greetd service

**Note:** Package configs use conservative syntax compatible with stable distribution versions. For bleeding-edge compositor features, use the install script or AUR instead.

### Fedora (.rpm)

```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.1.8/sysc-greet-1.1.8-1.x86_64.rpm
sudo dnf install ./sysc-greet-1.1.8-1.x86_64.rpm
```

After installation, **reboot** your system to see sysc-greet.

### Switching Compositors

The package auto-detects your compositor during installation. If you have multiple compositors installed or want to switch to a different one, edit `/etc/greetd/config.toml`:

```bash
sudo nano /etc/greetd/config.toml
```

Change the `command` line to your preferred compositor:

**Niri:**
```toml
[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

**Hyprland:**
```toml
[default_session]
command = "Hyprland -c /etc/greetd/hyprland-greeter-config.conf"
user = "greeter"
```

**Sway:**
```toml
[default_session]
command = "sway -c /etc/greetd/sway-greeter-config"
user = "greeter"
```

All compositor configs are installed at `/etc/greetd/`. Save the file and reboot to apply the change.

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

By default, the NixOS module does not install `niri`, `hyprland`, or `sway`.
Install your chosen compositor yourself, or set `niriPackage`, `hyprlandPackage`,
or `swayPackage` if you want the module to install and use a specific package.
If your compositor is managed elsewhere, set `compositorCommand` to the exact
command greetd should run.

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

## Uninstall

### Quick Install Script / Manual Build

Re-run the installer and select **Uninstall sysc-greet**:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The uninstaller removes:
- `/usr/local/bin/sysc-greet` — binary
- `/usr/share/sysc-greet/` — ASCII configs, wallpapers, data files
- `/etc/greetd/kitty.conf` and compositor config files
- `/var/cache/sysc-greet/` — saved preferences (optional step)

The greeter user and `/var/lib/greeter/` are intentionally preserved.

### Pre-built Packages

**Debian/Ubuntu:**
```bash
sudo apt remove sysc-greet
```

**Fedora:**
```bash
sudo dnf remove sysc-greet
```

### Arch Linux (AUR)

```bash
yay -R sysc-greet
# replace with sysc-greet-hyprland or sysc-greet-sway if applicable
```

### NixOS

Remove the sysc-greet module from `configuration.nix` and rebuild:

```bash
sudo nixos-rebuild switch --flake .#your-hostname
```

### After Uninstalling

To stop greetd from failing on next boot:

```bash
sudo systemctl disable greetd.service
```

---

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
