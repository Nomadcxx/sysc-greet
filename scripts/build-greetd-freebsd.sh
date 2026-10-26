#!/bin/sh
set -eu

[ "$(uname -s)" = FreeBSD ] || { echo 'Build this prerequisite on FreeBSD.' >&2; exit 1; }
source_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
output=$(mkdir -p "${1:-dist/freebsd-greetd}" && cd "${1:-dist/freebsd-greetd}" && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/greetd-freebsd.XXXXXXXX")
trap 'rm -rf "$work"' EXIT
CARGO_BUILD_JOBS=${CARGO_BUILD_JOBS:-2}
export CARGO_BUILD_JOBS
revision=d5f27c3a8b1758da9e45608d2cff20303f21c2ff

fetch -o "$work/greetd.tar.gz" "https://github.com/kennylevinsen/greetd/archive/$revision.tar.gz"
sha256 -c ff1ebe00a04b9d39c2f2deab66c40a16c7492c4d8674a74ea4bb518978ebfe12 "$work/greetd.tar.gz"
fetch -o "$work/pam-sys.crate" https://static.crates.io/crates/pam-sys/pam-sys-0.5.6.crate
sha256 -c cd4858311a097f01a0006ef7d0cd50bca81ec430c949d7bf95cbefd202282434 "$work/pam-sys.crate"
mkdir "$work/greetd" "$work/pam-sys"
tar -xzf "$work/greetd.tar.gz" -C "$work/greetd" --strip-components 1
tar -xzf "$work/pam-sys.crate" -C "$work/pam-sys" --strip-components 1
patch -d "$work/greetd" -p1 < "$source_dir/packaging/freebsd/greetd.patch"
patch -d "$work/pam-sys" -p1 < "$source_dir/packaging/freebsd/pam-sys.patch"
cat >> "$work/greetd/Cargo.toml" <<'TOML'

[patch.crates-io]
pam-sys = { path = "../pam-sys" }
TOML
cd "$work/greetd"
# Resolving the local binding changes only its registry source in Cargo.lock.
cargo test -p pam-sys --test freebsd_abi
cargo test --workspace --locked
cargo build --release --locked -p greetd -p agreety
# CARGO_TARGET_DIR may point to a shared build cache.
target_dir=${CARGO_TARGET_DIR:-$work/greetd/target}
install -m 755 "$target_dir/release/greetd" "$target_dir/release/agreety" "$output/"
printf 'Built native greetd and agreety in %s\n' "$output"
