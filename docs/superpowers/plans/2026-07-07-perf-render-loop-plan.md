# Render Loop Performance — Plan

> **Status:** Implemented; awaiting visual verification on real kitty
> **Results (240x67 pty):** idle 53% → 31% active / **5.9% after 60s idle**;
> fire 84% → 44%; matrix 60% → 44%; plasma 77% → 69%. Fire render
> 5.3ms/24,459 allocs → 0.13ms/3 allocs per frame.
> **Spec:** [2026-07-07 perf audit](../audits/2026-07-07-perf-audit.md)
> **Branch:** `perf/render-loop` (worktree-isolated; merge target `development`)
> **Baseline (240x67 pty):** idle 53% CPU / 1.0 MiB/s output; fire 84%; plasma 77%

## Regression constraints (from git history)

These shaped the current code; violating them re-opens closed bugs:

1. **View() and all render functions stay pure** — no `Update()`, `Resize()`,
   or any state mutation from render paths. Value-receiver mutation from View
   caused the "hyperspeed on first boot" bug (7e1c316, issue #45). All
   changes here live in `Update()` handlers and tick scheduling.
2. **The dual tick chain is intentional.** `bgTickMsg` exists so animSpeed
   (slow 60ms / normal 30ms / fast 15ms) only affects background effects
   (ff8ca52, c5bfb7e). The 30ms UI tick was picked for typewriter-ticker
   smoothness (a02cb history note in main.go:409). Both features must behave
   identically when animations are active.
3. **Ghosting protections stay intact.** The full-terminal blank backing layer
   (e265c41) and full-width effect rows are load-bearing: every effect Render()
   row must remain exactly `width` visible cells, and `render_test.go` must
   pass. Rows end with an SGR reset so colors can't bleed into padding.
4. **Lazy inits ride the tick** (0d1299c, 6f5ab3f): aquarium/sonar/cracktro/
   plasma creation and the gslapper launch happen on tickMsg once dimensions
   are known. Tick gating must not run before these have fired.

## Changes

### 1. Cache screensaver config (main.go:925, screensaver.go)

`loadScreensaverConfig()` opens and parses screensaver.conf on every 30ms tick.
Cache it in the model at startup; refresh at most once per 60s from the tick
handler (preserves live-edit of idle_timeout without the 33/s file opens).

### 2. Skip bgTick work when nothing is animating

`doBgTick` reschedules at full rate with background "none". Keep the chain
alive (self-healing, avoids kick-on-every-selection-path bugs) but reschedule
at 1s when no effect is selected or the mode can't display it. On background
selection in the menu, the key-event handler kicks an immediate fast bgTick so
switching feels instant.

### 3. Idle-gate the UI tick

The 30ms tick drives: typewriter ticker, banner gradient, border animation
counters, screensaver idle check, print/beams/pour effects, lazy inits.
Adaptive interval, decided in the tickMsg handler:

- **30ms (unchanged):** any of — last key input < 60s ago; ticker background
  selected; print/beams/pour active; screensaver print animation running;
  lazy inits still pending.
- **1s:** otherwise (idle at login, screensaver clock). Animation counters
  (`animationFrame`, `borderFrame`, `pulseColor`) do NOT advance on slow
  ticks — the banner gradient freezes instead of stepping at 1fps. Screensaver
  activation and clock only need 1s resolution.
- Any key press returns to 30ms immediately (key handler already reschedules
  the chain implicitly — verify; if not, kick one).

### 4. Run-length SGR rendering in effect Render() methods

Replace per-cell `lipgloss.NewStyle().Foreground().Render()` with direct SGR
writing: single `strings.Builder` grown upfront, emit `38;2;r;g;bm` only when
the cell color differs from the previous cell, reset at each row end. Prototype
benched 39x on fire (5.3ms → 0.13ms, 24,459 → 3 allocs).

Safety: layer content is parsed into cells by ultraviolet
(`StyledString.Draw`), and the terminal writer re-emits per the detected color
profile — raw truecolor SGR in layer content does not bypass TTY fallback.
Verify by inspecting emitted output under `TERM=linux`.

Effects, in order: fire, matrix, rain (shared helper), then plasma, aquarium,
fireworks, beams, beams_text, pour, blackhole, print as applicable — audit each
for the same pattern before converting.

### 5. Palette updates on theme change only — DROPPED

The per-tick palette sync costs ~2μs/frame; moving it to the theme-selection
handler risks missing a theme-change path and leaving an effect wrongly
colored until restart. Self-healing beats the micro-win. Not done on purpose.

## Verification per step

- `go test ./...` (render_test.go guards layer alignment)
- Effect benchmarks before/after (`audit_bench_test.go` pattern)
- pty CPU measurement (ptyrun harness): idle, fire, matrix, plasma —
  record in the PR body against the baseline table
- Effect output invariants test: every Render() row = exactly `width` cells,
  `height` rows, row-end reset — added as a real test in this PR
- Visual check on real kitty by Nomadcxx before merge (gradients, ticker
  smoothness, menu responsiveness, screensaver entry/exit, power menu
  ghosting)

## Out of scope

- Plasma's 111 MiB RSS (separate issue if it survives the render rewrite)
- bubbletea/ultraviolet-level optimizations
- Any visual redesign; output must look identical at active frame rates
