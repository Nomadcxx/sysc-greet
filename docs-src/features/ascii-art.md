# ASCII Art Configuration

Each session type can have custom ASCII art with multiple variants. Users can cycle through variants using Page Up/Down keys at the greeter.

## Configuration Location

`/usr/share/sysc-greet/ascii_configs/`

Each session has a `.conf` file with the session name as the filename.

## Format

```ini
name=Session Name

# ASCII variants (can have multiple: ascii_1, ascii_2, etc.)
ascii_1=
  ╔════════════════════════════╗
  ║    SESSION NAME ART          ║
  ╚════════════════════════════╝

ascii_2=
   _____  __   _____  _____   __
  |   _/ /  / __ \ / __ \ / /
  |  /  / _/  /_/ / /_/ /_/ 
```

# Optional: Custom colors for rainbow effect (comma-separated hex)
colors=#4285f4,#34a853,#fbbc05,#ea4335,#9c27b0,#ff9800
```

## Fields

- **name** - Display name for the session (used in session dropdown)
- **ascii_1**, **ascii_2**, etc. - Multiple ASCII art variants that users can cycle through
- **colors** - Optional hex colors for rainbow gradient effect on ASCII art

## ASCII Variant Cycling

Press **Page Up** or **Page Down** at the greeter to cycle through ASCII variants for the selected session. The last selected variant is saved to preferences and restored on next login.

## Creating Custom ASCII

**ASCII Generators:**
- [patorjk.com/software/taag](http://patorjk.com/software/taag/) - Web-based generator
- [ASCII Art Archive](https://www.asciiart.eu/) - Browse existing art
- [Mobius](https://github.com/Mobius-Team/Mobius) - Advanced ASCII art tools

**Using figlet:**

Install figlet and figlet-fonts:

```bash
# Arch Linux
sudo pacman -S figlet

# Install additional fonts
git clone https://github.com/xero/figlet-fonts
sudo cp figlet-fonts/* /usr/share/figlet/

# Generate ASCII
figlet -f dos_rebel "HYPRLAND"
```

**Important:** Keep ASCII art under 80 columns wide for compatibility with most terminal sizes.

**Figlet fonts:** [github.com/xero/figlet-fonts](https://github.com/xero/figlet-fonts)

## Session Name Mapping

sysc-greet maps session names from XDG session files to config filenames:

| Session Name | Config Filename |
|-------------|-----------------|
| GNOME Desktop | gnome_desktop |
| KDE Plasma | kde |
| Hyprland | hyprland |
| Sway | sway |
| i3 | i3wm |
| BSPWM | bspwm_manager |
| Xmonad | xmonad |
| Openbox | openbox |
| Xfce | xfce |
| Cinnamon | cinnamon |
| IceWM | icewm |
| Qtile | qtile |
| Weston | weston |

For sessions not listed, the lowercase first word of the session name is used as the config filename.
