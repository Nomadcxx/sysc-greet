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

## Custom Wallpapers

The wallpaper menu provides access to both static images and video backgrounds stored in the user wallpaper directory.

### Location

`/var/lib/greeter/Pictures/wallpapers/`

### Supported Formats

**Static Images:**
- PNG
- JPG / JPEG
- WebP
- GIF

**Video:**
- MP4
- WebM
- MKV
- AVI
- MOV

### Accessing Wallpapers

Press **F1** Settings then **Wallpaper** to browse available wallpapers. The menu lists both static images and video files.

### Adding Wallpapers

```bash
# Copy static image to greeter's wallpaper directory
sudo cp ~/Pictures/my-bg.png /var/lib/greeter/Pictures/wallpapers/
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/my-bg.png

# Copy video to greeter's wallpaper directory
sudo cp ~/Videos/my-animation.mp4 /var/lib/greeter/Pictures/wallpapers/
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/my-animation.mp4
```

### Stop Video Wallpaper

From the wallpaper menu, select **Stop Video Wallpaper** to pause video playback and revert to a static wallpaper. This uses gSlapper's IPC to pause the video without restarting the daemon.

## Wallpaper Priority

1. Custom wallpapers (gSlapper) - Highest priority. When a wallpaper is selected from the menu (video or static), it overrides themed wallpapers and background effects
2. Themed wallpapers - Auto-selected when theme changes, unless a custom wallpaper is active
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
