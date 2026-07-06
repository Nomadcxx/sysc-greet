# Cagebreak Setup

Cagebreak is a tiling kiosk compositor forked from Cage. It replaces Hyprland for the greetd greeter session, which is [deprecated](hyprland.md). Niri remains the default greeter compositor.

Cagebreak supports wlr-layer-shell, so gSlapper wallpapers work the same as under niri and sway. Wallpapers appear on secondary monitors; the primary output is covered by the greeter TUI, which draws its own background effects.

The greeter config is ~10 lines, defines no keybindings, and quits the compositor through cagebreak's IPC socket after login. Tracked in [issue #69](https://github.com/Nomadcxx/sysc-greet/issues/69).

## Install Cagebreak

=== "Arch Linux"

    Cagebreak is in the AUR, not the official repos. socat is required to quit the compositor after login.

    ```bash
    paru -S cagebreak   # or: yay -S cagebreak
    sudo pacman -S socat
    ```

=== "NixOS"

    ```nix
    services.sysc-greet = {
      enable = true;
      compositor = "cagebreak";
      # optional: cagebreakPackage = pkgs.cagebreak;
    };
    ```

=== "Other distros"

    Check your package manager or build from [project-repo/cagebreak](https://github.com/project-repo/cagebreak).

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
