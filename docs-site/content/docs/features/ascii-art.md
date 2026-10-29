---
title: "ASCII Art Customization"
description: "Each session type can have its own ASCII art with multiple variants. Press Page Up/Down at the greeter to cycle through them."
---

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

## Per-Session Color Override

Override the ASCII art color for a specific session, independent of your selected theme:

```ini
name=Hyprland
color=#89b4fa

ascii_1=
...
```

For plain art, `color=` sets its color. If omitted, sysc-greet uses the active theme's primary color.

## Theme-aware ANSI art

For art with embedded SGR colors, set `ansi_style=true`. sysc-greet maps the 16 ANSI color slots to roles in the active theme and strips unsupported terminal control sequences.

## Creating Custom ASCII

Install [sysclet v0.3.0 or later](https://github.com/Nomadcxx/sysclet/releases). In its TUI, select the sysc-greet export, choose a profile and `ascii_N` slot, then press `Ctrl+S`. sysclet writes the profile directly and sets `ansi_style=true`.

To print a profile snippet from the command line:

```sh
sysclet export --font-name Bleach --slot ascii_1 COWBOY
```

Add `--config PROFILE` to update a profile directly. Use elevated privileges for system-owned files under `/usr/share/sysc-greet/ascii_configs/`.

Browse the [ASCII Art Archive](https://www.asciiart.eu/) for existing artwork. Keep art under 80 columns wide for compatibility.

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
| Mango | mango.conf |

For sessions not listed, the lowercase first word of the session name is used as the config filename.
