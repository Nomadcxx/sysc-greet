package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBorderPreferencesRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	want := UserPreferences{
		BorderStyle:       "nfo02",
		BorderEffect:      "off",
		BorderPaletteMode: "vga",
		FormMode:          "frosted",
	}
	if err := SavePreferences(want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("LoadPreferences() returned nil after saving preferences")
	}
	if got.BorderStyle != want.BorderStyle || got.BorderEffect != want.BorderEffect ||
		got.BorderPaletteMode != want.BorderPaletteMode || got.FormMode != want.FormMode {
		t.Fatalf("border preferences = %+v, want style=%q effect=%q palette=%q form=%q",
			got, want.BorderStyle, want.BorderEffect, want.BorderPaletteMode, want.FormMode)
	}
}

func TestSavePreferencesUsesPrivateAtomicFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	want := UserPreferences{BorderStyle: "nfo02", Username: "first"}
	if err := SavePreferences(want); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, cacheDir, preferencesFile)
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	want.Username = "second"
	if err := SavePreferences(want); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(before, after) {
		t.Fatal("SavePreferences rewrote the existing file in place")
	}

	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Errorf("cache directory mode = %04o, want 0700", got)
	}
	if got := after.Mode().Perm(); got != 0o600 {
		t.Errorf("preferences file mode = %04o, want 0600", got)
	}
	got, err := LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Username != "second" {
		t.Fatalf("loaded preferences = %+v, want second username", got)
	}
}

func TestLoadPreferencesHardensLegacyModes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, cacheDir)
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(UserPreferences{Username: "remembered"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cachePath, preferencesFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Username != "remembered" {
		t.Fatalf("loaded preferences = %+v, want remembered username", got)
	}
	dirInfo, err := os.Stat(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("cache directory mode after load = %04o, want 0700", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Errorf("preferences file mode after load = %04o, want 0600", got)
	}
}

func TestPreferencesRejectSymlinkPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cachePath := filepath.Join(home, cacheDir)
	if err := os.MkdirAll(cachePath, 0o700); err != nil {
		t.Fatal(err)
	}

	targetDir := filepath.Join(home, "outside")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	targetFile := filepath.Join(targetDir, preferencesFile)
	if err := os.WriteFile(targetFile, []byte(`{"username":"untouched"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetFile, filepath.Join(cachePath, preferencesFile)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPreferences(); err == nil {
		t.Fatal("LoadPreferences accepted a symlink preference file")
	}

	if err := os.Remove(filepath.Join(cachePath, preferencesFile)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(cachePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetDir, cachePath); err != nil {
		t.Fatal(err)
	}
	if err := SavePreferences(UserPreferences{Username: "changed"}); err == nil {
		t.Fatal("SavePreferences accepted a symlink cache directory")
	}
	if _, err := LoadPreferences(); err == nil {
		t.Fatal("LoadPreferences accepted a symlink cache directory")
	}
	got, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"username":"untouched"}` {
		t.Fatalf("symlink target changed: %s", got)
	}
}
