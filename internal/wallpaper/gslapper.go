// internal/wallpaper/gslapper.go
package wallpaper

import (
	"fmt"
	"net"
	"os"
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

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(string(buf[:n])), nil
}

// ChangeWallpaper changes the current wallpaper with fade transition
func ChangeWallpaper(path string) error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	// Set fade transition
	SendCommand("set-transition fade")
	SendCommand("set-transition-duration 0.5")

	// Change wallpaper
	resp, err := SendCommand("change " + path)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(resp, "OK") {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
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

	if resp != "OK" {
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

	if resp != "OK" {
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
