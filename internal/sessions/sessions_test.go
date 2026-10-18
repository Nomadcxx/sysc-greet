package sessions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSessionsFromPathsDiscoversSyscCompositorWaylandEntry(t *testing.T) {
	waylandDir := filepath.Join(t.TempDir(), "wayland-sessions")
	if err := os.MkdirAll(waylandDir, 0o755); err != nil {
		t.Fatal(err)
	}

	entryPath := filepath.Join(waylandDir, "sysc-compositor.desktop")
	execLine := "/usr/local/bin/sysc-compositor -c /home/nomadx/.config/sysc-compositor/config.kdl"
	entry := "[Desktop Entry]\nName=sysc-compositor\nExec=" + execLine + "\nType=Application\n"
	if err := os.WriteFile(entryPath, []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := loadSessionsFromPaths([]string{waylandDir})
	if err != nil {
		t.Fatalf("loadSessionsFromPaths(): %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("session count = %d, want 1", len(got))
	}

	want := Session{
		Name: "sysc-compositor",
		Exec: execLine,
		Type: "Wayland",
		Path: entryPath,
	}
	if got[0] != want {
		t.Errorf("session = %#v, want %#v", got[0], want)
	}
}
