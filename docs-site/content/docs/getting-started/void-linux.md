---
title: "Void Linux"
description: "Install sysc-greet on Void Linux with xbps packages and runit service management."
---

sysc-greet installs on **Void Linux** through the Go installer (`cmd/installer`), with **xbps** packages and **runit** service management.

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
SYSC_COMPOSITOR=cagebreak sudo ./install.sh   # simplest path on Void
SYSC_COMPOSITOR=niri sudo ./install.sh         # builds gSlapper from source
```

## What the installer does on Void

| Step | Void behaviour |
|------|----------------|
| Package manager | `xbps-install -Sy` |
| greetd account | **`_greeter`**, created by Void's greetd package |
| Greeter home | `/var/lib/_greeter` |
| Service enable | links `/etc/sv/greetd` into `/etc/runit/runsvdir/default` |
| VT conflict | removes `agetty-tty1` and writes its `down` file (greetd uses vt=1) |
| gSlapper | built from source, it is not in the Void repos |

## Backends on Void

**cagebreak** is the simplest path — greetd, cagebreak, kitty and socat are all in the official Void repos with no source builds.

**niri** and **sway** are packaged in Void as well. Both need gSlapper, which is not, so the installer compiles it against the GStreamer development packages.

**mango** and **hyprland** are not in the Void repos; the installer refuses them.

## niri on Void — important note

Do **not** add `niri --session` to greeter or desktop configs on Void ([niri#2448](https://github.com/niri-wm/niri/issues/2448)). sysc-greet's niri greeter config uses `niri -c /etc/greetd/niri-greeter-config.kdl` only.

## Manual verification after install

```bash
ls -la /etc/runit/runsvdir/default/greetd   # should link to /etc/sv/greetd
cat /etc/greetd/config.toml                 # user = "_greeter", vt = 1
id _greeter                                 # the account from Void's greetd package
sv status greetd                            # if runit-sv is installed
```

Reboot from a TTY to test the greeter (not over SSH).

## Troubleshooting

**Installer says missing init**

Void must use runit. The installer looks for `/etc/runit/runsvdir`; install `runit-void` if it is missing.

**Black screen / no greeter**

- Check `agetty-tty1` is not still enabled on vt 1
- `journalctl` is unavailable on Void — use `sv log greetd` or `/var/log/sv/greetd/current`

**gSlapper build fails on niri**

Install build deps manually:

```bash
xbps-install -Sy meson ninja git pkg-config gstreamer1-devel gst-plugins-base1-devel \
  gst-plugins-good1 gst-plugins-bad1-devel wayland-devel wayland-protocols
```

Or switch to cagebreak: re-run the installer with `SYSC_COMPOSITOR=cagebreak`.

## See also

- [Cagebreak backend](../compositors/cagebreak)
- [Niri backend](../compositors/niri)
- [Installation](../getting-started/installation)
