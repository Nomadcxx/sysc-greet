# Void Linux installer support — design

Date: 2026-10-21
Status: implemented, runtime validation on Void hardware still outstanding

## Goal

Let the Go installer (`cmd/installer`) set up sysc-greet on Void Linux, where the
package manager is `xbps` and the init system is `runit` instead of systemd.

## Facts about Void that the design depends on

These were read from the official package index and the packages themselves
(`repo-default.voidlinux.org/current`, `greetd-0.10.3_2`, `runit-void-20250212_4`,
`runit-2.3.1_2`):

- `greetd` ships `/etc/sv/greetd/{run,supervise}` — the runit service lives in
  `/etc/sv`.
- The live runit service tree is `/etc/runit/runsvdir/default`, with `current`
  pointing at it. `/var/service` is the upstream runit package's symlink to
  `/var/run/runit/runsvdir/current`, not the tree Void populates.
- `runit-void` enables `/etc/sv/agetty-tty1` … `agetty-tty6` in that tree
  unless the service directory carries a `down` file.
- The `greetd` package's install script creates the `_greeter` system account
  with the home `/var/lib/_greeter`, the group `_greeter` and the `video`
  group. Its own `/etc/greetd/config.toml` greets `user = "_greeter"` on
  `vt = 7`.
- Package names in the Void repos: `greetd`, `kitty`, `niri`, `sway`, `cage`,
  `cagebreak`, `socat`, `gstreamer1-devel`, `gst-plugins-base1-devel`,
  `gst-plugins-good1-devel`, `gst-plugins-bad1-devel`, `wayland-devel`,
  `wayland-protocols`, `mesa`, `pkg-config`, `meson`, `ninja`, `git`. There is
  no `gslapper`, no `mango` and no `hyprland`.

## Design

`cmd/installer/platform.go` owns everything Void-specific:

- `detectPlatform` records the init system, the distro id and the greeter
  account before the install tasks run. The distro id comes from the installer's
  existing `osRelease()` helper, so no second `/etc/os-release` parser exists.
- `greeterAccount` maps the distro id to the account: `_greeter` on Void,
  `greeter` everywhere else. The account is not probed on other distros, so a
  stray `_greeter` account cannot move a Debian or Arch install onto the Void
  paths.
- `greeterGroups` only lists groups the system actually has, because `useradd`
  and `usermod` fail on unknown groups.
- `substituteGreeterPaths` rewrites the hardcoded `/var/lib/greeter` paths and
  the `user = "greeter"` lines in the greeter configs.
- `liveRunitDirs` finds the service directories that exist under a root and
  follows symlinks, so the same code works for Void's tree, `/var/service` and
  `/etc/service`.
- `linkRunitService` and `unlinkRunitService` add and remove the service link
  in every live tree.
- `markServiceDown` writes the `down` file that keeps `runit-void` from
  re-enabling a service.
- `enableRunitGreetd` links the greetd service and then clears `agetty-tty1`
  from the greeter's virtual terminal, which is what Void's greetd README asks
  for.

`cmd/installer/main.go` gains the `xbps` branches: package detection, package
queries, greetd, kitty, the compositors Void packages, the GStreamer build
dependencies and the build tools. The service enable and disable tasks pick the
runit or the systemd path from the detected init system.

## Behaviour on Void

- greetd and kitty install with `xbps-install -Sy`.
- niri, sway and cagebreak install from the Void repos; mango and hyprland are
  refused with a clear message.
- gSlapper is built from source against the GStreamer development packages.
- The greeter session runs as `_greeter` in `/var/lib/_greeter`, and the greeter
  configs are rewritten to match.
- The greetd service is linked into `/etc/runit/runsvdir/default`, and
  `agetty-tty1` is removed from that tree and marked down.

## Testing

`cmd/installer/platform_test.go` covers the account mapping, the path
substitution, the xbps argument builders and the runit tree handling against a
temporary root. `scripts/verify-void-installer.sh` runs the static checks.

## Open item

The installer has not been run on Void hardware yet. The remaining checks are
the ones in the pull request: a fresh Void VM, the greeter on vt 1 after a
reboot, and a login through the cagebreak backend.
