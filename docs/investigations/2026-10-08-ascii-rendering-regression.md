# ASCII rendering failure after fullscreen resize

## Cause

Bubble Tea v2 at `997384b0b35e` resizes its terminal renderer with the previous dimensions:

```go
s.scr.Resize(s.width, s.height)
s.width, s.height = w, h
```

The underlying Ultraviolet renderer uses that width to resize its cached tab stops. Kitty initially opens at approximately 86 columns, then Niri expands it to 174 columns. After that resize, the model and screen buffer use the new size while cursor movement optimizations use the old tab grid. Repeated animation updates place glyphs outside the art and login border.

The upstream correction assigns the new dimensions before resizing the renderer. Commit `63fc263` pins Bubble Tea to upstream `d7c95b8a2778` and adapts the three moved terminal settings to the View API.

## Project history

| Commit | Date | Evidence |
| --- | --- | --- |
| `f72cc9d` | 2025-10-10 | Initial commit imports Bubble Tea v2, returns `tea.View`, enables fullscreen, and pins the affected version in go.mod. |
| `d976278` | 2025-10-25 | Introduces beams and places active-theme refresh inside the production-only preference-save guard. |
| `68faff8` | 2025-10-26 | Adds pour and its refresh under the same guard. |
| `2635f6b` | 2026-04-10 | Promotes x/ansi from indirect to direct for keyboard handling; does not change the Bubble Tea or Ultraviolet versions. |
| `f7f9526` | 2026-10-08 | Refreshes active text-effect palettes in test mode. |
| `63fc263` | 2026-10-08 | Applies the upstream renderer resize fix. |

The faulty dependency was present from the first repository commit. The new theme files exposed the failure during testing. History establishes the earliest checked-in exposure; the development baseline reproduces the runtime failure. We have not identified a recent operating-system or Kitty change that explains when it first became noticeable.

## Reproduction and isolation

Bench: SSH port 7777, `nomadx@192.168.0.64`; native Niri, Kitty 0.49.1. Run a newly built binary in fullscreen Kitty with `-c NONE` and `--test`. Select beams or pour through F1, then inspect at 5, 10, and 15 seconds.

- Development baseline and feature build both displace glyphs with Dracula.
- Blue also fails before the renderer update.
- The composed model frame has the correct 38-column XFCE art and placement.
- Reporting `TERM=xterm-256color` does not resolve the failure.
- Disabling only tab movement in a temporary Ultraviolet build resolves it.
- Changing only the resize call to `s.scr.Resize(w, h)` in a temporary Bubble Tea build resolves it.
- The upstream dependency update resolves it without a local dependency fork.

These experiments isolate stale renderer tab stops. The same failure occurs without a nested compositor.

## Coverage gap and checks

Existing animation tests check frame sizes and color resets. They do not exercise a terminal's differential cursor movements after a fullscreen resize. Screenshots immediately after effect selection also miss accumulated corruption.

The runnable hardware check is [check-ascii-effects.py](../../scripts/check-ascii-effects.py):

```sh
python3 scripts/check-ascii-effects.py /path/to/sysc-greet
python3 scripts/check-ascii-effects.py /path/to/sysc-greet --theme blue
```

It launches native fullscreen Kitty, selects the theme through the menu, and checks all four effects at 5, 10, and 15 seconds. The old binary fails at beams' five-second frame. The fixed build passes with Dracula and Blue. The script checks containment; visual captures supplement it for animation shape and palette.

Local `go test -count=1 -p 2 ./...`, `GOMAXPROCS=4 go vet ./...`, and the bench build/focused tests pass. GitHub CI also passes for `63fc263`.

Commit `7387be8` fixes the separate CLI theme-selection problem (`sysc-greet-dev-46`). Startup now applies the requested theme to the active palette, inputs and spinner, with case-insensitive menu names and saved-theme precedence. Native Blue startup and secondary palette propagation pass.
