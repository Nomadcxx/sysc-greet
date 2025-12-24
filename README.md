# sysc-greet

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go with the Bubble Tea framework.

![Preview](https://github.com/Nomadcxx/sysc-greet/raw/master/assets/showcase.gif)

## Quick Links

- [Full Documentation](https://nomadcxx.github.io/sysc-greet/) - Complete guides, configuration, and usage instructions

## Installation

### Quick Install

One-line installer that works on most Linux distributions:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The installer automatically detects your package manager and works on:
- **Arch Linux** (pacman)
- **Debian/Ubuntu** (apt)
- **Fedora** (dnf/yum)
- **openSUSE** (zypper)
- **Alpine Linux** (apk)

It handles compositor selection, dependency installation, and configuration automatically.

### Build from Source

Build and install using the interactive installer:

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go run ./cmd/installer/
```

The installer guides you through compositor selection and configuration.

### Arch Linux (AUR)

Three AUR packages available for different compositors:

```bash
# Recommended (niri)
yay -S sysc-greet

# Hyprland variant
yay -S sysc-greet-hyprland

# Sway variant
yay -S sysc-greet-sway
```

### NixOS (Flake)

For NixOS users, sysc-greet is available as a flake. Add to your configuration:

**flake.nix:**
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

**configuration.nix:**
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

Then rebuild:
```bash
sudo nixos-rebuild switch --flake .#your-hostname
```

## Documentation

For detailed documentation, configuration guides, troubleshooting, and usage instructions, see the [full documentation site](https://nomadcxx.github.io/sysc-greet/).

## License

MIT
