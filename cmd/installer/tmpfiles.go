package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// The installer runs as root, and /tmp is writable by every local account.
// Anything it downloads or builds goes in a fresh directory only root can
// enter, so a path another user planted can neither redirect a write nor
// swap a package or source tree before root installs it.

// newWorkDir makes the installer's private scratch directory (mode 0700,
// unpredictable name, created exclusively).
func newWorkDir() (string, error) {
	return os.MkdirTemp("", "sysc-greet-installer-")
}

// openInstallerLog opens the shared log for append without following a
// symlink, and refuses a file another account owns or has hard-linked.
func openInstallerLog(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !info.Mode().IsRegular() || !ok || int(st.Uid) != os.Geteuid() || st.Nlink != 1 {
		f.Close()
		return nil, fmt.Errorf("%s is not a regular file owned by this user", path)
	}
	return f, nil
}

// maxSumsBytes bounds the SHA256SUMS download; the real file is under 2 KB.
const maxSumsBytes = 1 << 20

// sha256For returns the digest SHA256SUMS lists for name, in the
// "<hex>  <name>" form sha256sum writes (a "*" before the name marks binary
// mode and is accepted).
func sha256For(sums []byte, name string) (string, error) {
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != name {
			continue
		}
		digest := strings.ToLower(fields[0])
		if _, err := hex.DecodeString(digest); err != nil || len(digest) != sha256.Size*2 {
			return "", fmt.Errorf("SHA256SUMS has a malformed digest for %s", name)
		}
		return digest, nil
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("SHA256SUMS does not list %s", name)
}

// downloadVerified fetches artifact from the release at base into dir and
// checks it against the release's SHA256SUMS before returning its path. A
// file that does not match is removed.
func downloadVerified(base, dir, artifact string) (string, error) {
	if artifact == "" || filepath.Base(artifact) != artifact {
		return "", fmt.Errorf("invalid artifact name %q", artifact)
	}
	sums, err := fetch(base+"SHA256SUMS", maxSumsBytes)
	if err != nil {
		return "", fmt.Errorf("download SHA256SUMS: %w", err)
	}
	want, err := sha256For(sums, artifact)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, artifact)
	got, err := downloadTo(base+artifact, path)
	if err != nil {
		os.Remove(path)
		return "", err
	}
	if got != want {
		os.Remove(path)
		return "", fmt.Errorf("%s: SHA-256 %s does not match the release's %s", artifact, got, want)
	}
	return path, nil
}

func fetch(url string, limit int64) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("response is too large")
	}
	return data, nil
}

// downloadTo writes url to a new file at dest, never an existing one or a
// symlink, and returns the SHA-256 of what it wrote.
func downloadTo(url, dest string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, h), resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
