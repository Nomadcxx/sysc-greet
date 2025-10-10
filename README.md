# sysc-greet

A graphical console greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go.

## Features

- **Themes**: Dracula, Gruvbox, Material, Nord, Tokyo Night, Catppuccin, Solarized, Monochrome, TransIsHardJob
- **Background Animations**: Fire (DOOM PSX effect), Matrix (digital rain), Rain (ASCII drops), Static Pattern
- **Border Styles**: Classic, Modern, Minimal, ASCII-1, ASCII-2, Wave, Pulse
- **Preference Caching**: Remembers theme, background, border style, and session
- **Multiple ASCII Variants**: Page Up/Down to cycle through different ASCII art per session
- **Multi-Background Support**: Enable multiple effects simultaneously

## Installation

### Prerequisites

- Go 1.21+
- greetd
- Terminal emulator (kitty recommended)

### Build and Install

```bash
git clone https://github.com/Nomadcxx/sysc-greet
cd sysc-greet
./install.sh
```

The installer builds the greeter, installs it to `/usr/local/bin/sysc-greet`, and copies required assets to `/usr/share/sysc-greet/`.

### Configure greetd

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "kitty --class=greeter -e sysc-greet"
user = "greeter"
```

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
