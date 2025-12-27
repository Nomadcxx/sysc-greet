#!/usr/bin/env python3
"""
Generate sysc-greet theme wallpapers from SYSC ASCII art.
Uses Pillow to render ASCII art onto themed backgrounds.

Usage:
    python3 generate-wallpapers.py [--ascii /path/to/ascii.txt]
"""

from PIL import Image, ImageDraw, ImageFont
import os
import sys
import argparse

# Theme definitions: name -> (bg_color, text_color)
THEMES = {
    "default": ("#1a1a2e", "#e94560"),
    "gruvbox": ("#282828", "#fe8019"),
    "material": ("#263238", "#80cbc4"),
    "nord": ("#2e3440", "#81a1c1"),
    "dracula": ("#282a36", "#bd93f9"),
    "catppuccin": ("#1e1e2e", "#cba6f7"),
    "tokyo-night": ("#1a1b26", "#7aa2f7"),
    "solarized": ("#002b36", "#268bd2"),
    "monochrome": ("#1a1a1a", "#ffffff"),
    "transishardjob": ("#1a1a1a", "#5BCEFA"),
    "rama": ("#2b2d42", "#ef233c"),
    "eldritch": ("#212337", "#37f499"),
    "dark": ("#000000", "#ffffff"),
}

# Image dimensions (3840x2160 for 4K UHD - scales down nicely to all displays)
WIDTH = 3840
HEIGHT = 2160


def hex_to_rgb(hex_color):
    """Convert hex color to RGB tuple."""
    hex_color = hex_color.lstrip('#')
    return tuple(int(hex_color[i:i+2], 16) for i in (0, 2, 4))


def load_ascii_art(filepath):
    """
    Load ASCII art from file, preserving exact spacing.
    Returns list of lines with original formatting intact.
    """
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Split into lines, keeping exact content
    lines = content.split('\n')

    # Remove trailing empty lines only
    while lines and lines[-1] == '':
        lines.pop()

    # Remove leading empty lines only
    while lines and lines[0] == '':
        lines.pop(0)

    return lines


def get_monospace_font(size):
    """Try to load a monospace font, fallback to default."""
    font_paths = [
        "/usr/share/fonts/TTF/JetBrainsMono-Regular.ttf",
        "/usr/share/fonts/TTF/DejaVuSansMono.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
        "/usr/share/fonts/liberation-mono/LiberationMono-Regular.ttf",
        "/usr/share/fonts/TTF/FiraCode-Regular.ttf",
        "/usr/share/fonts/noto/NotoSansMono-Regular.ttf",
    ]

    for path in font_paths:
        if os.path.exists(path):
            try:
                return ImageFont.truetype(path, size)
            except Exception:
                continue

    # Fallback to default font
    try:
        return ImageFont.truetype("DejaVuSansMono.ttf", size)
    except Exception:
        print("Warning: Using default font (may not be monospace)")
        return ImageFont.load_default()


def find_max_line_width(lines):
    """Find the maximum character width among all lines."""
    return max(len(line) for line in lines) if lines else 0


def generate_wallpaper(theme_name, bg_color, text_color, lines, output_dir):
    """Generate a wallpaper for a specific theme."""

    # Create image with background color
    bg_rgb = hex_to_rgb(bg_color)
    img = Image.new('RGB', (WIDTH, HEIGHT), bg_rgb)
    draw = ImageDraw.Draw(img)

    # Load font
    font_size = 18
    font = get_monospace_font(font_size)

    # Calculate dimensions using a reference character for consistent spacing
    # Use block character to get accurate monospace width
    ref_bbox = draw.textbbox((0, 0), "█", font=font)
    char_width = ref_bbox[2] - ref_bbox[0]
    char_height = ref_bbox[3] - ref_bbox[1]
    line_height = int(char_height * 0.85)  # Overlap slightly for connected look

    # Find the widest line in characters
    max_chars = find_max_line_width(lines)

    # Calculate total text block dimensions
    text_block_width = max_chars * char_width
    text_block_height = len(lines) * line_height

    # Center the text block
    start_x = (WIDTH - text_block_width) // 2
    start_y = (HEIGHT - text_block_height) // 2

    # Draw each line
    text_rgb = hex_to_rgb(text_color)

    for i, line in enumerate(lines):
        # Expand tabs to spaces (8-space tabs to match terminal default)
        expanded_line = line.replace('\t', '        ')
        y_pos = start_y + (i * line_height)
        draw.text((start_x, y_pos), expanded_line, font=font, fill=text_rgb)

    # Save the image
    output_path = os.path.join(output_dir, f"sysc-greet-{theme_name}.png")
    img.save(output_path, 'PNG', optimize=True)
    print(f"Generated: {output_path}")

    return output_path


def main():
    parser = argparse.ArgumentParser(description='Generate sysc-greet theme wallpapers')
    script_dir = os.path.dirname(os.path.abspath(__file__))
    default_ascii = os.path.join(script_dir, 'ASCII.txt')
    parser.add_argument('--ascii', '-a', type=str,
                        default=default_ascii,
                        help='Path to ASCII art file (default: scripts/ASCII.txt)')
    parser.add_argument('--output', '-o', type=str, default=None,
                        help='Output directory (default: ../wallpapers/)')
    parser.add_argument('--theme', '-t', type=str, default=None,
                        help='Generate only specific theme')
    args = parser.parse_args()

    # Determine output directory
    if args.output:
        output_dir = args.output
    else:
        script_dir = os.path.dirname(os.path.abspath(__file__))
        project_dir = os.path.dirname(script_dir)
        output_dir = os.path.join(project_dir, "wallpapers")

    # Create output directory if needed
    os.makedirs(output_dir, exist_ok=True)

    # Load ASCII art
    ascii_path = args.ascii
    if not os.path.exists(ascii_path):
        print(f"Error: ASCII art file not found: {ascii_path}")
        sys.exit(1)

    print(f"Loading ASCII art from: {ascii_path}")
    lines = load_ascii_art(ascii_path)

    print(f"Loaded {len(lines)} lines, max width: {find_max_line_width(lines)} chars")
    print(f"Output directory: {output_dir}")
    print(f"Image size: {WIDTH}x{HEIGHT}")
    print()

    # Preview first few lines
    print("ASCII preview:")
    for i, line in enumerate(lines[:3]):
        print(f"  {i+1}: '{line[:60]}...'") if len(line) > 60 else print(f"  {i+1}: '{line}'")
    print()

    # Generate wallpapers
    themes_to_generate = THEMES.items()
    if args.theme:
        if args.theme not in THEMES:
            print(f"Error: Unknown theme '{args.theme}'")
            print(f"Available themes: {', '.join(THEMES.keys())}")
            sys.exit(1)
        themes_to_generate = [(args.theme, THEMES[args.theme])]

    for theme_name, (bg_color, text_color) in themes_to_generate:
        generate_wallpaper(theme_name, bg_color, text_color, lines, output_dir)

    print()
    print(f"Done! Generated {len(list(themes_to_generate))} wallpaper(s).")
    print()
    print("To install, copy to /usr/share/sysc-greet/wallpapers/:")
    print(f"  sudo cp {output_dir}/*.png /usr/share/sysc-greet/wallpapers/")


if __name__ == "__main__":
    main()
