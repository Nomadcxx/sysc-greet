# Wallpapers

sysc-greet supports two types of wallpapers: themed static images and video backgrounds.

## Themed Wallpapers

Themed wallpapers are static images that automatically match the selected theme.

### Location

`/usr/share/sysc-greet/wallpapers/`

### Naming Convention

Images must follow the naming convention: `sysc-greet-{theme}.png`

Examples:
- `sysc-greet-dracula.png` - Displayed when Dracula theme is active
- `sysc-greet-nord.png` - Displayed when Nord theme is active
- `sysc-greet-catppuccin.png` - Displayed when Catppuccin theme is active

### Automatic Theme Switching

When you change themes using F1 Settings, sysc-greet automatically switches to the matching themed wallpaper if it exists.

### Included Themed Wallpapers

sysc-greet ships with themed wallpapers for all built-in themes:
- Dracula
- Gruvbox
- Material
- Nord
- Tokyo Night
- Catppuccin
- Solarized
- Monochrome
- TransIsHardJob
- Eldritch
- RAMA
- Default

### Adding or Replacing Themed Wallpapers

```bash
# Replace an existing theme wallpaper
sudo cp ~/my-custom-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-dracula.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-dracula.png

# Add a wallpaper for a new theme
sudo cp ~/my-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-mytheme.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-mytheme.png
```

## Video Wallpapers

Video wallpapers provide animated backgrounds with multi-monitor support using gSlapper.

### Location

`/var/lib/greeter/Pictures/wallpapers/`

### Supported Formats

- MP4
- WebM
- MKV
- AVI
- MOV

### Accessing Wallpapers

Press **F1** Settings then **Wallpaper** to browse available video wallpapers.

### Adding Video Wallpapers

```bash
# Copy video to greeter's wallpaper directory
sudo cp ~/Videos/my-animation.mp4 /var/lib/greeter/Pictures/wallpapers/
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/my-animation.mp4
```

### Stop Wallpaper

From the wallpaper menu, select **Stop Wallpaper** to stop the gSlapper process and clear the wallpaper preference.

## Wallpaper Priority

1. Video wallpapers (gSlapper) - Highest priority. When selected, they override themed wallpapers and background effects
2. Themed wallpapers - Auto-selected when theme changes, unless video wallpaper is active
3. Background effects - Fire, Matrix, etc. Only active when no wallpaper is set

## Troubleshooting

**gSlapper not starting:**
```bash
# Check gSlapper debug log
cat /tmp/sysc-greet-wallpaper.log

# Verify gSlapper is installed
which gslapper

# Check compositor config has gSlapper startup command
cat /etc/greetd/niri-greeter-config.kdl
```

**Wallpaper not changing:**
- Verify gSlapper socket is running: `ls /tmp/sysc-greet-wallpaper.sock`
- Check compositor is actually starting gSlapper in its config
- Review debug logs: `journalctl -u greetd -n 50`
