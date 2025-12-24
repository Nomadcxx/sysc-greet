# Niri Compositor Configuration

sysc-greet works with the Niri Wayland compositor. This guide covers the configuration required for sysc-greet to function properly.

## greetd Configuration

### config.toml

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

## Niri Config

sysc-greet provides a pre-configured Niri configuration file that starts sysc-greet with kitty.

### niri-greeter-config.kdl

Create or edit `/etc/greetd/niri-greeter-config.kdl`:

```kdl
input {
    keyboard {
        // Optional: Configure keyboard layout
        xkb {
            layout "us"
        }
    }
}

output {
    // Auto-start gSlapper with default wallpaper
    // Comment this out if using swww or no wallpaper
    spawn-at-startup "gslapper" "-s" "-o" "loop panscan=1.0" "*" "/usr/share/sysc-greet/wallpapers/sysc-greet-default.png"
}

// Hide cursor after inactivity
cursor {
    hide-when-typing
    hide-after-inactive-ms 1000
}

// Start kitty with sysc-greet
spawn-at-startup "kitty" "--config" "/etc/greetd/kitty.conf" "--override" "hide_window_decorations=yes" "-e" "/usr/local/bin/sysc-greet"
```

### Key Fields

| Setting | Value | Description |
|----------|--------|-------------|
| spawn-at-startup | gslapper | Start gSlapper wallpaper daemon |
| spawn-at-startup | kitty | Start sysc-greet in kitty terminal |
| -s | Daemon mode | Run gSlapper in background |
| -o "loop panscan=1.0" | GStreamer options | Loop and panscan settings |
| -o hide-window-decorations | Kitty option | Hide window title bar |

## Wallpaper Setup

sysc-greet configures gSlapper to use the Unix socket at `/tmp/sysc-greet-wallpaper.sock` for IPC communication.

### Themed Wallpapers

When you change themes, sysc-greet automatically switches to themed wallpapers in `/usr/share/sysc-greet/wallpapers/`.

### Video Wallpapers

Video wallpapers are stored in `/var/lib/greeter/Pictures/wallpapers/` and can be selected from the F1 Wallpaper menu.

## Permissions

Ensure proper file permissions:

```bash
sudo chown -R greeter:greeter /var/cache/sysc-greet
sudo chown -R greeter:greeter /var/lib/greeter/Pictures/wallpapers
sudo chmod 755 /var/lib/greeter
```

## Verification

After configuration, verify Niri starts correctly:

```bash
# Restart greetd
sudo systemctl restart greetd

# View greetd logs (includes compositor output)
journalctl -u greetd -n 50
```

## Keyboard Layout

For non-US layouts, see [Keyboard Layout Configuration](../configuration/keyboard-layout.md).
