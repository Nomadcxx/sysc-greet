---
title: "Backgrounds & Effects"
description: "sysc-greet includes animated background effects and ASCII text effects."
---

sysc-greet includes animated background effects and ASCII text effects.

## Background Effects

Access via **F1** → **Backgrounds**:

| Effect | Description |
|--------|-------------|
| Fire | DOOM PSX-style fire animation |
| Matrix | Falling green characters |
| ASCII Rain | Falling characters in theme colors |
| Fireworks | Particle explosions |
| Aquarium | Fish, bubbles, and seaweed |
| Sonar | Scanning sonar display |
| Cracktro | Retro scene-style animated art |
| Plasma | Animated plasma field |
| AK-47 | Rotating model |
| Australian SLR | Rotating L2A1 SAR model |
| Higgs Field | Particle field with travelling waves and vortex motion |
| Rotating models | Statues, landmarks, and aircraft |

Higgs Field and the rotating models render as Braille text through the shared sysc-Go mesh pipeline.
Rotating model frames display the model title, creator, and license.
`THIRD_PARTY_NOTICES.md` lists their source links and full license terms.

## ASCII Effects

Access via **F1** → **ASCII Effects**:

| Effect | Description |
|--------|-------------|
| Typewriter | Character-by-character text animation |
| Print | Top-down dot-matrix printer style |
| Beams | Horizontal color beam scanning |
| Pour | Characters pour down with gradient |

## Custom Theme Support

All effects automatically use colors from custom themes. When you select a custom theme, effects generate palettes from your theme's `primary`, `secondary`, `accent`, `warning`, and `danger` colors.

See [Custom Themes](../configuration/themes#custom-themes) for creating your own themes.

## Behavior

- Background effects are mutually exclusive
- ASCII effects are mutually exclusive
- Background and ASCII effects can run together
- A GSlapper video suppresses secondary background effects; with no effect selected, the existing wallpaper remains
- The configured screensaver takes precedence over wallpapers and effects, including video, on secondary outputs

The greeter runs fullscreen in Kitty under the compositor that greetd starts. For visual checks, use Kitty in that real compositor; a raw TTY does not exercise the greeter window or compositor output behavior.
