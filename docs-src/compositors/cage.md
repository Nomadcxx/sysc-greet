# Cage Setup

!!! tip "Recommended greeter backend"
    Cage is the **recommended** Wayland backend for the sysc-greet greeter session. It replaces Hyprland for greetd, which is being [phased out over the next ~3 months](hyprland.md).

!!! warning "Cage Lite — no boot wallpapers"
    Cage is a **single-client kiosk** backend. Unlike niri, Hyprland, and sway, it cannot run gSlapper as a second Wayland client.

    **What you lose:** boot static/video wallpapers, in-greeter wallpaper changes via gSlapper  
    **What you keep:** login, themes, TUI background effects, ASCII art, session selection

See the [design spec](https://github.com/Nomadcxx/sysc-greet/blob/master/docs/superpowers/specs/2026-06-18-cage-compositor-design.md) for the full investigation.

## Why Cage?

- **Faster, simpler greeter session** — no Hyprland config, window rules, layer rules, or `hyprctl` exit chain
- **Fewer moving parts** — one compositor binary, one launcher script
- **Good fit for kiosk-style greeters** — fullscreen Kitty is exactly what Cage is designed for

Tracked in [issue #69](https://github.com/Nomadcxx/sysc-greet/issues/69).

## Install Cage

=== "Arch Linux"

```bash
sudo pacman -S cage
```

=== "NixOS"

```nix
services.sysc-greet = {
  enable = true;
  compositor = "cage";
};
```

=== "Other distros"

Build from [cage-kiosk/cage](https://github.com/cage-kiosk/cage) or check your package manager.

## greetd Config

=== "Installer"

```bash
SYSC_COMPOSITOR=cage sudo ./install.sh
```

=== "Manual"

Edit `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "cage -s -m extend -- /etc/greetd/cage-greeter-session.sh"
user = "greeter"

[initial_session]
command = "cage -s -m extend -- /etc/greetd/cage-greeter-session.sh"
user = "greeter"
```

Install the launcher script from `scripts/cage-greeter-session.sh` to `/etc/greetd/cage-greeter-session.sh` (mode `755`).

### Flags

| Flag | Purpose |
|------|---------|
| `-s` | Allow VT switching |
| `-m extend` | Span the greeter across all connected outputs |

Cage exits automatically when Kitty (its sole client) exits after a successful login.

## Launcher Script

`/etc/greetd/cage-greeter-session.sh`:

```sh
#!/bin/sh
export XDG_CACHE_HOME=/tmp/greeter-cache
export HOME=/var/lib/greeter
exec kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet
```

## Keyboard Layout

Cage has **no compositor keyboard configuration**. Set XKB environment variables on the Kitty line (in the launcher script):

```sh
export XKB_DEFAULT_LAYOUT=fr
export XKB_DEFAULT_VARIANT=oss   # if needed
exec kitty ...
```

See [Keyboard Layout](../configuration/keyboard-layout.md) for layout codes.

## Feature Comparison

| Feature | niri / Hyprland / sway | Cage Lite |
|---------|------------------------|-----------|
| Login via greetd | ✅ | ✅ |
| TUI background effects | ✅ | ✅ |
| Theme colors / ASCII | ✅ | ✅ |
| Boot static wallpaper | ✅ | ❌ |
| Video wallpaper | ✅ | ❌ |
| gSlapper layer-shell | ✅ | ❌ |
| Compositor config file | Required | None |
| Hyprland maintenance | Ongoing | N/A |

## Verification

```bash
# Static checks (from repo root)
./scripts/verify-cage-greeter.sh

# Restart greetd (from a TTY)
sudo systemctl restart greetd
journalctl -u greetd -b --no-pager | tail -30
```

## Cagebreak (future)

[Cagebreak](https://github.com/project-repo/cagebreak) is a tiling compositor forked from Cage. It may support multiple clients and wlr-layer-shell, which would restore wallpaper parity. This is **not yet implemented** — see the implementation plan for the Phase 2 spike.

## Migrating from Hyprland

If you currently use Hyprland for the greetd greeter session:

```bash
sudo pacman -S cage   # or your distro's cage package
SYSC_COMPOSITOR=cage sudo ./install.sh
sudo systemctl restart greetd
```

Your Hyprland **desktop session** is unchanged — only the boot greeter moves to Cage.

**Trade-off:** Cage Lite has no gSlapper boot wallpapers. Use TUI background effects (F3) instead, or switch to niri if you need video/static wallpapers at the greeter.

## Troubleshooting

**Greeter does not appear**

- Confirm `cage` is in PATH for the greeter user
- Check `journalctl -u greetd` for cage startup errors

**Password rejected with non-US keyboard**

- Add `XKB_DEFAULT_LAYOUT` (and variant if needed) to the launcher script

**Expected wallpapers missing**

- This is by design in Cage Lite mode. Use TUI background effects (F3) or switch to niri/Hyprland/sway for gSlapper wallpapers.
