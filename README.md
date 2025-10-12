# sysc-greet

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go.
![Preview](https://github.com/Nomadcxx/sysc-greet/raw/master/assets/showcase.gif)

## Features

- **Themes**: Dracula, Gruvbox, Material, Nord, Tokyo Night, Catppuccin, Solarized, Monochrome, TransIsHardJob
- **Background Animations**: Fire (DOOM PSX effect), Matrix (digital rain), Rain (ASCII drops), Static Pattern
- **Border Styles**: Classic, Modern, Minimal, ASCII-1, ASCII-2, Wave, Pulse
- **Preference Caching**: Remembers theme, background, border style, and session
- **Multiple ASCII Variants**: Page Up/Down to cycle through different ASCII art per session
- **Multi-Background Support**: Enable multiple effects simultaneously

## Installation

### Quick Install (One-Line)

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

### Manual Install

**Clone and run installer:**

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
go build -o install-sysc-greet ./cmd/installer/
sudo ./install-sysc-greet
```

**What the installer does:**
- Checks dependencies (Go, systemd, package manager)
- Installs greetd and gslapper (optional, for video wallpapers)
- Builds sysc-greet binary
- Installs to `/usr/local/bin/sysc-greet`
- Copies configs to `/usr/share/sysc-greet/`
- Configures greetd with niri compositor
- Enables greetd.service

After installation, reboot to see sysc-greet as your login screen.

### Configuration

The installer automatically configures greetd at `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

Manual configuration not required unless using a different compositor.

## Usage

### Key Bindings

- **F2**: Settings menu (themes, borders, backgrounds)
- **F3**: Session selection
- **F4**: Power menu (shutdown/reboot)
- **F5**: Release notes
- **Page Up/Down**: Cycle ASCII variants

### Test Mode

```bash
sysc-greet --test
```

## ASCII Art Configuration

Add custom ASCII art to `/usr/share/sysc-greet/ascii_configs/`:

```ini
# mysession.conf
name=mysession

ascii_1=
  Your ASCII art here
  Line 2
  Line 3

ascii_2=
  Alternative variant
  Line 2

colors=#color1,#color2,#color3
animation_style=rainbow
animation_speed=1.0
```

## License

MIT
