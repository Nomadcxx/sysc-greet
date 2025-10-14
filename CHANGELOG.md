# Changelog

All notable changes to sysc-greet will be documented in this file.

## [Unreleased]

### Removed
- **sysc-greet.conf system** - Removed unused config file loading system that only loaded custom color palettes
  - Removed `loadConfig()` function and `Config.Palettes` field
  - Removed help text references to sysc-greet.conf
  - Hardcoded `sessionPalettes` already provide all needed color schemes (GNOME, KDE, Hyprland, Sway, i3, Xfce)
- **Animation options from ASCII configs** - Removed non-functional `animation_style`, `animation_speed`, and `animation_direction` fields from all ASCII config files
- **Animation documentation** - Cleaned up CONFIGURATION.md to remove references to unimplemented animation features

### Fixed
- Removed confusing "0 custom palettes loaded" message that appeared in help text

## [1.0.0] - 2025-10-14

### Added
- Initial public release
- 9 themes: Dracula, Catppuccin, Nord, Tokyo Night, Gruvbox, Material, Solarized, Monochrome, TransIsHardJob
- Background effects: Fire (PSX DOOM), Matrix rain, ASCII rain, Static patterns
- 7 border styles: Classic, Modern, Minimal, ASCII-1, ASCII-2, Wave, Pulse
- Screensaver with configurable idle timeout and ASCII art cycling
- Multi-ASCII variant support with Page Up/Down navigation
- Video wallpaper support via gslapper (multi-monitor)
- Session management with X11/Wayland auto-detection
- Preference caching (theme, background, border, session)
- greetd integration
- Built with Go + Bubble Tea framework

### Key Bindings
- F2 - Settings menu (themes, borders, backgrounds)
- F3 - Session selection
- F4 - Power menu (shutdown/reboot)
- F5 - Release notes
- Page Up/Down - Cycle ASCII variants
- Tab - Navigate fields
- Enter - Submit/Continue
- Esc - Cancel/Return

[Unreleased]: https://github.com/Nomadcxx/sysc-greet/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Nomadcxx/sysc-greet/releases/tag/v1.0.0
