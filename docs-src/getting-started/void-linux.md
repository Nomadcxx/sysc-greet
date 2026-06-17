# Void Linux Installation

sysc-greet is installable on **Void Linux** through the **Go installer** (`cmd/installer`), with **runit** service management and **xbps** packages.

## Quick install

```bash
git clone https://github.com/Nomadcxx/sysc-greet.git
cd sysc-greet
go run ./cmd/installer/
```

Or via the one-line wrapper (requires Go):

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

Pre-select a Wayland backend:

```bash
SYSC_COMPOSITOR=cage sudo ./install.sh    # recommended on Void
SYSC_COMPOSITOR=niri sudo ./install.sh    # builds gSlapper from source
```

## What the installer does on Void

| Step | Void behaviour |
|------|----------------|
| Package manager | `xbps-install -Sy` |
| greetd user | **`_greeter`** (from Void's greetd package) |
| Greeter home | `/var/lib/_greeter` |
| Service enable | `ln -s /etc/sv/greetd /var/service/` |
| VT conflict | Disables `agetty-tty1` if present (greetd uses vt=1) |
| cage backend | Skips gSlapper (TUI backgrounds only) |
| niri/sway | Builds gSlapper from source (not in Void repos) |

## Recommended backend on Void

**cage** is the simplest path — greetd, cage, and kitty are all in official Void repos with no source builds.

**niri** works and supports full wallpapers, but the installer must compile gSlapper and GStreamer build deps.

**hyprland** is not available in Void repos via xbps; the installer will refuse it.

## niri on Void — important note

Do **not** add `niri --session` to greeter or desktop configs on Void ([niri#2448](https://github.com/niri-wm/niri/issues/2448)). sysc-greet's niri greeter config uses `niri -c /etc/greetd/niri-greeter-config.kdl` only.

## Manual verification after install

```bash
ls -la /var/service/greetd          # should symlink to /etc/sv/greetd
cat /etc/greetd/config.toml
sv status greetd                    # if runit-sv is installed
```

Reboot from a TTY to test the greeter (not over SSH).

## Troubleshooting

**Installer says missing init**

Void must use runit (`/etc/sv` present). Install `runit` if missing.

**Black screen / no greeter**

- Check `agetty-tty1` is not still enabled on vt 1
- `journalctl` is unavailable on Void — use `sv log greetd` or `/var/log/sv/greetd/current`

**gSlapper build fails on niri**

Install build deps manually: `xbps-install -Sy meson ninja git pkg-config gstreamer1-devel gst-plugins-base1-devel wayland-devel`

Or switch to cage: re-run installer with `SYSC_COMPOSITOR=cage`.

## See also

- [Cage backend](../compositors/cage.md)
- [Niri backend](../compositors/niri.md)
- [Investigation spec](../../superpowers/specs/2026-06-18-issue-53-void-chimera-bsd-design.md)
