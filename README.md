![sysc-greet](assets/header.webp)

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go with the Bubble Tea framework.

<img src="assets/showcase.webp" alt="sysc-greet tour: compositor headers, Higgs Field, Napoleon and other particle backgrounds, themes, and border styles" width="100%">

## Quick Links

- [Full Documentation](https://nomadcxx.github.io/sysc-greet/) - Complete guides, configuration, and usage instructions

> [!NOTE]
> The greeter supports Niri, Cagebreak, Sway, and Mango. Hyprland 0.55+ uses the Lua config and requires explicit opt-in with `SYSC_COMPOSITOR=hyprland`. Test startup, login, and quit on the target machine before enabling it for greetd.
>
> sysc-compositor support is experimental. sysc-compositor is unreleased and WIP; when its binary is installed, the greeter runs it with the Niri config and IPC.

## Installation

### Quick Install

One-line installer that works on most Linux distributions:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The installer detects the package manager on Arch Linux, Debian/Ubuntu, Fedora, openSUSE, Void Linux, and Gentoo. It installs the greeter dependencies and configures the selected backend.

### Build from Source

If you want to build it yourself, just clone the repo and run the installer:

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
sudo go run ./cmd/installer/
```

The installer walks you through compositor selection and configuration.
Set `SYSC_COMPOSITOR=hyprland` before running the installer to show the Hyprland 0.55+ option.

### Void Linux

Native XBPS packages are available for x86_64 glibc and musl. Download sysc-greet and gSlapper for the same architecture:

| Architecture | sysc-greet 1.3.0 | gSlapper 1.5.2 |
|---|---|---|
| x86_64 (glibc) | [Download](https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.0/sysc-greet-1.3.0_1.x86_64.xbps) | [Download](https://github.com/Nomadcxx/gSlapper/releases/download/v1.5.2/gslapper-1.5.2_1.x86_64.xbps) |
| x86_64 (musl) | [Download](https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.0/sysc-greet-1.3.0_1.x86_64-musl.xbps) | [Download](https://github.com/Nomadcxx/gSlapper/releases/download/v1.5.2/gslapper-1.5.2_1.x86_64-musl.xbps) |

Follow the [Void Linux guide](https://nomadcxx.github.io/sysc-greet/docs/getting-started/void-linux/#native-xbps-packages) to verify checksums, install the packages and configure greetd. These are upstream release packages; [submission to Void's repositories](https://github.com/void-linux/void-packages/pull/63202) is pending.

Release CI builds and install-checks both XBPS variants and publishes them with source archives and `SHA256SUMS-void`.

The Go installer also supports XBPS and runit. The tested setup uses elogind, D-Bus and polkit. Void glibc x86_64 VM tests passed desktop login, reboot and shutdown.

### Gentoo

The installer supports Portage and Gentoo's native OpenRC `display-manager` service. Use Sway from the main tree, or install another backend before selecting it. The amd64 desktop/OpenRC VM passed clean installation, login/logout, reboot and shutdown on VT 7.

See the [Gentoo guide](https://nomadcxx.github.io/sysc-greet/docs/getting-started/gentoo/) for prerequisites, USE settings and recovery commands. Gentoo systemd and minimal profiles have not received VM validation.

### Arch Linux (AUR)

Five AUR packages are available, one for each packaged greeter backend:

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

The packages conflict with each other — install exactly one. `sysc-greet-cagebreak` pulls in cagebreak and socat from the AUR.

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

By default, the NixOS module does not install `niri`, `cagebreak`, `hyprland`, `sway`, or `mango`.
Install your chosen backend yourself, or set `niriPackage`, `cagebreakPackage`, `hyprlandPackage`,
`swayPackage`, or `mangoPackage` if you want the module to install and use a specific package.
Mango is not in nixpkgs; `mangoPackage` must come from your own flake.
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

**Latest: v1.3.0**

**Debian/Ubuntu:**
```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.0/sysc-greet_1.3.0_amd64.deb
sudo apt install ./sysc-greet_1.3.0_amd64.deb
```

**Fedora:**
```bash
wget https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.0/sysc-greet-1.3.0-1.x86_64.rpm
sudo dnf install ./sysc-greet-1.3.0-1.x86_64.rpm
```

See [Installation Guide](https://nomadcxx.github.io/sysc-greet/docs/getting-started/installation/) for details.


## Startup theme

Use `--theme blue` to select a theme at startup. Names from the Themes menu are case insensitive. An explicit theme overrides the saved theme; test mode ignores saved preferences.

## Follow the shell theme

The greeter follows sysc-shell's committed named palette by default. The account chosen during installation controls the machine-wide greeter theme; each user's sysc-lock follows their own shell. The next greeter or lock start uses the selection.

The greeter installer uses `SUDO_USER` as the theme account. Set `SYSC_THEME_USER=username` when installing as root to choose it explicitly. It creates `/var/lib/sysc-greet/shell-theme/` owned by that account and readable by the greeter. Sysc-shell publishes the theme name there and in its user's XDG config directory. This keeps the user's home private.

Select **Follow sysc-shell** in the Themes menu to toggle following. Choosing a greeter theme switches to an independent selection. `--theme dracula` overrides the shell for that launch; `--follow-shell=false` disables following. `--shell-theme-file PATH` selects an exported name file for testing. Missing, invalid or unsupported selections retain the saved greeter theme. Generated wallpaper, hex and custom palettes use this fallback until color transport supports them.

## Optional login status

Enable the bordered CPU and Memory panels inside the login panel:

```sh
sysc-greet --test --metrics --secondary-backgrounds=false
```

Add `--weather-location LAT,LON` for weather at your configured coordinates. Use `--weather-units fahrenheit` for Fahrenheit; Celsius is the default. Both sources default to disabled. Each panel has its own subdued border and label. Unavailable readings show `--`; cached weather shows `stale` after a failed refresh. Weather hides first on narrow screens; the panels hide when authentication needs the space.

For persistent settings, create `/etc/sysc-greet/widgets.json` (for greetd) or `$XDG_CONFIG_HOME/sysc-greet/widgets.json` (default `~/.config/sysc-greet/widgets.json`). The first existing user file overrides the system file; launch flags override file settings. `--widget-config PATH` selects a different file.

```json
{
  "metrics": true,
  "weather_location": "-37.8136,144.9631",
  "weather_units": "celsius"
}
```

To reuse sysc-shell coordinates and units, set `"weather_location": "sysc-shell"` and omit `weather_units`. The default source is the running user's `sysc-shell/config.json` under the XDG config directory. Set `"shell_config": "/path/to/sysc-shell/config.json"` or `--weather-shell-config PATH` to choose another source. Greetd runs as `greeter`, so use a config that this account can read; do not assume it can access your home directory. Direct coordinates in the system widget file work without sysc-shell. Shell configs with a city name need explicit coordinates in the greeter file. If the shell config is unavailable or invalid, the greeter warns and hides weather while keeping login available. Set `"weather_location": ""` or pass `--weather-location=` to disable weather.

## Secondary monitors

On Niri, Cagebreak, Sway, Mango, and Hyprland 0.55+, the greeter can draw the selected background effect on each active secondary output with `sysc-terminal`. Experimental sysc-compositor builds use the same path through the Niri-compatible IPC exposed over `NIRI_SOCKET`. With no effect selected, the existing wallpaper remains. A selected GSlapper video suppresses secondary effects. `--secondary-effect pour` opts into a renderer effect only when the greeter background is None; the default is empty. Effects the installed `sysc-terminal` does not provide are skipped without substituting another effect.

`--secondary-backgrounds=false` disables secondary background effects, and `--secondary-exclude DP-3,HDMI-A-1` excludes outputs from those effects. The configured screensaver takes precedence over wallpapers and effects and appears in a read-only fullscreen Kitty child on every secondary output. Input in any saver window wakes the primary greeter and closes the other saver windows. Background helpers are removed when outputs disappear or the greeter exits. Install `sysc-terminal` separately to mirror effects; Kitty is already required by the greeter.

## Border styles

Minimal is the default and the fallback when no valid border style is saved; valid saved styles continue to load unchanged. Open **F1 → Borders** to choose Classic, Modern, Minimal, ASCII-1, ASCII-2, NFO 01–03, or Braille Hairline. Sextant Bevel appears in Kitty, the terminal used by greetd for sysc-greet. The hug layout keeps the frame around the form; `--border-layout=fill` grows it toward the terminal edges.

The border effect starts Off. Shimmer starts after ten seconds without input, travels at 24 cells per second, pauses four seconds after each lap, and stops after one minute idle.

Auto form mode uses see-through for AK-47, Australian SLR, Sonar, Cracktro, and Plasma; frosted for Matrix, Fire, ASCII Rain, Fireworks, and Aquarium; and opaque for None, Pattern, and Typewriter. Set the form mode and choose the Theme, VGA, or Terminal palette in F1 → Borders.

Built-in `.ans` assets live in `cmd/sysc-greet/assets/borders/` and ship inside the binary. The greeter also loads system assets from `/usr/share/sysc-greet/borders/`. Add a SAUCE `nineslice` comment with the border insets and asset name. Preview an asset at several sizes with:

```sh
sysc-greet --test --border-preview path/to/border.ans --size 80x24,132x40,200x55
```

The preview prints ANSI output to the terminal. The greeter rejects test mode when greetd sets `GREETD_SOCK`.

## Documentation

For detailed docs, configuration guides, troubleshooting, and usage instructions, check out the [full documentation site](https://nomadcxx.github.io/sysc-greet/).

## License

sysc-greet's own source code is licensed under [GPL-3.0-only](LICENSE).

See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for notices covering third-party code and model assets supplied through sysc-Go.

---

<a href="https://github.com/Nomadcxx"><img src="https://raw.githubusercontent.com/Nomadcxx/Nomadcxx/main/assets/rama-mark.svg" height="22" alt="RAMA"></a> — terminal-native tooling for the linux desktop.
[More projects →](https://github.com/Nomadcxx) · [Sponsor](https://github.com/sponsors/Nomadcxx) ❤️
