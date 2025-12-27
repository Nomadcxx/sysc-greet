# ASCII Art Customization

Each session type can have its own ASCII art with multiple variants. Press `Page Up/Down` at the greeter to cycle through them.

## Configuration Location

`/usr/share/sysc-greet/ascii_configs/`

Each session gets a `.conf` file (e.g., `hyprland.conf`, `kde.conf`, `gnome_desktop.conf`).

## Format

```ini
name=Hyprland

# Multiple variants (user can cycle through these)
ascii_1=
_____________________________  ________  .______________
\______   \_   _____/\______ \ \______ \ |   \__    ___/
 |       _/|    __)_  |    |  \ |    |  \|   | |    |
 |    |   \|        \ |    `   \|    `   \   | |    |
 |____|_  /_______  //_______  /_______  /___| |____|
        \/        \/         \/        \/
ascii_2=
 ________  ___  ___  ________  ___  __    ________
|\   ____\|\  \|\  \|\   ____\|\  \|\  \ |\   ____\
\ \  \___|\ \  \\\  \ \  \___|\ \  \/  /|\ \  \___|_
 \ \_____  \ \  \\\  \ \  \    \ \   ___  \ \_____  \
  \|____|\  \ \  \\\  \ \  \____\ \  \\ \  \|____|\  \
    ____\_\  \ \_______\ \_______\ \__\\ \__\____\_\  \
   |\_________\|_______|\|_______|\|__| \|__|\_________\
   \|_________|                             \|_________|
```

## Creating Custom ASCII

**ASCII generators:**

- [bit](https://github.com/superstarryeyes/bit) - Terminal ASCII art generator with many fonts
- [patorjk.com/software/taag](http://patorjk.com/software/taag/) - Web-based generator
- [ASCII Art Archive](https://www.asciiart.eu/) - Browse existing art

**Using bit:**

```bash
# Quick install (Linux/macOS)
curl -sfL https://raw.githubusercontent.com/superstarryeyes/bit/main/install.sh | sh

# Or build from source
git clone https://github.com/superstarryeyes/bit
cd bit
make build
./bit  # Interactive TUI mode

# CLI mode - generate ASCII art
bit "HYPRLAND"

# List available fonts
bit -list

# Use specific font with gradient
bit -font banner -color 31 -gradient 34 "HYPRLAND"
```

**Important:** Keep ASCII art under 80 columns wide for compatibility.

**Test your config:**

```bash
sysc-greet --test
```

## Roast Messages

Custom messages displayed with typewriter and scrolling ticker effects.

| Field | Type | Description |
|-------|------|-------------|
| roasts | String | Messages separated by `│` character |

**Example:**

```ini
roasts=First message │ Second message │ Third message
```

**How to customize roasts:**

1. Open the matching session config in `/usr/share/sysc-greet/ascii_configs/` (e.g., `hyprland.conf`, `kde.conf`, `openbox.conf`). The `name=` should reflect the session/WM.
2. Add or edit a single `roasts=` line with your messages separated by `│` (pipe). Keep it on one line.
3. If `roasts=` is omitted or left empty, the greeter falls back to built-in defaults per WM.
4. Use straight ASCII text; emojis may not render in all terminals.
5. Leave `screensaver.conf` without `roasts=` (screensaver uses clock/ASCII only).

The typewriter and scrolling ticker effects will cycle through these messages.

## Session Name Mapping

sysc-greet maps session names from XDG session files to config filenames:

| Session Name | Config Filename |
|--------------|-----------------|
| GNOME Desktop | gnome_desktop.conf |
| KDE Plasma | kde.conf |
| Hyprland | hyprland.conf |
| Sway | sway.conf |
| i3 | i3wm.conf |
| BSPWM | bspwm_manager.conf |
| Xmonad | xmonad.conf |
| Openbox | openbox.conf |
| Xfce | xfce.conf |
| Cinnamon | cinnamon.conf |
| IceWM | icewm.conf |
| Qtile | qtile.conf |
| Weston | weston.conf |

For sessions not listed, the lowercase first word of the session name is used as the config filename.
