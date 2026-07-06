# Cagebreak Compositor Pivot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Cage Lite greeter path in PR #78 with Cagebreak, giving full gSlapper wallpaper parity, so Hyprland greeter support can be deprecated.

**Architecture:** Config-driven, mirroring the niri pattern: greetd runs `cagebreak -e -c /etc/greetd/cagebreak-greeter-config`; the config `exec`s the gSlapper wallpaper daemon (layer-shell) and the kitty→sysc-greet chain, which sends `quit` to cagebreak's IPC socket after login. No Go application changes to sysc-greet itself; installer/flake/packaging/docs swap cage→cagebreak.

**Tech Stack:** Go (installer), POSIX sh, cagebreak 3.2.1, greetd, gSlapper, Nix flake, nfpm.

**Spec:** `docs/superpowers/specs/2026-07-06-cagebreak-compositor-design.md`

## Global Constraints

- Branch: `feat/cage-compositor-investigation` in worktree `.claude/worktrees/cagebreak-pivot`. Merge target is `development`, NOT `master`. Never merge or push without explicit user approval.
- Commit style per CLAUDE.md: brief, action-focused, NO Co-Authored-By / AI attribution, explicit file paths in `git add` (never `git add .`).
- greetd command everywhere: `cagebreak -e -c /etc/greetd/cagebreak-greeter-config`
- Greeter config path everywhere: `/etc/greetd/cagebreak-greeter-config` (mode 0644)
- Cage (the compositor) is fully removed from the PR: `scripts/cage-greeter-session.sh` and `docs-src/compositors/cage.md` deleted, all `"cage"` cases replaced by `"cagebreak"`.
- Tasks 3+ are GATED on Task 2 (VT spike) passing its P0 checks. If the wallpaper check fails, STOP and consult the user.
- Deprecation copy (use verbatim where a menu/warning mentions it): Hyprland greeter support ending in ~3 months; migrate to cagebreak.

---

### Task 1: Headless spike (SSH-safe, automated)

Verifies over SSH, without a display: cagebreak parses our config, `$CAGEBREAK_SOCKET` propagates to `exec` children, gSlapper (layer-shell client) starts and stays alive, and `quit` over the socket shuts cagebreak down cleanly.

**Files:**
- Create: `/tmp/claude-1000/-home-nomadx-Documents-sysc-greet-dev/1cb7821e-6425-4a7f-a87c-74180e296830/scratchpad/spike/config` (throwaway, not committed)
- Create: `/tmp/claude-1000/-home-nomadx-Documents-sysc-greet-dev/1cb7821e-6425-4a7f-a87c-74180e296830/scratchpad/spike/run.sh` (throwaway, not committed)

**Interfaces:**
- Produces: a pass/fail result for `SOCKET_PROPAGATES` (drives the quit mechanism in Task 3) and `GSLAPPER_ALIVE` (early layer-shell signal before the VT test).

- [ ] **Step 1: Write the spike config**

`$SCRATCHPAD/spike/config` (use the scratchpad path above; `$SCRATCHPAD` below means that directory):

```
background 0.1 0.1 0.1
exec env > /tmp/cagebreak-spike-env.txt
exec gslapper -f '*' /home/nomadx/Documents/sysc-greet-dev/.claude/worktrees/cagebreak-pivot/wallpapers/sysc-greet-example.png 2> /tmp/cagebreak-spike-gslapper.log
```

Note: if `wallpapers/sysc-greet-example.png` does not exist in the worktree, use any PNG on disk (e.g. one from `/usr/share/sysc-greet/wallpapers/`); adjust the path in the config.

- [ ] **Step 2: Write the runner script**

`$SCRATCHPAD/spike/run.sh`:

```bash
#!/usr/bin/env bash
set -u
SPIKE_DIR="$(cd "$(dirname "$0")" && pwd)"
rm -f /tmp/cagebreak-spike-env.txt /tmp/cagebreak-spike-gslapper.log

WLR_BACKENDS=headless WLR_LIBINPUT_NO_DEVICES=1 WLR_RENDERER=pixman \
  cagebreak -e -c "${SPIKE_DIR}/config" &
CB_PID=$!

sleep 3

FAIL=0

# 1. cagebreak still running?
if kill -0 "$CB_PID" 2>/dev/null; then echo "PASS cagebreak running"; else echo "FAIL cagebreak exited early"; FAIL=1; fi

# 2. CAGEBREAK_SOCKET propagated to exec children?
if grep -q '^CAGEBREAK_SOCKET=' /tmp/cagebreak-spike-env.txt 2>/dev/null; then
  echo "PASS SOCKET_PROPAGATES"
  SOCK=$(grep '^CAGEBREAK_SOCKET=' /tmp/cagebreak-spike-env.txt | cut -d= -f2)
else
  echo "FAIL SOCKET_PROPAGATES (env dump: $(cat /tmp/cagebreak-spike-env.txt 2>/dev/null | head -c 200))"
  SOCK=""
  FAIL=1
fi

# 3. gslapper (layer-shell client) still alive?
if pgrep -x gslapper >/dev/null; then echo "PASS GSLAPPER_ALIVE"; else echo "FAIL GSLAPPER_ALIVE"; cat /tmp/cagebreak-spike-gslapper.log 2>/dev/null; FAIL=1; fi

# 4. quit via socket
if [ -n "$SOCK" ] && [ -S "$SOCK" ]; then
  echo quit | socat - UNIX-CONNECT:"$SOCK"
  sleep 2
  if kill -0 "$CB_PID" 2>/dev/null; then echo "FAIL quit-via-socket (still running)"; FAIL=1; else echo "PASS quit-via-socket"; fi
fi

# Cleanup whatever is left
kill "$CB_PID" 2>/dev/null; pkill -x gslapper 2>/dev/null
exit "$FAIL"
```

- [ ] **Step 3: Run it**

Run: `bash $SCRATCHPAD/spike/run.sh`
Expected: all four `PASS` lines, exit 0.

- If `FAIL cagebreak exited early`: read cagebreak's stderr (rerun without `&`, capture output). Common causes: `XDG_RUNTIME_DIR` unset, config parse error (cagebreak is strict — one command per line).
- If `WLR_RENDERER=pixman` fails, retry without it (allow GLES on the server GPU).
- If `FAIL SOCKET_PROPAGATES`: record it — Task 3 then uses the `pkill` quit variant.
- If `FAIL GSLAPPER_ALIVE`: read `/tmp/cagebreak-spike-gslapper.log`. If it shows a missing `zwlr_layer_shell_v1` global, the pivot's premise is broken — STOP, report to user (fall back to Cage Lite per spec).

- [ ] **Step 4: Record results**

Note the outcomes of `SOCKET_PROPAGATES` and `GSLAPPER_ALIVE` in the task report. Nothing to commit (spike artifacts are scratchpad-only).

---

### Task 2: VT spike on real hardware (USER-ASSISTED GATE)

Visual confirmation on a real DRM session: wallpaper renders *behind* fullscreen kitty running sysc-greet, keyboard input works, chain-quit works.

**Files:**
- Create: `/tmp/cagebreak-vt-spike/config` (throwaway, on-disk where a TTY shell can reach it)

**Interfaces:**
- Consumes: Task 1's `SOCKET_PROPAGATES` result (pick the quit line accordingly).
- Produces: the go/no-go decision for Tasks 3–10.

- [ ] **Step 1: Build sysc-greet and stage the VT spike config**

```bash
cd /home/nomadx/Documents/sysc-greet-dev/.claude/worktrees/cagebreak-pivot
mkdir -p /tmp/cagebreak-vt-spike
go build -o /tmp/cagebreak-vt-spike/sysc-greet ./cmd/sysc-greet/
```

Write `/tmp/cagebreak-vt-spike/config` (socket variant — swap the trailing `; echo quit | socat ...` for `; pkill cagebreak` if Task 1 failed SOCKET_PROPAGATES):

```
background 0.0 0.0 0.0
exec gslapper -f '*' /usr/share/sysc-greet/wallpapers/sysc-greet-example.png
exec kitty --start-as=fullscreen -o background_opacity=0.6 /tmp/cagebreak-vt-spike/sysc-greet --test; echo quit | socat - UNIX-CONNECT:"$CAGEBREAK_SOCKET"
```

(Adjust the PNG path to any wallpaper that exists on the machine. `background_opacity=0.6` makes the layer-shell wallpaper visibly show through — that's the P0 check.)

- [ ] **Step 2: Hand off to the user**

Ask the user to run, from a physical TTY (not SSH):

```bash
cagebreak -e -c /tmp/cagebreak-vt-spike/config
```

And confirm:
1. **P0:** wallpaper image visible behind the translucent kitty/sysc-greet UI
2. **P0:** keyboard input works in sysc-greet (arrow keys, typing)
3. **P0:** quitting sysc-greet (Ctrl+C in test mode) exits kitty AND cagebreak, returning to the TTY — no hang, no loop
4. **P1:** `XKB_DEFAULT_LAYOUT=de cagebreak -e -c ...` gives a German layout (optional)

- [ ] **Step 3: GATE — wait for user confirmation**

Do not proceed to Task 3 until the user reports the P0 results. If wallpaper fails: STOP, report, fall back to Cage Lite per spec.

---

### Task 3: Greeter config file; delete cage launcher

**Files:**
- Create: `config/cagebreak-greeter-config`
- Delete: `scripts/cage-greeter-session.sh`

**Interfaces:**
- Produces: `/etc/greetd/cagebreak-greeter-config` content consumed verbatim by Task 4 (installer embeds the same content) and shipped by Tasks 5–6.

- [ ] **Step 1: Write `config/cagebreak-greeter-config`**

Socket-quit variant (default; if Task 1 failed SOCKET_PROPAGATES, replace everything after `/usr/local/bin/sysc-greet` on the kitty line with `; pkill cagebreak` and drop the socat mention in the comment):

```
# sysc-greet cagebreak greeter config — used ONLY by the greetd greeter session
# greetd runs: cagebreak -e -c /etc/greetd/cagebreak-greeter-config
# -e enables the IPC socket used to quit cagebreak after login (requires socat)
# Keyboard layout: set XKB_DEFAULT_LAYOUT/XKB_DEFAULT_VARIANT in greetd's command env

background 0.0 0.0 0.0

# gSlapper wallpaper daemon (wlr-layer-shell client, same as niri/sway setup)
exec HOME=/var/lib/greeter /usr/local/bin/sysc-greet --wallpaper-daemon

# Greeter UI; when it exits (login or shutdown), quit cagebreak so greetd can start the session
exec XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; echo quit | socat - UNIX-CONNECT:"$CAGEBREAK_SOCKET"

# No bind/definekey lines on purpose: no compositor keybindings reachable from the greeter
```

- [ ] **Step 2: Validate the config parses**

Run: `WLR_BACKENDS=headless WLR_LIBINPUT_NO_DEVICES=1 timeout 5 cagebreak -e -c config/cagebreak-greeter-config; echo "exit=$?"`
Expected: cagebreak starts (timeout kills it, exit=124) with no config parse errors on stderr. The `exec` lines will log failures for `/usr/local/bin/sysc-greet` if not installed — that's fine; only parse errors matter.

- [ ] **Step 3: Delete the cage launcher**

```bash
git rm scripts/cage-greeter-session.sh
```

- [ ] **Step 4: Commit**

```bash
git add config/cagebreak-greeter-config
git commit -m "Add cagebreak greeter config, drop cage launcher"
```

---

### Task 4: Installer swap cage → cagebreak

**Files:**
- Modify: `cmd/installer/main.go` (comment at ~line 226, compositor list ~line 376, validation map ~line 381 and ~line 959, menu render ~line 564, pacman/apt/dnf install cases ~lines 1003–1035, `configureGreetd` cage case ~line 1628, config write mode ~line 1648, `removeConfigs` ~line 1737)

**Interfaces:**
- Consumes: config content from Task 3 (embedded verbatim, socket or pkill variant to match).
- Produces: installer option string `"cagebreak"`, greetd command `cagebreak -e -c /etc/greetd/cagebreak-greeter-config`.

- [ ] **Step 1: Replace identifiers and menu entries**

In `cmd/installer/main.go`:

1. Struct comment: `// "niri", "hyprland", "sway", or "cage"` → `// "cagebreak", "niri", "sway", or "hyprland"`
2. Selection list: `compositors := []string{"cage", "niri", "sway", "hyprland"}` → `compositors := []string{"cagebreak", "niri", "sway", "hyprland"}`
3. Both validation maps (`Update` and `installCompositor`): `"cage": {"cage"},` → `"cagebreak": {"cagebreak"},`
4. Menu render block:

```go
	compositors := []struct {
		name string
		desc string
	}{
		{"cagebreak", "Recommended — minimal tiling kiosk; full gSlapper wallpapers"},
		{"niri", "Tiling compositor with scrollable workspaces + gSlapper wallpapers"},
		{"sway", "Stable i3-compatible tiling compositor + gSlapper wallpapers"},
		{"hyprland", "Deprecated — greeter support ending in ~3 months; migrate to cagebreak"},
	}
```

5. Footer line: `"Hyprland greeter support is being phased out — cage is the recommended path"` → `"Hyprland greeter support is being phased out — cagebreak is the recommended path"` and the index-3 warning text `Use cage or niri instead.` → `Use cagebreak or niri instead.`

- [ ] **Step 2: Package-manager install cases**

pacman case — replace:

```go
		case "cage":
			cmd = exec.Command("pacman", "-S", "--noconfirm", "cage")
```

with:

```go
		case "cagebreak":
			return installCagebreakArch(m)
```

apt and dnf cases — replace the two `case "cage":` returns with:

```go
		case "cagebreak":
			return fmt.Errorf("cagebreak not in standard repos — build from https://github.com/project-repo/cagebreak")
```

Add this function near `installCompositor`:

```go
// installCagebreakArch installs cagebreak on Arch: official repos first, then
// an AUR helper run as the invoking user (AUR helpers refuse to run as root).
func installCagebreakArch(m *model) error {
	if err := exec.Command("pacman", "-S", "--noconfirm", "cagebreak").Run(); err == nil {
		return nil
	}
	sudoUser := os.Getenv("SUDO_USER")
	for _, helper := range []string{"paru", "yay"} {
		helperPath, err := exec.LookPath(helper)
		if err != nil {
			continue
		}
		var cmd *exec.Cmd
		if sudoUser != "" {
			cmd = exec.Command("sudo", "-u", sudoUser, helperPath, "-S", "--noconfirm", "cagebreak")
		} else {
			cmd = exec.Command(helperPath, "-S", "--noconfirm", "cagebreak")
		}
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("cagebreak is AUR-only — install it first: paru -S cagebreak (or yay -S cagebreak)")
}
```

- [ ] **Step 3: `configureGreetd` — replace the cage case**

Replace the whole `case "cage":` block with (config string must byte-match Task 3's file):

```go
	case "cagebreak":
		compositorConfig = `# sysc-greet cagebreak greeter config — used ONLY by the greetd greeter session
# greetd runs: cagebreak -e -c /etc/greetd/cagebreak-greeter-config
# -e enables the IPC socket used to quit cagebreak after login (requires socat)
# Keyboard layout: set XKB_DEFAULT_LAYOUT/XKB_DEFAULT_VARIANT in greetd's command env

background 0.0 0.0 0.0

# gSlapper wallpaper daemon (wlr-layer-shell client, same as niri/sway setup)
exec HOME=/var/lib/greeter /usr/local/bin/sysc-greet --wallpaper-daemon

# Greeter UI; when it exits (login or shutdown), quit cagebreak so greetd can start the session
exec XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; echo quit | socat - UNIX-CONNECT:"$CAGEBREAK_SOCKET"

# No bind/definekey lines on purpose: no compositor keybindings reachable from the greeter
`
		configPath = "/etc/greetd/cagebreak-greeter-config"
		greetdCommand = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
```

Then remove the cage 0755 special case — restore:

```go
	// Write compositor config
	if err := os.WriteFile(configPath, []byte(compositorConfig), 0644); err != nil {
		return fmt.Errorf("compositor config write failed")
	}
```

- [ ] **Step 4: Uninstall cleanup**

In `removeConfigs`, replace `"/etc/greetd/cage-greeter-session.sh",` with `"/etc/greetd/cagebreak-greeter-config",`.

- [ ] **Step 5: Build and vet**

Run: `go build ./... && go vet ./cmd/installer/`
Expected: clean build, no vet errors. Also run `grep -rn '"cage"' cmd/installer/main.go` — expected: no matches.

- [ ] **Step 6: Commit**

```bash
git add cmd/installer/main.go
git commit -m "Switch installer greeter backend from cage to cagebreak"
```

---

### Task 5: Nix flake swap cage → cagebreak

**Files:**
- Modify: `flake.nix` (package build script cage block ~line 90, `compositorPackage` ~line 148, `defaultCompositorCommand` ~line 157, enum ~line 170, `cagePackage` option ~line 211, `environment.etc` ~line 272)

**Interfaces:**
- Consumes: `config/cagebreak-greeter-config` from Task 3.
- Produces: NixOS options `compositor = "cagebreak"` and `cagebreakPackage`.

- [ ] **Step 1: Package build — replace the cage launcher block**

Replace:

```nix
            # Cage lite launcher (single-client kiosk — no gSlapper)
            cp scripts/cage-greeter-session.sh $out/etc/greetd/
            substituteInPlace $out/etc/greetd/cage-greeter-session.sh \
              --replace '/usr/local/bin/sysc-greet' "$out/bin/sysc-greet" \
              --replace 'kitty ' "${pkgs.kitty}/bin/kitty "
            chmod +x $out/etc/greetd/cage-greeter-session.sh
```

with:

```nix
            # Cagebreak greeter config (gSlapper layer-shell wallpapers supported)
            cp config/cagebreak-greeter-config $out/etc/greetd/
            substituteInPlace $out/etc/greetd/cagebreak-greeter-config \
              --replace '/usr/local/bin/sysc-greet' "$out/bin/sysc-greet" \
              --replace 'kitty ' "${pkgs.kitty}/bin/kitty " \
              --replace 'socat ' "${pkgs.socat}/bin/socat "
```

- [ ] **Step 2: Module options**

1. `compositorPackage`: `else if cfg.compositor == "cage" then cfg.cagePackage` → `else if cfg.compositor == "cagebreak" then cfg.cagebreakPackage`
2. `defaultCompositorCommand`:

```nix
            else if cfg.compositor == "cagebreak" then
              "${compositorExecutable cfg.cagebreakPackage "cagebreak"} -e -c /etc/greetd/cagebreak-greeter-config"
```

3. Enum: `types.enum [ "niri" "hyprland" "sway" "cage" ]` → `types.enum [ "niri" "hyprland" "sway" "cagebreak" ]`, description: `Use cagebreak for a minimal tiling kiosk greeter with full gSlapper wallpaper support (recommended).`
4. Rename option `cagePackage` → `cagebreakPackage`:

```nix
            cagebreakPackage = mkOption {
              type = types.nullOr types.package;
              default = null;
              defaultText = literalExpression "null";
              description = "cagebreak package to use and install for the greeter compositor. When null, the cagebreak command is resolved from PATH.";
            };
```

5. `environment.etc`: `"greetd/cage-greeter-session.sh".source = ...cage-greeter-session.sh";` → `"greetd/cagebreak-greeter-config".source = "${package}/etc/greetd/cagebreak-greeter-config";`

- [ ] **Step 3: Validate**

Run: `command -v nix >/dev/null && nix flake check --no-build 2>&1 | tail -5 || nix-instantiate --parse flake.nix >/dev/null 2>&1 || echo "nix unavailable — grep fallback"`
If nix is unavailable, verify by inspection: `grep -n 'cage' flake.nix` — expected: only `cagebreak` matches, no bare-cage references.

- [ ] **Step 4: Commit**

```bash
git add flake.nix
git commit -m "Use cagebreak in Nix module"
```

---

### Task 6: Packaging (nfpm) + postinstall auto-detect

**Files:**
- Modify: `nfpm.yaml` (description ~line 10, cage launcher entry ~line 61)
- Modify: `scripts/postinstall.sh` (auto-detect block ~lines 26–54)

**Interfaces:**
- Consumes: greetd command string from Global Constraints.

- [ ] **Step 1: nfpm.yaml**

Description: `Supports Niri, Hyprland, Sway, and experimental Cage compositors.` → `Supports Cagebreak, Niri, and Sway compositors (Hyprland deprecated).`

Replace:

```yaml
  # Cage lite launcher (experimental)
  - src: ./scripts/cage-greeter-session.sh
    dst: /etc/greetd/cage-greeter-session.sh
    file_info:
      mode: 0755
```

with:

```yaml
  # Cagebreak greeter config
  - src: ./config/cagebreak-greeter-config
    dst: /etc/greetd/cagebreak-greeter-config
    file_info:
      mode: 0644
```

- [ ] **Step 2: postinstall.sh auto-detect**

Replace the cage branch:

```sh
if command -v cage &>/dev/null; then
    COMPOSITOR="cage"
    GREETD_COMMAND="cage -s -m extend -- /etc/greetd/cage-greeter-session.sh"
```

with:

```sh
if command -v cagebreak &>/dev/null; then
    COMPOSITOR="cagebreak"
    GREETD_COMMAND="cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
```

Update the two user-facing strings: `(cage, niri, sway, hyprland)` → `(cagebreak, niri, sway, hyprland)`, `Please install cage (recommended)` → `Please install cagebreak (recommended)`, and in the hyprland warning `Migrate to cage: SYSC_COMPOSITOR=cage sudo ./install.sh` → `Migrate to cagebreak: SYSC_COMPOSITOR=cagebreak sudo ./install.sh` plus doc URL `compositors/cage/` → `compositors/cagebreak/`.

- [ ] **Step 3: Validate**

Run: `bash -n scripts/postinstall.sh && grep -c cagebreak scripts/postinstall.sh nfpm.yaml`
Expected: syntax OK; cagebreak referenced in both files; `grep -n '\bcage\b' scripts/postinstall.sh nfpm.yaml` returns nothing.

- [ ] **Step 4: Commit**

```bash
git add nfpm.yaml scripts/postinstall.sh
git commit -m "Package cagebreak greeter config and prefer it in postinstall"
```

---

### Task 7: Verify script

**Files:**
- Create: `scripts/verify-cagebreak-greeter.sh`
- Delete: `scripts/verify-cage-greeter.sh`

- [ ] **Step 1: Write `scripts/verify-cagebreak-greeter.sh`**

```bash
#!/usr/bin/env bash
# verify-cagebreak-greeter.sh — smoke checks for the cagebreak greeter integration
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG="${ROOT}/config/cagebreak-greeter-config"
FAIL=0

pass() { echo "✓ $*"; }
fail() { echo "✗ $*"; FAIL=1; }

echo "=== Cagebreak greeter verification ==="

if [[ -f "${CONFIG}" ]]; then
  pass "greeter config exists: ${CONFIG}"
else
  fail "greeter config missing: ${CONFIG}"
fi

if command -v cagebreak >/dev/null 2>&1; then
  pass "cagebreak binary found: $(command -v cagebreak) ($(cagebreak -v 2>&1))"
  if WLR_BACKENDS=headless WLR_LIBINPUT_NO_DEVICES=1 timeout 5 cagebreak -e -c "${CONFIG}" >/dev/null 2>&1 || [[ $? -eq 124 ]]; then
    pass "config parses under headless cagebreak"
  else
    fail "cagebreak rejected the greeter config"
  fi
else
  echo "⚠ cagebreak not installed — skip runtime test (Arch: paru -S cagebreak)"
fi

if command -v socat >/dev/null 2>&1; then
  pass "socat available (needed for quit-via-socket)"
else
  fail "socat missing — greeter cannot quit cagebreak after login"
fi

if grep -q '"cagebreak"' "${ROOT}/cmd/installer/main.go"; then
  pass "installer references cagebreak"
else
  fail "installer does not reference cagebreak"
fi

if [[ -f "${ROOT}/docs-src/compositors/cagebreak.md" ]]; then
  pass "cagebreak docs present"
else
  fail "docs-src/compositors/cagebreak.md missing"
fi

echo
if [[ "${FAIL}" -eq 0 ]]; then
  echo "All static checks passed."
  echo "Manual: SYSC_COMPOSITOR=cagebreak sudo ./install.sh → systemctl restart greetd → test login"
else
  echo "Some checks failed."
  exit 1
fi
```

Then: `chmod +x scripts/verify-cagebreak-greeter.sh && git rm scripts/verify-cage-greeter.sh`

- [ ] **Step 2: Run it**

Run: `bash scripts/verify-cagebreak-greeter.sh`
Expected: all checks pass except the docs check (✗ until Task 8) — acceptable interim state; re-run in Task 9.

- [ ] **Step 3: Commit**

```bash
git add scripts/verify-cagebreak-greeter.sh
git commit -m "Add cagebreak verify script"
```

---

### Task 8: Documentation swap

**Files:**
- Create: `docs-src/compositors/cagebreak.md`
- Delete: `docs-src/compositors/cage.md`
- Modify: `mkdocs.yml:63`, `docs-src/compositors/hyprland.md:1-12`, `docs-src/getting-started/installation.md`, `README.md`
- Modify: `docs/superpowers/specs/2026-06-18-cage-compositor-design.md:1-5`, `docs/superpowers/plans/2026-06-18-cage-compositor-plan.md:1-5` (superseded banners)

- [ ] **Step 1: Write `docs-src/compositors/cagebreak.md`**

```markdown
# Cagebreak Setup

!!! tip "Recommended greeter backend"
    Cagebreak is the **recommended** Wayland backend for the sysc-greet greeter session. It replaces Hyprland for greetd, which is being [phased out over the next ~3 months](hyprland.md).

Unlike the earlier Cage experiment, Cagebreak supports **wlr-layer-shell**, so gSlapper boot wallpapers (static and video) work exactly as they do under niri and sway — full feature parity with the Hyprland greeter.

Tracked in [issue #69](https://github.com/Nomadcxx/sysc-greet/issues/69).

## Why Cagebreak?

- **Minimal tiling kiosk** — a Cage fork built for exactly this: a fullscreen app plus support clients
- **Full wallpapers** — gSlapper runs as a layer-shell client, same as niri/sway
- **No config churn** — a ~10 line static config vs Hyprland's window rules, layer rules, and exit chains
- **Locked down** — the greeter config defines zero keybindings

## Install Cagebreak

=== "Arch Linux"

    Cagebreak is in the AUR (not the official repos):

    ```bash
    paru -S cagebreak   # or: yay -S cagebreak
    ```

    Also install socat (used to quit the compositor after login):

    ```bash
    sudo pacman -S socat
    ```

=== "NixOS"

    ```nix
    services.sysc-greet = {
      enable = true;
      compositor = "cagebreak";
      # optional: cagebreakPackage = pkgs.cagebreak;
    };
    ```

=== "Other distros"

    Check your package manager or build from [project-repo/cagebreak](https://github.com/project-repo/cagebreak).

## greetd Config

=== "Installer"

    ```bash
    SYSC_COMPOSITOR=cagebreak sudo ./install.sh
    ```

=== "Manual"

    Edit `/etc/greetd/config.toml`:

    ```toml
    [terminal]
    vt = 1

    [default_session]
    command = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
    user = "greeter"
    ```

    `-e` enables cagebreak's IPC socket; the greeter config uses it to quit the compositor after login.

The greeter config lives at `/etc/greetd/cagebreak-greeter-config` and is installed by the installer or package.

## Keyboard Layout

Cagebreak uses the standard XKB environment variables. For a non-US layout, wrap the greetd command:

```toml
command = "env XKB_DEFAULT_LAYOUT=de XKB_DEFAULT_VARIANT=nodeadkeys cagebreak -e -c /etc/greetd/cagebreak-greeter-config"
```

## Feature Matrix

| Feature | Cagebreak | niri / sway | Hyprland (deprecated) |
|---------|-----------|-------------|-----------------------|
| Login via greetd IPC | ✅ | ✅ | ✅ |
| TUI background effects | ✅ | ✅ | ✅ |
| Static/video boot wallpapers (gSlapper) | ✅ | ✅ | ✅ |
| Non-US keyboard | ✅ XKB env | ✅ compositor config | ✅ compositor config |
| Greeter config size | ~10 lines | ~80 lines | ~100+ lines |

## Migrating from Hyprland

1. Install cagebreak (see above)
2. Re-run the installer: `SYSC_COMPOSITOR=cagebreak sudo ./install.sh`
3. Reboot (or `sudo systemctl restart greetd` from a TTY)

Your daily Hyprland desktop session is unaffected — only the boot greeter changes.

## Troubleshooting

- **Black screen, no wallpaper:** check gSlapper is installed and `/var/cache/sysc-greet/` is readable by the greeter user
- **Greeter restarts in a loop:** run `cagebreak -e -c /etc/greetd/cagebreak-greeter-config` from a TTY as your user to see the error; a config parse error makes cagebreak exit immediately and greetd relaunches it
- **Stuck after login:** verify `socat` is installed — without it the compositor never quits and greetd waits forever
```

Then: `git rm docs-src/compositors/cage.md`

- [ ] **Step 2: mkdocs.yml nav**

`- Cage (experimental): compositors/cage.md` → `- Cagebreak: compositors/cagebreak.md`

- [ ] **Step 3: Cross-references in hyprland.md / installation.md / README.md**

All in the deprecation copy added by the cage draft — swap cage → cagebreak:

- `docs-src/compositors/hyprland.md`: `**Migrate to [Cage](cage.md)** (recommended — faster, simpler)` → `**Migrate to [Cagebreak](cagebreak.md)** (recommended — full wallpaper parity)`
- `docs-src/getting-started/installation.md`: `(cage recommended, niri, sway, or hyprland deprecated)` → `(cagebreak recommended, niri, sway, or hyprland deprecated)`; `Migrate to [Cage](../compositors/cage.md)` → `Migrate to [Cagebreak](../compositors/cagebreak.md)`; prerequisites line `cage (recommended), niri, sway` → `cagebreak (recommended), niri, sway`; the AUR block

  ```
  # Cage (recommended): install cage from repos, then:
  sudo pacman -S cage
  SYSC_COMPOSITOR=cage curl -fsSL .../install.sh | sudo bash
  ```

  becomes

  ```
  # Cagebreak (recommended): install from AUR, then:
  paru -S cagebreak && sudo pacman -S socat
  SYSC_COMPOSITOR=cagebreak curl -fsSL .../install.sh | sudo bash
  ```

- `README.md`: `New installs should choose **cage** (recommended) or **niri**.` → `New installs should choose **cagebreak** (recommended) or **niri**.`; the AUR comment `# Recommended (cage — install script, or: pacman -S cage && SYSC_COMPOSITOR=cage ./install.sh)` → `# Recommended (cagebreak — AUR: paru -S cagebreak, then SYSC_COMPOSITOR=cagebreak ./install.sh)`; `(deprecated — migrate to cage)` → `(deprecated — migrate to cagebreak)`; Nix example `compositor = "cage";  # recommended` → `compositor = "cagebreak";  # recommended` and `set \`cagePackage\`, \`niriPackage\`` → `set \`cagebreakPackage\`, \`niriPackage\``; `does not install \`cage\`, \`niri\`` → `does not install \`cagebreak\`, \`niri\``

- [ ] **Step 4: Superseded banners on the June 18 docs**

Insert directly under the title of `docs/superpowers/specs/2026-06-18-cage-compositor-design.md`:

```markdown
> **SUPERSEDED (2026-07-06):** This investigation pivoted from Cage to Cagebreak (layer-shell → full gSlapper wallpaper parity). See [2026-07-06-cagebreak-compositor-design.md](2026-07-06-cagebreak-compositor-design.md).
```

And under the title of `docs/superpowers/plans/2026-06-18-cage-compositor-plan.md`:

```markdown
> **SUPERSEDED (2026-07-06):** See [2026-07-06-cagebreak-compositor-plan.md](2026-07-06-cagebreak-compositor-plan.md).
```

- [ ] **Step 5: Verify no stale references**

Run: `grep -rn --include='*.md' -i '\bcage\b' README.md docs-src/ mkdocs.yml | grep -v -i cagebreak`
Expected: no hits outside the two superseded June 18 docs (docs/superpowers is excluded from the grep paths on purpose).

- [ ] **Step 6: Commit**

```bash
git add docs-src/compositors/cagebreak.md mkdocs.yml docs-src/compositors/hyprland.md docs-src/getting-started/installation.md README.md docs/superpowers/specs/2026-06-18-cage-compositor-design.md docs/superpowers/plans/2026-06-18-cage-compositor-plan.md
git commit -m "Document cagebreak greeter backend"
```

---

### Task 9: Full verification pass

**Files:** none new.

- [ ] **Step 1: Build + tests + vet**

Run: `go build ./... && go test ./... && go vet ./...`
Expected: build OK, `ok github.com/Nomadcxx/sysc-greet/cmd/sysc-greet`, no vet errors.

- [ ] **Step 2: Verify script clean**

Run: `bash scripts/verify-cagebreak-greeter.sh`
Expected: ALL checks ✓ (docs now exist), exit 0.

- [ ] **Step 3: Whole-repo stale-cage sweep**

Run: `grep -rn '\bcage\b' --include='*.go' --include='*.sh' --include='*.yaml' --include='*.yml' --include='*.nix' cmd/ scripts/ config/ flake.nix nfpm.yaml mkdocs.yml`
Expected: no matches (cagebreak only).

- [ ] **Step 4: Commit any stragglers**

Only if steps 1–3 forced fixes; use explicit paths, message `Fix cagebreak review nits`.

---

### Task 10: Update PR #78 (user approval required to push)

**Files:** none (GitHub metadata + push).

- [ ] **Step 1: ASK THE USER before pushing**

CLAUDE.md: never push automatically. Present the commit list (`git log --oneline origin/feat/cage-compositor-investigation..HEAD`) and ask for approval to push.

- [ ] **Step 2: Push and retarget (after approval)**

```bash
git push origin feat/cage-compositor-investigation
gh pr edit 78 --repo Nomadcxx/sysc-greet --base development --title "feat: Cagebreak greeter path + Hyprland deprecation (draft)"
```

- [ ] **Step 3: Update the PR body (after approval)**

```bash
gh pr edit 78 --repo Nomadcxx/sysc-greet --body "$(cat <<'EOF'
## Summary

Implements [#69](https://github.com/Nomadcxx/sysc-greet/issues/69): add **Cagebreak** as the recommended greetd backend and begin **phasing out Hyprland** for the greeter session.

Pivoted from the earlier Cage Lite draft: Cagebreak supports **wlr-layer-shell**, so gSlapper boot wallpapers (static + video) work — **full feature parity with the Hyprland greeter**, none of Cage Lite's trade-offs.

**Hyprland is NOT removed in this PR** — it remains functional but is marked deprecated with a ~3 month removal timeline.

## What changed

- **Cagebreak path:** `cagebreak -e -c /etc/greetd/cagebreak-greeter-config` → gSlapper (layer-shell wallpapers) + Kitty → sysc-greet; compositor quits via IPC socket after login
- **Installer:** cagebreak first in menu (recommended); AUR-helper fallback on Arch; hyprland last (deprecated warning)
- **Nix:** `compositor = "cagebreak"` + `cagebreakPackage` option
- **postinstall:** prefer cagebreak → niri → sway → hyprland; warn on hyprland auto-detect
- **Docs:** `docs-src/compositors/cagebreak.md` (replaces cage.md), migration guide, deprecation banners
- **Design:** `docs/superpowers/specs/2026-07-06-cagebreak-compositor-design.md`
- **Verified on hardware:** wallpaper renders behind the greeter, login works, compositor exits cleanly (spike results in plan doc)

## Manual test checklist

- [ ] `paru -S cagebreak && sudo pacman -S socat`
- [ ] `SYSC_COMPOSITOR=cagebreak sudo ./install.sh`
- [ ] `sudo systemctl restart greetd` (from TTY)
- [ ] Greeter appears **with wallpaper**, login works
- [ ] Non-US keyboard: `XKB_DEFAULT_LAYOUT` in greetd command env
EOF
)"
```

- [ ] **Step 4: Report**

Link the PR, summarize spike results, and remind: merge target is `development`, merge only on explicit user sign-off.
