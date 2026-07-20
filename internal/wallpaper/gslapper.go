// internal/wallpaper/gslapper.go
package wallpaper

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GSlapperSocket is the path to the greeter's gSlapper IPC socket
const GSlapperSocket = "/tmp/sysc-greet-wallpaper.sock"

// IsGSlapperRunning checks if gSlapper IPC socket exists
func IsGSlapperRunning() bool {
	_, err := os.Stat(GSlapperSocket)
	return err == nil
}

// SendCommand sends a command to gSlapper via Unix socket and returns the response
func SendCommand(cmd string) (string, error) {
	conn, err := net.DialTimeout("unix", GSlapperSocket, 2*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect to gSlapper socket: %w", err)
	}
	defer conn.Close()

	// Set read/write deadline
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send command
	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response (increased buffer for status queries with file paths)
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(string(buf[:n])), nil
}

// isOKResponse checks if a gSlapper response indicates success
// gSlapper may respond with "OK" or "OK <details>"
func isOKResponse(resp string) bool {
	return resp == "OK" || strings.HasPrefix(resp, "OK ")
}

// ChangeWallpaper changes the current wallpaper with fade transition
func ChangeWallpaper(path string) error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	// Validate file exists before sending to gSlapper
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("wallpaper file not found: %s", path)
	}

	// Set fade transition (best effort - don't fail if these don't work)
	if _, err := SendCommand("set-transition fade"); err != nil {
		// Log but continue - transition settings are optional
	}
	if _, err := SendCommand("set-transition-duration 0.5"); err != nil {
		// Log but continue - transition settings are optional
	}

	// Change wallpaper
	resp, err := SendCommand("change " + path)
	if err != nil {
		return err
	}

	if !isOKResponse(resp) {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// CHANGED 2026-07-20 - Retry IPC with backoff - Problem: at boot the socket file appears before the
// gSlapper IPC server accepts commands, so the first ChangeWallpaper always failed and every boot fell
// through to the pkill+restart path (issue #83). Failed dials on a dead socket return immediately, so
// retrying is cheap; on a live server the first attempt succeeds and adds no latency.
// ChangeWallpaperWithRetry retries ChangeWallpaper with exponential backoff.
func ChangeWallpaperWithRetry(path string) error {
	var err error
	delay := 100 * time.Millisecond
	for i := 0; i < 5; i++ {
		if err = ChangeWallpaper(path); err == nil {
			return nil
		}
		time.Sleep(delay)
		delay *= 2
	}
	return err
}

// Quit asks the greeter's gSlapper instance to exit cleanly via IPC
func Quit() error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	resp, err := SendCommand("quit")
	if err != nil {
		return err
	}

	if !isOKResponse(resp) {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// CHANGED 2026-07-20 - Stop via IPC quit, pkill scoped to our socket as last resort - Problem:
// "pkill -f gslapper" SIGTERMs every gSlapper on the machine, including user-session instances,
// and the blunt kill path is what produced orphaned greeter processes and shutdown crashes
// (issue #83). The quit command shuts down exactly the instance holding our socket.
// StopInstance stops the greeter's gSlapper instance, preferring a clean IPC quit.
func StopInstance() {
	if err := Quit(); err == nil {
		// Wait briefly for the process to exit and remove its socket
		for i := 0; i < 10; i++ {
			if !IsGSlapperRunning() {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Last resort: match only processes whose command line references our socket
	exec.Command("pkill", "-f", "gslapper.*"+GSlapperSocket).Run()
	time.Sleep(100 * time.Millisecond) // brief pause for process cleanup
}

// PauseVideo pauses video playback
func PauseVideo() error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	resp, err := SendCommand("pause")
	if err != nil {
		return err
	}

	if !isOKResponse(resp) {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// ResumeVideo resumes video playback
func ResumeVideo() error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	resp, err := SendCommand("resume")
	if err != nil {
		return err
	}

	if !isOKResponse(resp) {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// QueryStatus returns current gSlapper status
func QueryStatus() (string, error) {
	if !IsGSlapperRunning() {
		return "", fmt.Errorf("gSlapper is not running")
	}

	return SendCommand("query")
}
