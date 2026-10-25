package main

import (
	"reflect"
	"testing"
)

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
	for _, manager := range []string{"pacman", "dnf", "yum", "zypper", "xbps"} {
		if cmd := refreshCommand(manager); cmd != nil {
			t.Errorf("refreshCommand(%q) = %v; want nil", manager, cmd.Args)
		}
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
