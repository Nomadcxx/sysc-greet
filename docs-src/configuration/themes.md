# Themes

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

An example theme is provided in the repository at `examples/themes/example.toml`.

### TTY Compatibility

sysc-greet uses the `colorprofile` library to detect terminal capabilities and fall back gracefully:

- **TrueColor terminals** - Full 24-bit color support
- **ANSI256 terminals** - 256-color palette support
- **Basic TTY** - Falls back to basic ANSI 16 colors

This ensures consistent appearance across different terminal emulators and TTY.
