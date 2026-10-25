---
title: "Void Linux"
description: "Install sysc-greet on Void Linux with xbps packages and runit service management."
---

sysc-greet installs on **Void Linux** through the Go installer (`cmd/installer`), with **xbps** packages and **runit** service management.

## Quick install

```bash
sudo xbps-install -Sy go git
git clone https://github.com/Nomadcxx/sysc-greet.git
cd sysc-greet
sudo go run ./cmd/installer/
```

Or via the one-line wrapper (requires Go):

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

Pre-select a Wayland backend:

```bash
sudo SYSC_COMPOSITOR=cagebreak go run ./cmd/installer/
sudo SYSC_COMPOSITOR=niri go run ./cmd/installer/
```

## What the installer does on Void

| Step | Void behaviour |
|------|----------------|
| Package manager | `xbps-install -Sy` |
| greetd account | **`_greeter`**, created by Void's greetd package |
| Greeter home | `/var/lib/_greeter` |
| Seat and power | installs elogind, D-Bus and polkit; D-Bus activates elogind when PAM opens a session |
| Service enable | links `dbus` and `greetd` into `/etc/runit/runsvdir/default` |
| VT conflict | removes `agetty-tty1` and writes its `down` file (greetd uses vt=1) |
| Kitty font | installs `dejavu-fonts-ttf`, which kitty does not pull in on a minimal host |
| gSlapper | builds from source for each backend; installs GCC and the GStreamer development packages |
| Power menu | calls `loginctl --no-ask-password reboot` or `poweroff`; polkit allows `_greeter` |

## Backends on Void

**cagebreak**, **niri** and **sway** come from the official Void repos. The installer also installs socat for cagebreak so it can exit after login.

Each backend uses gSlapper for wallpaper. Void does not package gSlapper, so the installer compiles it from source.

Install a desktop session before logging in. Void's cagebreak package does not ship a session `.desktop` file, so installing the cagebreak greeter alone leaves the session list empty. A packaged desktop such as niri provides one.

**mango** and **hyprland** are not in the Void repos; the installer refuses them.

## niri on Void — important note

Keep the greeter command as `niri -c /etc/greetd/niri-greeter-config.kdl`. It uses a separate config from your desktop and does not need `--session`. Void's niri 26.04 desktop entry uses `niri --session`; login handoff to that session has passed VM validation.

## Manual verification after install

```bash
ls -l /etc/runit/runsvdir/default/{dbus,greetd}
cat /etc/greetd/config.toml                 # user = "_greeter", vt = 1
id _greeter                                 # the account from Void's greetd package
sv status dbus greetd
loginctl list-sessions                       # _greeter on seat0, tty1
cat /etc/polkit-1/rules.d/85-greeter.rules     # subject.user == "_greeter"
test ! -e /etc/runit/runsvdir/default/agetty-tty1
```

Reboot, confirm the greeter appears on tty1, then log in to your desktop. Log out and test F4 reboot and shutdown. Unit tests and source checks cannot prove seat access or power actions.

## Troubleshooting

**Installer says missing init**

Void must use runit. The installer looks for `/etc/runit/runsvdir`; install `runit-void` if it is missing.

**Black screen / no greeter**

- Check `agetty-tty1` is not still enabled on vt 1
- Check `sv status dbus greetd` and `loginctl list-sessions`
- If you use `socklog-void`, read `/var/log/socklog/daemon/current` and `/var/log/socklog/secure/current`. Void does not provide `journalctl` or `sv log`.
- If kitty reports `FcFontMatch() failed`, install `dejavu-fonts-ttf`.

**gSlapper build fails**

Install build deps manually:

```bash
sudo xbps-install -Sy gcc meson ninja git pkg-config gstreamer1-devel gst-plugins-base1-devel \
  gst-plugins-good1 gst-plugins-bad1-devel wayland-devel wayland-protocols
```

See `/tmp/sysc-greet-installer.log` for the failing build command. Switching the compositor does not skip this build.

## See also

- [Cagebreak backend](../compositors/cagebreak)
- [Niri backend](../compositors/niri)
- [Installation](../getting-started/installation)
