# Issue #53 — Void / Chimera / BSD Support — Investigation

> **Issue:** [#53 add void and chimera support](https://github.com/Nomadcxx/sysc-greet/issues/53)  
> **Status:** Investigation complete — Void is viable first target; BSD is harder  
> **Date:** 2026-06-18

## Request

Original reporter uses **Void Linux** and plans to move to **Chimera**. They want the install script to support these distros. Owner previously noted interest in Void and BSD/GhostBSD support.

## Executive summary

| Platform | Verdict | First usable path | Effort |
|----------|---------|-------------------|--------|
| **Void Linux** | ✅ **Yes — recommended Phase 1** | Manual install today; installer Phase 1 | Medium |
| **Chimera Linux** | ✅ Yes — Phase 2 | Manual install; shares apk-like tooling | Medium |
| **GhostBSD / FreeBSD** | ⚠️ Partial — Phase 3 | Manual install; greetd port very new | High |

**Bottom line:** We can get Void working without huge rewrites. The blockers are **init system** (not package names) and **gSlapper absence** (Cage Lite from PR #78 mitigates this). Chimera is structurally similar. GhostBSD is viable for sway/cage but greetd packaging is immature and paths differ.

---

## Current installer assumptions (blockers)

The Go installer (`cmd/installer/main.go`) is built for **systemd Linux**:

| Assumption | Void | Chimera | FreeBSD/GhostBSD |
|------------|------|---------|------------------|
| Package manager | ❌ `xbps` not detected | ❌ `apk` not detected | ❌ `pkg` not detected |
| Init / service | ❌ requires `systemctl` | ❌ uses **dinit** | ❌ uses **rc.d** / `service` |
| greetd user | `greeter` | `greeter` (manual) | `greeter` (manual) |
| greetd package user | **`_greeter`** (Void official) | TBD | N/A (new port) |
| Greeter home | `/var/lib/greeter` | `/var/lib/greeter` | `/var/lib/greeter` |
| Binary path | `/usr/local/bin/sysc-greet` | `/usr/bin` typical | `/usr/local/bin` |
| polkit shutdown rules | `/etc/polkit-1/rules.d/` | likely similar | different stack |
| gSlapper build | ❌ no Void template | unknown | unlikely packaged |

**Hardest gap:** `checkDependencies()` fails without `systemd` — Void/Chimera/BSD installs die before compositor selection.

---

## Dependency availability matrix

### Void Linux (xbps — all in official repos unless noted)

| Package | Void repo | Notes |
|---------|-----------|-------|
| greetd 0.10.3 | ✅ | runit service (`/etc/sv/greetd`); user **`_greeter`** |
| kitty 0.47.1 | ✅ | |
| cage 0.3.0 | ✅ | pulls `xorg-server-xwayland` |
| niri 25.11 | ✅ | **Avoid `niri --session` on Void** (known lockup) |
| sway | ✅ | |
| hyprland | ❓ | not verified in void-packages search |
| gSlapper | ❌ | **not in void-packages** — build from source or skip |
| go | ✅ | |
| polkit | ✅ | |
| seatd | ✅ | |

### Chimera Linux (apk — main + user repos)

| Package | Chimera | Notes |
|---------|---------|-------|
| greetd 0.10.3 | ✅ user repo | `apk add greetd` |
| kitty 0.46.2 | ✅ user repo | |
| sway 1.12 | ✅ main | |
| cage | ❓ | verify in cports |
| niri | ❓ | verify |
| gSlapper | ❓ | likely absent |
| init | **dinit** | `dinitctl enable greetd` |

### GhostBSD / FreeBSD (pkg)

| Package | FreeBSD ports | Notes |
|---------|---------------|-------|
| greetd | ⚠️ **new port Jan 2026** (bug 292588) | may not be in GhostBSD stable yet |
| kitty 0.47 | ✅ x11/kitty | quarterly repo may lag — use `latest` |
| cage 0.3.0 | ✅ x11-wm/cage | |
| sway | ✅ | handbook documents greetd + tuigreet pattern |
| hyprland | ✅ | community installs exist |
| seatd | ✅ required | `sysrc seatd_enable=YES` |
| gSlapper | ❌ | not found in ports |
| init | **rc.d** | no systemd |

---

## Void-specific gotchas

### 1. greetd system account naming

Void's greetd package creates:

- User: **`_greeter`** (not `greeter`)
- Home: **`/var/lib/_greeter`**
- Service: runit via `ln -s /etc/sv/greetd /var/service/`

sysc-greet installer creates **`greeter`** at `/var/lib/greeter`. Both can coexist if we own the greetd config, but we must pick one consistently in `config.toml` and launcher scripts.

**Recommendation:** On Void, detect `_greeter` from greetd package and use it OR document using sysc-greet's `greeter` user and skip void's default config.

### 2. runit, not systemd

```bash
# Enable greetd on Void
ln -s /etc/sv/greetd /var/service/

# Disable conflicting agetty on greeter VT (from void greetd README)
rm /var/service/agetty-tty1   # if tty1 used in config.toml
```

Installer needs `enableService()` branch for runit.

### 3. niri `--session` flag

Void users report niri 25.08+ lockups when launched with `--session`. sysc-greet niri greeter config correctly uses `niri -c ...` without `--session`. ✅ No change needed, but document for Void niri desktop sessions separately.

### 4. No gSlapper

Void has no gslapper package. **Cage Lite** (PR #78) is the natural Void default:

```bash
xbps-install -S greetd cage kitty go git meson ninja
# build sysc-greet, SYSC_COMPOSITOR=cage manual install
```

For niri/sway on Void, gslapper must be built from source (optional task) or wallpapers limited to TUI effects.

### 5. musl vs glibc

Void supports both. Go binary is generally portable. Kitty/cage/niri have musl-specific build notes in void templates — use distro packages, don't bundle.

---

## Chimera-specific notes

- Package manager: `apk` (v3 — Chimera fork, Alpine-compatible CLI)
- greetd in **`user`** repo — enable `[user]` in `/etc/apk/repositories`
- Service: **dinit** (`dinitctl enable greetd` per Chimera docs)
- Reporter's planned migration path: Void → Chimera means **apk support covers both** long-term
- Chimera uses elogind (sway depends on `libelogind` — visible in package deps)

---

## GhostBSD / FreeBSD notes

- **seatd** must be running before Wayland greeter
- Paths: `/usr/local/bin` vs `/usr/bin` — configs need substitution (like Nix flake already does)
- **greetd port is weeks old** — GhostBSD may need ports build or wait for package
- **quarterly vs latest** pkg repos — kitty/cage may be missing on quarterly (document `pkg meta -l latest`)
- Polkit / shutdown: may need different approach than Linux polkit rules
- No `render` group — FreeBSD uses different DRM permission model

**Realistic GhostBSD Phase 3 deliverable:** documented manual install with sway or cage, not full installer automation initially.

---

## Recommended approach

### Phase 1 — Void Linux (highest ROI, satisfies #53 reporter)

**Goal:** `xbps-install` path + runit service + Cage Lite default

1. Add `xbps` to `detectPackageManager()` (check `/usr/bin/xbps-install`)
2. Add `initSystem` detection: `systemd` | `runit` | `dinit` | `rc`
3. Relax `checkDependencies()` — require greetd-capable init, not systemd specifically
4. `installGreetd/kitty/cage/niri/sway` cases for `xbps-install -Sy`
5. `enableService()` runit branch: symlink `/etc/sv/greetd` → `/var/service/`
6. Void greeter user detection: `_greeter` vs `greeter`
7. Skip gslapper install on Void (or optional source build task)
8. Docs: `docs-src/getting-started/void-linux.md`
9. `scripts/verify-void-greeter.sh`

**Manual install recipe (works today without code changes):**

```bash
# Void — Cage Lite manual path
sudo xbps-install -Sy greetd cage kitty git go
git clone https://github.com/Nomadcxx/sysc-greet.git
cd sysc-greet
git checkout feat/cage-compositor-investigation   # or master once merged
go build -o sysc-greet ./cmd/sysc-greet/
go build -o install-sysc-greet ./cmd/installer/

# Installer will fail on systemd check — workaround: manual config
sudo install -Dm755 sysc-greet /usr/local/bin/sysc-greet
sudo install -Dm755 scripts/cage-greeter-session.sh /etc/greetd/
sudo install -Dm644 config/kitty-greeter.conf /etc/greetd/kitty.conf
sudo cp -r ascii_configs fonts wallpapers /usr/share/sysc-greet/

# greetd config — use _greeter if void package installed, else create greeter
sudo tee /etc/greetd/config.toml <<'EOF'
[terminal]
vt = 1

[default_session]
command = "cage -s -m extend -- /etc/greetd/cage-greeter-session.sh"
user = "_greeter"
EOF

sudo ln -sf /etc/sv/greetd /var/service/
sudo rm -f /var/service/agetty-tty1   # if using vt=1
```

### Phase 2 — Chimera Linux

- Add `apk` detection (careful: distinguish Chimera apk from Alpine in `/etc/os-release`)
- dinit service enable
- user repo for greetd/kitty
- Reuse most Void logic (non-systemd, no gslapper default)

### Phase 3 — GhostBSD / FreeBSD

- Add `pkg` detection
- seatd + rc.d enable
- Path substitution in configs (`/usr/local/bin`)
- greetd from ports when available in GhostBSD repos
- Document quarterly vs latest repo requirement

---

## Relationship to Cage Lite (PR #78)

Cage Lite is a **strong enabler** for Void/BSD:

| Feature | niri/sway on Void | cage on Void |
|---------|-------------------|--------------|
| gSlapper wallpapers | needs source build | not needed |
| Compositor config file | required | launcher script only |
| Package count | higher | greetd + cage + kitty |

**Recommend Void/BSD default to cage** in new platform docs, same as Linux deprecation direction.

---

## What NOT to do

- Don't claim full Chimera support in Phase 1 — reporter is on Void now
- Don't bundle gslapper build into installer initially — high failure rate across distros
- Don't assume GhostBSD has greetd in pkg yet — verify per release
- Don't use `niri --session` anywhere in Void docs

---

## Success criteria

### Phase 1 (Void)
- [ ] `SYSC_COMPOSITOR=cage ./install.sh` works on Void without systemd
- [ ] greetd starts via runit after reboot
- [ ] Login succeeds; TUI backgrounds work
- [ ] Issue #53 reporter can install without manual config hacking

### Phase 2 (Chimera)
- [ ] Same flow with `apk` + dinit

### Phase 3 (GhostBSD)
- [ ] Documented manual path verified on GhostBSD 24.x+
- [ ] Installer automation optional follow-up

---

## Open questions

1. Use Void's `_greeter` or sysc-greet's `greeter` on Void? (Recommend: detect and adapt)
2. Submit gslapper template to void-packages? (Upstream contribution — separate effort)
3. Chimera in Phase 1 or Phase 2? (Recommend Phase 2 — same apk code, less test surface)
4. AUR-style `sysc-greet-void` XBPS template? (Phase 1.5 packaging)

## Suggested issue #53 triage comment

> Investigation complete. **Void is viable** as Phase 1 (xbps + runit + Cage Lite). Chimera Phase 2 (apk + dinit). GhostBSD/FreeBSD Phase 3 (pkg + rc.d, greetd port very new). Main blockers are systemd-only installer and missing gslapper on Void — Cage path from #78 mitigates the latter. Design spec: `docs/superpowers/specs/2026-06-18-issue-53-void-chimera-bsd-design.md`
