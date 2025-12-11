#!/bin/bash
# Script to prepare AUR release with correct GitHub tarball checksums

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.9"
    exit 1
fi

VERSION="$1"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Preparing release v${VERSION}"

# Ensure tag exists
if ! git rev-parse "v${VERSION}" >/dev/null 2>&1; then
    echo "Error: Tag v${VERSION} does not exist"
    echo "Create it first with: git tag v${VERSION} -m 'Release v${VERSION}'"
    exit 1
fi

# Download GitHub tarball and get sha256sum
echo "==> Downloading GitHub tarball..."
curl -L "https://github.com/Nomadcxx/sysc-greet/archive/v${VERSION}.tar.gz" \
    -o "/tmp/sysc-greet-${VERSION}.tar.gz"

SHA256=$(sha256sum "/tmp/sysc-greet-${VERSION}.tar.gz" | awk '{print $1}')
echo "==> GitHub tarball SHA256: ${SHA256}"

# Update PKGBUILDs
echo "==> Updating PKGBUILDs..."
for PKGBUILD in PKGBUILD PKGBUILD-hyprland PKGBUILD-sway; do
    sed -i "s/^pkgver=.*/pkgver=${VERSION}/" "${REPO_ROOT}/${PKGBUILD}"
    sed -i "s/^pkgrel=.*/pkgrel=1/" "${REPO_ROOT}/${PKGBUILD}"
    sed -i "s/^sha256sums=.*/sha256sums=('${SHA256}')/" "${REPO_ROOT}/${PKGBUILD}"
done

# Generate .SRCINFO files
echo "==> Generating .SRCINFO files..."
cd "${REPO_ROOT}"
makepkg --printsrcinfo > .SRCINFO
makepkg --printsrcinfo -p PKGBUILD-hyprland > .SRCINFO-hyprland
makepkg --printsrcinfo -p PKGBUILD-sway > .SRCINFO-sway

echo "==> Done! PKGBUILDs updated with correct GitHub tarball checksum"
echo ""
echo "Next steps:"
echo "  1. Test builds: makepkg -si (for each variant)"
echo "  2. Commit: git add PKGBUILD* .SRCINFO*"
echo "  3. Update AUR repos"
