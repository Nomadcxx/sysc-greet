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

The installer automatically detects your package manager and works on Arch Linux, Debian/Ubuntu, Fedora, and openSUSE. It'll handle greeter backend selection, install dependencies, and set everything up for you.

> Hyprland greeter support ends in ~3 months. New installs should choose niri (default) or cagebreak, its replacement.

### Build from Source

If you want to build it yourself, just clone the repo and run the installer:

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go run ./cmd/installer/
```

The installer walks you through compositor selection and configuration.

### Arch Linux (AUR)

Three AUR packages available depending on which greeter backend you're using:

```bash
# niri (default)
yay -S sysc-greet

# Hyprland variant (deprecated, replaced by cagebreak)
yay -S sysc-greet-hyprland

# Sway variant
yay -S sysc-greet-sway
```

For cagebreak, run the installer with `SYSC_COMPOSITOR=cagebreak` — it installs cagebreak and socat itself (AUR on Arch, prebuilt release packages on Debian/Ubuntu/Fedora, source build elsewhere).

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
    compositor = "niri";  # or "cagebreak", "sway", "hyprland" (deprecated)
  };

  # Optional: Set initial session for auto-login
  services.sysc-greet.settings.initial_session = {
    command = "Hyprland";
    user = "your-username";
  };
}
```

By default, the NixOS module does not install `niri`, `cagebreak`, `hyprland`, or `sway`.
Install your chosen backend yourself, or set `niriPackage`, `cagebreakPackage`, `hyprlandPackage`,
or `swayPackage` if you want the module to install and use a specific package.
If your compositor is managed elsewhere, set `compositorCommand` to the exact
command greetd should run.

Then rebuild:
```bash
sudo nixos-rebuild switch --flake .#your-hostname
```

> **Note:** On NixOS, sysc-greet automatically uses the correct data directory
> via build-time path injection. The Nix store path is injected at build time,
> so no manual configuration or symlinks are needed (in theory)

### Pre-built Packages (Debian/Ubuntu/Fedora)

Download pre-built packages from [GitHub Releases](https://github.com/Nomadcxx/sysc-greet/releases/latest):

**Latest: v1.1.8**

**Debian/Ubuntu:**
```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.1.8/sysc-greet_1.1.8_amd64.deb
sudo apt install ./sysc-greet_1.1.8_amd64.deb
```

**Fedora:**
```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.1.8/sysc-greet-1.1.8-1.x86_64.rpm
sudo dnf install ./sysc-greet-1.1.8-1.x86_64.rpm
```

See [Installation Guide](https://nomadcxx.github.io/sysc-greet/getting-started/installation/) for details.


## Documentation

For detailed docs, configuration guides, troubleshooting, and usage instructions, check out the [full documentation site](https://nomadcxx.github.io/sysc-greet/).

## License

GPL-3.0
