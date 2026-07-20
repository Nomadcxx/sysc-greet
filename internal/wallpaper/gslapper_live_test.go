package wallpaper

// Live tests against a real gslapper instance. They reproduce the issue #83
// boot race: the IPC socket file appears before the server accepts commands,
// so a single ChangeWallpaper attempt right after spawn always failed.
// Skipped when gslapper or a Wayland session is unavailable.

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

const testPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func requireLiveEnv(t *testing.T) (gslapperPath string) {
	t.Helper()
	gslapperPath, err := exec.LookPath("gslapper")
	if err != nil {
		t.Skip("gslapper not in PATH")
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/run/user/" + strconv.Itoa(os.Getuid())
	}
	display := os.Getenv("WAYLAND_DISPLAY")
	if display == "" {
		display = "wayland-1"
	}
	if _, err := os.Stat(filepath.Join(runtimeDir, display)); err != nil {
		t.Skipf("no Wayland display socket at %s/%s", runtimeDir, display)
	}
	if IsGSlapperRunning() {
		t.Skipf("socket %s already in use; not touching a live greeter instance", GSlapperSocket)
	}
	// Sweep stale instances from earlier runs. gslapper builds with the
	// signal-handler bug (gSlapper issue #24) can hang on SIGTERM, so a
	// plain pkill is not enough to guarantee a clean slate.
	exec.Command("pkill", "-9", "-f", "gslapper.*"+GSlapperSocket).Run()
	time.Sleep(100 * time.Millisecond)
	return gslapperPath
}

func writeTestPNG(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(testPNGBase64)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func spawnGslapper(t *testing.T, gslapperPath, wallpaper string) {
	t.Helper()
	cmd := exec.Command(gslapperPath, "-f", "-I", GSlapperSocket, "*", wallpaper)
	// Isolate state saving from the user's real gslapper state
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		cmd.Env = append(cmd.Env, "WAYLAND_DISPLAY=wayland-1")
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start gslapper: %v", err)
	}
	go cmd.Wait() // reap the forking parent
	t.Cleanup(func() {
		exec.Command("pkill", "-f", "gslapper.*"+GSlapperSocket).Run()
		time.Sleep(300 * time.Millisecond)
		// SIGKILL stragglers: pre-#24-fix gslapper can hang on SIGTERM
		exec.Command("pkill", "-9", "-f", "gslapper.*"+GSlapperSocket).Run()
		os.Remove(GSlapperSocket)
	})
}

// TestChangeWallpaperWithRetryDuringStartup reproduces the boot race: call
// IPC immediately after spawning gslapper, while the server may not yet be
// accepting commands.
func TestChangeWallpaperWithRetryDuringStartup(t *testing.T) {
	gslapperPath := requireLiveEnv(t)
	dir := t.TempDir()
	first := writeTestPNG(t, dir, "first.png")
	second := writeTestPNG(t, dir, "second.png")

	spawnGslapper(t, gslapperPath, first)

	// Wait only for the socket file, exactly like setThemeWallpaper does,
	// then change immediately - the single-attempt version failed here.
	deadline := time.Now().Add(5 * time.Second)
	for !IsGSlapperRunning() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !IsGSlapperRunning() {
		t.Fatal("gslapper socket never appeared")
	}

	if err := ChangeWallpaperWithRetry(second); err != nil {
		t.Fatalf("ChangeWallpaperWithRetry failed: %v", err)
	}
}

// TestStopInstance verifies the clean IPC quit path removes the instance and
// its socket without resorting to signals.
func TestStopInstance(t *testing.T) {
	gslapperPath := requireLiveEnv(t)
	dir := t.TempDir()
	png := writeTestPNG(t, dir, "wall.png")

	spawnGslapper(t, gslapperPath, png)

	deadline := time.Now().Add(5 * time.Second)
	for !IsGSlapperRunning() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !IsGSlapperRunning() {
		t.Fatal("gslapper socket never appeared")
	}
	// Let the server finish coming up so quit is a fair test of the IPC path
	time.Sleep(500 * time.Millisecond)

	StopInstance()

	if IsGSlapperRunning() {
		t.Fatal("socket still present after StopInstance")
	}
	// The socket disappears early in gslapper's graceful shutdown; give the
	// process a moment to finish tearing down its pipeline and exit
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := exec.Command("pgrep", "-f", "gslapper.*"+GSlapperSocket).Run(); err != nil {
			return // process gone
		}
		time.Sleep(100 * time.Millisecond)
	}
	matched, _ := exec.Command("pgrep", "-af", "gslapper.*"+GSlapperSocket).Output()
	t.Fatalf("gslapper process still running 3s after StopInstance:\n%s", matched)
}
