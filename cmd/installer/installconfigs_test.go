package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerStartsInstallForMissingPackagedCompositor(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("SYSC_COMPOSITOR", "")
	for _, name := range []string{"niri", "sway"} {
		t.Run(name, func(t *testing.T) {
			m := newModel(false, nil)
			m.step = stepCompositorSelect
			m.packageManager = "xbps"
			for i, option := range compositorPickerOptions() {
				if option.name == name {
					m.compositorIndex = i
				}
			}
			next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			got := next.(model)
			if got.step != stepInstalling || cmd == nil || len(got.errors) != 0 {
				t.Fatalf("missing %s: step=%v, cmd nil=%v, errors=%v", name, got.step, cmd == nil, got.errors)
			}
		})
	}
}

func TestPickerRequiresNiriWhereAutomaticInstallIsUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("SYSC_COMPOSITOR", "niri")
	for _, manager := range []string{"apt", "dnf", "zypper", "pacman"} {
		t.Run(manager, func(t *testing.T) {
			m := newModel(false, nil)
			m.packageManager = manager
			next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			got := next.(model)
			if got.step != stepCompositorSelect || cmd != nil || len(got.errors) == 0 {
				t.Fatalf("%s: step=%v, cmd nil=%v, errors=%v; want refusal before installation", manager, got.step, cmd == nil, got.errors)
			}
		})
	}
}

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
			if !strings.Contains(string(data), "-r THIRD_PARTY_NOTICES.md /usr/share/sysc-greet/") {
				t.Error("third-party notices were not copied to the sysc-greet data directory")
			}
			if strings.Contains(string(data), "ignored.txt") {
				t.Fatal("installed a non-TOML file")
			}
		})
	}
}

func TestSyscCompositorGreeterBackend(t *testing.T) {
	pickerHas := func() *greeterCompositor {
		for _, compositor := range compositorPickerOptions() {
			if compositor.name == "sysc-compositor" {
				return &compositor
			}
		}
		return nil
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("SYSC_COMPOSITOR", "")
	if pickerHas() != nil {
		t.Fatal("experimental sysc-compositor appeared without its binary or an explicit request")
	}

	t.Setenv("SYSC_COMPOSITOR", "sysc-compositor")
	selected := pickerHas()
	if selected == nil {
		t.Fatal("SYSC_COMPOSITOR=sysc-compositor did not enable the picker entry")
	}
	if options := compositorPickerOptions(); options[0].name != "niri" || options[1].name != "sysc-compositor" {
		t.Fatalf("picker order = %s, %s; want niri, sysc-compositor", options[0].name, options[1].name)
	}
	if !strings.Contains(selected.label, "Experimental") {
		t.Fatalf("sysc-compositor label %q does not mark it experimental", selected.label)
	}

	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "sysc-compositor"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("SYSC_COMPOSITOR", "")
	if pickerHas() == nil {
		t.Fatal("installed sysc-compositor binary was not detected")
	}
	if selected.installable {
		t.Fatal("sysc-compositor must be installed separately")
	}
	if got := compositorBinaries(selected.name); len(got) != 1 || got[0] != "sysc-compositor" {
		t.Fatalf("compositorBinaries(%q) = %v", selected.name, got)
	}

	for _, tc := range []struct {
		compositor string
		want       string
	}{
		{"niri", "niri -c /etc/greetd/niri-greeter-config.kdl"},
		{"sysc-compositor", "sysc-compositor -c /etc/greetd/niri-greeter-config.kdl"},
	} {
		if got := niriGreeterCommand(tc.compositor); got != tc.want {
			t.Errorf("niriGreeterCommand(%q) = %q, want %q", tc.compositor, got, tc.want)
		}
	}
}

func TestHyprlandGreeterRequiresExplicitOptIn(t *testing.T) {
	t.Setenv("SYSC_COMPOSITOR", "")
	for _, compositor := range compositorPickerOptions() {
		if compositor.name == "hyprland" {
			t.Fatal("Hyprland appeared in the default installer picker")
		}
	}

	t.Setenv("SYSC_COMPOSITOR", "hyprland")
	var found bool
	for _, compositor := range compositorPickerOptions() {
		if compositor.name == "hyprland" {
			found = true
			if len(compositor.binaries) != 1 || compositor.binaries[0] != "start-hyprland" {
				t.Fatalf("Hyprland binaries = %v, want [start-hyprland]", compositor.binaries)
			}
		}
	}
	if !found {
		t.Fatal("explicit SYSC_COMPOSITOR=hyprland did not enable the Hyprland picker entry")
	}
	if got, want := hyprlandGreetdCommand, "start-hyprland -- --config /etc/greetd/hyprland-greeter-config.lua"; got != want {
		t.Fatalf("Hyprland greetd command = %q, want %q", got, want)
	}
}

func TestNiriGreeterQuitCommand(t *testing.T) {
	for _, tc := range []struct {
		compositor string
		want       string
	}{
		{"niri", "niri msg action quit --skip-confirmation"},
		{"sysc-compositor", "sysc-compositor msg action quit --skip-confirmation"},
	} {
		if got := niriGreeterQuitCommand(tc.compositor); got != tc.want {
			t.Errorf("niriGreeterQuitCommand(%q) = %q, want %q", tc.compositor, got, tc.want)
		}
	}
}
