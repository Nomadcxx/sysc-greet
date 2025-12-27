# Niri Setup

Configuration for running sysc-greet with the niri Wayland compositor.

## greetd Config

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"

[initial_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

## Compositor Config

Edit `/etc/greetd/niri-greeter-config.kdl`:

```kdl
hotkey-overlay {
    skip-at-startup
}

input {
    keyboard {
        xkb {
            layout "us"
        }
        repeat-delay 400
        repeat-rate 40
    }

    touchpad {
        tap
    }
}

layer-rule {
    match namespace="^wallpaper$"
    place-within-backdrop true
}

layout {
    gaps 0
    center-focused-column "never"

    focus-ring {
        off
    }

    border {
        off
    }
}

animations {
    off
}

window-rule {
    match app-id="kitty"
    opacity 0.90
}

// Start gslapper with default wallpaper (forked to background with IPC socket)
spawn-at-startup "gslapper" "-f" "-I" "/tmp/sysc-greet-wallpaper.sock" "*" "/usr/share/sysc-greet/wallpapers/sysc-greet-default.png"

spawn-sh-at-startup "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"

binds {
}
```

## Cursor Visibility

To hide the cursor, uncomment in the config:

```kdl
cursor {
    hide-when-typing
    hide-after-inactive-ms 1000
}
```

## Keyboard Layout

Change `layout "us"` to your preferred layout. See [Keyboard Layout](../configuration/keyboard-layout.md) for details.

## Verification

```bash
# Restart greetd
sudo systemctl restart greetd

# Check logs
journalctl -u greetd -n 50
```
