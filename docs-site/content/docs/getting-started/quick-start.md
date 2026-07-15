---
title: "Quick Start"
description: "Quick guide to get sysc-greet running on your system."
---

Quick guide to get sysc-greet running on your system.

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```

The interactive installer will guide you through:
1. Choosing your greeter compositor: niri (default), cagebreak, sway, or hyprland (deprecated)
2. Configuring compositor settings
3. Installing dependencies automatically

### Manual Install

If you prefer manual setup, see [Installation Guide](installation).

## First Launch

After installation, sysc-greet will start automatically when you boot your system.

To test the greeter without restarting:

```bash
sysc-greet --test
```

For fullscreen testing:
```bash
kitty --start-as=fullscreen sysc-greet --test
```

## Basic Usage

### Login

1. Enter your username (or press Tab to select session)
2. Press Enter or Tab to advance to password field
3. Enter your password
4. Press Enter to login

### Settings Menu (F1)

Press F1 to open settings menu. Navigate with Up/Down arrows:
- **Themes** - Change color theme
- **Borders** - Change border style
- **Backgrounds** - Change background effects
- **ASCII Effects** - Change text animation
- **Wallpaper** - Select video wallpapers

### Session Selection (F2)

Press F2 to open session dropdown. Use Up/Down to select a different session.

### Power Menu (F4)

Press F4 to access power options:
- Reboot
- Shutdown
- Cancel

## Configuration

sysc-greet automatically saves your preferences:
- Selected theme
- Selected background effect
- Selected border style
- Selected wallpaper
- Last used session (if username caching enabled)

Preferences are stored in `/var/cache/sysc-greet/` and restored on next login.

### Common Tasks

**Change Theme**
1. Press F1
2. Select Themes
3. Navigate to desired theme
4. Press Enter to apply

**Set Video Wallpaper**
1. Press F1
2. Select Wallpaper
3. Navigate to desired video file
4. Press Enter to apply

**Cycle ASCII Art Variant**
1. Press Page Up or Page Down
2. ASCII art changes to next/previous variant

For detailed configuration options, see [Themes](../configuration/themes), [Backgrounds](../configuration/backgrounds), and [Keyboard Layout](../configuration/keyboard-layout).

## Non-US Keyboard Layouts

If you use a non-US keyboard layout (Dvorak, German, French, etc.), you may find that passwords are rejected at the greeter even though your compositor layout is configured correctly. This is because Kitty (the terminal running sysc-greet) needs explicit XKB environment variables — it does not inherit the compositor's layout automatically.

See the [Keyboard Layout guide](../configuration/keyboard-layout) for per-compositor setup instructions.
