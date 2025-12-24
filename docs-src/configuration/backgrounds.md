# Backgrounds

sysc-greet supports multiple background types for the login screen.

## Background Effects

### Fire Effect

DOOM PSX-style fire effect renders at top of screen. This is a particle-based animation that runs independently of the main interface.

**Configuration:** Not configurable - static implementation

**Access:** F1 Settings > Backgrounds > Fire

### Matrix Effect

Matrix rain effect displays falling green characters similar to The Matrix movie.

**Configuration:** Not configurable - static implementation

**Access:** F1 Settings > Backgrounds > Matrix

### ASCII Rain Effect

Falling ASCII characters rain down the screen. This effect uses the current theme's color palette instead of the Matrix green.

**Configuration:** Not configurable - static implementation

**Access:** F1 Settings > Backgrounds > ASCII Rain

### Fireworks Effect

Firework explosion animations appear at random screen positions. This is a lightweight particle system.

**Configuration:** Not configurable - static implementation

**Access:** F1 Settings > Backgrounds > Fireworks

### Aquarium Effect

Animated aquarium scene with swimming fish, bubbles, and seaweed. Colors adapt to the selected theme.

**Configuration:** Not configurable - static implementation

**Access:** F1 Settings > Backgrounds > Aquarium

## Background Effect Priority

Background effects are mutually exclusive. When you enable one, others are automatically disabled. The priority order is:

1. User-selected background (highest priority)
2. Last active background (if not explicitly disabled)

## TTY Compatibility

All background effects render using `lipgloss` for styling, which provides TTY compatibility through the `colorprofile` library:
- TrueColor terminals get full 24-bit color support
- ANSI256 terminals get 256-color palette support
- Basic TTY falls back to 16 ANSI colors

## Video Wallpapers

Video wallpapers take priority over all background effects when active. The wallpaper system uses gSlapper for video playback.

For details on video wallpapers, see [Wallpapers Feature](../features/wallpapers.md).
