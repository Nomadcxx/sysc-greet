---
title: "Troubleshooting"
description: "Common issues and solutions for sysc-greet."
---

Common issues and solutions for sysc-greet.

## Greeter Won't Start

### Check greetd Service Status

```bash
sudo systemctl status greetd
```

Expected output: `Active: active (running)`

If inactive, start the service:
```bash
sudo systemctl start greetd
```

### View greetd Logs

```bash
journalctl -u greetd -n 50
```

Look for:
- IPC socket creation errors
- Permission denied errors
- Compositor startup failures

### Verify Configuration

Check that greetd config exists at `/etc/greetd/config.toml`:

```bash
cat /etc/greetd/config.toml
```

Expected content:
```toml
[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
```

## IPC Client Errors

### "FATAL: Failed to create IPC client"

This error means sysc-greet cannot communicate with greetd.

**Causes:**
1. Running sysc-greet directly without greetd (not in production mode)
2. `GREETD_SOCK` environment variable not set
3. greetd socket doesn't exist

**Solutions:**

1. **Verify greetd is running:**
   ```bash
   sudo systemctl status greetd
   ```

2. **Check if socket exists:**
   ```bash
   ls -la /run/greetd/
   ```

3. **Test with greetd directly:**
   ```bash
   sudo -u greeter sysc-greet
   ```

## Wallpaper Issues

### gSlapper Not Working

**Check gSlapper status:**

```bash
ps aux | grep gslapper
```

**View gSlapper debug log:**

```bash
cat /tmp/sysc-greet-wallpaper.log
```

**Verify compositor config starts gSlapper:**

```bash
# For niri
cat /etc/greetd/niri-greeter-config.kdl

# For cagebreak
cat /etc/greetd/cagebreak-greeter-config

# For sway
cat /etc/greetd/sway-greeter-config

# For mango
cat /etc/greetd/mango-greeter-config.conf

# For Hyprland 0.55+
cat /etc/greetd/hyprland-greeter-config.lua
```

Look for gSlapper startup command.

**Install gSlapper if missing:**

```bash
# Arch Linux
sudo pacman -S gslapper

# Build from source
git clone https://github.com/Nomadcxx/gSlapper
cd gSlapper
make
sudo make install
```

### Wallpapers Not Appearing

**Check for themed wallpapers:**

```bash
ls -la /usr/share/sysc-greet/wallpapers/
```

Ensure `sysc-greet-{theme}.png` files exist for your selected theme.

**Check permissions:**

```bash
ls -la /var/lib/greeter/Pictures/wallpapers/
sudo chown -R greeter:greeter /var/lib/greeter/Pictures/wallpapers/
```

## Compositor Issues

### Screen Remains Black

**Check compositor logs:**

Compositors run as user processes, not systemd services. Check greetd logs which include compositor output:

```bash
journalctl -u greetd -n 50
```

For compositor-specific debugging, check their log files or stderr output in greetd logs.

**Verify compositor config syntax:**

```bash
# Test niri config
niri -c /etc/greetd/niri-greeter-config.kdl --check

# Check Hyprland 0.55+ Lua config syntax
luac -p /etc/greetd/hyprland-greeter-config.lua
```

Cagebreak has no syntax-check flag. Run it from a TTY to see parse errors — a bad config makes it exit immediately, so greetd relaunches the greeter in a loop:

```bash
cagebreak -e -c /etc/greetd/cagebreak-greeter-config
```

**Stuck at a black screen after login (cagebreak):** verify socat is installed. Without it the compositor never quits and greetd waits forever.

Check the mango greeter config (unknown keys fail the parse):

```bash
mango -c /etc/greetd/mango-greeter-config.conf -p
```

**Stuck at a black screen after login (mango):** verify `mmsg` is installed (it ships with `mangowm`) and `/etc/greetd/mango-greeter-session.sh` is executable.

### Kitty Terminal Issues

If kitty doesn't start:

```bash
# Check kitty config exists
cat /etc/greetd/kitty.conf

# Test kitty independently
kitty --start-as=fullscreen --config /etc/greetd/kitty.conf sysc-greet --test
```

The greeter config disables kitty's keyboard shortcuts and mouse actions. sysc-greet still receives keyboard input and mouse events.

An upgrade needs the updated kitty and compositor configs as well as the binary. If your package manager keeps a modified config, merge the changes while preserving your keyboard layout. The kitty config needs `clear_all_shortcuts yes`, `clear_all_mouse_actions yes` and `allow_remote_control no`. Compositor launches should use `XDG_CACHE_HOME=/var/lib/greeter/.cache`. Void packages use `/etc/greetd/sysc-greet/` and `/var/lib/_greeter/` instead.

## Debug Mode

sysc-greet opens a debug log only with `--debug`:

```bash
sysc-greet --test --debug
```

View the log under the home directory of the account running sysc-greet:

```bash
cat "$HOME/.cache/sysc-greet/debug.log"
```

The production greeter uses the greeter account's home. If sysc-greet cannot open that log, it tries `/tmp/sysc-greet-debug.log` with the same private-file checks.

Debug logs append and can contain usernames and session details. Remove old logs if you no longer need them; disabling debug mode leaves existing files intact.

Debug logs include:
- Key press events
- Mode transitions
- Animation state changes
- IPC communication
- Wallpaper commands

## Cache Issues

If preferences aren't being saved:

```bash
# Check cache directory exists and has correct permissions
ls -la /var/cache/sysc-greet/
sudo chown -R greeter:greeter /var/cache/sysc-greet/
sudo chmod 755 /var/cache/sysc-greet/
```

Clear cache if needed:
```bash
sudo rm -rf /var/cache/sysc-greet/*
```

## Getting Help

For unresolved issues:
1. Check existing [GitHub Issues](https://github.com/Nomadcxx/sysc-greet/issues)
2. Enable debug mode and collect logs
3. Review [Architecture Documentation](../development/architecture)
4. Check greetd documentation: https://git.sr.ht/~kennylevinsen/greetd/
