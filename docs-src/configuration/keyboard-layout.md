# Keyboard Layout

sysc-greet runs inside a compositor, so keyboard layout is set there.

## niri

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

## sway

Edit `/etc/greetd/sway-greeter-config`:

```bash
input * {
    xkb_layout "de"
}
```

## hyprland

Edit `/etc/greetd/hyprland-greeter-config.conf`:

```ini
input {
    kb_layout = de
}
```

Replace `de` with your layout (`us`, `fr`, `es`, `uk`, etc). Full list in `/usr/share/X11/xkb/rules/base.lst`.

## Non-US Layouts with Kitty

If your layout doesn't work correctly in Kitty (e.g., Shift key reverts to QWERTY), set XKB environment variables in the compositor config's Kitty exec line.

*Thanks to [@morganorix](https://github.com/morganorix) for discovering this solution!*

**niri** (`/etc/greetd/niri-greeter-config.kdl`):

```kdl
spawn-sh-at-startup "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter XKB_DEFAULT_LAYOUT=fr XKB_DEFAULT_VARIANT=oss kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"
```

**sway** (`/etc/greetd/sway-greeter-config`):

```bash
exec "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter XKB_DEFAULT_LAYOUT=fr XKB_DEFAULT_VARIANT=oss kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; swaymsg exit"
```

**hyprland** (`/etc/greetd/hyprland-greeter-config.conf`):

```ini
exec-once = XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter XKB_DEFAULT_LAYOUT=fr XKB_DEFAULT_VARIANT=oss kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch exit
```

Replace `fr` with your layout and `oss` with your variant (or omit `XKB_DEFAULT_VARIANT` if not needed).

Restart greetd after changes: `sudo systemctl restart greetd`
