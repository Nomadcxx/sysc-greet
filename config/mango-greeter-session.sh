#!/bin/sh
# Started by mango -s. When kitty (the greeter) exits, quit mango so greetd can start the session.
cd /var/lib/greeter || exit 1
export HOME=/var/lib/greeter
export XDG_CACHE_HOME=/var/lib/greeter/.cache
export XDG_CONFIG_HOME=/var/lib/greeter/.config
export XDG_STATE_HOME=/var/lib/greeter/.local/state

setsid /usr/local/bin/sysc-greet --wallpaper-daemon &
wallpaper_pid=$!
cleanup_wallpaper() {
    kill -TERM -- "-$wallpaper_pid" 2>/dev/null || true
    wait "$wallpaper_pid" 2>/dev/null || true
}
trap cleanup_wallpaper EXIT

kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet || true
mmsg dispatch quit
