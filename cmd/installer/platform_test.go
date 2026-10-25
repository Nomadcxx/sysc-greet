package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGreeterAccount(t *testing.T) {
	cases := []struct {
		distro string
		user   string
		home   string
	}{
		{"void", "_greeter", "/var/lib/_greeter"},
		{"arch", "greeter", "/var/lib/greeter"},
		{"debian", "greeter", "/var/lib/greeter"},
		{"fedora", "greeter", "/var/lib/greeter"},
		{"", "greeter", "/var/lib/greeter"},
	}
	for _, tc := range cases {
		user, home := greeterAccount(tc.distro)
		if user != tc.user || home != tc.home {
			t.Errorf("greeterAccount(%q) = %q, %q; want %q, %q", tc.distro, user, home, tc.user, tc.home)
		}
	}
}

func TestSubstituteGreeterPaths(t *testing.T) {
	const config = `cd /var/lib/greeter && HOME=/var/lib/greeter
user = "greeter"
subject.user == "greeter"
`
	got := substituteGreeterPaths(config, "_greeter", "/var/lib/_greeter")
	want := `cd /var/lib/_greeter && HOME=/var/lib/_greeter
user = "_greeter"
subject.user == "_greeter"
`
	if got != want {
		t.Errorf("substituteGreeterPaths =\n%s\nwant\n%s", got, want)
	}

	// The default account must not change.
	same := substituteGreeterPaths(config, "greeter", "/var/lib/greeter")
	if same != config {
		t.Errorf("substituteGreeterPaths rewrote the default account:\n%s", same)
	}
}

// writeServiceTree builds a runit tree under root: the service directory in
// /etc/sv and the live service links in the given directories.
func writeServiceTree(t *testing.T, root string, service string, liveDirs ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "etc/sv", service), 0755); err != nil {
		t.Fatalf("mkdir service dir: %v", err)
	}
	for _, dir := range liveDirs {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatalf("mkdir live dir %s: %v", dir, err)
		}
	}
}

func TestLiveRunitDirs(t *testing.T) {
	root := t.TempDir()
	writeServiceTree(t, root, "greetd", "etc/runit/runsvdir/default")

	dirs := liveRunitDirs(root)
	if len(dirs) != 1 || dirs[0] != filepath.Join(root, "etc/runit/runsvdir/default") {
		t.Fatalf("liveRunitDirs = %v; want the void service tree", dirs)
	}
}

func TestLiveRunitDirsFollowsServiceSymlink(t *testing.T) {
	root := t.TempDir()
	writeServiceTree(t, root, "greetd", "etc/service")
	if err := os.MkdirAll(filepath.Join(root, "var"), 0755); err != nil {
		t.Fatalf("mkdir var: %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "etc/service"), filepath.Join(root, "var/service")); err != nil {
		t.Fatalf("symlink var/service: %v", err)
	}

	dirs := liveRunitDirs(root)
	// /var/service is a link to /etc/service here, so both paths are live.
	if len(dirs) != 2 {
		t.Fatalf("liveRunitDirs = %v; want the var/service link and etc/service", dirs)
	}
	found := false
	for _, dir := range dirs {
		if dir == filepath.Join(root, "var/service") {
			found = true
		}
	}
	if !found {
		t.Errorf("liveRunitDirs = %v; want the var/service link", dirs)
	}
}

func TestLinkRunitService(t *testing.T) {
	root := t.TempDir()
	writeServiceTree(t, root, "greetd", "etc/runit/runsvdir/default")

	if err := linkRunitService(root, "greetd"); err != nil {
		t.Fatalf("linkRunitService: %v", err)
	}
	link := filepath.Join(root, "etc/runit/runsvdir/default/greetd")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink %s: %v", link, err)
	}
	if want := filepath.Join(root, "etc/sv/greetd"); target != want {
		t.Errorf("link target = %q; want %q", target, want)
	}

	// Linking twice must not fail.
	if err := linkRunitService(root, "greetd"); err != nil {
		t.Fatalf("linkRunitService again: %v", err)
	}
}

func TestLinkRunitServiceWithoutServiceTree(t *testing.T) {
	root := t.TempDir()
	if err := linkRunitService(root, "greetd"); err == nil {
		t.Fatal("linkRunitService succeeded without a runit service directory")
	}
}

func TestUnlinkRunitService(t *testing.T) {
	root := t.TempDir()
	writeServiceTree(t, root, "agetty-tty1", "etc/runit/runsvdir/default")
	if err := linkRunitService(root, "agetty-tty1"); err != nil {
		t.Fatalf("linkRunitService: %v", err)
	}
	if err := unlinkRunitService(root, "agetty-tty1"); err != nil {
		t.Fatalf("unlinkRunitService: %v", err)
	}
	link := filepath.Join(root, "etc/runit/runsvdir/default/agetty-tty1")
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("%s still exists", link)
	}
}

func TestMarkServiceDown(t *testing.T) {
	root := t.TempDir()
	writeServiceTree(t, root, "agetty-tty1")

	if err := markServiceDown(root, "agetty-tty1"); err != nil {
		t.Fatalf("markServiceDown: %v", err)
	}
	down := filepath.Join(root, "etc/sv/agetty-tty1/down")
	if _, err := os.Stat(down); err != nil {
		t.Errorf("down file missing: %v", err)
	}

	// A service the system does not have is not an error.
	if err := markServiceDown(root, "agetty-tty9"); err != nil {
		t.Errorf("markServiceDown for a missing service: %v", err)
	}
}

func TestWriteGreeterScriptRewritesThePathsAndKeepsTheExecutableBit(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "session.sh")
	const script = "#!/bin/sh\ncd /var/lib/greeter || exit 1\nexport HOME=/var/lib/greeter\n"
	if err := os.WriteFile(src, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "etc", "greetd", "session.sh")

	if err := writeGreeterScript(src, dst, "_greeter", "/var/lib/_greeter"); err != nil {
		t.Fatalf("writeGreeterScript: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	want := "#!/bin/sh\ncd /var/lib/_greeter || exit 1\nexport HOME=/var/lib/_greeter\n"
	if string(data) != want {
		t.Errorf("session script =\n%s\nwant\n%s", data, want)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("session script mode = %v; want 0755", info.Mode().Perm())
	}
}

func TestWriteGreeterScriptSetsTheExecutableBitOnAnExistingFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "session.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "session.sh")
	// An older install left the script without the executable bit.
	if err := os.WriteFile(dst, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := writeGreeterScript(src, dst, "greeter", "/var/lib/greeter"); err != nil {
		t.Fatalf("writeGreeterScript: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("session script mode = %v; want 0755", info.Mode().Perm())
	}
}

func TestLoginServicePackages(t *testing.T) {
	if got := loginServicePackages(initSystemd); got != nil {
		t.Errorf("loginServicePackages(systemd) = %v; want nil, systemd-logind covers it", got)
	}
	want := []string{"elogind", "dbus", "polkit"}
	if got := loginServicePackages(initRunit); !reflect.DeepEqual(got, want) {
		t.Errorf("loginServicePackages(runit) = %v; want %v", got, want)
	}
}

func TestEnableRunitGreetdLinksTheLoginServices(t *testing.T) {
	root := t.TempDir()
	const live = "etc/runit/runsvdir/default"
	for _, service := range []string{"greetd", "dbus", "elogind", "agetty-tty1"} {
		writeServiceTree(t, root, service, live)
	}
	if err := linkRunitService(root, "agetty-tty1"); err != nil {
		t.Fatal(err)
	}

	if err := enableRunitGreetd(root); err != nil {
		t.Fatalf("enableRunitGreetd: %v", err)
	}
	for _, service := range []string{"dbus", "greetd"} {
		if _, err := os.Lstat(filepath.Join(root, live, service)); err != nil {
			t.Errorf("%s not linked: %v", service, err)
		}
	}
	if _, err := os.Lstat(filepath.Join(root, live, "elogind")); !os.IsNotExist(err) {
		t.Error("elogind linked: runit races D-Bus activation and retries the duplicate daemon")
	}
	if _, err := os.Lstat(filepath.Join(root, live, "agetty-tty1")); !os.IsNotExist(err) {
		t.Error("agetty-tty1 still holds the greeter VT")
	}
}

func TestEnableRunitGreetdUsesDBusActivation(t *testing.T) {
	root := t.TempDir()
	const live = "etc/runit/runsvdir/default"
	for _, service := range []string{"greetd", "dbus"} {
		writeServiceTree(t, root, service, live)
	}
	if err := enableRunitGreetd(root); err != nil {
		t.Fatalf("enableRunitGreetd without an elogind runit service: %v", err)
	}
}

func TestEnableRunitGreetdNeedsDBus(t *testing.T) {
	root := t.TempDir()
	const live = "etc/runit/runsvdir/default"
	writeServiceTree(t, root, "greetd", live)
	err := enableRunitGreetd(root)
	if err == nil || !strings.Contains(err.Error(), "dbus") {
		t.Fatalf("enableRunitGreetd without dbus = %v; want an error naming dbus", err)
	}
	if _, err := os.Lstat(filepath.Join(root, live, "greetd")); !os.IsNotExist(err) {
		t.Error("greetd enabled without a seat provider")
	}
}
