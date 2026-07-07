# Cagebreak Packaging for Debian/Fedora — Plan

> **Status:** Implementation plan — approach decided 2026-07-06, research complete 2026-07-07
> **Follows:** PR #78 (cagebreak greeter backend, merged to `development`)
> **Blocks:** Hyprland greeter removal on apt/dnf distros

## Problem

Cagebreak is packaged for Arch (AUR), NixOS, Alpine, Void, and the BSDs. It is
absent from Debian, Ubuntu, and Fedora repos. sysc-greet supports those distros,
so the hyprland → cagebreak migration has no install path there.

Cagebreak is C, dynamically linked against wlroots. A single binary is not
portable across distros the way the static sysc-greet Go binary is, and distro
wlroots versions (0.17–0.20) pin which cagebreak release can build.

## Decision (2026-07-06)

1. **Primary — release CI artifacts.** Extend `.github/workflows/release.yml`
   with per-distro container jobs that build cagebreak against each distro's
   wlroots and package it with nfpm as a separate
   `cagebreak_<ver>_<distro>.deb/.rpm` attached to sysc-greet GitHub releases.
   The installer on apt/dnf installs that artifact.
2. **Fallback — installer source build.** Mirror the gslapper pattern
   (`buildGslapperFromSource` in `cmd/installer/main.go`): per-distro dep
   lists plus wlroots-version → cagebreak-tag selection. Used only where no
   artifact matches.
3. Upstream packaging PR to project-repo/cagebreak: rejected as redundant —
   distro inclusion needs maintainers/sponsorship a PR cannot shortcut, and our
   artifacts serve users either way.

## Research findings (2026-07-07)

### wlroots requirement per cagebreak tag

Verified against each tag's `meson.build`. Cagebreak has **no meson subproject
fallback for wlroots** — the system wlroots version dictates the buildable tag.
The stub's "test wlroots as meson subproject" task is dead; removed.

| cagebreak tag | wlroots requirement |
|---|---|
| 3.2.1 (2026-06) | wlroots-0.20 |
| 3.1.0 | wlroots-0.19 |
| 3.0.1 | wlroots-0.19 |
| 2.4.0 | wlroots-0.18 |
| 2.3.1 | >= 0.17.0, < 0.18.0 |

### wlroots shipped per target distro

| Distro | wlroots | Buildable cagebreak |
|---|---|---|
| Ubuntu 24.04 LTS | 0.17.1 | 2.3.1 |
| Debian 13 (trixie, stable) | 0.18.2 | 2.4.0 |
| Ubuntu 25.04 / 25.10 | 0.18.2 (+0.17.4 compat) | 2.4.0 |
| Fedora 41 | 0.18.2 | 2.4.0 |
| Fedora 42 | 0.19.2 (+0.18.2 compat) | 3.1.0 |
| Debian sid / Fedora rawhide / Tumbleweed | 0.20.1 | 3.2.1 |
| openSUSE Leap | 0.14.1 | none — unsupported |

Old tags are fine: the greeter config uses only `background`, `exec`, `-c`,
and `-e`, all present since 2.x. Feature drift between 2.3.1 and 3.2.1 is
irrelevant to the greeter.

### Build dependencies

Common to all tags: meson, ninja, wlroots (devel), wayland-server/client/cursor,
wayland-protocols, xkbcommon, libinput, cairo, pango + pangocairo, fontconfig,
libevdev, libudev, pixman-1, scdoc (man pages, optional via
`-Dman-pages=false`).

Debian/Ubuntu package names: `meson ninja-build libwlroots-dev libwayland-dev
wayland-protocols libxkbcommon-dev libinput-dev libcairo2-dev libpango1.0-dev
libfontconfig-dev libevdev-dev libudev-dev libpixman-1-dev`
(trixie names the wlroots package `libwlroots-0.18-dev`).

Fedora package names: `meson ninja-build wlroots-devel wayland-devel
wayland-protocols-devel libxkbcommon-devel libinput-devel cairo-devel
pango-devel fontconfig-devel libevdev-devel systemd-devel pixman-devel`.

### wlroots detection (for the source-build fallback)

pkg-config module names differ by version: `wlroots-0.20`, `wlroots-0.19`,
`wlroots-0.18` for modern releases; plain `wlroots` for 0.17. Probe from newest
to oldest with `pkg-config --exists`, map the first hit to a tag via the table
above.

## Phase A — CI artifacts (primary)

### A1. Build matrix in `release.yml`

New `package-cagebreak` job (runs alongside the existing `build` job on tag
push), one matrix entry per target:

```yaml
package-cagebreak:
  runs-on: ubuntu-latest
  container: ${{ matrix.image }}
  strategy:
    matrix:
      include:
        - { image: "ubuntu:24.04", cagebreak: "2.3.1", packager: deb, suffix: ubuntu24.04 }
        - { image: "debian:13",    cagebreak: "2.4.0", packager: deb, suffix: debian13 }
        - { image: "fedora:42",    cagebreak: "3.1.0", packager: rpm, suffix: fedora42 }
  steps:
    # 1. install build deps (apt/dnf per image)
    # 2. git clone --depth 1 --branch <tag> https://github.com/project-repo/cagebreak
    # 3. meson setup build -Dman-pages=false && ninja -C build
    # 4. smoke test: WLR_BACKENDS=headless WLR_LIBINPUT_NO_DEVICES=1 \
    #      timeout 5 ./build/cagebreak -e -c config/cagebreak-greeter-config
    #    (exit 124 = parsed and ran; anything else fails the job)
    # 5. nfpm package --config nfpm-cagebreak.yaml --packager <packager>
    # 6. upload-artifact
```

The existing release-create step gains a download-artifact step and attaches
`cagebreak_*` packages plus their checksums to the same release.

Notes:
- Not added to the sysc-greet package itself (see Non-goals).
- Debian 13 artifact also serves Ubuntu 25.x and is documented as such;
  we do not build a fourth image for it.
- Fedora 41 skipped: EOL before this ships; users there get the source-build
  fallback (wlroots 0.18 → tag 2.4.0).

### A2. `nfpm-cagebreak.yaml`

Modeled on `nfpm.yaml`, but packaging the compositor binary only:

```yaml
name: cagebreak
arch: amd64
version: ${CAGEBREAK_VERSION}     # the cagebreak tag, not the sysc-greet tag
maintainer: Nomadcxx
description: Cagebreak Wayland compositor (built for sysc-greet greeter use)
license: MIT
depends: ${CAGEBREAK_DEPENDS}     # per-distro runtime libs, set by the matrix
contents:
  - src: ./cagebreak/build/cagebreak
    dst: /usr/bin/cagebreak
```

Runtime `depends` must be explicit per target (nfpm does not compute shared-lib
deps the way debhelper does): e.g. trixie `libwlroots-0.18, libpango-1.0-0,
libcairo2, libinput10, libxkbcommon0, libevdev2, libfontconfig1`, Fedora
`wlroots, pango, cairo, libinput, libxkbcommon, libevdev, fontconfig`.
Verify exact soname-package names inside each container during implementation
(`dpkg -S`/`dnf provides` on the linked libs from `ldd`).

Output naming: `cagebreak_<version>_<suffix>_amd64.deb` /
`cagebreak-<version>-1.<suffix>.x86_64.rpm` so the installer can select by
distro string.

### A3. Tasks

- [ ] `nfpm-cagebreak.yaml` with env-driven version/depends
- [ ] `package-cagebreak` matrix job in `release.yml` per A1
- [ ] Headless smoke test step against `config/cagebreak-greeter-config`
- [ ] Attach artifacts + extend SHA256SUMS in the release job
- [ ] Dry-run via `workflow_dispatch` trigger before tagging a real release

## Phase B — installer integration

### B1. Artifact fetch path (apt/dnf)

In `cmd/installer/main.go`, new `installCagebreakDebian` / `installCagebreakFedora`
called from the compositor-install switch where the current "build from
https://github.com/project-repo/cagebreak" error sits:

1. Detect distro + version (`/etc/os-release` ID + VERSION_ID, already parsed
   for gslapper deps).
2. Map to artifact suffix: ubuntu 24.04 → `ubuntu24.04`; debian 13 / ubuntu
   25.x → `debian13`; fedora 42+ → `fedora42`.
3. Download from the latest sysc-greet release
   (`https://github.com/Nomadcxx/sysc-greet/releases/latest/download/<name>`),
   verify against SHA256SUMS, install via `apt install ./<file>` /
   `dnf install ./<file>` (resolves the declared lib deps).
4. No matching artifact (unknown distro, arm64, sid-class wlroots 0.20) →
   fall through to B2.

### B2. Source-build fallback `buildCagebreakFromSource`

Mirror `buildGslapperFromSource` (~`cmd/installer/main.go:1093`):

1. Install build deps via the per-distro lists from Research findings
   (extend the existing dep-install helpers the way `getGStreamerDeps` does).
2. Probe wlroots: `pkg-config --exists wlroots-0.20 / -0.19 / -0.18 / wlroots`
   newest-first; map hit → tag (0.20→3.2.1, 0.19→3.1.0, 0.18→2.4.0,
   0.17→2.3.1). No hit → error naming the missing `-dev`/`-devel` package.
3. `git clone --depth 1 --branch <tag>`, `meson setup build -Dman-pages=false`,
   `ninja -C build`, install binary to `/usr/local/bin/cagebreak`.
4. Verify with the same headless parse test used by
   `scripts/verify-cagebreak-greeter.sh`.

### B3. Tasks

- [ ] `installCagebreakDebian` / `installCagebreakFedora` artifact fetch
- [ ] `buildCagebreakFromSource` fallback with wlroots probe
- [ ] Replace the current manual-build error message with the above flow
- [ ] `socat` added to the cagebreak dep set on apt/dnf (already handled on Arch)
- [ ] Docs: per-distro install section in `docs-src/compositors/cagebreak.md`

## Phase C — validation

- [ ] Debian 13 container/VM: installer → artifact install → greetd config →
      headless greeter smoke test
- [ ] Ubuntu 24.04 container/VM: same via the 2.3.1 artifact
- [ ] Fedora 42 container/VM: same via the 3.1.0 rpm
- [ ] Source-build fallback exercised once on a target without artifacts
- [ ] openSUSE Tumbleweed: verify packaged cagebreak (2.3.1 in repos) or the
      0.20 source build works with the greeter config
- [ ] Real-hardware login test on at least one apt distro before flipping any
      hyprland-removal switch

## Non-goals

- Bundling cagebreak inside the sysc-greet package (file conflicts if a distro
  ever packages it; couples our release to a compositor we don't maintain)
- Getting cagebreak into official Debian/Fedora repos
- openSUSE Leap support (wlroots 0.14 predates every cagebreak tag we can use)
- arm64 artifacts (source-build fallback covers it; revisit on demand)
