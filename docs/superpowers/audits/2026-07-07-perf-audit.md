# sysc-greet Performance Audit — 2026-07-07

Measured on development @ a8189e5, Ryzen 7 5700X, 240x67 pty (fullscreen kitty
on 1080p), test mode. CPU is % of one core for the sysc-greet process only —
kitty pays a comparable cost again to parse and composite the output stream.

## Measurements

| Config | CPU | Terminal output | RSS |
|---|---|---|---|
| Idle login, background "none" | 53% | 1.0 MiB/s | 57 MiB |
| fire | 84% | 1.4 MiB/s | 57 MiB |
| matrix | 60% | 1.5 MiB/s | 67 MiB |
| plasma | 77% | 2.0 MiB/s | 111 MiB |

Effect benchmarks at 240x67 (`internal/animations/audit_bench_test.go`):

| Benchmark | per frame | allocs/frame |
|---|---|---|
| FireEffect.Render (240x26) | 5.3 ms | 24,459 |
| MatrixEffect.Render | 2.6 ms | 9,891 |
| RainEffect.Render | 1.8 ms | 5,870 |
| Plasma Update+Render | 2.1 ms | 10,267 |
| FireEffect.Update | 0.16 ms | 0 |
| Fire render, run-length prototype | **0.13 ms** | **3** |

## Why the idle login screen burns half a core

CPU profile attribution: 23% of samples inside `model.View` (20% of that in
`renderDualBorderLayout`), the rest in bubbletea/ultraviolet cell diffing and
terminal writing. Nothing is individually slow — the app just produces a new,
different full frame 33 times a second, forever:

1. **The tick never stops and never slows.** `doTick()` reschedules
   unconditionally every 30 ms (main.go:956), and `doBgTick` reschedules even
   when `selectedBackground` is "none" (main.go:1002). Two overlapping tick
   loops at 15–60 ms each.
2. **Every frame differs, so nothing can be skipped.** The ASCII banner runs a
   monochrome gradient animation keyed to `m.animationFrame` (ascii.go:740),
   recoloring the banner region every tick. Border pulse/wave counters
   (`borderFrame`, `pulseColor`) also advance every tick. Result: ~30 KiB of
   changed cells per frame at 33 fps on a screen where the user sees an
   essentially static login form.
3. **Disk I/O on every tick.** `loadScreensaverConfig()` opens and parses
   `screensaver.conf` 33x/second while in login/password mode (main.go:925),
   only to check the idle timeout.
4. **Palettes rebuilt every bgTick.** Each effect calls
   `GetXPalette(m.currentTheme)` + `UpdatePalette` every 30 ms; the theme only
   changes on user action. (Minor: ~60 ns each, but it's pure waste.)

## Why the effects cost 2–5 ms per frame

Every effect renders with `lipgloss.NewStyle().Foreground(c).Render(char)`
**per cell per frame** (fire.go:134, same pattern in matrix, rain, plasma,
aquarium, fireworks, beams, pour, blackhole). That allocates a style object and
emits a full SGR sequence + reset for every character — ~24k allocations and
~660 KiB of garbage per fire frame.

The prototype in `audit_bench_test.go` (`renderFireRunLength`) writes SGR codes
directly and only when the color changes along a row: **39x faster, 3 allocs**
(5.3 ms → 0.13 ms). The same transformation applies to every effect.

## Recommendations, in order of impact

1. **Idle-gate the render loop.** When mode is login/password, no background
   effect is active, and no key was pressed recently, drop to a slow tick
   (250 ms–1 s) or stop ticking entirely. The banner gradient at 33 fps is
   imperceptible; run it at 8–10 fps or freeze it after N seconds idle. This is
   the difference between a greeter that idles at ~0% and one that pins half a
   core (plus kitty's share) 24/7 on every machine at the login screen.
2. **Don't schedule `doBgTick` when no effect is selected**, and don't tick
   effects whose mode can't display them (menus, power screen).
3. **Cache the screensaver config.** Load once at startup, reload when the
   settings menu changes it. Removes 33 file opens/second.
4. **Run-length color rendering in effect Render() methods** per the
   prototype: direct SGR emission, new sequence only on color change, single
   `strings.Builder` sized upfront. 39x on fire; similar expected elsewhere.
5. **Update palettes on theme change**, not per tick.
6. Plasma also holds 111 MiB RSS — worth a look at its buffers while touching
   its render path.

Items 1–3 fix the idle cost (the one every user pays); item 4 fixes the
animated-background cost. All are localized: tick scheduling in main.go's
Update, Render() bodies in internal/animations.
