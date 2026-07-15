---
title: "Themes"
description: "sysc-greet includes multiple built-in color themes. Themes affect the entire color scheme of the greeter including backgrounds, borders, text, and accent ele..."
---

sysc-greet includes multiple built-in color themes. Themes affect the entire color scheme of the greeter including backgrounds, borders, text, and accent elements.

## Available Themes

| Theme | Primary Color | Description |
|--------|--------------|-------------|
| Dracula | #bd93f9 | Dark purple-blue theme |
| Gruvbox | #fe8019 | Warm dark theme |
| Material | #80cbc4 | Material Design dark theme |
| Nord | #81a1c1 | Arctic blue-toned dark theme |
| Tokyo Night | #7aa2f7 | Modern dark theme |
| Catppuccin | #cba6f7 | Soft pastel dark theme |
| Solarized | #268bd2 | Solarized dark theme |
| Monochrome | #ffffff | Black and white minimal theme |
| TransIsHardJob | #5BCEFA | Transgender flag colors |
| Eldritch | #37f499 | Purple and green theme |
| RAMA | #ef233c | RAMA keyboard aesthetics |
| Dark | #ffffff | True black and white minimal theme |

## Changing Themes

Press **F1** → **Themes** to cycle through available themes. Your selection is saved automatically.

## Custom Themes

Create custom themes by placing TOML files in:

- `/usr/share/sysc-greet/themes/` (system-wide)
- `~/.config/sysc-greet/themes/` (user)

Custom themes appear in F1 → Themes alongside built-in themes.

### Format

```toml
# my-theme.toml
name = "My Theme"

[colors]
bg_base = "#1a1a2e"
bg_active = "#2a2a3e"
primary = "#e94560"
secondary = "#0f3460"
accent = "#16213e"
warning = "#f59e0b"
danger = "#ef4444"
fg_primary = "#ffffff"
fg_secondary = "#cccccc"
fg_muted = "#888888"
border_focus = "#e94560"
```

All color fields are required. Use hex format (`#RRGGBB`).

Example themes are provided in the repository at `examples/themes/`.

### Community Examples

#### Piink

A vibrant pink and magenta theme by [Pat8998](https://github.com/Pat8998).

```toml
# piink.toml
name = "Piink"

[colors]
bg_base = "#000000"
bg_active = "#ad1a5d"
primary = "#f72585"
secondary = "#a71aad"
accent = "#f72585"
warning = "#ffff00"
danger = "#ff0000"
fg_primary = "#ee33ee"
fg_secondary = "#a125f7"
fg_muted = "#ee33ee"
border_focus = "#711aad"
```

![Piink theme](https://github.com/user-attachments/assets/c0d3ede0-5945-45da-b0d8-48ab9adf9484)

This theme also pairs well with a custom ASCII config using braille-art. See `examples/themes/piink.toml` in the repository.

Have a custom theme you'd like to share? Open an [issue](https://github.com/Nomadcxx/sysc-greet/issues) and we'll add it here.

### Generating Wallpapers

After adding a new theme, regenerate wallpapers to include theme-matched backgrounds:

```bash
cd scripts/
python3 generate-wallpapers.py
```

This creates wallpapers for all themes in `/usr/share/sysc-greet/wallpapers/`. See [Wallpapers](../features/wallpapers) for more options.

### Background Effects

Custom themes automatically work with all background effects (fire, matrix, rain, fireworks). The effects generate color palettes from your theme's colors:

- **Fire** uses `bg_base` → `warning` → `danger` → `primary` gradient
- **Matrix** uses `bg_base` → `secondary` → `primary` for the falling characters
- **Rain** uses `primary`, `secondary`, `accent` for the drops
- **Fireworks** uses all accent colors for variety

No additional configuration needed - just select your custom theme and enable an effect.

### TTY Compatibility

sysc-greet uses the `colorprofile` library to detect terminal capabilities and fall back gracefully:

- **TrueColor terminals** - Full 24-bit color support
- **ANSI256 terminals** - 256-color palette support
- **Basic TTY** - Falls back to basic ANSI 16 colors

This ensures consistent appearance across different terminal emulators and TTY.
