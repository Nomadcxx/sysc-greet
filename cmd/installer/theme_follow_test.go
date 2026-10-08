package main

import (
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"testing"
)

func TestShellThemeDirectory(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	if u.Uid == "0" {
		t.Skip("needs a non-root installation account")
	}
	root := filepath.Join(t.TempDir(), "sysc-greet")
	if err := setupShellThemeDir(root, u.Username); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filepath.Join(root, "shell-theme"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 || int(info.Sys().(*syscall.Stat_t).Uid) != os.Getuid() {
		t.Fatal("wrong theme ownership or readability")
	}
	if err := setupShellThemeDir(root, u.Username); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, bad); err != nil {
		t.Fatal(err)
	}
	if err := setupShellThemeDir(bad, u.Username); err == nil {
		t.Fatal("followed directory symlink")
	}
}
