package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
)

func preferenceTestModel(session string) model {
	m := initialModel(Config{TestMode: true, RememberUsername: true}, false)
	m.config.TestMode = false
	m.selectedSession = &sessions.Session{Name: session}
	return m
}

func TestSavePreferencesLogsWriteFailures(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home-file")
	if err := os.WriteFile(home, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	previous := debugLog
	var output strings.Builder
	debugLog = log.New(&output, "", 0)
	t.Cleanup(func() { debugLog = previous })

	model{}.savePreferences()
	if !strings.Contains(output.String(), "failed to save sysc-greet preferences") {
		t.Fatalf("save failure was not logged: %q", output.String())
	}
}

func TestSavePreferencesPreservesOnlyLastSuccessfulUsernameForSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := cache.SavePreferences(cache.UserPreferences{Session: "niri", Username: "alice"}); err != nil {
		t.Fatal(err)
	}

	m := preferenceTestModel("niri")
	m.usernameInput.SetValue("unverified")
	m.currentTheme = "ayu"
	m.savePreferences()

	prefs, err := cache.LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if prefs.Username != "alice" {
		t.Errorf("saved username = %q, want last successful username %q", prefs.Username, "alice")
	}
	if prefs.Session != "niri" {
		t.Errorf("saved session = %q, want %q", prefs.Session, "niri")
	}
	if prefs.Theme != "ayu" {
		t.Errorf("saved theme = %q, want newly selected theme %q", prefs.Theme, "ayu")
	}
}

func TestSavePreferencesDoesNotPairUsernameWithDifferentSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := cache.SavePreferences(cache.UserPreferences{Session: "niri", Username: "alice"}); err != nil {
		t.Fatal(err)
	}

	m := preferenceTestModel("sway")
	m.usernameInput.SetValue("unverified")
	m.savePreferences()

	prefs, err := cache.LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if prefs.Username != "" {
		t.Errorf("saved username = %q, want empty after selecting a different session", prefs.Username)
	}
	if prefs.Session != "sway" {
		t.Errorf("saved session = %q, want %q", prefs.Session, "sway")
	}
}

func TestSaveSuccessfulLoginRemembersAuthenticatedUsername(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := preferenceTestModel("niri")
	m.saveSuccessfulLogin("alice", "niri")

	prefs, err := cache.LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if prefs.Username != "alice" || prefs.Session != "niri" {
		t.Fatalf("saved login = (%q, %q), want (%q, %q)", prefs.Username, prefs.Session, "alice", "niri")
	}
}

func TestDebugLogUsesPrivateCachePermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, ".cache", "sysc-greet")
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cachePath, "debug.log")
	if err := os.WriteFile(path, []byte("old log"), 0o644); err != nil {
		t.Fatal(err)
	}

	previous := debugLog
	debugLog = nil
	t.Cleanup(func() { debugLog = previous })
	initDebugLog()
	if debugLog == nil {
		t.Fatal("initDebugLog() did not open the cache log")
	}
	if file, ok := debugLog.Writer().(*os.File); ok {
		t.Cleanup(func() { file.Close() })
	}

	dirInfo, err := os.Stat(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("debug directory mode = %04o, want 0700", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Errorf("debug file mode = %04o, want 0600", got)
	}
}

func TestOpenPrivateLogRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("untouched"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "debug.log")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if file, err := openPrivateLog(path); err == nil {
		file.Close()
		t.Fatal("openPrivateLog accepted a symlink")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "untouched" {
		t.Fatalf("symlink target changed: %q", got)
	}
}

func TestValidatePrivateLogOwnerRejectsOtherUID(t *testing.T) {
	otherUID := uint32(os.Geteuid()) + 1
	if err := validatePrivateLogOwner("/tmp/sysc-greet-debug.log", otherUID); err == nil {
		t.Fatal("private log owner validation accepted another UID")
	}
	if err := validatePrivateLogOwner("/tmp/sysc-greet-debug.log", uint32(os.Geteuid())); err != nil {
		t.Fatalf("private log owner validation rejected current UID: %v", err)
	}
}
