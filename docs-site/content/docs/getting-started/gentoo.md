---
title: "Gentoo"
description: "Install sysc-greet with Portage and Gentoo's native OpenRC display-manager service."
---

The Gentoo installer route uses Portage, Sway and the native OpenRC `display-manager` service. It installs elogind, D-Bus and polkit for seat access and power actions.

The target configuration is amd64 with the desktop/OpenRC profile. Gentoo systemd and minimal profiles need separate validation.

## Install

Configure graphics drivers and Wayland support for your hardware first. Kitty and Mesa need Wayland support. gSlapper needs `media-libs/gst-plugins-base` with EGL, Wayland and a GL API (`gles2` or `opengl`).

Run these commands from a root shell:

```bash
emerge --ask dev-lang/go dev-vcs/git
git clone https://github.com/Nomadcxx/sysc-greet.git
cd sysc-greet
SYSC_COMPOSITOR=sway go run ./cmd/installer/
```

Portage may compile dependencies, including Kitty's shader compiler and the multimedia libraries. Your configured binary-package settings still apply. The installer leaves keywords and USE settings to you; resolve any Portage requests through your normal configuration, then retry.

The installer can install Sway from Gentoo's main tree. Install other compositor backends separately before selecting them.

## OpenRC setup

| Setting | Gentoo value |
|---|---|
| Greeter account | `greetd`, from `acct-user/greetd` |
| Greeter home | `/var/lib/greetd` |
| Greeter VT | 7, leaving the default tty1–6 gettys in place |
| Service package | `gui-libs/display-manager-init` |
| Display-manager config | `/etc/conf.d/display-manager` with `DISPLAYMANAGER="greetd"` and `CHECKVT=7` |
| Boot services | `dbus`, `elogind` and `display-manager` in the default runlevel |

If `/etc/rc.conf` sets `DISPLAYMANAGER`, update or remove that override: Gentoo gives it precedence over the service config. Keep SSH or another TTY available while changing login services.

Reboot after installation. The Sway greeter uses `/etc/greetd/sway-greeter-config`; your desktop uses its own session configuration.

## Verify and recover

```bash
rc-status default
rc-service display-manager status
cat /etc/greetd/config.toml
cat /etc/conf.d/display-manager
id greetd
loginctl list-sessions
```

Confirm the greeter runs as `greetd` on VT 7, log in to a desktop session, then check F4 reboot and shutdown after logging out.

From SSH or another TTY, stop the greeter before repairing its configuration:

```bash
rc-service display-manager stop
# Correct the configuration, then restart:
rc-service display-manager start
```

For installer failures, read `/tmp/sysc-greet-installer.log`. For package build failures, use the Portage build log named in that output.

## See also

- [Sway backend](../compositors/sway)
- [Installation](../getting-started/installation)
