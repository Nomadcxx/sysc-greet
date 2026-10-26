package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestInstallKittyVoidIncludesFont(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run("kitty_installed="+strconv.FormatBool(existing), func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("PATH", dir)
			argsFile := filepath.Join(dir, "args")
			t.Setenv("KITTY_INSTALL_ARGS", argsFile)
			if err := os.WriteFile(filepath.Join(dir, "xbps-install"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$KITTY_INSTALL_ARGS\"\n"), 0755); err != nil {
				t.Fatal(err)
			}
			if existing {
				if err := os.WriteFile(filepath.Join(dir, "kitty"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			if err := installKitty(&model{packageManager: "xbps"}); err != nil {
				t.Fatal(err)
			}
			args, err := os.ReadFile(argsFile)
			if err != nil {
				t.Fatalf("kitty installed=%v: font install did not run: %v", existing, err)
			}
			if got := strings.Fields(string(args)); !reflect.DeepEqual(got, []string{"-Sy", "kitty", "dejavu-fonts-ttf"}) {
				t.Fatalf("install args = %v; want kitty and a monospace font", got)
			}
		})
	}
}

func TestInstallCommand(t *testing.T) {
	cases := []struct {
		manager  string
		packages []string
		want     []string
	}{
		{"pacman", []string{"greetd"}, []string{"pacman", "-S", "--noconfirm", "--needed", "greetd"}},
		{"apt", []string{"greetd", "kitty"}, []string{"apt-get", "install", "-y", "greetd", "kitty"}},
		{"dnf", []string{"sway"}, []string{"dnf", "install", "-y", "sway"}},
		{"yum", []string{"sway"}, []string{"yum", "install", "-y", "sway"}},
		{"zypper", []string{"sway"}, []string{"zypper", "install", "-y", "sway"}},
		// xbps-install is the standalone install front end, so it takes no
		// "install" subcommand.
		{"xbps", []string{"greetd", "kitty"}, []string{"xbps-install", "-Sy", "greetd", "kitty"}},
		// The cagebreak greeter quits through socat, so Void has to install it.
		{"xbps", []string{"socat"}, []string{"xbps-install", "-Sy", "socat"}},
		{"emerge", []string{"greetd", "kitty"}, []string{"emerge", "--noreplace", "--ask=n", "--autounmask=n", "gui-libs/greetd", "x11-terms/kitty"}},
	}
	for _, tc := range cases {
		cmd := installCommand(tc.manager, tc.packages...)
		if cmd == nil {
			t.Errorf("installCommand(%q) = nil", tc.manager)
			continue
		}
		if !reflect.DeepEqual(cmd.Args, tc.want) {
			t.Errorf("installCommand(%q, %v) = %v; want %v", tc.manager, tc.packages, cmd.Args, tc.want)
		}
	}
}

func TestInstalledCommand(t *testing.T) {
	cases := []struct {
		manager string
		want    []string
	}{
		{"pacman", []string{"pacman", "-Q", "greetd"}},
		{"apt", []string{"dpkg-query", "-W", "-f=${Status}", "greetd"}},
		{"dnf", []string{"rpm", "-q", "greetd"}},
		{"yum", []string{"rpm", "-q", "greetd"}},
		{"zypper", []string{"rpm", "-q", "greetd"}},
		// Only the show mode reports presence through its exit status.
		{"xbps", []string{"xbps-query", "-S", "greetd"}},
		{"emerge", []string{"portageq", "has_version", "/", "gui-libs/greetd"}},
	}
	for _, tc := range cases {
		cmd := installedCommand(tc.manager, "greetd")
		if cmd == nil {
			t.Errorf("installedCommand(%q) = nil", tc.manager)
			continue
		}
		if !reflect.DeepEqual(cmd.Args, tc.want) {
			t.Errorf("installedCommand(%q) = %v; want %v", tc.manager, cmd.Args, tc.want)
		}
	}
}

func TestAvailableCommand(t *testing.T) {
	cases := []struct {
		manager string
		want    []string
	}{
		{"pacman", []string{"pacman", "-Si", "greetd"}},
		{"apt", []string{"apt-cache", "show", "greetd"}},
		{"dnf", []string{"dnf", "info", "greetd"}},
		{"yum", []string{"yum", "info", "greetd"}},
		{"zypper", []string{"zypper", "info", "greetd"}},
		{"xbps", []string{"xbps-query", "-R", "-S", "greetd"}},
		{"emerge", []string{"emerge", "--pretend", "--nodeps", "--quiet", "--autounmask=n", "gui-libs/greetd"}},
	}
	for _, tc := range cases {
		cmd := availableCommand(tc.manager, "greetd")
		if cmd == nil {
			t.Errorf("availableCommand(%q) = nil", tc.manager)
			continue
		}
		if !reflect.DeepEqual(cmd.Args, tc.want) {
			t.Errorf("availableCommand(%q) = %v; want %v", tc.manager, cmd.Args, tc.want)
		}
	}
}

func TestRefreshCommand(t *testing.T) {
	cmd := refreshCommand("apt")
	if cmd == nil || !reflect.DeepEqual(cmd.Args, []string{"apt-get", "update"}) {
		t.Errorf("refreshCommand(apt) = %v; want [apt-get update]", cmd)
	}
	for _, manager := range []string{"pacman", "dnf", "yum", "zypper", "xbps", "emerge"} {
		if cmd := refreshCommand(manager); cmd != nil {
			t.Errorf("refreshCommand(%q) = %v; want nil", manager, cmd.Args)
		}
	}
}

func TestGentooPackageAtoms(t *testing.T) {
	for name, atom := range map[string]string{
		"greetd": "gui-libs/greetd", "kitty": "x11-terms/kitty",
		"socat": "net-misc/socat", "sway": "gui-wm/sway",
		"elogind": "sys-auth/elogind", "dbus": "sys-apps/dbus", "polkit": "sys-auth/polkit",
		"gcc": "sys-devel/gcc", "meson": "dev-build/meson", "ninja": "dev-build/ninja",
		"git": "dev-vcs/git", "pkg-config": "virtual/pkgconfig",
		"display-manager-init": "gui-libs/display-manager-init",
		"media-libs/gstreamer": "media-libs/gstreamer",
	} {
		cmd := installCommand("emerge", name)
		if cmd == nil || cmd.Args[len(cmd.Args)-1] != atom {
			t.Errorf("installCommand(emerge, %q) = %v; want atom %s", name, cmd, atom)
		}
	}
}

func TestGentooGStreamerPackages(t *testing.T) {
	want := []string{
		"media-libs/gstreamer", "media-libs/gst-plugins-base",
		"media-libs/gst-plugins-good", "media-libs/gst-plugins-bad",
		"dev-libs/wayland", "dev-libs/wayland-protocols",
		"media-libs/mesa", "virtual/pkgconfig",
	}
	if got := getGStreamerDeps("emerge"); !reflect.DeepEqual(got, want) {
		t.Fatalf("getGStreamerDeps(emerge) = %v; want %v", got, want)
	}
}

func TestInstallGentooSwayUsesPortage(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	t.Setenv("PATH", dir)
	t.Setenv("EMERGE_ARGS", argsFile)
	if err := os.WriteFile(filepath.Join(dir, "emerge"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$EMERGE_ARGS\"\n"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := installCompositor(&model{packageManager: "emerge", selectedCompositor: "sway"}); err != nil {
		t.Fatalf("installCompositor(sway): %v", err)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"--noreplace", "--ask=n", "--autounmask=n", "gui-wm/sway"}
	if got := strings.Fields(string(args)); !reflect.DeepEqual(got, want) {
		t.Fatalf("emerge args = %v; want %v", got, want)
	}
}

func TestInstallGentooGreetdInstallsDisplayManagerInit(t *testing.T) {
	for _, greetdInstalled := range []bool{false, true} {
		t.Run("greetd_installed="+strconv.FormatBool(greetdInstalled), func(t *testing.T) {
			dir := t.TempDir()
			argsFile := filepath.Join(dir, "args")
			t.Setenv("PATH", dir)
			t.Setenv("EMERGE_ARGS", argsFile)
			if err := os.WriteFile(filepath.Join(dir, "emerge"), []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$EMERGE_ARGS\"\n"), 0755); err != nil {
				t.Fatal(err)
			}

			m := &model{greetdInstalled: greetdInstalled, packageManager: "emerge", initSystem: initOpenRC}
			if err := installGreetd(m); err != nil {
				t.Fatalf("installGreetd: %v", err)
			}
			args, err := os.ReadFile(argsFile)
			if err != nil {
				t.Fatal(err)
			}
			var want []string
			if !greetdInstalled {
				want = append(want, "--noreplace --ask=n --autounmask=n gui-libs/greetd")
			}
			want = append(want, "--noreplace --ask=n --autounmask=n gui-libs/display-manager-init")
			if got := strings.Split(strings.TrimSpace(string(args)), "\n"); !reflect.DeepEqual(got, want) {
				t.Fatalf("emerge invocations = %v; want %v", got, want)
			}
		})
	}
}

func TestUnknownPackageManager(t *testing.T) {
	if installCommand("", "greetd") != nil {
		t.Error("installCommand with no package manager should be nil")
	}
	if installedCommand("portage", "greetd") != nil {
		t.Error("installedCommand with an unknown package manager should be nil")
	}
	if availableCommand("portage", "greetd") != nil {
		t.Error("availableCommand with an unknown package manager should be nil")
	}
}

func TestGslapperBuildToolsIncludeACompiler(t *testing.T) {
	found := false
	for _, tool := range gslapperBuildTools {
		if tool == "cc" {
			found = true
		}
	}
	if !found {
		t.Fatalf("gslapperBuildTools = %v; want a C compiler, Void's base-system has none", gslapperBuildTools)
	}
	for manager := range packageManagerCommands {
		if got := getBinaryPackageName("cc", manager); got != "gcc" {
			t.Errorf("getBinaryPackageName(cc, %s) = %q; want gcc", manager, got)
		}
	}
}
