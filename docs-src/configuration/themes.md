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
| RAMA | #ef233c | RAMA keyboard aesthetics |
| Dark | #ffffff | True black and white minimal theme |
| Default | #8b5cf6 | Original Crush-inspired theme |

## Changing Themes

Press **F1** → **Themes** to cycle through available themes. Your selection is saved automatically.

## Custom Themes

Custom themes are not currently supported. All themes are compiled into the binary.

If you want a specific color scheme added, open a [feature request](https://github.com/Nomadcxx/sysc-greet/issues/new) with:

- Theme name
- Primary color (hex, e.g., `#bd93f9`)
- Secondary color
- Accent color
- Background color

### TTY Compatibility

sysc-greet uses the `colorprofile` library to detect terminal capabilities and fall back gracefully:

- **TrueColor terminals** - Full 24-bit color support
- **ANSI256 terminals** - 256-color palette support
- **Basic TTY** - Falls back to basic ANSI 16 colors

This ensures consistent appearance across different terminal emulators and TTY.
