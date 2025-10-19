# sysc-greet

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go with the Bubble Tea framework.

![Preview](https://github.com/Nomadcxx/sysc-greet/raw/master/assets/showcase.gif)

## Features

- **9 Themes**: Dracula, Gruvbox, Material, Nord, Tokyo Night, Catppuccin, Solarized, Monochrome, TransIsHardJob
- **Background Effects**: Fire (DOOM PSX), Matrix rain, ASCII rain, Static patterns
- **7 Border Styles**: Classic, Modern, Minimal, ASCII-1, ASCII-2, Wave, Pulse
- **Screensaver**: Configurable idle timeout with ASCII art cycling
- **Multiple ASCII Variants**: Page Up/Down navigation per session
- **Video Wallpapers**: Multi-monitor support via gslapper
- **Session Management**: Auto-detection of X11/Wayland sessions
- **Preference Caching**: Theme, background, border, session persistence

## Installation

### Arch Linux (AUR)

```bash
yay -S sysc-greet
# or
paru -S sysc-greet
```

### Quick Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

### Manual Build

**Requirements:**
- Go 1.21+
- greetd
- niri (compositor)
- kitty (terminal)
- swww (wallpaper daemon)
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

```toml
[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

**Create niri greeter config** (`/etc/greetd/niri-greeter-config.kdl`):

```kdl
// SYSC-Greet Niri config for greetd greeter session
hotkey-overlay {
    skip-at-startup
}

input {
    keyboard {
        xkb {
            layout "us"
        }
        repeat-delay 400
        repeat-rate 40
    }
    touchpad {
        tap;
    }
}

layer-rule {
    match namespace="^wallpaper$"
    place-within-backdrop true
}

layout {
    gaps 0
    center-focused-column "never"
    focus-ring { off }
    border { off }
}

animations {
    off
}

window-rule {
    match app-id="kitty"
    opacity 0.90
}

spawn-at-startup "swww-daemon"
spawn-sh-at-startup "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"

binds {
}
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

## Configuration

**For detailed configuration options, see [CONFIGURATION.md](https://github.com/Nomadcxx/sysc-greet/blob/master/CONFIGURATION.md)**

### ASCII Art Format

Custom ASCII art configs in `/usr/share/sysc-greet/ascii_configs/`:

```ini
# mysession.conf
name=My Session

ascii_1=
  Your ASCII art here
  Line 2
  Line 3

ascii_2=
  Alternative variant
  Line 2

colors=#ff5555,#50fa7b,#bd93f9
```

**Note:** `colors` define theme color overrides (accent, success, info)

## Usage

### Key Bindings

- **F2** - Settings menu (themes, borders, backgrounds)
- **F3** - Session selection
- **F4** - Power menu (shutdown/reboot)
- **F5** - Release notes
- **Page Up/Down** - Cycle ASCII variants
- **Tab** - Navigate fields
- **Enter** - Submit/Continue
- **Esc** - Cancel/Return

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
```

## Acknowledgements

- [tuigreet](https://github.com/apognu/tuigreet) by apognu - Original inspiration and base
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm - Terminal styling
- [greetd](https://git.sr.ht/~kennylevinsen/greetd) by kennylevinsen - Login manager

## License

MIT
