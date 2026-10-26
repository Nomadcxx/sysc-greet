package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestKittyGreeterDisablesActions(t *testing.T) {
	kitty, err := exec.LookPath("kitty")
	if err != nil {
		t.Skip("kitty is required to check the effective greeter action maps")
	}
	home := t.TempDir()
	configHome := filepath.Join(home, ".config")
	hostileConfig := filepath.Join(configHome, "kitty", "kitty.conf")
	if err := os.MkdirAll(filepath.Dir(hostileConfig), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hostileConfig, []byte("map ctrl+shift+f3 command_palette\nmouse_map ctrl+left press grabbed,ungrabbed mouse_handle_click link\nallow_remote_control yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	shippedConfig, err := filepath.Abs("../../config/kitty-greeter.conf")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		config string
	}{
		{name: "primary", config: shippedConfig},
		{name: "secondary_with_config", config: hostileConfig},
		{name: "secondary_without_config"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var overrides []string
			if tc.name != "primary" {
				const executable = "/usr/bin/sysc-greet"
				args := secondarySaverCommandArgs(tc.config, executable, "nord", "2026-10-26T12:00:00Z", "ready")
				for i := 0; i < len(args) && args[i] != executable; i++ {
					if args[i] == "-o" {
						i++
						overrides = append(overrides, args[i])
					}
				}
			}
			data, err := json.Marshal(map[string]any{"config": tc.config, "overrides": overrides})
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(kitty, "+runpy", `
import json, os
from kitty.config import load_config

case = json.loads(os.environ["SYSC_TEST_KITTY_CASE"])
errors = []
paths = (case["config"],) if case["config"] else ()
overrides = [option.replace("=", " ", 1) for option in case["overrides"] or []]
opts = load_config(*paths, overrides=overrides, accumulate_bad_lines=errors)
if errors:
    raise ValueError(errors)
modes = getattr(opts, "keyboard_modes", None)
maps = [mode.keymap for mode in modes.values()] if modes is not None else [opts.keymap]
if any(maps):
    raise ValueError("kitty keyboard actions remain enabled")
if opts.mousemap:
    raise ValueError("kitty mouse actions remain enabled")
if opts.allow_remote_control not in ("no", False):
    raise ValueError("kitty remote control is enabled")
`)
			cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+configHome, "XDG_CACHE_HOME="+filepath.Join(home, ".cache"), "SYSC_TEST_KITTY_CASE="+string(data))
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("kitty rejected the greeter security policy: %v\n%s", err, output)
			}
		})
	}
}
