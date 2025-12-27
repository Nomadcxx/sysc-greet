# Sway Setup

Configuration for running sysc-greet with the Sway Wayland compositor.

## greetd Config

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "sway -c /etc/greetd/sway-greeter-config"
user = "greeter"

[initial_session]
command = "sway -c /etc/greetd/sway-greeter-config"
user = "greeter"
```

## Compositor Config

Edit `/etc/greetd/sway-greeter-config`:

```bash
# SYSC-Greet Sway config for greetd greeter session
# Monitors auto-detected by Sway at runtime

# Disable window borders
default_border none
default_floating_border none

# No gaps needed for greeter
gaps inner 0
gaps outer 0

# Input configuration
input * {
    xkb_layout "us"
    repeat_delay 400
    repeat_rate 40
}

input type:touchpad {
    tap enabled
}

# Window rules for kitty
for_window [app_id="kitty"] fullscreen enable

# Startup applications
# Start gslapper with default wallpaper (forked to background with IPC socket)
exec gslapper -f -I /tmp/sysc-greet-wallpaper.sock '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png
exec "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; swaymsg exit"
```

## Cursor Visibility

To hide the cursor after inactivity, add:

```bash
seat * hide_cursor 1000
```

## Keyboard Layout

Change `xkb_layout "us"` to your preferred layout. See [Keyboard Layout](../configuration/keyboard-layout.md) for details.

## Verification

```bash
# Restart greetd
sudo systemctl restart greetd

# Check logs
journalctl -u greetd -n 50
```
