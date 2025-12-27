# Wallpapers

sysc-greet supports two types of wallpapers: themed static images and custom wallpapers (including video backgrounds).

## Themed Wallpapers (Static Images)

**Location:** `/usr/share/sysc-greet/wallpapers/`

**Managed by:** [gSlapper](https://github.com/Nomadcxx/gSlapper) (Wayland wallpaper daemon)

These auto-match your selected theme using the naming convention `sysc-greet-{theme}.png`.

### Included Themed Wallpapers

- `sysc-greet-dracula.png`
- `sysc-greet-gruvbox.png`
- `sysc-greet-nord.png`
- `sysc-greet-tokyo-night.png`
- `sysc-greet-catppuccin.png`
- `sysc-greet-material.png`
- `sysc-greet-solarized.png`
- `sysc-greet-monochrome.png`
- `sysc-greet-eldritch.png`
- `sysc-greet-transishardjob.png`
- `sysc-greet-rama.png`
- `sysc-greet-default.png`
- `sysc-greet-dark.png`

### Adding or Replacing Themed Wallpapers

```bash
# Replace an existing theme wallpaper
sudo cp ~/my-nord-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-nord.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-nord.png

# Add a wallpaper for a new theme
sudo cp ~/my-bg.png /usr/share/sysc-greet/wallpapers/sysc-greet-mytheme.png
sudo chown greeter:greeter /usr/share/sysc-greet/wallpapers/sysc-greet-mytheme.png
```

## Custom Wallpapers (Videos & Images)

**Location:** `/var/lib/greeter/Pictures/wallpapers/`

**Managed by:** [gSlapper](https://github.com/Nomadcxx/gSlapper) (Video wallpaper manager)

Video wallpapers provide animated backgrounds with multi-monitor support.

### Requirements

```bash
# Install gSlapper (if not already installed)
yay -S gslapper
# or build from source: https://github.com/Nomadcxx/gSlapper
```

### Adding Custom Wallpapers

```bash
# Copy video to greeter's wallpaper directory
sudo cp ~/Videos/cool-animation.mp4 /var/lib/greeter/Pictures/wallpapers/

# Set correct ownership
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/cool-animation.mp4
```

### Supported Formats

**Static Images:** PNG, JPG, JPEG, WebP, GIF

**Video:** MP4, WebM, MKV, AVI, MOV

## Accessing Wallpapers in Greeter

Press `F1` (Settings) → Backgrounds → Select your wallpaper or video

Both static and video wallpapers will appear in the same menu.

## Stop Video Wallpaper

From the wallpaper menu, select **Stop Video Wallpaper** to pause video playback. This uses gSlapper's IPC to pause without restarting the daemon.

## Troubleshooting

**Wallpaper not displaying:**

```bash
# Check gSlapper is installed
which gslapper

# Verify gSlapper socket exists
ls -la /tmp/sysc-greet-wallpaper.sock

# Check compositor config
cat /etc/greetd/niri-greeter-config.kdl | grep gslapper
```

**Multi-monitor issues:**

gSlapper should display wallpapers on all monitors. If only one monitor shows the wallpaper, ensure you're using gSlapper v1.0.9+ which fixed the multi-monitor rendering bug.
