package main

import (
	"fmt"
	"os"
	"time"
)

func parseSecondarySaverStart(value string) (time.Time, error) {
	startedAt, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid screensaver start time: %w", err)
	}
	return startedAt, nil
}

func secondarySaverCommandArgs(configFile, executable, theme, startedAt, readyFile string) []string {
	args := []string{"--start-as=fullscreen", "--class=" + secondarySaverWindowClass, "--single-instance=no"}
	if configFile != "" {
		args = append(args, "--config="+configFile)
	}
	args = append(args, "-o", "clear_all_shortcuts=yes", "-o", "clear_all_mouse_actions=yes",
		"-o", "allow_remote_control=no", "-o", "close_on_child_death=yes", executable,
		"--secondary-screensaver", "--follow-shell=false", "--theme", theme,
		"--screensaver-started-at", startedAt, "--secondary-ready-file", readyFile)
	return args
}

func writeSecondarySaverReady(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create secondary screensaver ready file: %w", err)
	}
	if _, err := file.WriteString("ready\n"); err != nil {
		file.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write secondary screensaver ready file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(path)
		return fmt.Errorf("sync secondary screensaver ready file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("close secondary screensaver ready file: %w", err)
	}
	return nil
}
