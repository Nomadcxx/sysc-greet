---
title: "Mango Setup"
description: "Mango is a dwl-based tiling Wayland compositor. sysc-greet supports it as a greeter backend on Arch Linux."
---

Mango is a dwl-based tiling Wayland compositor. sysc-greet supports it as a greeter backend on Arch Linux, where `mangowm` is in the `extra` repo. Other distros have no mango package; use [niri](niri), [cagebreak](cagebreak), or [sway](sway). Niri remains the default greeter compositor, and cagebreak remains the replacement for the deprecated [Hyprland](hyprland) greeter.

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

=== "Other distros"

    Not supported. The installer refuses `SYSC_COMPOSITOR=mango` outside Arch.

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
- **Greeter restarts in a loop:** run `mango -c /etc/greetd/mango-greeter-config.conf -p` to check the config. Unknown keys fail the parse.
- **Compositor keybindings work on the login screen:** the greeter config contains `bind=` lines or sources `/etc/mango/config.conf`. Remove them.
- **Testing:** don't start mango from inside a logged-in graphical session. On exit it stops `graphical-session.target`, which ends that session. Test from a TTY or a spare machine.
