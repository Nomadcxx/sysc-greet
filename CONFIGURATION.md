# Configuration Guide

## Table of Contents
- [ASCII Art Customization](#ascii-art-customization)
- [Screensaver Configuration](#screensaver-configuration)
- [Wallpapers](#wallpapers)
- [Key Locations](#key-locations)

---

## ASCII Art Customization

Each session type can have its own ASCII art with multiple variants. Press `Page Up/Down` at the greeter to cycle through them.

### Configuration Location

`/usr/share/sysc-greet/ascii_configs/`

Each session gets a `.conf` file (e.g., `hyprland.conf`, `kde.conf`, `gnome_desktop.conf`).

### Format

```ini
name=Hyprland

# Multiple variants (user can cycle through these)
ascii_1=
  ██╗  ██╗██╗   ██╗██████╗ ██████╗ ██╗      █████╗ ███╗   ██╗██████╗
  ██║  ██║╚██╗ ██╔╝██╔══██╗██╔══██╗██║     ██╔══██╗████╗  ██║██╔══██╗
  ███████║ ╚████╔╝ ██████╔╝██████╔╝██║     ███████║██╔██╗ ██║██║  ██║
  ██╔══██║  ╚██╔╝  ██╔═══╝ ██╔══██╗██║     ██╔══██║██║╚██╗██║██║  ██║
  ██║  ██║   ██║   ██║     ██║  ██║███████╗██║  ██║██║ ╚████║██████╔╝
  ╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═════╝

ascii_2=
   _  ___   ______  ____  __    ___   _  ______
  / |/ / | / / __ \/ __ \/ /   / _ | / |/ / __ \
 /    /| |/ / /_/ / /_/ / /__ / __ |/    / /_/ /
/_/|_/ |___/ .___/\____/____//_/ |_/_/|_/_____/
          /_/

# Color gradient (hex colors - used for theme color overrides)
colors=#89b4fa,#a6e3a1,#f9e2af,#fab387,#f38ba8,#cba6f7
```

### Creating Custom ASCII

**ASCII generators:**
- [patorjk.com/software/taag](http://patorjk.com/software/taag/)
- `figlet` command-line tool

**Important:** Keep ASCII art under 80 columns wide for compatibility.

**Test your config:**
```bash
sysc-greet --test
```

---

## Screensaver Configuration

The login screen has a screensaver because waiting for authentication can be an aesthetic experience.

### Configuration File

`/usr/share/sysc-greet/ascii_configs/screensaver.conf`

```ini
# Idle time before activation (minutes)
idle_timeout=5

# Time/Date formats (Go time format)
time_format=3:04:05 PM
date_format=Monday, January 2, 2006

# Clock size: small, medium, large
clock_size=medium

# Animation on screensaver start
animate_on_start=true
animation_type=print
animation_speed=20

# ASCII variants (cycles every 5 minutes)
ascii_1=
  ▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
   ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
  ▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀
  //  SEE YOU SPACE COWBOY //

ascii_2=
  ███████╗██╗     ███████╗███████╗██████╗
  ██╔════╝██║     ██╔════╝██╔════╝██╔══██╗
  ███████╗██║     █████╗  █████╗  ██████╔╝
  ╚════██║██║     ██╔══╝  ██╔══╝  ██╔═══╝
  ███████║███████╗███████╗███████╗██║
  ╚══════╝╚══════╝╚══════╝╚══════╝╚═╝
```

### Time Format Reference

Go uses the reference time `01/02 03:04:05PM '06 -0700` (1234567 - memorable, right?).

**Common formats:**
- `3:04:05 PM` - 12-hour with seconds
- `15:04:05` - 24-hour with seconds
- `Monday, January 2, 2006` - Full date
- `2006-01-02` - ISO format

### Animation Types
- `print` - Typewriter reveal effect
- `none` - Instant appearance

### Behavior
- Activates after `idle_timeout` minutes
- Exits on any keyboard/mouse input
- Cycles through ASCII variants every 5 minutes

---

## Wallpapers

### Location
`/usr/share/sysc-greet/wallpapers/`

### Supported Formats
- **Static:** PNG, JPG (via `swww`)
- **Video:** MP4, WebM (via `gslapper`)

### Theme Wallpapers
Images named `sysc-greet-{theme}.png` are automatically matched to themes.

### Accessing in Greeter
Press `F2` → Backgrounds → Select your wallpaper or background effect

---

## Key Locations

### Configuration Files
- **greetd config:** `/etc/greetd/config.toml`
- **Niri config:** `/etc/greetd/niri-greeter-config.kdl`
- **Kitty config:** `/etc/greetd/kitty.conf`

### Data Directories
- **ASCII configs:** `/usr/share/sysc-greet/ascii_configs/`
- **Fonts:** `/usr/share/sysc-greet/fonts/`
- **Wallpapers:** `/usr/share/sysc-greet/wallpapers/`
- **Cache:** `/var/cache/sysc-greet/`
- **Greeter home:** `/var/lib/greeter/`

### Binary Location
- **Executable:** `/usr/local/bin/sysc-greet`

### Logs
- **greetd logs:** `journalctl -u greetd`
- **Debug log:** `/tmp/sysc-greet-debug.log` (when using `--debug` flag)

### Permissions
- **Greeter user:** `greeter` (created during install)
- **Cache ownership:** `greeter:greeter` on `/var/cache/sysc-greet`

---

## Troubleshooting

**Greeter won't start:**
```bash
sudo systemctl status greetd
journalctl -u greetd -n 50
```

**ASCII art broken:**
- Keep width ≤ 80 columns
- Test first: `cat /usr/share/sysc-greet/ascii_configs/yourfile.conf`

**Screensaver not working:**
- Verify `/usr/share/sysc-greet/ascii_configs/screensaver.conf` exists
- Test: `sysc-greet --test --screensaver`

**Preferences not saving:**
```bash
sudo chown -R greeter:greeter /var/cache/sysc-greet
sudo chmod 755 /var/cache/sysc-greet
```

---

*Made with questionable amounts of caffeine.*
