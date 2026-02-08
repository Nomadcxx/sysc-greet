# Package Installation Guide

This guide covers installing sysc-greet from .deb and .rpm packages.

## Download Packages

Download the appropriate package from the [GitHub Releases](https://github.com/Nomadcxx/sysc-greet/releases) page:

- **Debian/Ubuntu**: `sysc-greet_vX.X.X_amd64.deb`
- **Fedora/openSUSE**: `sysc-greet-X.X.X-1.x86_64.rpm`

## Install

### Debian/Ubuntu

```bash
sudo apt update
sudo apt install ./sysc-greet_vX.X.X_amd64.deb
```

The package will:
1. Install sysc-greet to `/usr/local/bin/`
2. Install configs to `/usr/share/sysc-greet/`
3. Detect your compositor and configure greetd
4. Enable the greetd service

### Fedora

```bash
sudo dnf install ./sysc-greet-X.X.X-1.x86_64.rpm
```

### openSUSE

```bash
sudo zypper install ./sysc-greet-X.X.X-1.x86_64.rpm
```

## Post-Installation

After installation, **reboot** your system to see sysc-greet.

The installer automatically:
- Creates a `greeter` user with appropriate permissions
- Configures greetd to use your installed compositor
- Enables the greetd service

## Troubleshooting

### No compositor detected

If you see "No supported compositor detected", install Niri, Hyprland, or Sway:

```bash
# Ubuntu (Niri from PPA or build from source)
# Fedora
sudo dnf install niri
# or
sudo dnf install hyprland
# or
sudo dnf install sway
```

Then reinstall sysc-greet or manually edit `/etc/greetd/config.toml`.

### Existing greetd config

If you already have a greetd configuration, the installer won't overwrite it. You can manually update it:

```toml
[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

Replace the command with the appropriate one for your compositor:
- Niri: `niri -c /etc/greetd/niri-greeter-config.kdl`
- Hyprland: `Hyprland -c /etc/greetd/hyprland-greeter-config.conf`
- Sway: `sway -c /etc/greetd/sway-greeter-config`

## Uninstall

### Debian/Ubuntu

```bash
sudo apt remove sysc-greet
```

### Fedora

```bash
sudo dnf remove sysc-greet
```

### openSUSE

```bash
sudo zypper remove sysc-greet
```

## Notes

- Package configs use conservative syntax compatible with stable distribution versions
- For bleeding-edge compositor features, use the Go installer instead
- The greeter user is not removed on uninstall for safety
