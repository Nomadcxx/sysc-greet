---
title: "Testing"
description: "Guide for testing sysc-greet during development."
---

Guide for testing sysc-greet during development.

## Test Mode

sysc-greet includes a test mode that bypasses greetd authentication, allowing you to test the greeter without locking your session.

### Running Test Mode

```bash
# Basic test mode
sysc-greet --test

# Test mode with debug logging
sysc-greet --test --debug

# Test mode with screensaver enabled
sysc-greet --test --screensaver
```

### Fullscreen Testing

For accurate preview, test in fullscreen:

```bash
kitty --start-as=fullscreen sysc-greet --test
```

## Debug Logging

Enable debug logging to diagnose issues:

```bash
sysc-greet --test --debug
```

Debug logs are written to `/tmp/sysc-greet-debug.log` and include:
- Key press events
- Mode transitions
- Animation state changes
- IPC communication (if not in test mode)
- Wallpaper commands
- Theme changes
- Preference saves

### Viewing Debug Logs

```bash
# View entire log
cat /tmp/sysc-greet-debug.log

# Follow log in real-time
tail -f /tmp/sysc-greet-debug.log

# View last 50 lines
tail -n 50 /tmp/sysc-greet-debug.log
```

## Test Mode Restrictions

**Important:** Test mode is automatically disabled if `GREETD_SOCK` environment variable is set. This prevents accidental use of test mode in production environments.

## Command Line Options

### Theme Testing

Test with a specific theme:

```bash
sysc-greet --test --theme dracula
sysc-greet --test --theme nord
sysc-greet --test --theme tokyo-night
```

Available theme names: dracula, gruvbox, material, nord, tokyo-night, catppuccin, solarized, monochrome, transishardjob, eldritch


### Username Caching

Test username caching:

```bash
sysc-greet --test --remember-username
```

## Testing Checklist

When testing changes, verify:

- [ ] All themes render correctly
- [ ] Border styles display properly
- [ ] Background effects animate smoothly
- [ ] ASCII effects work for all variants
- [ ] Screensaver activates after timeout
- [ ] Preferences save and restore correctly
- [ ] Session selection works
- [ ] Keyboard navigation functions properly
- [ ] Test mode works without greetd
- [ ] Debug logging captures expected events

## Integration Testing

For full integration testing with greetd:

1. Build and install sysc-greet
2. Configure greetd service
3. Restart greetd: `sudo systemctl restart greetd`
4. Test actual login flow
5. Check greetd logs: `journalctl -u greetd -n 50`

**Note:** Integration testing requires root access and will lock your session. Use test mode for most development work.

## Performance Testing

Monitor resource usage during testing:

```bash
# Monitor CPU/memory
htop

# Monitor in separate terminal
watch -n 1 'ps aux | grep sysc-greet'
```

Background effects (fire, matrix, rain) are CPU-intensive. Verify they don't cause excessive resource usage on target hardware.

## TTY Compatibility Testing

Test on different terminal types:

```bash
# TrueColor terminal (kitty)
kitty sysc-greet --test

# ANSI256 terminal
TERM=xterm-256color sysc-greet --test

# Basic TTY (requires actual TTY)
# Switch to TTY: Ctrl+Alt+F2
# Run: sysc-greet --test
```

Verify color fallbacks work correctly across terminal types.

## Common Test Scenarios

### Test Theme Switching
1. Start greeter: `sysc-greet --test`
2. Press F1 (Settings)
3. Navigate to Themes
4. Cycle through all themes
5. Verify colors update correctly

### Test Background Effects
1. Start greeter: `sysc-greet --test`
2. Press F1 (Settings)
3. Navigate to Backgrounds
4. Test each effect (Fire, Matrix, Rain, Fireworks, Aquarium)
5. Verify animations render correctly

### Test Screensaver
1. Start greeter: `sysc-greet --test --screensaver`
2. Wait for idle timeout (or manually trigger)
3. Verify screensaver displays
4. Press any key to exit
5. Verify login state is preserved

### Test ASCII Variants
1. Start greeter: `sysc-greet --test`
2. Press Page Up/Down
3. Verify ASCII art cycles through variants
4. Verify selected variant persists after restart

## Troubleshooting Test Mode

**Test mode doesn't start:**
- Check `GREETD_SOCK` is not set: `unset GREETD_SOCK`
- Verify binary is executable: `chmod +x sysc-greet`

**Debug log not created:**
- Check `/tmp/` is writable
- Verify `--debug` flag is used
- Check file permissions: `ls -la /tmp/sysc-greet-debug.log`

**Colors look wrong:**
- Test in fullscreen kitty for accurate preview
- Check terminal color profile support
- Verify theme files are correct

For more troubleshooting, see [Troubleshooting Guide](../getting-started/troubleshooting).
