package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// packageCommands holds the commands one package manager uses to install and
// look up packages. Every install command runs without asking for a
// confirmation.
type packageCommands struct {
	install   []string
	installed []string
	// installedStatus is text the installed query has to print; empty means
	// its exit status alone answers.
	installedStatus string
	available       []string
	// refresh updates the package lists before an install, for managers
	// that do not do it themselves.
	refresh []string
}

var packageManagerCommands = map[string]packageCommands{
	"pacman": {
		install:   []string{"pacman", "-S", "--noconfirm", "--needed"},
		installed: []string{"pacman", "-Q"},
		available: []string{"pacman", "-Si"},
	},
	"apt": {
		install: []string{"apt-get", "install", "-y"},
		// dpkg-query also knows removed packages that left config files, so
		// only the status text tells an installed package apart.
		installed:       []string{"dpkg-query", "-W", "-f=${Status}"},
		installedStatus: "install ok installed",
		available:       []string{"apt-cache", "show"},
		refresh:         []string{"apt-get", "update"},
	},
	"dnf": {
		install:   []string{"dnf", "install", "-y"},
		installed: []string{"rpm", "-q"},
		available: []string{"dnf", "info"},
	},
	"yum": {
		install:   []string{"yum", "install", "-y"},
		installed: []string{"rpm", "-q"},
		available: []string{"yum", "info"},
	},
	"zypper": {
		install:   []string{"zypper", "install", "-y"},
		installed: []string{"rpm", "-q"},
		available: []string{"zypper", "info"},
	},
	// Void Linux. xbps-install is the standalone install front end, so it
	// takes no "install" subcommand. Only the xbps-query show mode reports
	// presence through its exit status: it resolves the exact name and
	// returns ENOENT when the package is absent. The search mode matches the
	// pkgver loosely and exits 0 when it finds nothing.
	"xbps": {
		install:   []string{"xbps-install", "-Sy"},
		installed: []string{"xbps-query", "-S"},
		available: []string{"xbps-query", "-R", "-S"},
	},
}

// commandWith builds base followed by args, or nil when base is empty.
func commandWith(base []string, args ...string) *exec.Cmd {
	if len(base) == 0 {
		return nil
	}
	full := append(append([]string{}, base[1:]...), args...)
	return exec.Command(base[0], full...)
}

// installCommand returns the command that installs the packages, or nil for
// an unknown package manager.
func installCommand(manager string, packages ...string) *exec.Cmd {
	return commandWith(packageManagerCommands[manager].install, packages...)
}

// installedCommand returns the query that tells whether a package is
// installed, or nil for an unknown package manager.
func installedCommand(manager, pkg string) *exec.Cmd {
	return commandWith(packageManagerCommands[manager].installed, pkg)
}

// availableCommand returns the query that tells whether the repos carry a
// package, or nil for an unknown package manager.
func availableCommand(manager, pkg string) *exec.Cmd {
	return commandWith(packageManagerCommands[manager].available, pkg)
}

// refreshCommand returns the command that updates the package lists, or nil
// when the package manager needs none.
func refreshCommand(manager string) *exec.Cmd {
	return commandWith(packageManagerCommands[manager].refresh)
}

// checkPackageInstalled checks if a package is already installed
func checkPackageInstalled(pkg, packageManager string) bool {
	cmd := installedCommand(packageManager, pkg)
	if cmd == nil {
		return false
	}
	status := packageManagerCommands[packageManager].installedStatus
	if status == "" {
		return cmd.Run() == nil
	}
	output, err := cmd.Output()
	return err == nil && strings.Contains(string(output), status)
}

// checkPackageExists checks if a package exists in repos (can be installed)
func checkPackageExists(pkg, packageManager string) bool {
	cmd := availableCommand(packageManager, pkg)
	if cmd == nil {
		return false
	}
	return cmd.Run() == nil
}

// installPackages installs the packages with the detected package manager,
// updating the package lists once per installer run where the manager needs
// it. The command and its output go to the installer log under task.
func installPackages(m *model, task string, packages ...string) error {
	cmd := installCommand(m.packageManager, packages...)
	if cmd == nil {
		return fmt.Errorf("unsupported package manager '%s' - install %s manually",
			m.packageManager, strings.Join(packages, ", "))
	}
	if !m.packageListsRefreshed {
		if refresh := refreshCommand(m.packageManager); refresh != nil {
			runCommand("Refresh package lists", refresh, m) // Best effort; the install reports failures
		}
		m.packageListsRefreshed = true
	}
	return runCommand(task, cmd, m)
}
