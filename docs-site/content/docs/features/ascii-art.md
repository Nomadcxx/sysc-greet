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

Unquoted art skips blank lines and lines starting with `#`. To keep every line as written, open the variant with `"""` and close it with `"""` on its own line:

```ini
ascii_3="""
 _____   _____  ___
/ __\ \ / / __|/ __|
\__ \\ V /\__ \ (__
|___/ |_| |___/\___|

  ___ ___ ___ ___ _____
 / __| _ \ __| __|_   _|
| (_ |   / _|| _|  | |
 \___|_|_\___|___| |_|
"""
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

For art with embedded SGR colors, set `ansi_style=true`. The setting covers the whole file. Variants without escape sequences keep the `color=` or theme color; variants with them ignore `color=`.

sysc-greet accepts the ESC byte or the spelled-out `\033`. It maps the 16 ANSI colors to colors of the active theme, so the art follows theme changes:

| SGR foreground | SGR background | Theme color |
|----------------|----------------|-------------|
| 30 | 40 | `bg_base` |
| 31, 91 | 41 | `danger` |
| 32, 35, 92, 95 | 42, 45 | `accent` |
| 33, 93 | 43 | `warning` |
| 34, 36 | 44, 46 | `secondary` |
| 37 | 47 | `fg_secondary` |
| 90 | | `fg_muted` |
| 94, 96 | | `primary` |
| 97 | | `fg_primary` |

`0` resets the text to `fg_primary` on `bg_base`; `39` and `49` reset one of the two. sysc-greet skips 256-color and 24-bit codes and drops other control sequences.

The Beams and Pour effects draw the art in their own colors. Print shows ANSI art whole instead of revealing it line by line.

## Creating Custom ASCII

Install [sysclet v0.3.0 or later](https://github.com/Nomadcxx/sysclet/releases). It renders text with TheDraw fonts. In its TUI, select the sysc-greet export, choose a profile and `ascii_N` slot, then press `Ctrl+S`. sysclet writes the profile directly and sets `ansi_style=true`.

From the command line:

```bash
# List the bundled fonts, or those whose names contain a filter
sysclet fonts
sysclet fonts bleach

# Preview
sysclet render --font-name Bleach NIRI

# Print a profile snippet to paste into a .conf file
sysclet export --font-name Bleach --slot ascii_1 NIRI

# Or update a system-owned profile directly
sudo "$(command -v sysclet)" export --font-name Bleach --slot ascii_1 \
  --config /usr/share/sysc-greet/ascii_configs/niri.conf NIRI
```

sysclet snippets set `ansi_style=true` and wrap the art in `"""`. With `--config`, sysclet replaces the slot or appends it, sets `ansi_style=true`, and first saves the original beside it as `niri.conf.sysclet-<timestamp>.bak`. It refuses to write through a symlink.

The installer and the AUR package copy the bundled profiles into `/usr/share/sysc-greet/ascii_configs/` on each install, so a reinstall or upgrade replaces an edited profile. Keep a copy of your changes.

Browse the [ASCII Art Archive](https://www.asciiart.eu/) for existing artwork. Plain text art needs no `ansi_style`; art with SGR colors needs `ansi_style=true`.

Keep art under 80 columns wide for compatibility. The framed layouts need a terminal of at least 80 columns and 24 rows, and sysc-greet uses them only when the frame and art fit. Otherwise it shows the art above a compact login form, and drops the art when that does not fit either.

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
| Plasma | kde.conf |
| Miracle | miracle-wm.conf |

For sessions not listed, the lowercase first word of the session name is used as the config filename.
