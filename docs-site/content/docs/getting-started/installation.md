---
title: "Installation"
description: "This guide covers installation methods for sysc-greet across different Linux distributions."
---

This guide covers installation methods for sysc-greet across different Linux distributions.

## Quick Install Script

One-line install for most systems:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The interactive installer will prompt you to:
1. Choose your greeter backend: Niri (default), Cagebreak, Sway, or Mango. An installed experimental sysc-compositor is also listed
2. Configure backend settings
3. Install dependencies automatically

> Set `SYSC_COMPOSITOR=hyprland` when running the installer to opt into Hyprland 0.55+. The installer then shows its Lua based backend option. Check startup, login, and quit on the target machine before enabling greetd.

### Void Linux

Void supports native XBPS release packages for x86_64 glibc and musl, plus the Go installer. Both use runit, elogind, D-Bus and polkit. See the [Void Linux guide](../getting-started/void-linux) for package downloads, greetd setup and existing seatd/turnstile configurations.

```bash
sudo SYSC_COMPOSITOR=cagebreak go run ./cmd/installer/
```

## Manual Build

### Prerequisites

- Go 1.26+
- greetd
- Wayland backend: Niri (default), Cagebreak, Sway, Mango, or Hyprland 0.55+ (sysc-compositor: experimental)
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
sudo go run ./cmd/installer/
```

## Arch Linux (AUR)

sysc-greet provides five AUR packages, one per greeter backend:

```bash
# niri (default)
yay -S sysc-greet

# Cagebreak variant
yay -S sysc-greet-cagebreak

# Sway variant
yay -S sysc-greet-sway

# Mango variant
yay -S sysc-greet-mango

# Hyprland 0.55+ Lua variant
yay -S sysc-greet-hyprland
```

The packages conflict with each other, so install exactly one. Each pulls in its compositor as a dependency. The Hyprland package requires version 0.55 or newer; the Cagebreak package also installs socat so it can quit Cagebreak after login.

## Pre-built Packages

Download pre-built packages from [GitHub Releases](https://github.com/Nomadcxx/sysc-greet/releases):

### Debian/Ubuntu (.deb)

```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.0/sysc-greet_1.3.0_amd64.deb
sudo apt install ./sysc-greet_1.3.0_amd64.deb
```

The package will:
1. Install sysc-greet to `/usr/local/bin/`
2. Install configs to `/usr/share/sysc-greet/`
3. Detect your compositor and configure greetd
4. Enable the greetd service

**Note:** Package configs use conservative syntax compatible with stable distribution versions. For bleeding-edge compositor features, use the install script or AUR instead.

### Fedora (.rpm)

```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.0/sysc-greet-1.3.0-1.x86_64.rpm
sudo dnf install ./sysc-greet-1.3.0-1.x86_64.rpm
```

After installation, **reboot** your system to see sysc-greet.

The releases page also carries cagebreak `.deb`/`.rpm` packages (Debian 13, Ubuntu 24.04/25.x, Fedora 42) since no distro packages exist — see [Cagebreak Setup](../compositors/cagebreak).

### Void Linux (.xbps)

Download matching sysc-greet and gSlapper packages for x86_64 glibc or musl and follow the [native XBPS instructions](../getting-started/void-linux#native-xbps-packages). The packages use `/usr/bin/sysc-greet` and `/etc/greetd/sysc-greet/`; configure greetd and enable its runit service after installation.

### Switching Compositors

The Debian and Fedora packages auto-detect your compositor during installation. If you have multiple compositors installed or want to switch to a different one, edit `/etc/greetd/config.toml`:

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

**sysc-compositor (experimental, WIP):**
```toml
[default_session]
command = "sysc-compositor -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```
sysc-compositor is unreleased and not packaged by sysc-greet. Install a development build separately and make it available on greetd's `PATH`. The installer lists it only when the binary is found, and reuses the Niri greeter config and its Niri IPC quit command.
Set `SYSC_COMPOSITOR=sysc-compositor` to list and preselect it in the installer.

**Cagebreak:**
```toml
[default_session]
command = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
user = "greeter"
```

`-e` enables cagebreak's IPC socket, which the greeter config uses to quit the compositor after login — socat must be installed. See [Cagebreak Setup](../compositors/cagebreak).

**Sway:**
```toml
[default_session]
command = "sway -c /etc/greetd/sway-greeter-config"
user = "greeter"
```

**Mango:**
```toml
[default_session]
command = "mango -c /etc/greetd/mango-greeter-config.conf -s /etc/greetd/mango-greeter-session.sh"
user = "greeter"
```

The session script starts the greeter and quits mango with `mmsg` after login. See [Mango Setup](../compositors/mango).

**Hyprland 0.55+:**
```toml
[default_session]
command = "start-hyprland -- --config /etc/greetd/hyprland-greeter-config.lua"
user = "greeter"
```

The Lua config installs at `/etc/greetd/hyprland-greeter-config.lua`. The AUR package sets this command when you choose the Hyprland variant. For the source installer, set `SYSC_COMPOSITOR=hyprland` before starting it.

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
{ lib, ... }:
{
  services.sysc-greet = {
    enable = true;
    compositor = "niri";  # or "cagebreak", "sway", "mango", "hyprland" (0.55+)
  };

  # Required for the CC BY-NC model assets included through sysc-Go.
  nixpkgs.config.allowUnfreePredicate = pkg: builtins.elem (lib.getName pkg) [ "sysc-greet" ];

  # Optional: Set initial session for auto-login
  services.sysc-greet.settings.initial_session = {
    command = "Hyprland";
    user = "your-username";
  };
}
```

sysc-greet's own source code remains GPL-3.0-only. This package-specific exception covers the current binary's sysc-Go model assets; the module uses your system's Nixpkgs license policy.

By default, the NixOS module does not install `niri`, `cagebreak`, `hyprland`, `sway`, or `mango`.
Install your chosen backend yourself, or set `niriPackage`, `cagebreakPackage`, `hyprlandPackage`,
`swayPackage`, or `mangoPackage` if you want the module to install and use a specific package.
Mango is not in nixpkgs; `mangoPackage` must come from your own flake.
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
# command = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
# command = "sway --unsupported-gpu -c /etc/greetd/sway-greeter-config"
# command = "mango -c /etc/greetd/mango-greeter-config.conf -s /etc/greetd/mango-greeter-session.sh"
# command = "start-hyprland -- --config /etc/greetd/hyprland-greeter-config.lua"
user = "greeter"
```

### Install Compositor Config

Copy the appropriate config to `/etc/greetd/`:

```bash
# niri
sudo cp config/niri-greeter-config.kdl /etc/greetd/

# cagebreak
sudo cp config/cagebreak-greeter-config /etc/greetd/

# sway
sudo cp config/sway-greeter-config /etc/greetd/

# mango
sudo cp config/mango-greeter-config.conf /etc/greetd/
sudo install -m755 config/mango-greeter-session.sh /etc/greetd/

# Hyprland 0.55+
sudo cp config/hyprland-greeter-config.lua /etc/greetd/
```

### Create Greeter User

```bash
sudo useradd -M -d /var/lib/greeter -G video,render,input -s /usr/bin/nologin greeter
sudo mkdir -p /var/cache/sysc-greet /var/lib/greeter/Pictures/wallpapers \
    /var/lib/greeter/.cache /var/lib/greeter/.config /var/lib/greeter/.local/state
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

**Void Linux:**
```bash
sudo xbps-remove sysc-greet
```

### Arch Linux (AUR)

```bash
yay -R sysc-greet
# replace with sysc-greet-cagebreak, sysc-greet-sway, or sysc-greet-hyprland if applicable
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

For more help, see [Troubleshooting Guide](../getting-started/troubleshooting).
