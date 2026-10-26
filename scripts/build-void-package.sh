#!/bin/bash
set -euo pipefail

pkgname=sysc-greet
tag=${1:-}
tag=${tag#v}
if [[ ! $tag =~ ^[0-9]+(\.[0-9]+)*(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
    printf 'Invalid release version: %s\n' "$tag" >&2
    exit 1
fi
version=${tag%%+*}
version=${version//-/}
source_dir=$(cd "$(dirname "$0")/.." && pwd)
output=$(mkdir -p "${2:-dist/void}" && cd "${2:-dist/void}" && pwd)
work=${VOID_BUILD_DIR:-$(mktemp -d /tmp/sysc-greet-void.XXXXXXXX)}
repo=$work/void-packages
void_revision=69f1e62affe9690848728b523f18620054a4e63b

arch=$(xbps-uhelper arch)
case "$arch" in x86_64|x86_64-musl) ;; *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; esac
mkdir "$repo"
git init -q "$repo"
git -C "$repo" fetch -q --depth=1 https://github.com/void-linux/void-packages.git "$void_revision"
git -C "$repo" checkout -q --detach FETCH_HEAD
mkdir -p "$repo/srcpkgs/$pkgname"
cp -R "$source_dir/packaging/void/." "$repo/srcpkgs/$pkgname/"
archive=$pkgname-$version-void-source.tar.gz
git -C "$source_dir" archive --format=tar --prefix="$pkgname-$version/" HEAD | gzip -n > "$output/$archive"
checksum=$(sha256sum "$output/$archive")
checksum=${checksum%% *}
sed -i -e "s|^version=.*|version=$version|" \
    -e "s|^checksum=.*|checksum=$checksum|" \
    -e "s|^distfiles=.*|distfiles=\"https://github.com/Nomadcxx/sysc-greet/releases/download/v$tag/$archive\"|" \
    "$repo/srcpkgs/$pkgname/template"
mkdir -p "$repo/hostdir/sources/$pkgname-$version"
cp "$output/$archive" "$repo/hostdir/sources/$pkgname-$version/"
printf 'XBPS_CHROOT_CMD=uchroot\nXBPS_MAKEJOBS=2\n' > "$repo/etc/conf"
# Use the released dependency until Void carries gslapper.
dependency=gslapper-1.5.2_1.$arch.xbps
mkdir -p "$repo/hostdir/binpkgs"
(
    cd "$repo/hostdir/binpkgs"
    release=https://github.com/Nomadcxx/gSlapper/releases/download/v1.5.2
    curl --fail --location --retry 3 -o "$dependency" "$release/$dependency"
    curl --fail --location --retry 3 -o SHA256SUMS-void "$release/SHA256SUMS-void"
    awk -v file="$dependency" '$2 == file { print; found=1 } END { if (!found) exit 1 }' \
        SHA256SUMS-void > dependency.sha256
    sha256sum -c dependency.sha256
    xbps-rindex -a "$dependency"
)
cd "$repo"
./xbps-src binary-bootstrap
./xbps-src -Q pkg "$pkgname"
cp "hostdir/binpkgs/$pkgname-${version}_1.$arch.xbps" "$output/"
