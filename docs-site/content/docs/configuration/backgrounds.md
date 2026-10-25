---
title: "Backgrounds"
description: "sysc-greet supports background effects and wallpapers."
---

sysc-greet supports background effects and wallpapers.

## Background Effects

Access via **F1** → **Backgrounds**:

| Effect | Description |
|--------|-------------|
| Fire | DOOM PSX-style fire effect |
| Matrix | Matrix rain (green characters) |
| ASCII Rain | Falling ASCII characters (theme colors) |
| Fireworks | Random firework explosions |
| Aquarium | Swimming fish and bubbles |
| Sonar | Scanning sonar display |
| Cracktro | Retro scene-style animated art |
| Plasma | Animated plasma field |
| AK-47 | Rotating AK-47 model |
| Australian SLR | Rotating L2A1 SAR model |
| Higgs Field | Particle field with travelling waves and vortex motion |
| Rotating models | Statues, landmarks, and aircraft from the Backgrounds menu |
| Typewriter, Print, Beams, Pour | Text and ASCII-art animations in the ASCII Effects submenu |
| None | Disable background effect |

## Wallpapers

Custom image and video wallpapers are managed by [gSlapper](https://github.com/Nomadcxx/gSlapper).

### Themed Wallpapers

**Location:** `/usr/share/sysc-greet/wallpapers/`

Auto-switch when theme changes. Named `sysc-greet-{theme}.png`.

### Custom Wallpapers

**Location:** `/var/lib/greeter/Pictures/wallpapers/`

Add your own images or videos:

```bash
sudo cp ~/my-wallpaper.mp4 /var/lib/greeter/Pictures/wallpapers/
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/my-wallpaper.mp4
```

Access via **F1** → **Wallpaper**.

## Secondary outputs

On Niri, Cagebreak, Sway, Mango, and Hyprland 0.55+, sysc-greet can use an installed `sysc-terminal` to draw the selected effect and theme on secondary outputs. Experimental sysc-compositor builds use the same path through the Niri-compatible IPC exposed over `NIRI_SOCKET`. If no effect is selected, the compositor wallpaper remains. A selected GSlapper video suppresses secondary effects. An explicit `--secondary-effect` supplies a fallback only when the selected effect is None; its default is empty. Effects absent from the installed renderer are skipped without substitution.

| Option | Behavior |
| --- | --- |
| `--secondary-backgrounds=false` | Disable secondary background effects; enabled by default |
| `--secondary-exclude DP-3,HDMI-A-1` | Exclude named outputs from background effects |
| `--secondary-effect pour` | Use this sysc-terminal effect when the greeter background is None |

When the configured screensaver activates, it appears on every secondary output in a read-only fullscreen Kitty child and takes precedence over wallpapers and effects, including GSlapper video. A key or mouse event in any saver window wakes the primary greeter and closes the other saver windows. Disabling secondary background effects or excluding an output does not disable its screensaver. Install sysc-terminal separately to mirror effects; a missing or failed helper leaves login available.
