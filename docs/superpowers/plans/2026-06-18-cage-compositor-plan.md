# Cage Compositor — Implementation Plan

> **Design:** [2026-06-18-cage-compositor-design.md](../specs/2026-06-18-cage-compositor-design.md)  
> **Branch:** `feat/cage-compositor-investigation`  
> **Type:** Draft PR / investigation spike

## Summary

Add **Cage Lite** greeter support and begin **phasing out Hyprland** for the greetd session. Hyprland remains functional in this PR — deprecation is communicated via installer UI, docs, and postinstall warnings. Cagebreak spike follows separately.

---

## Phase 1 — Draft PR (this branch)

### Task 1: Launcher script

- [x] Create `scripts/cage-greeter-session.sh`
- [x] Installer writes to `/etc/greetd/cage-greeter-session.sh` (mode 755)
- [x] nfpm.yaml ships script

### Task 2: Installer integration

- [x] Add `cage` to compositor menu (first position — recommended)
- [x] Mark `hyprland` as deprecated in menu description
- [x] Update arrow-key bounds (`compositorIndex < 3`)
- [x] `configureGreetd()` case `"cage"`
- [x] `installCompositor()` cage package mapping (pacman)
- [x] `SYSC_COMPOSITOR=cage` env pre-select
- [x] Uninstall cleanup: remove cage launcher script
- [ ] Installer Hyprland deprecation banner when hyprland selected
- [ ] Skip gslapper install when cage selected (optional — defer if complex)

### Task 3: Nix flake

- [x] Add `"cage"` to enum
- [x] Add `cagePackage` option
- [x] `defaultCompositorCommand` for cage
- [x] Install launcher via `environment.etc`
- [x] `environment.systemPackages` includes cage when selected

### Task 4: Documentation

- [x] `docs-src/compositors/cage.md`
- [x] `mkdocs.yml` nav entry
- [ ] `docs-src/getting-started/installation.md` — cage + Hyprland deprecation
- [ ] `docs-src/compositors/hyprland.md` — deprecation banner
- [ ] `README.md` — cage variant + deprecation notice
- [ ] Link from issue #69

### Task 5: postinstall + packaging

- [ ] `scripts/postinstall.sh` — prefer cage in auto-detect; warn on hyprland
- [x] `scripts/verify-cage-greeter.sh`

### Task 6: Hyprland deprecation comms

- [ ] Installer footer: "~3 months until Hyprland greeter removal"
- [ ] postinstall warning when hyprland auto-detected
- [ ] docs migration guide: Hyprland → Cage

### Task 7: Draft PR

- [ ] Push branch
- [ ] Open draft PR with deprecation timeline + manual test checklist

---

## Phase 2 — Cagebreak spike (follow-up, not this PR)

### Investigation tasks

- [ ] Install cagebreak on test machine
- [ ] Test: `gslapper` + kitty as two clients
- [ ] Test: wlr-layer-shell namespace `wallpaper`
- [ ] Test: greetd session exit (`cagebreak` IPC socket / config quit command)
- [ ] Test: XKB layout via cagebreak config
- [ ] Document results in `docs/superpowers/specs/2026-06-18-cagebreak-spike-results.md`
- [ ] Decision: promote cagebreak to supported compositor OR stay cage-lite only

### If cagebreak passes layer-shell test

- [ ] Add `config/cagebreak-greeter-config`
- [ ] Mirror niri startup pattern (wallpaper daemon + kitty chain)
- [ ] Installer 5th option or replace cage with cagebreak

---

## Phase 3 — Hyprland removal (after Cage stable)

- [ ] Stop shipping `sysc-greet-hyprland` AUR variant
- [ ] Remove Hyprland from installer menu
- [ ] Remove `config/hyprland-greeter-config*.conf` from packages
- [ ] Archive `docs-src/compositors/hyprland.md` with migration link
- [ ] Close #69 if Cagebreak not needed

---

## Phase 4 — Polish (post-approval)

- [ ] AUR package `sysc-greet-cage` variant
- [ ] postinstall: **do not** auto-select cage (keep explicit opt-in)
- [ ] Optional: `SYSC_GREETER_WALLPAPER_MODE=tui-only` env for logging clarity
- [ ] Close #69 with link to cage docs (or keep open for cagebreak)

---

## Manual Test Checklist

```bash
# 1. Install cage
sudo pacman -S cage   # or nix profile install nixpkgs#cage

# 2. Build & install sysc-greet with cage compositor
SYSC_COMPOSITOR=cage sudo ./install.sh

# 3. Verify greetd config
grep -A2 default_session /etc/greetd/config.toml
cat /etc/greetd/cage-greeter-session.sh

# 4. Restart greetd (from TTY, not SSH)
sudo systemctl restart greetd

# 5. Verify
# - Greeter appears on boot
# - TUI backgrounds/effects work (F3 menu)
# - Login succeeds → user session starts
# - No Hyprland splash / config errors in journal
journalctl -u greetd -b --no-pager | tail -50

# 6. Non-US keyboard (if applicable)
# Edit launcher script or installer output to set XKB_DEFAULT_LAYOUT=fr
# Confirm password entry works

# 7. Multi-monitor (if available)
# Confirm kitty spans outputs with cage -m extend
```

---

## Files Touched (Phase 1)

| File | Change |
|------|--------|
| `scripts/cage-greeter-session.sh` | **NEW** |
| `scripts/verify-cage-greeter.sh` | **NEW** |
| `docs-src/compositors/cage.md` | **NEW** |
| `docs/superpowers/specs/2026-06-18-cage-compositor-design.md` | **NEW** |
| `docs/superpowers/plans/2026-06-18-cage-compositor-plan.md` | **NEW** |
| `cmd/installer/main.go` | compositor menu + configureGreetd |
| `flake.nix` | cage enum + package option |
| `mkdocs.yml` | nav entry |
| `nfpm.yaml` | ship launcher script (if applicable) |

| `scripts/postinstall.sh` | cage-first auto-detect + hyprland deprecation warning |
| `README.md` | cage variant + deprecation notice |
| `docs-src/getting-started/installation.md` | cage + migration |
| `docs-src/compositors/hyprland.md` | deprecation banner |

**Unchanged:** `cmd/sysc-greet/*` (backend-neutral)
