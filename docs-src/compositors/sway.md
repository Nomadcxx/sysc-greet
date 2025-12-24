# Sway Compositor Configuration

sysc-greet works with the Sway Wayland compositor. This guide covers the configuration required for sysc-greet to function properly.

## greetd Configuration

### config.toml

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "sway --unsupported-gpu -c /etc/greetd/sway-greeter-config"
user = "greeter"
```

## Sway Config

sysc-greet provides a pre-configured Sway configuration file.

### sway-greeter-config

Edit `/etc/greetd/sway-greeter-config`:

```
# Source default Sway config
include /etc/sway/config

# Auto-start gSlapper with default wallpaper
# Comment this out if using swww or no wallpaper
exec gSlapper -s -o "loop panscan=1.0" '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png

# Hide cursor after inactivity
seat * hide_cursor 1000

# Start kitty with sysc-greet
exec kitty --config /etc/greetd/kitty.conf --override hide_window_decorations=yes -e /usr/local/bin/sysc-greet
```

### Key Commands

| Command | Description |
|----------|-------------|
| exec gSlapper | Start gSlapper wallpaper daemon |
| exec kitty | Start sysc-greet in kitty terminal |
| -s | Daemon mode (run in background) |
| -o "loop panscan=1.0" | GStreamer options for video playback |
| seat * hide_cursor 1000 | Hide cursor after 1000ms inactivity |
| hide_window_decorations | Kitty option to hide title bar |

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

After configuration, verify Sway starts correctly:

```bash
# Restart greetd
sudo systemctl restart greetd

# View greetd logs (includes compositor output)
journalctl -u greetd -n 50
```

## Keyboard Layout

For non-US layouts, see [Keyboard Layout Configuration](../configuration/keyboard-layout.md).
