# Cagebreak Packaging for Debian/Fedora — Plan (stub)

> **Status:** Stub — approach decided, implementation not started
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
   with per-distro container jobs (debian, ubuntu LTS, fedora) that build
   cagebreak against each distro's wlroots and package it with nfpm as a
   separate `cagebreak_<ver>_<distro>.deb/.rpm` attached to sysc-greet GitHub
   releases. The installer on apt/dnf installs that artifact.
2. **Fallback — installer source build.** Mirror the gslapper pattern
   (`buildGslapperFromSource`): distro dep lists (meson, ninja, libwlroots-dev,
   libxkbcommon-dev, wayland-protocols, scdoc) plus wlroots-version →
   cagebreak-tag selection. Used only where no artifact exists.
3. Upstream packaging PR to project-repo/cagebreak: rejected as redundant —
   distro inclusion needs maintainers/sponsorship a PR cannot shortcut, and our
   artifacts serve users either way.

## Tasks

### Phase A — CI artifacts (primary)

- [ ] Map wlroots version per target: debian:13, ubuntu:24.04, fedora:42
- [ ] Pick buildable cagebreak tag per wlroots version (greeter config needs
      only `background`, `exec`, `-c`, `-e`; 2.x suffices)
- [ ] Test wlroots as meson subproject — if cagebreak supports it, the matrix
      may collapse to one job per package format
- [ ] Container build jobs in `release.yml`; nfpm config `nfpm-cagebreak.yaml`
- [ ] Headless smoke test in CI (`WLR_BACKENDS=headless cagebreak -e -c ...`)
- [ ] Attach artifacts + SHA256SUMS to releases

### Phase B — installer integration

- [ ] apt/dnf path: fetch matching artifact from GitHub releases, install via
      `dpkg -i` / `dnf install`
- [ ] Source-build fallback `buildCagebreakFromSource` when no artifact matches
- [ ] Replace the current "build from https://github.com/project-repo/cagebreak"
      error with the above
- [ ] Docs: `docs-src/compositors/cagebreak.md` install section per distro

### Phase C — validation

- [ ] Debian VM/container: install → greetd → login
- [ ] Fedora VM/container: install → greetd → login
- [ ] openSUSE: verify the packaged 2.3.1 works with the greeter config

## Non-goals

- Bundling cagebreak inside the sysc-greet package (file conflicts if a distro
  ever packages it; couples our release to a compositor we don't maintain)
- Getting cagebreak into official Debian/Fedora repos
