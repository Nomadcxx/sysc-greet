#!/bin/sh
set -eu
repo=$(cd "$1" && pwd)
root=$(mktemp -d /tmp/sysc-greet-check.XXXXXXXX)
arch=$(xbps-uhelper arch)
export XBPS_ARCH="$arch"
set -- "$repo"/sysc-greet-*."$arch".xbps
test "$#" -eq 1 && test -f "$1"
expected=${1##*/}
expected=${expected%."$arch".xbps}
mirror=https://repo-default.voidlinux.org/current
[ "$arch" != x86_64-musl ] || mirror=$mirror/musl
mkdir -p "$root/var/db/xbps/keys"
cp /var/db/xbps/keys/* "$root/var/db/xbps/keys/"
xbps-install -r "$root" -R "$mirror" -Sy base-files greetd
config_before=$(sha256sum "$root/etc/greetd/config.toml")
xbps-install -r "$root" -R "$mirror" -R "$repo" -Sy sysc-greet
xbps-pkgdb -r "$root" sysc-greet
test "$(xbps-query -r "$root" -p pkgver sysc-greet)" = "$expected"
test "$(xbps-query -r "$root" -p architecture sysc-greet)" = "$arch"
xbps-pkgdb -r "$root" gslapper
test "$(sha256sum "$root/etc/greetd/config.toml")" = "$config_before"
test ! -e "$root/var/service/greetd"
test ! -L "$root/var/service/greetd"
test "$(xbps-query -r "$root" -p license sysc-greet)" = GPL-3.0-only
test "$(stat -c %u:%g "$root/usr/bin/sysc-greet")" = 0:0
grep -q '^_greeter:' "$root/etc/passwd"
for directory in var/cache/sysc-greet var/lib/_greeter/Pictures/wallpapers; do
    test "$(stat -c %u "$root/$directory")" = "$(awk -F: '$1 == "_greeter" {print $3}' "$root/etc/passwd")"
    test "$(stat -c %g "$root/$directory")" = "$(awk -F: '$1 == "_greeter" {print $4}' "$root/etc/passwd")"
done
test -s "$root/etc/greetd/sysc-greet/kitty.conf"
test -s "$root/usr/share/sysc-greet/THIRD_PARTY_NOTICES.md"
test -s "$root/usr/share/licenses/sysc-greet/LICENSE"
test ! -e "$root/usr/bin/go"
test ! -e "$root/usr/bin/gcc"
chroot "$root" /usr/bin/sysc-greet --version
printf 'Verified native %s sysc-greet package installation.\n' "$arch"
