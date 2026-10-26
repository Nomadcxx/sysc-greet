package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	initSystemd = "systemd"
	initRunit   = "runit"
	initOpenRC  = "openrc"
)

// greeterVT is the virtual terminal the greeter session runs on. The login
// prompt on this terminal has to give way to the greeter.
const greeterVT = 1

func greeterVTForInit(initSystem string) int {
	if initSystem == initOpenRC {
		return 7
	}
	return greeterVT
}

// Void's greetd package ships the _greeter system account with the home
// /var/lib/_greeter, and its own greetd config greets that account by name,
// so the greeter session has to run as the same user.
const (
	voidGreeterUser = "_greeter"
	voidGreeterHome = "/var/lib/_greeter"
)

// runitServiceRoots lists the directories that hold the live runit service
// links, relative to a root. Void keeps them in /etc/runit/runsvdir/default;
// other runit systems use /var/service or /etc/service.
var runitServiceRoots = []string{
	"etc/runit/runsvdir/default",
	"var/service",
	"etc/service",
}

// detectPlatform records the init system, the distro id and the account the
// greeter session runs as. It has to run before the install tasks.
func detectPlatform(m *model) {
	m.distroID, _ = osRelease()
	m.initSystem = detectInitSystem()
	resolveGreeterAccount(m)

	if m.debugMode {
		fmt.Fprintf(os.Stderr, "[DEBUG] Platform: distro=%s init=%s greeter=%s home=%s\n",
			m.distroID, m.initSystem, m.greeterUser, m.greeterHome)
	}
}

// detectInitSystem reports the init system that owns the greeter service.
func detectInitSystem() string {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return initSystemd
	}
	if info, err := os.Stat("/run/openrc"); err == nil && info.IsDir() {
		return initOpenRC
	}
	// Void ships runit with its live service tree under /etc/runit/runsvdir.
	if _, err := os.Stat("/etc/runit/runsvdir"); err == nil {
		return initRunit
	}
	if _, err := os.Stat("/var/service"); err == nil {
		return initRunit
	}
	return ""
}

// greeterAccount returns the account and home directory the greeter session
// runs as. Void and Gentoo package greetd with their own greeter accounts;
// other distros use the installer's account.
func greeterAccount(distroID string) (user string, home string) {
	switch distroID {
	case "void":
		return voidGreeterUser, voidGreeterHome
	case "gentoo":
		return "greetd", "/var/lib/greetd"
	default:
		return "greeter", "/var/lib/greeter"
	}
}

func resolveGreeterAccount(m *model) {
	m.greeterUser, m.greeterHome = greeterAccount(m.distroID)
}

// userExists reports whether the system knows the account.
func userExists(username string) bool {
	return exec.Command("id", username).Run() == nil
}

// groupExists reports whether the system knows the group.
func groupExists(name string) bool {
	return exec.Command("getent", "group", name).Run() == nil
}

// greeterGroups returns the groups the greeter account needs: video for the
// console, render for the GPU, input for the seat. Groups the system does not
// have are left out, because useradd and usermod fail on unknown groups.
func greeterGroups() string {
	groups := []string{"video"}
	for _, group := range []string{"render", "input"} {
		if groupExists(group) {
			groups = append(groups, group)
		}
	}
	return strings.Join(groups, ",")
}

// substituteGreeterPaths rewrites the paths and the account names that the
// greeter configs hardcode, so the same configs serve the Void _greeter
// account.
func substituteGreeterPaths(content, user, home string) string {
	if home != "" {
		content = strings.ReplaceAll(content, "/var/lib/greeter", home)
	}
	if user != "" {
		content = strings.ReplaceAll(content, `user = "greeter"`, fmt.Sprintf(`user = "%s"`, user))
		content = strings.ReplaceAll(content, `subject.user == "greeter"`, fmt.Sprintf(`subject.user == "%s"`, user))
	}
	return content
}

// writeGreeterScript writes a greeter script that has to keep its executable
// bit and that carries the greeter's own paths, so the greeter session also
// starts on a system whose greeter account is not /var/lib/greeter.
func writeGreeterScript(src, dst, user, home string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	content := substituteGreeterPaths(string(data), user, home)
	if err := os.WriteFile(dst, []byte(content), 0755); err != nil {
		return err
	}
	// WriteFile keeps the mode of a file that already exists, so the
	// executable bit has to be set again for an upgrade over an older file.
	return os.Chmod(dst, 0755)
}

// liveRunitDirs returns the runit service directories that exist under root.
func liveRunitDirs(root string) []string {
	dirs := []string{}
	for _, rel := range runitServiceRoots {
		path := filepath.Join(root, rel)
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil || !info.IsDir() {
			continue
		}
		dirs = append(dirs, path)
	}
	return dirs
}

// linkRunitService links a service from /etc/sv into every live runit service
// directory, so runit picks it up on the next scan.
func linkRunitService(root, name string) error {
	dirs := liveRunitDirs(root)
	if len(dirs) == 0 {
		return fmt.Errorf("no runit service directory found under %s", root)
	}
	for _, dir := range dirs {
		link := filepath.Join(dir, name)
		if _, err := os.Lstat(link); err == nil {
			continue
		}
		if err := os.Symlink(filepath.Join(root, "etc/sv", name), link); err != nil {
			return fmt.Errorf("failed to link %s: %w", link, err)
		}
	}
	return nil
}

// unlinkRunitService removes a service from every live runit service
// directory.
func unlinkRunitService(root, name string) error {
	for _, dir := range liveRunitDirs(root) {
		link := filepath.Join(dir, name)
		if _, err := os.Lstat(link); err != nil {
			continue
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("failed to remove %s: %w", link, err)
		}
	}
	return nil
}

// markServiceDown keeps a runit service from coming back. Void's runit-void
// enables every service in /etc/sv unless its directory carries a "down" file.
func markServiceDown(root, name string) error {
	dir := filepath.Join(root, "etc/sv", name)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	return os.WriteFile(filepath.Join(dir, "down"), nil, 0644)
}

// loginServicePackages returns the packages that provide the seat and the
// power actions on an init system without systemd-logind. polkit is started
// on demand by dbus, so it needs no service of its own.
func loginServicePackages(initSystem string) []string {
	if initSystem == initSystemd {
		return nil
	}
	return []string{"elogind", "dbus", "polkit"}
}

// enableRunitGreetd enables D-Bus and greetd, then frees the greeter VT.
// Void's D-Bus activates elogind when PAM opens the greeter session. Enabling
// elogind in runit too races that activation and retries a duplicate daemon.
func enableRunitGreetd(root string) error {
	services := []string{"dbus", "greetd"}
	for _, name := range services {
		if _, err := os.Stat(filepath.Join(root, "etc/sv", name)); err != nil {
			return fmt.Errorf("runit service missing at /etc/sv/%s - reinstall %s", name, name)
		}
	}
	for _, name := range services {
		if err := linkRunitService(root, name); err != nil {
			return err
		}
	}
	return disableAgetty(root, greeterVT)
}

// disableRunitGreetd disables the greetd runit service. The login services
// stay enabled: other sessions use them too.
func disableRunitGreetd() error {
	return unlinkRunitService("/", "greetd")
}

// disableAgetty stops the login prompt that would otherwise hold the greeter's
// virtual terminal. Void's greetd README asks for exactly this.
func disableAgetty(root string, vt int) error {
	name := fmt.Sprintf("agetty-tty%d", vt)
	if err := unlinkRunitService(root, name); err != nil {
		return err
	}
	return markServiceDown(root, name)
}

// enableOpenRCGreetd configures Gentoo's display-manager-init service to run
// greetd on its dedicated VT, then enables the services it needs at boot.
func enableOpenRCGreetd(root string) error {
	services := []string{"dbus", "elogind", "display-manager"}
	for _, name := range services {
		if _, err := os.Stat(filepath.Join(root, "etc/init.d", name)); err != nil {
			return fmt.Errorf("OpenRC service missing at /etc/init.d/%s: %w", name, err)
		}
	}

	configPath := filepath.Join(root, "etc/conf.d/display-manager")
	config, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read /etc/conf.d/display-manager: %w", err)
	}

	settings := []string{`DISPLAYMANAGER="greetd"`, "CHECKVT=7"}
	present := make(map[string]bool)
	for _, line := range strings.Split(string(config), "\n") {
		present[strings.TrimSpace(line)] = true
	}
	missing := make([]string, 0, len(settings))
	for _, setting := range settings {
		if !present[setting] {
			missing = append(missing, setting)
		}
	}
	if len(missing) > 0 {
		updated := string(config)
		if updated != "" {
			if !strings.HasSuffix(updated, "\n") {
				updated += "\n"
			}
			if !strings.HasSuffix(updated, "\n\n") {
				updated += "\n"
			}
		}
		updated += strings.Join(missing, "\n") + "\n"
		if err := os.WriteFile(configPath, []byte(updated), 0644); err != nil {
			return fmt.Errorf("write /etc/conf.d/display-manager: %w", err)
		}
	}

	for _, name := range services {
		if err := exec.Command("rc-update", "add", name, "default").Run(); err != nil {
			return fmt.Errorf("rc-update add %s default failed: %w", name, err)
		}
	}
	return nil
}
