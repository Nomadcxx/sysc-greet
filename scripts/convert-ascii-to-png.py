#!/usr/bin/env python3
"""
Convert ASCII art to PNG with transparent background.
"""

from PIL import Image, ImageDraw, ImageFont
import sys
import os

def ascii_to_png(ascii_file, output_file, font_size=16, padding=20):
    """Convert ASCII art file to PNG image."""
    
    # Read ASCII art
    with open(ascii_file, 'r') as f:
        lines = [line.rstrip('\n') for line in f.readlines()]
    
    # Remove empty lines at start/end
    while lines and not lines[0].strip():
        lines.pop(0)
    while lines and not lines[-1].strip():
        lines.pop()
    
    if not lines:
        print("Error: No content in ASCII file")
        return False
    
    # Calculate dimensions
    max_width = max(len(line) for line in lines)
    num_lines = len(lines)
    
    # Try to use a monospace font
    try:
        # Try common monospace fonts
        font_paths = [
            '/usr/share/fonts/TTF/DejaVuSansMono.ttf',
            '/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf',
            '/usr/share/fonts/truetype/liberation/LiberationMono-Regular.ttf',
            '/System/Library/Fonts/Menlo.ttc',
        ]
        font = None
        for path in font_paths:
            if os.path.exists(path):
                font = ImageFont.truetype(path, font_size)
                break
        
        if font is None:
            # Fall back to default font
            font = ImageFont.load_default()
    except Exception:
        font = ImageFont.load_default()
    
    # Calculate text size
    # Use a test character to measure
    test_img = Image.new('RGBA', (1, 1), (0, 0, 0, 0))
    test_draw = ImageDraw.Draw(test_img)
    bbox = test_draw.textbbox((0, 0), 'M', font=font)
    char_width = bbox[2] - bbox[0]
    char_height = bbox[3] - bbox[1]
    
    # Calculate image dimensions
    img_width = max_width * char_width + (padding * 2)
    img_height = num_lines * char_height + (padding * 2)
    
    # Create transparent image
    img = Image.new('RGBA', (img_width, img_height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Draw ASCII art (white text)
    y = padding
    for line in lines:
        x = padding
        draw.text((x, y), line, fill=(255, 255, 255, 255), font=font)
        y += char_height
    
    # Save PNG
    img.save(output_file, 'PNG')
    print(f"Created {output_file} ({img_width}x{img_height})")
    return True

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage: convert-ascii-to-png.py <input.txt> <output.png> [font_size]")
        sys.exit(1)
    
    ascii_file = sys.argv[1]
    output_file = sys.argv[2]
    font_size = int(sys.argv[3]) if len(sys.argv) > 3 else 16
    
    if not os.path.exists(ascii_file):
        print(f"Error: File not found: {ascii_file}")
        sys.exit(1)
    
    success = ascii_to_png(ascii_file, output_file, font_size)
    sys.exit(0 if success else 1)
