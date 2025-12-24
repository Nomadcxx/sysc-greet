# Hyprland Compositor Configuration

sysc-greet works with the Hyprland Wayland compositor. This guide covers the configuration required for sysc-greet to function properly.

## greetd Configuration

### config.toml

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "Hyprland -c /etc/greetd/hyprland-greeter-config.conf"
user = "greeter"
```

## Hyprland Config

sysc-greet provides a pre-configured Hyprland configuration file.

### hyprland-greeter-config.conf

Edit `/etc/greetd/hyprland-greeter-config.conf`:

```ini
# Source default Hyprland config
source = ~/.config/hypr/hyprland.conf

# Auto-start gSlapper with default wallpaper
# Comment this out if using swww or no wallpaper
exec-once = gslapper -s -o "loop panscan=1.0" '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png

# Hide cursor
cursor {
    invisible = true
}

# Start kitty with sysc-greet
exec-once = kitty --config /etc/greetd/kitty.conf --override hide_window_decorations=yes -e /usr/local/bin/sysc-greet
```

### Key Fields

| Setting | Value | Description |
|----------|--------|-------------|
| exec-once | gslapper | Start gSlapper wallpaper daemon |
| exec-once | kitty | Start sysc-greet in kitty terminal |
| -s | Daemon mode | Run gSlapper in background |
| -o "loop panscan=1.0" | GStreamer options | Loop and panscan settings |
| cursor.invisible | true | Hide cursor completely |
| hide_window_decorations | Kitty option | Hide window title bar |

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

After configuration, verify Hyprland starts correctly:

```bash
# Restart greetd
sudo systemctl restart greetd

# View greetd logs (includes compositor output)
journalctl -u greetd -n 50
```

## Keyboard Layout

For non-US layouts, see [Keyboard Layout Configuration](../configuration/keyboard-layout.md).
