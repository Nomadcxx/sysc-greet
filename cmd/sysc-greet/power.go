package main

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// powerCommand builds the power request. Tests replace it so they never
// reboot the host.
var powerCommand = func(args ...string) *exec.Cmd {
	return exec.Command(args[0], args[1:]...)
}

// powerFailedMsg reports a reboot or shutdown request that login1 refused.
// It is not an error value, so Update does not take it for a failed login.
type powerFailedMsg struct {
	action string
	reason string
}

// requestPower asks login1 to reboot or power off. systemd and elogind both
// ship loginctl, and 85-greeter.rules lets the greeter account do this without
// a password. --no-ask-password keeps polkit from prompting on the greeter's
// terminal. An accepted request returns no message: the greeter stays alive
// because systemd 260+ kills the session scope when the greeter exits, which
// races with the request.
func requestPower(action, verb string) tea.Cmd {
	return func() tea.Msg {
		output, err := powerCommand("loginctl", "--no-ask-password", verb).CombinedOutput()
		if err == nil {
			return nil
		}
		reason := strings.TrimSpace(string(output))
		if reason == "" {
			reason = err.Error()
		}
		if line, _, found := strings.Cut(reason, "\n"); found {
			reason = line
		}
		return powerFailedMsg{action: action, reason: reason}
	}
}
