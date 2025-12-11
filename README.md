# sysc-greet

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go with the Bubble Tea framework.

![Preview](https://github.com/Nomadcxx/sysc-greet/raw/master/assets/showcase.gif)

## Features

- **Mucho themes**: Dracula, Gruvbox, Material, Nord, Tokyo Night, Catppuccin, Solarized, Monochrome, RAMA, DARK, TrainsIsHardJob, [Eldritch](https://github.com/eldritch-theme/eldritch).
- **Background Effects**: Fire (DOOM PSX), Matrix rain, ASCII rain, Fireworks, Aquarium
- **ASCII Effects**: Typewriter, Print, Beams, and Pour effects for session text (more ASCII animation in [sysc-Go](https://github.com/Nomadcxx/sysc-Go))
- **Border Styles**: Classic, Modern, Minimal (best), ASCII-1, ASCII-2, Wave, Pulse
- **Screensaver**: Configurable idle timeout with ASCII art cycling
- **Video Wallpapers**: Multi-monitor support via gslapper
- **Security Features**: Failed attempt counter with account lockout warnings, optional username caching

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

### Advanced: Manual Build from Source

**Requirements:**
- Go 1.21+
- greetd
- Wayland compositor (niri, hyprland, or sway)
- kitty (terminal)
- swww (wallpaper daemon) - **⚠️ Will be replaced with awww in next major release**
- gslapper (optional, for video wallpapers)

**Build and install:**

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go build -o sysc-greet ./cmd/sysc-greet/
sudo install -Dm755 sysc-greet /usr/local/bin/sysc-greet
```

**Install assets:**

```bash
sudo mkdir -p /usr/share/sysc-greet/{ascii_configs,fonts,wallpapers}
sudo cp -r ascii_configs/* /usr/share/sysc-greet/ascii_configs/
sudo cp -r fonts/* /usr/share/sysc-greet/fonts/
sudo cp -r wallpapers/* /usr/share/sysc-greet/wallpapers/
sudo cp config/kitty-greeter.conf /etc/greetd/kitty.conf
```

**Configure greetd** (`/etc/greetd/config.toml`):

Choose your compositor and update the command below:

```toml
[terminal]
vt = 1

[default_session]
# Pick one:
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
# command = "Hyprland -c /etc/greetd/hyprland-greeter-config.conf"
# command = "sway --unsupported-gpu -c /etc/greetd/sway-greeter-config"
user = "greeter"
```

**Create compositor config:**

Copy the appropriate config file to `/etc/greetd/`:

```bash
# For niri:
sudo cp config/niri-greeter-config.kdl /etc/greetd/

# For hyprland:
sudo cp config/hyprland-greeter-config.conf /etc/greetd/

# For sway:
sudo cp config/sway-greeter-config /etc/greetd/
```

**Create greeter user:**

```bash
sudo useradd -M -G video -s /usr/bin/nologin greeter
sudo mkdir -p /var/cache/sysc-greet /var/lib/greeter/Pictures/wallpapers
sudo chown -R greeter:greeter /var/cache/sysc-greet /var/lib/greeter
sudo chmod 755 /var/lib/greeter
```

**Enable service:**

```bash
sudo systemctl enable greetd.service
```

## Customization

### Wallpapers

There are two types of wallpapers you can use:

#### 1. Themed Wallpapers (Static Images)

These auto-match your selected theme and are stored in `/usr/share/sysc-greet/wallpapers/`.

**Format:** `sysc-greet-{theme}.png`

**Example:** `sysc-greet-nord.png` automatically shows when Nord theme is active.

**Adding themed wallpapers:**
```bash
sudo cp ~/my-nord-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-nord.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-nord.png
```

#### 2. Custom Wallpapers (Videos)

Video wallpapers are managed by [gSlapper](https://github.com/Nomadcxx/gSlapper) and stored in `/var/lib/greeter/Pictures/wallpapers/`.

**Supported formats:** MP4, WebM

**Adding video wallpapers:**
```bash
sudo cp ~/Videos/cool-animation.mp4 /var/lib/greeter/Pictures/wallpapers/
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/cool-animation.mp4
```

**Accessing wallpapers:**
Press `F1` (Settings) → Backgrounds → Select your wallpaper or video

### ASCII Art Format

Custom ASCII art configs in `/usr/share/sysc-greet/ascii_configs/`:

```
# cinnamon.conf
name=My Session

ascii_1=
 🬭🬭🬭🬭 🬞🬭🬭🬭🬏🬞🬭🬼 🬞🬭🬏🬞🬭🬼 🬞🬭🬏 🬭🬭🬭🬭 🬞🬭🬽  🭈🬭🬏 🬭🬭🬭🬭 🬞🬭🬼 🬞🬭🬏
▐▒▌ 🭣🬀 ▐▒▌ ▐▒🭌🬿▐▒▌▐▒🭌🬿▐▒▌▐▒▌▐▒▌▐▒█🭍🭂█▒▌▐▒▌▐▒▌▐▒🭌🬿▐▒▌
▐─▌    ▐─▌ ▐─▌🭥🭒─▌▐─▌🭥🭒─▌▐─🬛🬫─▌▐─▌🭣🭘▐─▌▐─▌▐─▌▐─▌🭥🭒─▌
▐░▌ 🭈🬏 ▐░▌ ▐░▌ ▐░▌▐░▌ ▐░▌▐░▌▐░▌▐░▌  ▐░▌▐░▌▐░▌▐░▌ ▐░▌
 🬂🬂🬂🬂 🬁🬂🬂🬂🬀🬁🬂🬀 🬁🬂🬀🬁🬂🬀 🬁🬂🬀🬁🬂🬀🬁🬂🬀🬁🬂🬀  🬁🬂🬀 🬂🬂🬂🬂 🬁🬂🬀 🬁🬂🬀
ascii_2=
𜺠𜵡𜶜𜺣 𜶜𜵡 ▄𜺣▗▖▄𜺣▗▖ 𜷋𜺣 ▄𜺣𜷋▖𜷋𜴧𜶜𜺣▄𜺣▗▖
█  𜺨 ▐▌ █𜴦𜷥▌█𜴦𜷥▌𜷥𜶬𜷖𜵈█𜴦▜▌█ ▐▌█𜴦𜷥▌
𜴦𜶻𜷋🯦 𜷕𜷀 █ ▐▌█ ▐▌█ ▐▌█ ▐▌𜶫▂𜷕𜴍█ ▐▌

```

### Roast Messages

Customize the typewriter/scrolling roast messages for each window manager:

**Location:** `/usr/share/sysc-greet/ascii_configs/{wm}.conf`

**Format:** Add a `roasts=` line with messages separated by `│`:

```
roasts=Your first roast │ Second hilarious message │ Third burn
```

**Example** (`hyprland.conf`):
```
roasts=Hyprland: The only WM with more config than code │ Still compiling? │ At least the animations are smooth
```

If no custom roasts are defined, sysc-greet uses built-in defaults.

**For more customization options (screensaver, compositor configs, etc.), see [CONFIGURATION.md](https://github.com/Nomadcxx/sysc-greet/blob/master/CONFIGURATION.md)**

## Usage

### Key Bindings

- **F1** - Settings menu (themes, borders, backgrounds)
- **F2** - Session selection
- **F3** - Release notes
- **F4** - Power menu (shutdown/reboot)
- **Page Up/Down** - Cycle ASCII variants
- **Tab** - Navigate fields
- **Enter** - Submit/Continue
- **Esc** - Cancel/Return to previous screen

### Test Mode

Test the greeter without locking your session:

```bash
sysc-greet --test

# Test in fullscreen (recommended for accurate preview)
kitty --start-as=fullscreen sysc-greet --test
```

### Additional Options

```bash
sysc-greet --theme dracula          # Start with specific theme
sysc-greet --border ascii-2         # Start with specific border
sysc-greet --screensaver            # Enable screensaver in test mode
sysc-greet --remember-username      # Cache username across sessions
```

## Related Projects

If you like ASCII animations, CLI and TUI terminal aesthetics check out my other projects:

- **[sysc-Go](https://github.com/Nomadcxx/sysc-Go)** - Terminal animation library for Go with effects like fire, matrix, beams, and text animations
- **[sysc-walls](https://github.com/Nomadcxx/sysc-walls)** - Terminal-based screensaver with idle detection and animated effects for Wayland/X11

## Acknowledgements

- [tuigreet](https://github.com/apognu/tuigreet) by apognu - Original inspiration and base
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm - Terminal styling
- [greetd](https://git.sr.ht/~kennylevinsen/greetd) by kennylevinsen - Login manager

## License

MIT
