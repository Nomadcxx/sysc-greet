package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

// Only the account chosen during installation can publish the machine theme.
// The greeter gets read access without access to that account's home directory.
func setupShellThemeDir(root, account string) error {
	if account == "" || account == "root" {
		return nil
	}
	u, err := user.Lookup(account)
	if err != nil {
		return fmt.Errorf("theme account: %w", err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil || uid <= 0 {
		return fmt.Errorf("theme account must be non-root")
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("theme account group: %w", err)
	}
	for _, path := range []string{root, filepath.Join(root, "shell-theme")} {
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			if err = os.Mkdir(path, 0755); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !info.IsDir() {
			return fmt.Errorf("theme directory %s must be a real directory", path)
		}
	}
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() {
		return fmt.Errorf("theme parent %s must belong to the installer account", root)
	}
	if err = os.Chmod(root, 0755); err != nil {
		return err
	}
	path := filepath.Join(root, "shell-theme")
	if err = os.Chown(path, uid, gid); err != nil {
		return err
	}
	return os.Chmod(path, 0755)
}
