![sysc-greet](assets/logo.png)

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go with the Bubble Tea framework.

![Preview](assets/showcase.gif)

## Quick Links

- [Full Documentation](https://nomadcxx.github.io/sysc-greet/) - Complete guides, configuration, and usage instructions

## Installation

### Quick Install

One-line installer that works on most Linux distributions:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The installer automatically detects your package manager and works on Arch Linux, Debian/Ubuntu, Fedora, and openSUSE. It'll handle compositor selection, install dependencies, and set everything up for you.

### Build from Source

If you want to build it yourself, just clone the repo and run the installer:

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go run ./cmd/installer/
```

The installer walks you through compositor selection and configuration.

### Arch Linux (AUR)

Three AUR packages available depending on which compositor you're using:

```bash
# Recommended (niri)
yay -S sysc-greet

# Hyprland variant
yay -S sysc-greet-hyprland

# Sway variant
yay -S sysc-greet-sway
```

### NixOS (Flake)

If you're on NixOS, add sysc-greet to your flake:

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

For detailed docs, configuration guides, troubleshooting, and usage instructions, check out the [full documentation site](https://nomadcxx.github.io/sysc-greet/).

## License

GPL-3.0
