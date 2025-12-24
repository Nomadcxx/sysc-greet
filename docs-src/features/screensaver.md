# Screensaver

sysc-greet includes a screensaver that activates after a configurable idle period.

## Configuration

Screensaver settings are stored in `/usr/share/sysc-greet/ascii_configs/screensaver.conf`.

### Format

```ini
# Idle time before activation (minutes)
idle_timeout=5

# Time/Date formats (Go time format)
time_format=3:04:05 PM
date_format=Monday, January 2, 2006

# Clock style: kompaktblk (default, 3 rows), phmvga (2 rows, crisp), dos_rebel (8 rows, retro), plain (single line)
clock_style=kompaktblk

# Animation settings
animate_on_start=false
animation_type=print
animation_speed=20

# ASCII variants (cycles every 5 minutes)
ascii_1=
  ╔═════════════════════╗
  ║    SEE YOU SPACE COWBOY ║
  ╚════════════════════════╝

ascii_2=
 .__________   .___  __ __   ___  __  .__________
  \___  _/ / __ \   / __ \ / __ \   \___  _/ __ \
   |   |   |   |  |   |  |   |  |    \  /   / /
   |___|___|___|___|___|___|___|___/   /___|___/
```

### Fields

| Field | Type | Default | Description |
|--------|--------|----------|-------------|
| idle_timeout | Integer | 5 | Minutes before screensaver activates |
| time_format | String | 3:04:05 PM | Go reference time format |
| date_format | String | Monday, January 2, 2006 | Go reference date format |
| clock_style | String | kompaktblk | Clock ASCII style |
| animate_on_start | Boolean | false | Animate ASCII on screensaver activation |
| animation_type | String | - | Print, Beams, Pour, etc. |
| animation_speed | Integer | 20 | Milliseconds per character (for print) |

### Clock Styles

| Style | Description |
|--------|-------------|
| kompaktblk | Default compact digital style (3 rows) |
| phmvga | Crisp VGA-style (2 rows) |
| dos_rebel | Retro 8-line DOS font |
| plain | Single line display |

## Activation

The screensaver activates automatically when no keyboard or mouse input is received for `idle_timeout` minutes.

**Exit conditions:**
- Any key press
- Mouse movement
- System activity

## Time Format

sysc-greet uses Go's reference time: `01/02 03:04:05PM '06 -0700`.

Common formats:
- `3:04:05 PM` - 12-hour format
- `15:04:05` - 24-hour format
- `Monday, January 2, 2006` - Full date
- `2006-01-02` - ISO format

## ASCII Animation

When screensaver activates, ASCII art variants cycle every 5 minutes. ASCII art is loaded from screensaver configuration file.

If `animate_on_start` is enabled and `animation_type` is set, the selected ASCII effect animates on screensaver activation.

## Behavior Notes

- Time updates every second while screensaver is active
- Screensaver state is independent of login/password modes
- Previous settings (username, password) remain intact when screensaver exits
- Preferences are not saved during screensaver mode
