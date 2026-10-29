---
title: "Void Linux"
description: "Install sysc-greet on Void Linux with xbps packages and runit service management."
---

Install sysc-greet on **Void Linux** with native XBPS packages or the Go installer. Both use **runit**, elogind, D-Bus and polkit.

The installer and native packages passed Void glibc x86_64 VM tests for desktop login, F4 reboot and shutdown. Both glibc and musl packages passed build and package checks; musl has no separate graphical login test.

## Native XBPS packages

Download sysc-greet 1.3.2 and gSlapper 1.5.3 for your architecture. These upstream release packages work while [Void's package submission](https://github.com/void-linux/void-packages/pull/63202) awaits review. You do not need Go or a compiler.

If you used the Go installer, run its Uninstall option before switching to XBPS packages.

### Download and install

Run this on an x86_64 glibc or musl system. `xbps-uhelper arch` selects the matching packages:

```bash
sudo xbps-install -S curl
mkdir -p sysc-greet-void
cd sysc-greet-void
arch=$(xbps-uhelper arch)
curl -fLO "https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.2/sysc-greet-1.3.2_1.${arch}.xbps"
curl -fLO "https://github.com/Nomadcxx/gSlapper/releases/download/v1.5.3/gslapper-1.5.3_1.${arch}.xbps"
curl -fL -o sysc-greet-SHA256SUMS-void https://github.com/Nomadcxx/sysc-greet/releases/download/v1.3.2/SHA256SUMS-void
curl -fL -o gslapper-SHA256SUMS-void https://github.com/Nomadcxx/gSlapper/releases/download/v1.5.3/SHA256SUMS-void
sha256sum --ignore-missing -c sysc-greet-SHA256SUMS-void &&
  sha256sum --ignore-missing -c gslapper-SHA256SUMS-void &&
  xbps-rindex -a ./*.xbps &&
  sudo xbps-install --repository "$PWD" sysc-greet
```

XBPS installs the runtime dependencies from Void's repositories. Release CI builds and install-checks both libc variants. New releases include a source archive with the build recipe and `SHA256SUMS-void`. Future updates require downloading and installing the new release packages.

### Configure greetd

The package installs `/usr/bin/sysc-greet` and configs under `/etc/greetd/sysc-greet/`. It leaves `/etc/greetd/config.toml` and service activation to you.

Install a supported compositor, graphics drivers for your hardware, and a desktop session. For example, cagebreak can host the greeter while niri provides the desktop session. Mesa systems also need `mesa-dri`.

```bash
sudo xbps-install -S cagebreak niri
sudo cp -an /etc/greetd/config.toml /etc/greetd/config.toml.before-sysc-greet
```

Set `/etc/greetd/config.toml` to:

```toml
[terminal]
vt = 1

[default_session]
command = "cagebreak -e -c /etc/greetd/sysc-greet/cagebreak-greeter-config"
user = "_greeter"
```

Keep SSH or another TTY available during setup. Disable your existing display manager and any agetty on the selected VT before enabling greetd. For an enabled `agetty-tty1`:

```bash
sudo touch /etc/sv/agetty-tty1/down
sudo sv down agetty-tty1
sudo unlink /var/service/agetty-tty1
```

Enable the `dbus` and `greetd` runit services if they are not enabled:

```bash
sudo ln -s /etc/sv/dbus /var/service/
sudo ln -s /etc/sv/greetd /var/service/
```

D-Bus activates elogind; leave elogind's runit service disabled. Reboot after first installing elogind to apply its udev seat rules, then verify login and power actions. See the [Void handbook](https://docs.voidlinux.org/config/services/index.html) for service management.

To remove the package, use `sudo xbps-remove sysc-greet`. Manage the greetd service separately if you switch greeters.

## Go installer

If you installed sysc-greet through XBPS, manage it with XBPS. To switch to the installer, first remove the package with `sudo xbps-remove sysc-greet`. The installer checks for it before installing or uninstalling.

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

Each backend uses gSlapper for wallpaper. The XBPS route installs the gSlapper release package; the Go installer compiles it from source.

Install a desktop session before logging in. Void's cagebreak package does not ship a session `.desktop` file, so installing the cagebreak greeter alone leaves the session list empty. A packaged desktop such as niri provides one.

**mango** and **hyprland** are not in the Void repos; the installer refuses them.

## Existing seatd / turnstile setups

The tested installer route uses elogind for sessions, seat access and power actions. Power commands for seatd/turnstile/acpid setups are not configurable. A failed power request returns to the login screen with the error.

If you keep turnstile alongside elogind, set `manage_rundir = no` in `/etc/turnstile/turnstiled.conf` before using this route, as the [Void handbook](https://docs.voidlinux.org/config/session-management.html#turnstile) requires. This leaves elogind in charge of `XDG_RUNTIME_DIR`.

## niri on Void — important note

With XBPS packages, use `niri -c /etc/greetd/sysc-greet/niri-greeter-config.kdl`; the Go installer uses `niri -c /etc/greetd/niri-greeter-config.kdl`. Each uses a separate config from your desktop and does not need `--session`. Void's niri 26.04 desktop entry uses `niri --session`; login handoff to that session has passed VM validation.

## Manual verification after install

```bash
ls -l /etc/runit/runsvdir/default/{dbus,greetd}
cat /etc/greetd/config.toml                 # user = "_greeter", vt = 1
id _greeter                                 # the account from Void's greetd package
sv status dbus greetd
loginctl list-sessions                       # _greeter on seat0, tty1
# XBPS package (Go installer: /etc/polkit-1/rules.d/85-greeter.rules)
cat /etc/polkit-1/rules.d/85-sysc-greet.rules  # subject.user == "_greeter"
test ! -e /etc/runit/runsvdir/default/agetty-tty1
```

Reboot, confirm the greeter appears on tty1, then log in to your desktop. Log out and test F4 reboot and shutdown. Unit tests and source checks cannot prove seat access or power actions.

## Recovery and VT switching

The cagebreak VT-switching report with seatd/turnstile in [#130](https://github.com/Nomadcxx/sysc-greet/issues/130) remains open. The clean VM checks above cover the elogind route.

From SSH or a working TTY, stop the greeter before repairing its configuration:

```bash
sudo sv down greetd
# Edit the configuration, then restart:
sudo sv up greetd
```

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
