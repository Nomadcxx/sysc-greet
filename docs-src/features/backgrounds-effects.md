# Backgrounds & Effects

sysc-greet includes animated background effects and ASCII text effects.

## Background Effects

Access via **F1** → **Backgrounds**:

| Effect | Description |
|--------|-------------|
| Fire | DOOM PSX-style fire animation |
| Matrix | Falling green characters (The Matrix) |
| ASCII Rain | Falling characters (theme colors) |
| Fireworks | Random particle explosions |
| Aquarium | Fish, bubbles, and seaweed |

## ASCII Effects

Access via **F1** → **ASCII Effects**:

| Effect | Description |
|--------|-------------|
| Typewriter | Character-by-character text animation |
| Print | Top-down dot-matrix printer style |
| Beams | Horizontal color beam scanning |
| Pour | Characters pour down with gradient |

## Behavior

- Background effects are mutually exclusive
- ASCII effects are mutually exclusive
- Both can run together
- Video wallpapers override all effects

## TTY Compatibility

All effects use automatic color profile detection:

- TrueColor: Full 24-bit color
- ANSI256: 256-color palette
- Basic TTY: 16 ANSI colors
