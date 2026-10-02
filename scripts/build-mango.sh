#!/usr/bin/env bash
# build-mango.sh — build mango (+ mmsg) for distros without a mangowm package
#
# Usage: scripts/build-mango.sh <system|bundle|bundle-full> <payload-dir>
#
#   system       link against the distro's wlroots 0.20 + scenefx 0.5 (Fedora 44)
#   bundle       build wlroots + scenefx as private shared libs (Fedora 43, Ubuntu 26.04)
#   bundle-full  also bundle libwayland, libdrm, pixman, xkbcommon (Debian 13)
#
# Private libs live in /usr/lib/sysc-greet-mango, reached only through mango's
# RUNPATH, so the system's own wlroots/libdrm/... are untouched. Static linking
# does not work: scenefx carries copies of wlroots internals and the archives
# collide.
#
# Runs as root in a throwaway container (installs build deps, installs the
# private libs into / so later builds find them). Writes the package payload
# to <payload-dir> and runtime dependencies, one per line, to
# <payload-dir>.depends: deb package names, or rpm soname capabilities.
#
# Never run this, or the mango it builds, inside a logged-in graphical
# session: on exit mango stops graphical-session.target.
set -euo pipefail

MANGO_VERSION=0.17.4
WLROOTS_VERSION=0.20.2
SCENEFX_VERSION=0.5
WAYLAND_VERSION=1.24.0
WAYLAND_PROTOCOLS_VERSION=1.47
LIBDRM_VERSION=2.4.131
PIXMAN_VERSION=0.46.4
XKBCOMMON_VERSION=1.8.1

MODE=${1:?usage: build-mango.sh <system|bundle|bundle-full> <payload-dir>}
PAYLOAD=$(realpath -m "${2:?usage: build-mango.sh <system|bundle|bundle-full> <payload-dir>}")
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PRIV=/usr/lib/sysc-greet-mango
SRC=/tmp/mango-src
STAGE=/tmp/mango-stage

case "${MODE}" in system | bundle | bundle-full) ;; *) echo "unknown mode: ${MODE}" >&2; exit 2 ;; esac

if command -v apt-get >/dev/null; then
  PKGMGR=apt
elif command -v dnf >/dev/null; then
  PKGMGR=dnf
else
  echo "need apt-get or dnf" >&2; exit 2
fi

install_deps() {
  if [[ "${PKGMGR}" == apt ]]; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    local pkgs=(git ca-certificates build-essential meson ninja-build pkg-config cmake bison
      libwayland-dev libwayland-bin wayland-protocols libinput-dev libudev-dev libdrm-dev libpixman-1-dev libxkbcommon-dev
      libpcre2-dev libcjson-dev libpango1.0-dev libcairo2-dev
      libxcb1-dev libxcb-icccm4-dev libxcb-randr0-dev libxcb-composite0-dev libxcb-render-util0-dev libxcb-ewmh-dev
      libxcb-res0-dev libxcb-xfixes0-dev libxcb-dri3-dev libxcb-present-dev libxcb-errors-dev libxcb-xinput-dev libxcb-xkb-dev
      libegl-dev libgbm-dev libgles-dev libvulkan-dev glslang-tools libdisplay-info-dev libliftoff-dev libseat-dev hwdata
      liblcms2-dev xwayland libpciaccess-dev xkb-data libxml2-dev)
    apt-get install -y -qq --no-install-recommends "${pkgs[@]}" >/dev/null
  else
    local pkgs=(git gcc meson ninja-build pkgconf cmake bison
      wayland-devel wayland-protocols-devel libinput-devel systemd-devel libdrm-devel pixman-devel libxkbcommon-devel
      pcre2-devel cjson-devel pango-devel cairo-devel
      libxcb-devel xcb-util-wm-devel xcb-util-renderutil-devel xcb-util-errors-devel
      mesa-libEGL-devel mesa-libgbm-devel libglvnd-devel vulkan-loader-devel glslang libdisplay-info-devel libliftoff-devel
      libseat-devel hwdata-devel lcms2-devel xorg-x11-server-Xwayland-devel)
    [[ "${MODE}" == system ]] && pkgs+=(wlroots-devel scenefx-devel)
    dnf -y -q install "${pkgs[@]}" >/dev/null
  fi
}

# lib <name> <git-url> <tag> [meson options...]: build into the private libdir
lib() {
  local name=$1 url=$2 tag=$3; shift 3
  git clone -q --depth 1 --branch "${tag}" "${url}" "${SRC}/${name}" 2>/dev/null
  meson setup "${SRC}/${name}/build" "${SRC}/${name}" --buildtype=release --prefix=/usr \
    --libdir=lib/sysc-greet-mango -Dc_link_args="-Wl,-rpath,${PRIV}" "$@" >"${SRC}/${name}.log" 2>&1 \
    || { tail -30 "${SRC}/${name}.log"; exit 1; }
  ninja -C "${SRC}/${name}/build" >>"${SRC}/${name}.log" 2>&1 || { tail -30 "${SRC}/${name}.log"; exit 1; }
  meson install -C "${SRC}/${name}/build" --strip >/dev/null
  # wayland-protocols is build-time only
  [[ "${name}" == wayland-protocols ]] || DESTDIR="${STAGE}" meson install -C "${SRC}/${name}/build" --strip >/dev/null
  echo "built ${name} ${tag}"
}

install_deps
rm -rf "${SRC}" "${STAGE}" "${PAYLOAD}"
mkdir -p "${SRC}" "${STAGE}"
export PKG_CONFIG_PATH="${PRIV}/pkgconfig:/usr/share/pkgconfig"

if [[ "${MODE}" == bundle-full ]]; then
  lib wayland https://gitlab.freedesktop.org/wayland/wayland.git "${WAYLAND_VERSION}" \
    -Ddocumentation=false -Dtests=false -Ddtd_validation=false
  lib wayland-protocols https://gitlab.freedesktop.org/wayland/wayland-protocols.git "${WAYLAND_PROTOCOLS_VERSION}" -Dtests=false
  lib libdrm https://gitlab.freedesktop.org/mesa/drm.git "libdrm-${LIBDRM_VERSION}" \
    -Dtests=false -Dman-pages=disabled -Dvalgrind=disabled -Dcairo-tests=disabled \
    -Dintel=disabled -Damdgpu=disabled -Dradeon=disabled -Dnouveau=disabled -Dvmwgfx=disabled
  lib pixman https://gitlab.freedesktop.org/pixman/pixman.git "pixman-${PIXMAN_VERSION}" \
    -Dtests=disabled -Ddemos=disabled -Dgtk=disabled -Dlibpng=disabled
  lib libxkbcommon https://github.com/xkbcommon/libxkbcommon.git "xkbcommon-${XKBCOMMON_VERSION}" \
    -Denable-docs=false -Denable-tools=false -Denable-x11=false -Denable-wayland=false -Denable-bash-completion=false
fi
if [[ "${MODE}" != system ]]; then
  lib wlroots https://gitlab.freedesktop.org/wlroots/wlroots.git "${WLROOTS_VERSION}" -Dexamples=false -Dxwayland=enabled
  lib scenefx https://github.com/wlrfx/scenefx.git "${SCENEFX_VERSION}" -Dexamples=false
fi

git clone -q --depth 1 --branch "${MANGO_VERSION}" https://github.com/mangowm/mango.git "${SRC}/mango" 2>/dev/null
rpath=()
[[ "${MODE}" != system ]] && rpath=(-Dc_link_args="-Wl,-rpath,${PRIV}")
meson setup "${SRC}/mango/build" "${SRC}/mango" --buildtype=release --prefix=/usr "${rpath[@]}" \
  >"${SRC}/mango.log" 2>&1 || { tail -30 "${SRC}/mango.log"; exit 1; }
ninja -C "${SRC}/mango/build" >>"${SRC}/mango.log" 2>&1 || { tail -30 "${SRC}/mango.log"; exit 1; }
DESTDIR="${STAGE}" meson install -C "${SRC}/mango/build" --strip >/dev/null
echo "built mango ${MANGO_VERSION}"

# Payload: only runtime files. The bundled builds also stage headers,
# pkg-config files and wayland-scanner, which would clash with distro packages.
mkdir -p "${PAYLOAD}"
for path in usr/bin/mango usr/bin/mmsg etc/mango usr/share/wayland-sessions usr/share/xdg-desktop-portal \
  usr/share/man/man1/mmsg.1 usr/lib/systemd/user; do
  [[ -e "${STAGE}/${path}" ]] && (cd "${STAGE}" && cp -a --parents "${path}" "${PAYLOAD}/")
done
if [[ "${MODE}" != system ]]; then
  mkdir -p "${PAYLOAD}${PRIV}"
  cp -a "${STAGE}${PRIV}"/*.so* "${PAYLOAD}${PRIV}/"
fi

# Checks against the installed private libs (same paths the package uses)
for bin in "${PAYLOAD}"/usr/bin/mango "${PAYLOAD}"/usr/bin/mmsg; do
  if ldd "${bin}" | grep 'not found'; then echo "unresolved libraries in ${bin}" >&2; exit 1; fi
done
if grep -nE '^\s*(bind|mousebind|axisbind|source)\s*=' "${ROOT}/config/mango-greeter-config.conf"; then
  echo "greeter config has bind/source lines" >&2; exit 1
fi
"${PAYLOAD}/usr/bin/mango" -c "${ROOT}/config/mango-greeter-config.conf" -p >/dev/null
echo "greeter config parses under $("${PAYLOAD}/usr/bin/mango" -v)"

# Runtime dependencies: every shared lib mango/mmsg/private libs load, minus the private ones
libs=$(ldd "${PAYLOAD}"/usr/bin/mango "${PAYLOAD}"/usr/bin/mmsg "${PAYLOAD}${PRIV}"/*.so* 2>/dev/null \
  | awk '/=> \// {print $1, $3}' | grep -v " ${PRIV}/" | sort -u || true)
# plus two non-library runtime needs mango crashes without: the XKB keymaps
# (a bundled xkbcommon does not pull them in; crash at startup) and the
# Xwayland binary (mango 0.17.4 segfaults on exit when it is missing; Arch's
# mangowm depends on xorg-xwayland too)
if [[ "${PKGMGR}" == apt ]]; then
  { awk '{print $2}' <<<"${libs}" | xargs -r readlink -f | xargs -r dpkg -S 2>/dev/null | cut -d: -f1
    echo xkb-data; echo xwayland; } | sort -u >"${PAYLOAD}.depends"
else
  { awk '{print $1 "()(64bit)"}' <<<"${libs}"; echo xkeyboard-config; echo xorg-x11-server-Xwayland; } \
    | sort -u >"${PAYLOAD}.depends"
fi
echo "runtime depends: $(wc -l <"${PAYLOAD}.depends") entries in ${PAYLOAD}.depends"
