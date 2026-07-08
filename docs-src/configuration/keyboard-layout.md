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

## cagebreak

Cagebreak has no layout directive; it reads the standard XKB environment variables. Set them on the greetd command in `/etc/greetd/config.toml`:

```toml
command = "env XKB_DEFAULT_LAYOUT=de cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
```

This also covers Kitty — processes launched by cagebreak inherit the environment, so no separate exec-line change is needed for non-US layouts or Dvorak (add `XKB_DEFAULT_VARIANT=...` to the same command).

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

**cagebreak**: not affected — the XKB variables set on the greetd command (see [cagebreak](#cagebreak) above) are inherited by Kitty.

Replace `fr` with your layout and `oss` with your variant (or omit `XKB_DEFAULT_VARIANT` if not needed).

Restart greetd after changes: `sudo systemctl restart greetd`

## Dvorak

Dvorak is a common case where users find their password is rejected at the greeter even though the compositor layout is set correctly. This happens because Kitty (the terminal running sysc-greet) needs XKB environment variables set explicitly — it does not inherit the compositor's keyboard config.

Set `XKB_DEFAULT_LAYOUT=us XKB_DEFAULT_VARIANT=dvorak` in the Kitty exec line for your compositor:

**niri** (`/etc/greetd/niri-greeter-config.kdl`):

```kdl
spawn-sh-at-startup "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter XKB_DEFAULT_LAYOUT=us XKB_DEFAULT_VARIANT=dvorak kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"
```

**sway** (`/etc/greetd/sway-greeter-config`):

```bash
exec "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter XKB_DEFAULT_LAYOUT=us XKB_DEFAULT_VARIANT=dvorak kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; swaymsg exit"
```

**hyprland** (`/etc/greetd/hyprland-greeter-config.conf`):

```ini
exec-once = XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter XKB_DEFAULT_LAYOUT=us XKB_DEFAULT_VARIANT=dvorak kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch exit
```

**cagebreak** (`/etc/greetd/config.toml`):

```toml
command = "env XKB_DEFAULT_LAYOUT=us XKB_DEFAULT_VARIANT=dvorak cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
```

The same pattern applies to other Dvorak-family layouts (Dvorak Programmer, etc.) — just change `XKB_DEFAULT_VARIANT` accordingly. Run `localectl list-x11-keymap-variants us` to see all available US variants.
