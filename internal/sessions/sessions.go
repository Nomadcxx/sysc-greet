package sessions

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Session struct {
	Name string
	Exec string
	Type string // "X11" or "Wayland"
	Path string
}

func (s Session) FilterValue() string {
	return s.Name
}

func (s Session) Title() string {
	return s.Name
}

func (s Session) Description() string {
	return s.Type + " session • " + s.Exec
}

func (s Session) String() string {
	return s.Name
}

func LoadSessions() ([]Session, error) {
	var sessions []Session

	// Default paths
	paths := []string{
		"/usr/share/xsessions",
		"/usr/share/wayland-sessions",
		"/run/current-system/sw/share/xsessions",
		"/run/current-system/sw/share/wayland-sessions",
	}

	for _, basePath := range paths {
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !strings.HasSuffix(path, ".desktop") {
				return nil
			}

			session, err := parseDesktopFile(path)
			if err != nil {
				return nil
			}
			sessions = append(sessions, session)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

func parseDesktopFile(path string) (Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var name, exec, sessionType string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Name=") {
			name = strings.TrimPrefix(line, "Name=")
		} else if strings.HasPrefix(line, "Exec=") {
			exec = strings.TrimPrefix(line, "Exec=")
		}
	}

	if strings.Contains(path, "xsessions") {
		sessionType = "X11"
	} else {
		sessionType = "Wayland"
	}

	return Session{
		Name: name,
		Exec: exec,
		Type: sessionType,
		Path: path,
	}, nil
}
