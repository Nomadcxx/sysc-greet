# Keyboard Layout

Keyboard layout is configured in the compositor, not in sysc-greet directly.

## Niri

Edit `/etc/greetd/niri-greeter-config.kdl`:

```kdl
input {
    keyboard {
        xkb {
            layout "us"  # Change to your layout
        }
    }
}
```

Available layouts are defined in `/usr/share/X11/xkb/rules/base.lst`. Common values:
- `us` - US English
- `de` - German
- `gb` - British
- `fr` - French
- `es` - Spanish

## Sway

Edit `/etc/greetd/sway-greeter-config`:

```
input * {
    xkb_layout "us"  # Change to your layout
}
```

## Hyprland

Edit `/etc/greetd/hyprland-greeter-config.conf`:

```ini
input {
    kb_layout = us  # Change to your layout
}
```

## Non-US Layouts with Kitty

If your layout doesn't work correctly in Kitty (e.g., Shift key reverts to QWERTY), set XKB environment variables in the compositor config.

**Niri example:**
```kdl
exec-once kitty --config /etc/greetd/kitty.conf --override hide_window_decorations=yes -e env XKB_DEFAULT_LAYOUT=fr XKB_DEFAULT_VARIANT=oss /usr/local/bin/sysc-greet
```

**Sway example:**
```
exec env XKB_DEFAULT_LAYOUT=fr XKB_DEFAULT_VARIANT=oss kitty --config /etc/greetd/kitty.conf -e /usr/local/bin/sysc-greet
```

**Hyprland example:**
```ini
exec-once = kitty --config /etc/greetd/kitty.conf -e env XKB_DEFAULT_LAYOUT=fr XKB_DEFAULT_VARIANT=oss /usr/local/bin/sysc-greet
```

Replace `fr` and `oss` with your layout and variant. Omit the variant line if not needed.

## Applying Changes

After modifying compositor config, restart greetd:

```bash
sudo systemctl restart greetd
```

## Verification

Test the keyboard layout at the login screen:
1. Enter username field
2. Type characters to verify correct mapping
3. Test Shift, Alt, and special keys
