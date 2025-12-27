# Screensaver Configuration

The login screen has a screensaver because waiting for authentication can be an aesthetic experience.

## Configuration File

`/usr/share/sysc-greet/ascii_configs/screensaver.conf`

```ini
# Idle time before activation (minutes)
idle_timeout=5

# Time/Date formats (Go time format)
time_format=3:04:05 PM
date_format=Monday, January 2, 2006

# Clock style: kompaktblk (default, 3 rows), phmvga (2 rows, crisp), dos_rebel (8 rows, retro), plain (single line)
clock_style=kompaktblk

# ASCII variants (cycles every 5 minutes)
ascii_1=
  ▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
   ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
  ▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀
  //  SEE YOU SPACE COWBOY //

ascii_2=
.________._______._______ .____/\ .________
|    ___/: .____/:_.  ___\:   /  \|    ___/
|___    \| : _/\ |  : |/\ |.  ___/|___    \
|       /|   /  \|    /  \|     \ |       /
|__:___/ |_.: __/|. _____/|      \|__:___/
   :        :/    :/      |___\  /   :
                  :            \/
```

## Time Format Reference

Go uses the reference time `01/02 03:04:05PM '06 -0700` (1234567 - memorable, right?).

**Common formats:**

| Format | Description |
|--------|-------------|
| `3:04:05 PM` | 12-hour with seconds |
| `15:04:05` | 24-hour with seconds |
| `Monday, January 2, 2006` | Full date |
| `2006-01-02` | ISO format |

## Clock Styles

| Style | Description |
|-------|-------------|
| kompaktblk | Default compact digital style (3 rows) |
| phmvga | Crisp VGA-style (2 rows) |
| dos_rebel | Retro 8-line DOS font |
| plain | Single line display |

## Behavior

- Activates after `idle_timeout` minutes of no input
- Exits on any keyboard or mouse input
- Cycles through ASCII variants every 5 minutes
- Time updates every second while active
- Previous login state (username, password) remains intact when exiting

## Testing

```bash
sysc-greet --test --screensaver
```

## Troubleshooting

**Screensaver not working:**

- Verify `/usr/share/sysc-greet/ascii_configs/screensaver.conf` exists
- Check file permissions: `ls -la /usr/share/sysc-greet/ascii_configs/screensaver.conf`
- Test in isolation: `sysc-greet --test --screensaver`
