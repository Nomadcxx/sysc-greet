package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallConfigsBundledThemes(t *testing.T) {
	for _, tc := range []struct {
		name         string
		themes, fail bool
	}{
		{"no themes", false, false},
		{"bundled themes", true, false},
		{"copy failure", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			log := filepath.Join(dir, "commands")
			t.Setenv("INSTALL_LOG", log)
			t.Setenv("PATH", dir)
			t.Setenv("FAIL_INSTALL", "")
			if tc.fail {
				t.Setenv("FAIL_INSTALL", "1")
			}
			// Stub external commands to keep this check away from system directories.
			const script = `#!/bin/sh
printf '%s\n' "$*" >> "$INSTALL_LOG"
if [ "$1" = "-m" ] && [ "$FAIL_INSTALL" = "1" ]; then exit 1; fi
`
			for _, name := range []string{"mkdir", "cp", "install"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if tc.themes {
				if err := os.Mkdir("themes", 0o755); err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"ayu.toml", "blue.toml", "ignored.txt"} {
					if err := os.WriteFile(filepath.Join("themes", name), nil, 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			err := installConfigs(nil)
			if tc.fail {
				if err == nil || !strings.Contains(err.Error(), "ayu.toml") {
					t.Fatalf("want theme copy error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"ayu.toml", "blue.toml"} {
				command := "-m 644 themes/" + name + " /usr/share/sysc-greet/themes/"
				if got := strings.Contains(string(data), command); got != tc.themes {
					t.Errorf("theme install %q present = %v, want %v", command, got, tc.themes)
				}
			}
			if strings.Contains(string(data), "ignored.txt") {
				t.Fatal("installed a non-TOML file")
			}
		})
	}
}
