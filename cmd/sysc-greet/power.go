package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// powerCommand builds the power request. Tests replace it so they never
// reboot the host.
var powerCommand = func(args ...string) *exec.Cmd {
	return exec.Command(args[0], args[1:]...)
}

func parsePowerCommand(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var encodedArgv []*string
	if err := json.Unmarshal([]byte(raw), &encodedArgv); err != nil {
		return nil, fmt.Errorf("expected a JSON argv array: %w", err)
	}
	if len(encodedArgv) == 0 || encodedArgv[0] == nil || strings.TrimSpace(*encodedArgv[0]) == "" {
		return nil, fmt.Errorf("argv must start with a non-empty executable")
	}
	argv := make([]string, len(encodedArgv))
	for i, arg := range encodedArgv {
		if arg == nil {
			return nil, fmt.Errorf("argv[%d] must be a string", i)
		}
		if strings.IndexByte(*arg, 0) >= 0 {
			return nil, fmt.Errorf("argv cannot contain NUL bytes")
		}
		argv[i] = *arg
	}
	return argv, nil
}

// powerFailedMsg reports a reboot or shutdown request that login1 refused.
// It is not an error value, so Update does not take it for a failed login.
type powerFailedMsg struct {
	action string
	reason string
}

// requestPower runs an override when configured; otherwise it asks login1 to
// reboot or power off. systemd uses systemctl; elogind's loginctl adds power
// verbs absent from systemd's loginctl. The 85-greeter.rules policy allows
// these actions without a password, and --no-ask-password keeps polkit from
// prompting in the greeter terminal. An accepted request returns no message:
// the greeter stays alive because systemd 260+ kills the session scope when
// the greeter exits, which races with the request.
func requestPower(action, verb string, override []string) tea.Cmd {
	return func() tea.Msg {
		args := override
		if len(args) == 0 {
			tool := "loginctl"
			if _, err := os.Stat("/run/systemd/system"); err == nil {
				tool = "systemctl"
			}
			args = []string{tool, "--no-ask-password", verb}
		}
		output, err := powerCommand(args...).CombinedOutput()
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
