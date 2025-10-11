#!/bin/bash
# CHANGED 2025-10-10 - Generate theme-aware wallpapers with SYSC branding - Problem: Need themed backgrounds for multi-monitor

set -e

# SYSC ASCII art for branding
SYSC_ASCII="  
▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀ 
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀   
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀     
//SEE YOU IN SPACE COWBOY//"

# Output directory
OUTPUT_DIR="wallpapers"
mkdir -p "$OUTPUT_DIR"

# Theme definitions: name, background, foreground
declare -A THEMES
THEMES[dracula]="#282a36 #bd93f9"
THEMES[catppuccin]="#1e1e2e #cba6f7"
THEMES[nord]="#2e3440 #81a1c1"
THEMES[tokyo-night]="#1a1b26 #7aa2f7"
THEMES[gruvbox]="#282828 #fe8019"
THEMES[material]="#263238 #80cbc4"
THEMES[solarized]="#002b36 #268bd2"
THEMES[monochrome]="#1a1a1a #ffffff"
THEMES[transishardjob]="#1a1a1a #5BCEFA"
THEMES[default]="#1a1a1a #8b5cf6"

# Check for ImageMagick
if ! command -v magick &> /dev/null; then
    echo "Error: ImageMagick not found. Install with: sudo pacman -S imagemagick"
    exit 1
fi

echo "Generating theme wallpapers..."

for theme in "${!THEMES[@]}"; do
    colors=(${THEMES[$theme]})
    bg_color="${colors[0]}"
    fg_color="${colors[1]}"

    echo "  Creating wallpaper for theme: $theme"

    # Create 1920x1080 wallpaper with ASCII branding
    magick -size 1920x1080 \
        xc:"$bg_color" \
        -font "DejaVu-Sans-Mono" \
        -pointsize 20 \
        -fill "$fg_color" \
        -gravity center \
        -annotate +0+0 "$SYSC_ASCII" \
        -alpha set -channel A -evaluate multiply 0.3 \
        "$OUTPUT_DIR/sysc-greet-${theme}.png"
done

echo "✓ Generated ${#THEMES[@]} theme wallpapers in $OUTPUT_DIR/"
echo ""
echo "Wallpapers created:"
ls -1 "$OUTPUT_DIR"/sysc-greet-*.png
