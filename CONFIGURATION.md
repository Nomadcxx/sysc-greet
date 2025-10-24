# Configuration Guide

## Table of Contents
- [ASCII Art Customization](#ascii-art-customization)
- [Screensaver Configuration](#screensaver-configuration)
- [Wallpapers](#wallpapers)
- [Keyboard Layout](#keyboard-layout)
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

# Color gradient (hex colors - used for theme color overrides)
colors=#89b4fa,#a6e3a1,#f9e2af,#fab387,#f38ba8,#cba6f7
```

### Creating Custom ASCII

**ASCII generators:**
- [patorjk.com/software/taag](http://patorjk.com/software/taag/) - Web-based generator
- [ASCII Art Archive](https://www.asciiart.eu/) - Browse existing art
- [Mobius](https://github.com/Mobius-Team/Mobius) - Advanced ASCII art tools

**Using figlet:**

Install figlet and figlet-fonts for local generation:

```bash
# Arch Linux
sudo pacman -S figlet

# Install additional fonts
git clone https://github.com/xero/figlet-fonts
sudo cp figlet-fonts/* /usr/share/figlet/

# Generate ASCII
figlet -f dos_rebel "HYPRLAND"
```

**Figlet fonts:** [github.com/xero/figlet-fonts](https://github.com/xero/figlet-fonts)

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

# Clock style: kompaktblk (default, 3 rows), phmvga (2 rows, crisp), dos_rebel (8 rows, retro), plain (single line)
clock_style=kompaktblk

# ASCII variants (cycles every 5 minutes)
ascii_1=
  ▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
   ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
  ▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀
  //  SEE YOU SPACE COWBOY //

ascii_2=
.________._______._______ .____/\ .________
|    ___/: .____/:_.  ___\:   /  \|    ___/
|___    \| : _/\ |  : |/\ |.  ___/|___    \
|       /|   /  \|    /  \|     \ |       /
|__:___/ |_.: __/|. _____/|      \|__:___/ 
   :        :/    :/      |___\  /   :     
                  :            \/ 
```

### Time Format Reference

Go uses the reference time `01/02 03:04:05PM '06 -0700` (1234567 - memorable, right?).

**Common formats:**
- `3:04:05 PM` - 12-hour with seconds
- `15:04:05` - 24-hour with seconds
- `Monday, January 2, 2006` - Full date
- `2006-01-02` - ISO format

### Behavior
- Activates after `idle_timeout` minutes
- Exits on any keyboard/mouse input
- Cycles through ASCII variants every 5 minutes

---

## Wallpapers

There are two types of wallpapers, each stored in different locations and managed by different tools:

### 1. Themed Wallpapers (Static Images)

**Location:** `/usr/share/sysc-greet/wallpapers/`
**Managed by:** [swww](https://github.com/LGFae/swww) (Wayland wallpaper daemon)

These auto-match your selected theme using the naming convention `sysc-greet-{theme}.png`.

**Included themed wallpapers:**
- `sysc-greet-dracula.png`
- `sysc-greet-gruvbox.png`
- `sysc-greet-nord.png`
- `sysc-greet-tokyo-night.png`
- `sysc-greet-catppuccin.png`
- `sysc-greet-material.png`
- `sysc-greet-solarized.png`
- `sysc-greet-monochrome.png`
- `sysc-greet-transishardjob.png`

**Adding/replacing themed wallpapers:**

```bash
# Replace an existing theme wallpaper
sudo cp ~/my-nord-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-nord.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-nord.png

# Add a wallpaper for a new theme
sudo cp ~/my-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-mytheme.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-mytheme.png
```

### 2. Custom Wallpapers (Videos)

**Location:** `/var/lib/greeter/Pictures/wallpapers/`
**Managed by:** [gSlapper](https://github.com/Nomadcxx/gSlapper) (Video wallpaper manager)

Video wallpapers provide animated backgrounds with multi-monitor support. Also if you are still using mpvpaper in 2025 for video backgrounds you need to reflect on your life choices and install gSlapper instead.

**Requirements:**
```bash
# Install gSlapper (if not already installed)
yay -S gslapper
# or build from source: https://github.com/Nomadcxx/gSlapper
```

**Adding video wallpapers:**

```bash
# Copy video to greeter's wallpaper directory
sudo cp ~/Videos/cool-animation.mp4 /var/lib/greeter/Pictures/wallpapers/

# Set correct ownership
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/cool-animation.mp4
```

### Accessing Wallpapers in Greeter

Press `F1` (Settings) → Backgrounds → Select your wallpaper or video

Both static and video wallpapers will appear in the same menu.

---

## Cursor Visibility

Kitty auto-hides the cursor after 3 seconds of inactivity. To hide it permanently, add compositor options:

### niri

Edit `/etc/greetd/niri-greeter-config.kdl`:

```kdl
cursor {
    hide-when-typing
    hide-after-inactive-ms 1000
}
```

### sway

Edit `/etc/greetd/sway-greeter-config`:

```bash
seat * hide_cursor 1000
```

### hyprland

Edit `/etc/greetd/hyprland-greeter-config.conf`:

```ini
cursor {
    invisible = true
}
```

Restart greetd after changes: `sudo systemctl restart greetd`

---

## Keyboard Layout

sysc-greet runs inside a compositor, so keyboard layout is set there.

### niri

Edit `/etc/greetd/niri-greeter-config.kdl`:

```kdl
input {
    keyboard {
        xkb {
            layout "de"
        }
    }
}
```

### sway

Edit `/etc/greetd/sway-greeter-config`:

```bash
input * {
    xkb_layout "de"
}
```

### hyprland

Edit `/etc/greetd/hyprland-greeter.conf`:

```ini
input {
    kb_layout = de
}
```

Replace `de` with your layout (`us`, `fr`, `es`, `uk`, etc). Full list in `/usr/share/X11/xkb/rules/base.lst`.

Restart greetd after changes: `sudo systemctl restart greetd`

---

## Key Locations

### Configuration Files
- **greetd config:** `/etc/greetd/config.toml`
- **Compositor configs:**
  - Niri: `/etc/greetd/niri-greeter-config.kdl`
  - Hyprland: `/etc/greetd/hyprland-greeter-config.conf`
  - Sway: `/etc/greetd/sway-greeter-config`
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

*See you in space cowboy
