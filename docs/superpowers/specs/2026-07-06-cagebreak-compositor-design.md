# Cagebreak Compositor — Design Spec

> **Status:** Approved design (pending spike gate)
> **Issue:** [#69 Feature request: cage](https://github.com/Nomadcxx/sysc-greet/issues/69)
> **PR:** [#78](https://github.com/Nomadcxx/sysc-greet/pull/78) — pivoted from Cage Lite to Cagebreak
> **Supersedes:** [2026-06-18-cage-compositor-design.md](2026-06-18-cage-compositor-design.md)
> **Branch:** `feat/cage-compositor-investigation` (worktree-isolated; merge target is `development`, NOT `master`)

## Problem

Hyprland is expensive as a *greeter* compositor: config churn, heavy cold-start, and
most greeter-specific bug reports trace to compositor integration rather than
sysc-greet Go code. We want a minimal, stable replacement so Hyprland greeter
support can be deprecated and eventually removed.

The previous draft of PR #78 chose **Cage** ("Cage Lite"). Cage works with greetd
(it is the standard `cage -s -- gtkgreet` pattern on the Arch wiki), but it has **no
wlr-layer-shell**, so gSlapper wallpapers can never work — users migrating from
Hyprland would silently lose boot wallpapers. That makes Cage Lite a weak
deprecation story.

## Decision

Replace Cage with **[Cagebreak](https://github.com/project-repo/cagebreak)**
(v3.2.1, released 2026-06-13, actively maintained). Cagebreak is a Cage fork with
tiling, a config file, an IPC socket, and — critically — **wlr-layer-shell scene
trees** (`layer_shell_background/bottom/top` in `output.h`), which should let
gSlapper render wallpapers exactly as it does under niri/sway/Hyprland.

Why this beats Cage Lite:

| Capability | Cage 0.3.x | Cagebreak 3.2.x |
|---|---|---|
| greetd session | ✅ | ✅ (needs empirical spike) |
| gSlapper wallpapers (layer-shell) | ❌ never | ✅ expected — spike gate |
| Config file (`-c <path>`) | ❌ flags only | ✅ |
| Startup exec of multiple clients | ❌ single app | ✅ `exec <command>` |
| Programmatic quit | exits with client | ✅ `quit` via IPC socket (`-e`) |
| XKB layout | `XKB_DEFAULT_*` env | `XKB_DEFAULT_*` env |
| Keybinding lockdown | n/a | ✅ define no binds |
| Arch packaging | extra repo | ⚠️ AUR only (`cagebreak`, `cagebreak-bin`) |

The single downside is packaging: cagebreak is AUR-only on Arch. The installer
handles this (see below). Nixpkgs ships `cagebreak`.

## Goals

1. Add **cagebreak** as a supported greeter backend with **full feature parity**
   (gSlapper wallpapers, videos, multi-monitor) alongside niri and sway
2. **Phase out Hyprland** for the greeter session — deprecated in this PR with
   user-facing notices, removed ~3 months after cagebreak ships stable
3. Preserve login reliability (greetd IPC, session start, keyboard input)
4. No Go application changes — sysc-greet is compositor-agnostic

## Non-goals

- Keeping Cage Lite as a parallel option (removed from this PR; can return later
  if someone asks for a no-wallpaper kiosk path)
- Removing Hyprland support in this PR
- Dropping niri or sway support

## Spike Gate (must pass before the PR rewrite ships)

The design hinges on one empirical unknown: **does gSlapper's layer-shell
wallpaper actually render under cagebreak?** Nobody has tested it. The spike
runs on real hardware (a VT, not SSH — wlroots needs DRM/seat):

1. Install `cagebreak` (AUR: `yay -S cagebreak` or `cagebreak-bin`) on the test
   machine. *Note: as of 2026-07-06 cagebreak is NOT installed on the dev
   server — only cage 0.3.0.*
2. Minimal test config exercising the exact greeter pattern:
   - `background 0.0 0.0 0.0`
   - `exec` gSlapper (via `sysc-greet --wallpaper-daemon`)
   - `exec` kitty → `sysc-greet --test`, then quit via socket
   - no `bind`/`definekey` lines
3. Run `cagebreak -e -c /tmp/cagebreak-spike-config` from a VT.
4. Verify, in order of importance:
   - **P0:** wallpaper renders behind kitty (layer-shell background works)
   - **P0:** login-shaped flow works: kitty fullscreen, keyboard input reaches
     sysc-greet
   - **P0:** compositor exits when the kitty chain sends `quit` to
     `$CAGEBREAK_SOCKET` (confirm the env var propagates to `exec` children;
     fallback if not: `pkill cagebreak` as last resort)
   - **P1:** `XKB_DEFAULT_LAYOUT` honored
   - **P1:** single view occupies the full output (no visible tiling chrome)
   - **P1:** multi-monitor behavior sane (if second output available)

**Gate outcome:** If the P0 wallpaper test fails, STOP — the pivot loses its
rationale; fall back to the existing Cage Lite implementation already on the
branch and re-evaluate.

## Architecture

Mirrors the **niri pattern** (config-driven), not the Cage launcher-script
pattern:

```
greetd → cagebreak -e -c /etc/greetd/cagebreak-greeter-config
           ├─ exec sysc-greet --wallpaper-daemon   (gSlapper, layer-shell)
           └─ exec kitty → sysc-greet ; quit-via-socket
```

### Components

| File | Change |
|---|---|
| `config/cagebreak-greeter-config` | **New.** Greeter config (see below). Replaces `scripts/cage-greeter-session.sh` (deleted). |
| `cmd/installer/main.go` | `cage` → `cagebreak` in menu/validation/install/greetd config/uninstall. AUR-aware install on Arch. |
| `flake.nix` | `compositor = "cagebreak"` enum value + `cagebreakPackage` option; ship the config via `environment.etc`; drop cage branches. |
| `nfpm.yaml` | Ship `/etc/greetd/cagebreak-greeter-config` (0644); drop cage launcher. |
| `scripts/postinstall.sh` | Auto-detect order: `cagebreak → niri → sway → hyprland` (deprecation warning on hyprland). |
| `scripts/verify-cagebreak-greeter.sh` | Rewritten smoke checks + spike helper. Replaces `verify-cage-greeter.sh`. |
| `docs-src/compositors/cagebreak.md` | Replaces `cage.md`: setup, AUR note, full-parity feature table, Hyprland migration guide. |
| `docs-src/compositors/hyprland.md`, `installation.md`, `README.md`, `mkdocs.yml` | Deprecation notices retained, cage references → cagebreak. |

### Greeter config (`config/cagebreak-greeter-config`)

```
# sysc-greet cagebreak greeter config — used ONLY by greetd
background 0.0 0.0 0.0
exec HOME=/var/lib/greeter /usr/local/bin/sysc-greet --wallpaper-daemon
exec XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; echo quit | socat - UNIX-CONNECT:"$CAGEBREAK_SOCKET"
# No bind/definekey lines: no compositor keybindings reachable from the greeter
```

Exact quit mechanism (socat vs alternative) is confirmed during the spike; if
`$CAGEBREAK_SOCKET` does not propagate to `exec` children, fall back to
`pkill cagebreak`. `socat` becomes a package dependency only if the socket path
is used (Arch: `socat` in core/extra; document in nfpm depends + docs).

### greetd command

```toml
[default_session]
command = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
user = "greeter"
```

`-e` enables the IPC socket needed for `quit`. XKB overrides documented as:
`command = "env XKB_DEFAULT_LAYOUT=de cagebreak -e -c ..."`.

### Installer UX

```
cagebreak (recommended) — Minimal tiling kiosk; full gSlapper wallpapers
niri                    — Tiling compositor with scrollable workspaces
sway                    — Stable i3-compatible tiling compositor
hyprland (deprecated)   — Greeter support ending in ~3 months; migrate to cagebreak
```

Arch install path for cagebreak: try `pacman -S cagebreak` (in case it enters
official repos), else detect `paru`/`yay` and install from AUR **as the invoking
user** (AUR helpers refuse root; use `SUDO_USER`), else print manual AUR
instructions and continue (compositor presence is validated before this step
anyway).

### NixOS module

```nix
services.sysc-greet = {
  compositor = "cagebreak";
  cagebreakPackage = pkgs.cagebreak;  # optional; PATH resolution when null
};
```

`defaultCompositorCommand`:
`"${cagebreak}/bin/cagebreak -e -c /etc/greetd/cagebreak-greeter-config"`.

### Go application changes

**None.** sysc-greet already launches gSlapper generically and uses greetd IPC
regardless of compositor.

## PR / branch mechanics

- Work continues on `feat/cage-compositor-investigation` in an isolated worktree
  (`.claude/worktrees/cagebreak-pivot`).
- PR #78 base is retargeted `master` → `development`.
- PR title/body rewritten for the cagebreak pivot.
- The old cage spec/plan docs remain in-tree with a "superseded" banner for
  the investigation record.
- **Never merge without explicit user sign-off; merge target is `development`.**

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Layer-shell doesn't actually work with gSlapper | Spike gate before rewrite; fall back to Cage Lite already on branch |
| `$CAGEBREAK_SOCKET` not visible to exec children | Spike verifies; `pkill cagebreak` fallback |
| AUR-only packaging on Arch | Installer AUR-helper fallback + manual instructions; document clearly |
| Cagebreak default keybindings reachable pre-login | Greeter config defines zero binds; spike confirms no built-in defaults remain |
| greetd restart loop on compositor crash | Config validated by `cagebreak -c <config>` parse; verify script checks config syntax |
| Smaller community than cage-kiosk/cage | Accepted: 3.2.1 June 2026 shows active maintenance; revisit at Hyprland-removal decision |

## Success Criteria

1. Spike passes all P0 checks on real hardware
2. `SYSC_COMPOSITOR=cagebreak sudo ./install.sh` produces a working greetd config
3. Boot → greeter with gSlapper wallpaper → login → session start, no loop
4. Hyprland deprecation notices present in installer, postinstall, docs
5. PR #78 updated (base `development`), out of draft only after criteria 1–4
