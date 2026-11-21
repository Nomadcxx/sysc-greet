package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSHA256For(t *testing.T) {
	good := strings.Repeat("ab", 32)
	sums := []byte(good + "  cagebreak_2.4.0_debian13_amd64.deb\n" +
		strings.Repeat("cd", 32) + " *mangowm_0.18.2_debian13_amd64.deb\n" +
		"zz  broken.deb\n")
	for _, tc := range []struct {
		name, artifact, want string
		ok                   bool
	}{
		{"listed", "cagebreak_2.4.0_debian13_amd64.deb", good, true},
		{"binary-mode marker", "mangowm_0.18.2_debian13_amd64.deb", strings.Repeat("cd", 32), true},
		{"absent", "sysc-greet_9.9.9_amd64.deb", "", false},
		{"malformed digest", "broken.deb", "", false},
		{"prefix is not a match", "cagebreak", "", false},
	} {
		got, err := sha256For(sums, tc.artifact)
		if (err == nil) != tc.ok || got != tc.want {
			t.Errorf("%s: sha256For = %q, %v", tc.name, got, err)
		}
	}
}

func releaseServer(t *testing.T, files map[string]string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := files[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/"
}

func digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestDownloadVerifiedChecksTheReleaseDigest(t *testing.T) {
	const pkg = "cagebreak_2.4.0_debian13_amd64.deb"
	base := releaseServer(t, map[string]string{
		"SHA256SUMS":   digest("package bytes") + "  " + pkg + "\n" + digest("other") + "  tampered.deb\n",
		pkg:            "package bytes",
		"tampered.deb": "not what was released",
	})
	dir := t.TempDir()

	path, err := downloadVerified(base, dir, pkg)
	if err != nil {
		t.Fatalf("matching package refused: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "package bytes" || filepath.Dir(path) != dir {
		t.Fatalf("downloaded %q to %s", data, path)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("package mode = %v, want 0600", info.Mode().Perm())
	}

	if _, err := downloadVerified(base, dir, "tampered.deb"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("tampered package: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tampered.deb")); !os.IsNotExist(err) {
		t.Fatal("a package that failed verification was left behind")
	}
	if _, err := downloadVerified(base, dir, "unlisted.deb"); err == nil {
		t.Fatal("a package missing from SHA256SUMS was accepted")
	}
	if _, err := downloadVerified(base, dir, "../escape.deb"); err == nil {
		t.Fatal("an artifact name with a path was accepted")
	}
}

func TestDownloadToNeverFollowsAPlantedPath(t *testing.T) {
	base := releaseServer(t, map[string]string{"a.deb": "bytes"})
	dir := t.TempDir()
	victim := filepath.Join(t.TempDir(), "victim")
	if err := os.WriteFile(victim, []byte("root-owned"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "a.deb")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	if _, err := downloadTo(base+"a.deb", link); err == nil {
		t.Fatal("download followed a planted symlink")
	}
	if data, _ := os.ReadFile(victim); string(data) != "root-owned" {
		t.Fatalf("symlink target was overwritten: %q", data)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("planted"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := downloadTo(base+"a.deb", link); err == nil {
		t.Fatal("download reused a planted file")
	}
}

func TestOpenInstallerLogRefusesPlantedFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installer.log")
	f, err := openInstallerLog(path)
	if err != nil {
		t.Fatalf("fresh log: %v", err)
	}
	f.Close()
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("log mode = %v, want 0600", info.Mode().Perm())
	}
	if f, err = openInstallerLog(path); err != nil {
		t.Fatalf("reopening own log: %v", err)
	}
	f.Close()

	victim := filepath.Join(dir, "victim")
	if err := os.WriteFile(victim, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.log")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	if _, err := openInstallerLog(link); err == nil {
		t.Fatal("log followed a symlink")
	}
	hard := filepath.Join(dir, "hard.log")
	if err := os.Link(victim, hard); err != nil {
		t.Fatal(err)
	}
	if _, err := openInstallerLog(hard); err == nil {
		t.Fatal("log opened a hard-linked file")
	}
}

func TestOpenInstallerLogTightensAndDoesNotHang(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.log")
	if err := os.WriteFile(old, []byte("earlier run\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := openInstallerLog(old)
	if err != nil {
		t.Fatalf("own older log: %v", err)
	}
	f.Close()
	if info, _ := os.Stat(old); info.Mode().Perm() != 0o600 {
		t.Fatalf("older log left at %v, want 0600", info.Mode().Perm())
	}

	fifo := filepath.Join(dir, "fifo.log")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		f, err := openInstallerLog(fifo)
		if f != nil {
			f.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a FIFO was accepted as the log")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("opening a planted FIFO hung")
	}
}

func TestAnUnsetWorkDirIsNeverTheCurrentDirectory(t *testing.T) {
	if _, err := scratchPath("", "gslapper-build"); err == nil {
		t.Fatal("scratchPath accepted an unset work directory")
	}
	if _, err := downloadVerified("http://127.0.0.1:1/", "", "a.deb"); err == nil {
		t.Fatal("downloadVerified accepted an unset work directory")
	}
}

func TestNewWorkDirIsPrivate(t *testing.T) {
	dir, err := newWorkDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		t.Fatalf("work dir %s: %v, %v", dir, info.Mode(), err)
	}
}
