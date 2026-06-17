#!/bin/sh
# cage-greeter-session.sh — greetd session launcher for Cage (lite mode)
#
# Cage is a single-client kiosk compositor. This script is the ONE application
# cage runs. It execs kitty → sysc-greet directly (no gSlapper wallpaper daemon).
#
# See docs-src/compositors/cage.md and docs/superpowers/specs/2026-06-18-cage-compositor-design.md

set -eu

export XDG_CACHE_HOME="${XDG_CACHE_HOME:-/tmp/greeter-cache}"
export HOME="${HOME:-/var/lib/greeter}"

KITTY_CONFIG="${KITTY_CONFIG:-/etc/greetd/kitty.conf}"
SYSC_GREET_BIN="${SYSC_GREET_BIN:-/usr/local/bin/sysc-greet}"

# XKB_DEFAULT_LAYOUT / XKB_DEFAULT_VARIANT may be set by installer or user.
# Cage has no compositor keyboard config — Kitty env vars are the only path.

exec kitty --start-as=fullscreen --config="${KITTY_CONFIG}" "${SYSC_GREET_BIN}"
