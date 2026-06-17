package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	initSystemd = "systemd"
	initRunit   = "runit"
)

// detectPlatform fills init system, distro id, and default greeter account hints.
func detectPlatform(m *model) {
	m.distroID = readOSRelease("ID")
	m.initSystem = detectInitSystem()
	resolveGreeterAccount(m)

	if m.debugMode {
		fmt.Fprintf(os.Stderr, "[DEBUG] Platform: distro=%s init=%s greeter=%s home=%s\n",
			m.distroID, m.initSystem, m.greeterUser, m.greeterHome)
	}
}

func readOSRelease(key string) string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer file.Close()

	prefix := key + "="
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			value := strings.TrimPrefix(line, prefix)
			return strings.Trim(value, `"`)
		}
	}
	return ""
}

func detectInitSystem() string {
	if _, err := exec.LookPath("systemctl"); err == nil {
		if _, err := os.Stat("/run/systemd/system"); err == nil {
			return initSystemd
		}
	}
	if _, err := os.Stat("/etc/sv"); err == nil {
		return initRunit
	}
	if _, err := os.Stat("/var/service"); err == nil {
		return initRunit
	}
	return ""
}

func resolveGreeterAccount(m *model) {
	// Void's greetd package provisions _greeter — use it when present.
	if userExists("_greeter") || m.distroID == "void" {
		m.greeterUser = "_greeter"
		m.greeterHome = "/var/lib/_greeter"
		return
	}
	m.greeterUser = "greeter"
	m.greeterHome = "/var/lib/greeter"
}

func userExists(username string) bool {
	return exec.Command("id", username).Run() == nil
}

func substituteGreeterPaths(content string, m *model) string {
	if m.greeterUser == "" || m.greeterHome == "" {
		resolveGreeterAccount(m)
	}
	out := strings.ReplaceAll(content, "/var/lib/greeter", m.greeterHome)
	out = strings.ReplaceAll(out, "user = \"greeter\"", fmt.Sprintf("user = \"%s\"", m.greeterUser))
	out = strings.ReplaceAll(out, "subject.user == \"greeter\"", fmt.Sprintf("subject.user == \"%s\"", m.greeterUser))
	return out
}

func xbpsInstallArgs(packages ...string) []string {
	args := []string{"install", "-Sy"}
	return append(args, packages...)
}

func enableGreeterService(m *model) error {
	switch m.initSystem {
	case initRunit:
		return enableRunitGreetd(m)
	case initSystemd:
		return enableSystemdGreetd()
	case "":
		return fmt.Errorf("no supported init system detected (need systemd or runit)")
	default:
		return fmt.Errorf("unsupported init system: %s", m.initSystem)
	}
}

func disableGreeterService(m *model) error {
	switch m.initSystem {
	case initRunit:
		_ = os.Remove("/var/service/greetd")
		return nil
	case initSystemd:
		_ = exec.Command("systemctl", "disable", "greetd.service").Run()
		return nil
	default:
		return nil
	}
}

func enableSystemdGreetd() error {
	symlinkPath := "/etc/systemd/system/display-manager.service"
	if _, err := os.Lstat(symlinkPath); err == nil {
		os.Remove(symlinkPath)
	}
	cmd := exec.Command("systemctl", "enable", "greetd.service")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemd enable failed")
	}
	return nil
}

func enableRunitGreetd(m *model) error {
	if _, err := os.Stat("/etc/sv/greetd"); err != nil {
		return fmt.Errorf("runit service missing at /etc/sv/greetd — reinstall greetd")
	}

	serviceLink := "/var/service/greetd"
	if _, err := os.Lstat(serviceLink); err == nil {
		if m.debugMode {
			fmt.Fprintf(os.Stderr, "[DEBUG] greetd runit service already enabled\n")
		}
	} else {
		if err := os.Symlink("/etc/sv/greetd", serviceLink); err != nil {
			return fmt.Errorf("failed to symlink /var/service/greetd: %w", err)
		}
	}

	// Void greetd README: disable agetty on the greeter VT (we use vt=1).
	agettyLinks := []string{
		"/var/service/agetty-tty1",
		"/etc/service/agetty-tty1",
	}
	for _, link := range agettyLinks {
		if _, err := os.Lstat(link); err == nil {
			if err := os.Remove(link); err != nil && m.debugMode {
				fmt.Fprintf(os.Stderr, "[DEBUG] could not remove %s: %v\n", link, err)
			} else if m.debugMode {
				fmt.Fprintf(os.Stderr, "[DEBUG] disabled conflicting service %s\n", link)
			}
		}
	}

	return nil
}

func greeterGroups() string {
	// Void greetd template only adds video; render may not exist.
	if _, err := exec.Command("getent", "group", "render").Output(); err == nil {
		return "video,render,input"
	}
	return "video,input"
}
