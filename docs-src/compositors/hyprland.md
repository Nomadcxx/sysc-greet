# Hyprland Setup

Configuration for running sysc-greet with the Hyprland Wayland compositor.

## greetd Config

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "start-hyprland -- -c /etc/greetd/hyprland-greeter-config.conf"
user = "greeter"

[initial_session]
command = "start-hyprland -- -c /etc/greetd/hyprland-greeter-config.conf"
user = "greeter"
```

## Compositor Config

Edit `/etc/greetd/hyprland-greeter-config.conf`:

```ini
# SYSC-Greet Hyprland config for greetd greeter session
# Monitors auto-detected by Hyprland at runtime

# No animations for faster greeter startup
animations {
    enabled = false
}

# Minimal decorations
decoration {
    rounding = 0
    blur {
        enabled = false
    }
}

# Greeter doesn't need gaps
general {
    gaps_in = 0
    gaps_out = 0
    border_size = 0
}

misc {
    disable_hyprland_logo = true
    disable_splash_rendering = true
    background_color = rgb(000000)
    # Suppress watchdog warning - greetd doesn't pass fd properly to start-hyprland
    disable_watchdog_warning = true
}

# Input configuration
input {
    kb_layout = us
    repeat_delay = 400
    repeat_rate = 40

    touchpad {
        tap-to-click = true
    }
}

# Window rules for kitty greeter
windowrule = match:class ^(kitty)$, fullscreen on, opacity 1.0

# Layer rules for wallpaper daemon
layerrule = match:namespace wallpaper, blur on

# Startup applications
# Start gslapper with default wallpaper (forked to background with IPC socket)
exec-once = gslapper -f -I /tmp/sysc-greet-wallpaper.sock '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png
exec-once = XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch exit
```

## Cursor Visibility

To hide the cursor completely, add:

```ini
cursor {
    invisible = true
}
```

## Keyboard Layout

Change `kb_layout = us` to your preferred layout. See [Keyboard Layout](../configuration/keyboard-layout.md) for details.

## Verification

```bash
# Restart greetd
sudo systemctl restart greetd

# Check logs
journalctl -u greetd -n 50
```
