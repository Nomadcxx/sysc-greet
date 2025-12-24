# sysc-greet

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go with the Bubble Tea framework.

![Preview](https://github.com/Nomadcxx/sysc-greet/raw/master/assets/showcase.gif)

## Installation

### Quick Install Script

One-line install for most systems:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

### Manual Build

The installer lets you choose your compositor and handles all configuration:

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go run ./cmd/installer/
```

### Arch Linux (AUR)

First, decide which compositor you want. sysc-greet will install the recommended default (niri), sysc-greet-hyperland installs the Hyprland variant, and sysc-greet-sway installs the Sway variant.

```bash
# Recommended (niri)
yay -S sysc-greet

# Hyprland variant
yay -S sysc-greet-hyprland

# Sway variant
yay -S sysc-greet-sway
```

For detailed installation instructions, configuration, and usage, see the [full documentation](https://nomadcxx.github.io/sysc-greet/).

### NixOS (Flake)

Add sysc-greet to your NixOS configuration using the flake:

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

Then rebuild your system:
```bash
sudo nixos-rebuild switch --flake .#your-hostname
```

## Documentation

For detailed documentation, configuration guides, and usage instructions, see the [full documentation](https://nomadcxx.github.io/sysc-greet/).

## License

MIT
