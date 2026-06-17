# Cage / Cagebreak Compositor — Design Spec

> **Status:** Draft / investigation  
> **Issue:** [#69 Feature request: cage](https://github.com/Nomadcxx/sysc-greet/issues/69)  
> **Branch:** `feat/cage-compositor-investigation`

## Problem

Hyprland is the default greeter compositor for many installs (AUR `sysc-greet-hyprland`, postinstall auto-detection). It works but carries ongoing cost:

- Config churn (window rules, layer rules, `start-hyprland` vs `Hyprland`, watchdog fd, ecosystem popups)
- Heavier cold-start than a kiosk compositor
- ~99% of greeter-specific issues trace to compositor integration, not sysc-greet Go code

[Cage](https://github.com/cage-kiosk/cage) is a minimal wlroots kiosk compositor designed to run **one maximized application**. That matches 80% of what the greeter needs: a fullscreen Kitty running sysc-greet.

## Goals

1. Add **Cage** as a supported greeter backend alongside niri, sway, and Hyprland
2. **Phase out Hyprland** for the greeter session over the next few months — not removed in this PR, but deprecated with clear user-facing notices
3. Preserve **login reliability** (greetd IPC, session start, keyboard input)
4. Document **feature trade-offs** honestly (no silent regressions)
5. Leave a clean extension point for **Cagebreak** investigation in Phase 2

## Hyprland Deprecation Policy

Hyprland greeter support remains **fully functional** in this release. We are not removing it yet.

| Timeline | Action |
|----------|--------|
| **Now (this PR)** | Cage available in installer, Nix module, docs. Hyprland labeled **deprecated** in UI and docs. |
| **Next few months** | Encourage migration to Cage (or niri/sway). Collect feedback on Cage Lite trade-offs. |
| **~3 months after Cage ships stable** | Remove Hyprland from installer default menu position; postinstall stops auto-selecting Hyprland. |
| **Later** | Remove Hyprland greeter configs, AUR `sysc-greet-hyprland` variant, and related docs — only after Cage (or Cagebreak) is proven stable. |

**Rationale:** Hyprland is excellent as a *user session* compositor but expensive as a *greeter* compositor. Most greeter bugs are Hyprland config churn, not sysc-greet logic. Users who run Hyprland daily can keep it for their desktop session — only the greetd greeter session moves to Cage.

## Non-goals (Phase 1)

- Removing Hyprland support in this PR
- Making Cage the silent default without user awareness
- Full feature parity with Hyprland/niri (video wallpapers, gSlapper layer-shell, multi-monitor wallpaper daemon)
- Dropping niri or sway support
- Cagebreak implementation (investigation only)

## Current Architecture (reference)

```
greetd → compositor → [gslapper daemon] + kitty → sysc-greet
                      ↑ second Wayland client (wlr-layer-shell)
```

All three supported compositors (niri, Hyprland, sway) are **multi-client** and provide:

| Capability | Used by sysc-greet |
|------------|-------------------|
| Multiple Wayland clients | gSlapper + Kitty |
| wlr-layer-shell | gSlapper `wallpaper` namespace |
| Window rules | Kitty fullscreen |
| Compositor exit command | `hyprctl` / `niri msg` / `swaymsg` |
| Compositor keyboard config | `kb_layout` / `xkb` (+ Kitty `XKB_*` env) |

## Cage Constraint (critical)

**Cage runs exactly one application.** It does not host a second Wayland client.

Implications:

| Feature | Cage support |
|---------|-------------|
| Kitty fullscreen TUI | ✅ Primary client |
| `sysc-greet --wallpaper-daemon` → gSlapper | ❌ Second client rejected / invisible |
| Video wallpapers at boot | ❌ Requires gSlapper |
| Theme PNG wallpapers at boot | ❌ Requires gSlapper |
| TUI background effects (matrix, sonar, etc.) | ✅ Rendered inside Kitty |
| Theme colors / ASCII art | ✅ Unchanged |
| greetd IPC login | ✅ Unchanged |
| XKB via `XKB_DEFAULT_*` on Kitty | ✅ Only path (no compositor xkb config) |
| Multi-monitor | ⚠️ Cage `-m extend` spans single client across outputs |
| Compositor exit | ✅ Automatic when Kitty exits |

**Conclusion:** Cage "Lite" mode = Kitty-only session. Wallpapers fall back to TUI effects + theme colors already implemented in sysc-greet.

## Cagebreak (Phase 2 — not in Phase 1)

[Cagebreak](https://github.com/project-repo/cagebreak) is a Ratpoison-inspired **tiling** compositor forked from Cage. It supports multiple windows/workspaces and may support wlr-layer-shell.

| Question | Investigation needed |
|----------|---------------------|
| Layer-shell for gSlapper? | Test `gslapper` + kitty under cagebreak |
| Config format | `~/.config/cagebreak/config` vs greetd inline command |
| Packaging | Arch `cagebreak` AUR; not in most distro repos |
| Maintenance | Smaller community than cage-kiosk/cage |
| Greeter fit | Better wallpaper parity if layer-shell works |

**Recommendation:** Ship Cage Lite in Phase 1. Run cagebreak spike as gated Phase 2 before promising wallpaper parity.

## Proposed Approaches

### Approach A — Cage Lite (recommended Phase 1)

Inline greetd command via launcher script:

```bash
cage -s -m extend -- /etc/greetd/cage-greeter-session.sh
```

`cage-greeter-session.sh` sets greeter env vars and `exec`s Kitty → sysc-greet.

**Pros:** Minimal code, fast boot, no Hyprland config, easy to reason about  
**Cons:** No gSlapper/video/static boot wallpapers

### Approach B — Shell wrapper spawning gSlapper inside Cage

```bash
cage -- sh -c 'sysc-greet --wallpaper-daemon & exec kitty ...'
```

**Pros:** Would restore wallpapers if it worked  
**Cons:** gSlapper is a **second Wayland client** — Cage kiosk model does not support this. Issue #69 experiments without gSlapper succeeded; with gSlapper likely fails silently.

**Verdict:** Reject for Phase 1 unless empirical testing on real hardware proves otherwise.

### Approach C — Cagebreak with compositor config

Add `config/cagebreak-greeter-config` mirroring niri structure.

**Pros:** Potential full wallpaper + multi-client parity  
**Cons:** Unknown layer-shell support, more packaging work, smaller user base

**Verdict:** Phase 2 spike, not Phase 1 deliverable.

## Selected Design: Approach A (Cage Lite)

### Components

| File | Purpose |
|------|---------|
| `scripts/cage-greeter-session.sh` | Greeter env + Kitty exec (installed to `/etc/greetd/`) |
| `docs-src/compositors/cage.md` | Experimental setup + trade-offs |
| `cmd/installer/main.go` | 4th compositor option `(experimental)` |
| `flake.nix` | `compositor = "cage"` enum + `cagePackage` option |
| `scripts/verify-cage-greeter.sh` | Smoke test (nested Wayland if available) |

### greetd command

```toml
[default_session]
command = "cage -s -m extend -- /etc/greetd/cage-greeter-session.sh"
user = "greeter"
```

Flags:

- `-s` — allow VT switching (match issue #69 / greetd expectations)
- `-m extend` — span greeter across all connected outputs

### Launcher script

```sh
#!/bin/sh
export XDG_CACHE_HOME=/tmp/greeter-cache
export HOME=/var/lib/greeter
# XKB_DEFAULT_LAYOUT / VARIANT set by installer or user override
exec kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet
```

No `pkill cage` — Cage exits when its sole client (Kitty) exits.

### Installer UX

Menu order (Cage first — recommended path):

```
cage (recommended) — Minimal kiosk; fast boot; TUI backgrounds only
niri               — Full wallpapers; tiling compositor
sway               — Full wallpapers; i3-compatible
hyprland (deprecated) — Full wallpapers; greeter support ending in ~3 months
```

- `SYSC_COMPOSITOR=cage` env pre-select supported
- **postinstall.sh:** prefer `cage` → `niri` → `sway` → `hyprland` (with deprecation warning on hyprland)
- Show Hyprland deprecation notice in installer footer and when hyprland is selected

### NixOS module

```nix
compositor = "cage";
# optional
cagePackage = pkgs.cage;
```

`defaultCompositorCommand` becomes:

```nix
"${cage}/bin/cage -s -m extend -- /etc/greetd/cage-greeter-session.sh"
```

### Go application changes

**None required for Phase 1.** sysc-greet already:

- Falls back to TUI backgrounds when gSlapper is absent
- Uses greetd IPC regardless of compositor

Optional Phase 1.5: log once at startup when `SYSC_GREETER_MODE=cage` to clarify wallpaper fallback (not in initial draft).

### Feature matrix (documented for users)

| Feature | Hyprland/niri/sway | Cage Lite |
|---------|-------------------|-----------|
| Login | ✅ | ✅ |
| TUI effects | ✅ | ✅ |
| Theme colors | ✅ | ✅ |
| Boot static/video wallpaper | ✅ | ❌ |
| In-greeter wallpaper change | ✅ | ⚠️ TUI effects only |
| Non-US keyboard | compositor + Kitty XKB | Kitty XKB only |
| Multi-monitor wallpaper | ✅ | ⚠️ Single client span |
| Hyprland config maintenance | ❌ ongoing | ✅ none |

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Users expect wallpapers | Bold EXPERIMENTAL label + docs trade-off table |
| Old Intel GPU + Kitty EGL | Issue #77 (foot) is separate; cage doesn't solve EGL |
| Cage not in repos | Document manual install; Nix has `pkgs.cage` |
| greetd user can't run cage | Polkit/capabilities unchanged from other compositors |
| Regression in login | `verify-cage-greeter.sh` + manual greetd test checklist |

## Success Criteria (Phase 1)

1. `SYSC_COMPOSITOR=cage` install produces working greetd config
2. Greeter login succeeds on at least one test machine (Arch + cage from repos)
3. Docs clearly state wallpaper limitation
4. Draft PR open for community feedback before marking stable

## Open Questions (for user / Phase 2)

1. Should Cage Lite become the **recommended** compositor for new installs (over niri)?
2. Is TUI-only wallpaper acceptable as permanent trade-off, or is Cagebreak spike a blocker?
3. Should we add `foot` terminal path alongside cage for old Intel GPUs (#77)?
