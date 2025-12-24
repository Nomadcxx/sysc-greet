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

Press **F1** to open the Settings menu, then select **Themes** to cycle through available themes.

### Theme Behavior

Themes are applied immediately when selected and saved to user preferences. The next time sysc-greet starts, the last selected theme will be loaded automatically.

### Theme Colors

Each theme defines the following color variables:

- **BgBase** - Main background color
- **BgElevated** - Elevated surface background (same as BgBase for consistency)
- **BgSubtle** - Subtle background color
- **BgActive** - Active element background color
- **Primary** - Primary brand color (borders, focused elements)
- **Secondary** - Secondary accent color
- **Accent** - Tertiary accent color
- **Warning** - Warning state color
- **Danger** - Error state color
- **FgPrimary** - Primary text color
- **FgSecondary** - Secondary text color
- **FgMuted** - Muted text color
- **FgSubtle** - Subtle text color
- **BorderFocus** - Border color when focused

### TTY Compatibility

sysc-greet uses the `colorprofile` library to detect terminal capabilities and fall back gracefully:

- **TrueColor terminals** - Full 24-bit color support
- **ANSI256 terminals** - 256-color palette support
- **Basic TTY** - Falls back to basic ANSI 16 colors

This ensures consistent appearance across different terminal emulators and TTY.
