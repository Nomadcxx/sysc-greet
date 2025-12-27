# Backgrounds

sysc-greet supports background effects and wallpapers.

## Background Effects

Access via **F1** → **Backgrounds**:

| Effect | Description |
|--------|-------------|
| Fire | DOOM PSX-style fire effect |
| Matrix | Matrix rain (green characters) |
| ASCII Rain | Falling ASCII characters (theme colors) |
| Fireworks | Random firework explosions |
| Aquarium | Swimming fish and bubbles |
| None | Disable background effect |

## Wallpapers

Wallpapers take priority over background effects. Managed by [gSlapper](https://github.com/Nomadcxx/gSlapper).

### Themed Wallpapers

**Location:** `/usr/share/sysc-greet/wallpapers/`

Auto-switch when theme changes. Named `sysc-greet-{theme}.png`.

### Custom Wallpapers

**Location:** `/var/lib/greeter/Pictures/wallpapers/`

Add your own images or videos:

```bash
sudo cp ~/my-wallpaper.mp4 /var/lib/greeter/Pictures/wallpapers/
sudo chown greeter:greeter /var/lib/greeter/Pictures/wallpapers/my-wallpaper.mp4
```

Access via **F1** → **Wallpaper**.

## Priority

1. Custom wallpaper (if selected)
2. Themed wallpaper (auto-matches theme)
3. Background effect (if no wallpaper)
