package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadASCIIConfigKeepsEqualsInsideTripleQuotedArt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.conf")
	content := `name=ansi-demo
ansi_style=true
ascii_1="""
\033[0;37mA=B\033[0m
second row
"""
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	config, err := loadASCIIConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !config.ANSIStyle {
		t.Fatal("ANSI style was not loaded")
	}
	if len(config.ASCIIVariants) != 1 {
		t.Fatalf("ASCII variants = %d, want 1", len(config.ASCIIVariants))
	}
	if !strings.Contains(config.ASCIIVariants[0], `A=B`) || !strings.Contains(config.ASCIIVariants[0], "second row") {
		t.Fatalf("ASCII variant was truncated: %q", config.ASCIIVariants[0])
	}
}
