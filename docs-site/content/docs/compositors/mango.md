---
title: "Mango Setup"
description: "Mango is a dwl-based tiling Wayland compositor. sysc-greet supports it as a greeter backend on Arch, Debian 13, Ubuntu 26.04, and Fedora 43/44."
---

Mango is a dwl-based tiling Wayland compositor. sysc-greet supports it as a greeter backend on Arch Linux, where `mangowm` is in the `extra` repo, and on Debian 13, Ubuntu 26.04, and Fedora 43/44 through `mangowm` packages attached to [sysc-greet releases](https://github.com/Nomadcxx/sysc-greet/releases/latest). Other distros: use [niri](niri), [cagebreak](cagebreak), or [sway](sway). Niri remains the default greeter compositor, and cagebreak remains the replacement for the deprecated [Hyprland](hyprland) greeter.

Mango supports wlr-layer-shell, so gSlapper wallpapers work the same as under niri and sway: they appear on secondary monitors, and the greeter TUI covers the primary output.

The greeter config defines no keybindings. Mango only adds its built-in `Ctrl+Alt+F1`–`F12` VT switching. A session script starts the greeter and quits mango with `mmsg` after login.

## Install Mango

=== "Arch Linux"

    The [sysc-greet-mango](https://aur.archlinux.org/packages/sysc-greet-mango) AUR package installs sysc-greet, mangowm, and the greetd config:

    ```bash
    paru -S sysc-greet-mango   # or: yay -S sysc-greet-mango
    ```

    It conflicts with the other sysc-greet variants; remove those first. With the install script instead:

    ```bash
    sudo SYSC_COMPOSITOR=mango ./install.sh
    ```

    The installer runs `pacman -S mangowm` if mango is missing.

=== "NixOS"

    Mango is not in nixpkgs. Pass a package that provides `mango` and `mmsg` from your own flake:

    ```nix
    services.sysc-greet = {
      enable = true;
      compositor = "mango";
      mangoPackage = inputs.mango.packages.${pkgs.stdenv.hostPlatform.system}.default;
    };
    ```

=== "Debian / Ubuntu / Fedora"

    These distros do not package mango, so each sysc-greet release carries a `mangowm` package built for them. `sudo SYSC_COMPOSITOR=mango ./install.sh` downloads and installs the right one. Manual install:

    ```bash
    # Debian 13
    wget https://github.com/Nomadcxx/sysc-greet/releases/latest/download/mangowm_0.17.4_debian13_amd64.deb
    sudo apt install ./mangowm_0.17.4_debian13_amd64.deb

    # Ubuntu 26.04
    wget https://github.com/Nomadcxx/sysc-greet/releases/latest/download/mangowm_0.17.4_ubuntu26.04_amd64.deb
    sudo apt install ./mangowm_0.17.4_ubuntu26.04_amd64.deb

    # Fedora 43 or 44 (replace 44 with your release)
    sudo dnf install https://github.com/Nomadcxx/sysc-greet/releases/latest/download/mangowm-0.17.4-1.fedora44.x86_64.rpm
    ```

    Fedora 44 links the distro's wlroots 0.20 and scenefx 0.5. Fedora 43 and Ubuntu 26.04 have older wlroots, so their packages carry wlroots 0.20 and scenefx 0.5 in `/usr/lib/sysc-greet-mango`, used only by mango. Debian 13 also needs newer libwayland, libdrm, pixman, and xkbcommon, which are bundled the same way; those copies do not receive Debian security updates until a new sysc-greet release rebuilds them.

=== "Other distros"

    Not supported: Ubuntu 24.04 and older releases lack the libraries mango needs. The installer refuses `SYSC_COMPOSITOR=mango` there.

## greetd Config

The AUR package and the installer write this config. Manual setup, in `/etc/greetd/config.toml`:

```toml
[terminal]
vt = 1

[default_session]
command = "mango -c /etc/greetd/mango-greeter-config.conf -s /etc/greetd/mango-greeter-session.sh"
user = "greeter"
```

| File | Purpose |
|---|---|
| `/etc/greetd/mango-greeter-config.conf` | Greeter compositor config. Starts the gSlapper wallpaper daemon. |
| `/etc/greetd/mango-greeter-session.sh` | Runs kitty with sysc-greet, then `mmsg dispatch quit`. Must be executable. |

`-c` makes mango read only the greeter config. Never `source=` `/etc/mango/config.conf` from it: that file binds `Alt+Return` to a terminal and `Super+M` to quit.

## Keyboard Layout

Set the layout in `/etc/greetd/mango-greeter-config.conf`:

```ini
xkb_rules_layout=de
xkb_rules_variant=nodeadkeys
```

## Troubleshooting

- **Stuck after login:** `mmsg` is missing or `mango-greeter-session.sh` is not executable. Both `mango` and `mmsg` come from `mangowm`.
- **mango crashes when the greeter exits:** Xwayland is missing. mango 0.17.4 segfaults on exit without it; the `mangowm` packages depend on it.
- **Greeter restarts in a loop:** run `mango -c /etc/greetd/mango-greeter-config.conf -p` to check the config. Unknown keys fail the parse.
- **Compositor keybindings work on the login screen:** the greeter config contains `bind=` lines or sources `/etc/mango/config.conf`. Remove them.
- **Testing:** don't start mango from inside a logged-in graphical session. On exit it stops `graphical-session.target`, which ends that session. Test from a TTY or a spare machine.
