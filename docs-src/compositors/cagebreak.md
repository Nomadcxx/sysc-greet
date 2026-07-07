# Cagebreak Setup

Cagebreak is a tiling kiosk compositor forked from Cage. It replaces Hyprland for the greetd greeter session, which is [deprecated](hyprland.md). Niri remains the default greeter compositor.

Cagebreak supports wlr-layer-shell, so gSlapper wallpapers work the same as under niri and sway. Wallpapers appear on secondary monitors; the primary output is covered by the greeter TUI, which draws its own background effects.

The greeter config is ~10 lines, defines no keybindings, and quits the compositor through cagebreak's IPC socket after login. Tracked in [issue #69](https://github.com/Nomadcxx/sysc-greet/issues/69).

## Install Cagebreak

The installer handles all of this — including socat, which is required to quit the compositor after login. The sections below are for manual installs.

=== "Arch Linux"

    Cagebreak is in the AUR, not the official repos. The installer uses paru or yay if present; manually:

    ```bash
    paru -S cagebreak   # or: yay -S cagebreak
    sudo pacman -S socat
    ```

=== "Debian / Ubuntu / Fedora"

    No distro packages exist. The installer downloads a cagebreak package built against your distro's wlroots from [sysc-greet releases](https://github.com/Nomadcxx/sysc-greet/releases/latest) (Debian 13, Ubuntu 24.04/25.x, Fedora 42), and builds from source when no package matches. Manual install:

    ```bash
    # Debian 13 / Ubuntu 25.x
    wget https://github.com/Nomadcxx/sysc-greet/releases/latest/download/cagebreak_2.4.0_debian13_amd64.deb
    sudo apt install ./cagebreak_2.4.0_debian13_amd64.deb socat
    ```

=== "NixOS"

    ```nix
    services.sysc-greet = {
      enable = true;
      compositor = "cagebreak";
      # optional: cagebreakPackage = pkgs.cagebreak;
    };
    ```

=== "openSUSE"

    Tumbleweed packages cagebreak:

    ```bash
    sudo zypper install cagebreak socat
    ```

=== "Other distros"

    Check your package manager or build from [project-repo/cagebreak](https://github.com/project-repo/cagebreak). The buildable release depends on your wlroots version: 0.17 → 2.3.1, 0.18 → 2.4.0, 0.19 → 3.1.0, 0.20 → 3.2.1.

## greetd Config

=== "Installer"

    ```bash
    SYSC_COMPOSITOR=cagebreak sudo ./install.sh
    ```

=== "Manual"

    Edit `/etc/greetd/config.toml`:

    ```toml
    [terminal]
    vt = 1

    [default_session]
    command = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
    user = "greeter"
    ```

    `-e` enables the IPC socket. The greeter config uses it to quit cagebreak after login.

The installer and packages place the greeter config at `/etc/greetd/cagebreak-greeter-config`.

## Keyboard Layout

Cagebreak reads the standard XKB environment variables. For a non-US layout, wrap the greetd command:

```toml
command = "env XKB_DEFAULT_LAYOUT=de XKB_DEFAULT_VARIANT=nodeadkeys cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
```

## Migrating from Hyprland

1. Install cagebreak and socat (see above)
2. Re-run the installer: `SYSC_COMPOSITOR=cagebreak sudo ./install.sh`
3. Reboot, or `sudo systemctl restart greetd` from a TTY

Your Hyprland desktop session is unaffected. Only the boot greeter changes.

## Troubleshooting

- **No wallpaper on a secondary monitor:** check gSlapper is installed and `/var/cache/sysc-greet/` is readable by the greeter user. Single-monitor setups never show a wallpaper; the TUI covers the only output.
- **Greeter restarts in a loop:** run `cagebreak -e -c /etc/greetd/cagebreak-greeter-config` from a TTY to see the error. A config parse error makes cagebreak exit immediately and greetd relaunches it.
- **Stuck after login:** verify socat is installed. Without it the compositor never quits and greetd waits forever.
