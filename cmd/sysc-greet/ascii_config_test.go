package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMangoConfigIncludesSubmittedLogo(t *testing.T) {
	config, err := loadASCIIConfig(filepath.Join("..", "..", "ascii_configs", "mango.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !config.ANSIStyle {
		t.Fatal("Mango config does not enable ANSI art")
	}
	if got := len(config.ASCIIVariants); got != 13 {
		t.Fatalf("Mango config has %d variants, want 13", got)
	}
	if !strings.Contains(config.ASCIIVariants[12], `\033[32m=`) {
		t.Fatal("variant 13 does not contain the submitted Mango logo")
	}
}

func TestLoadASCIIConfigPreservesEqualsInArt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "quoted export",
			content: "ansi_style=true\n" +
				"ascii_1=\"\"\"\n" +
				"top\nname=inside art\nleft = right\nbottom\n" +
				"\"\"\"\nroasts=one│two\n",
			want: []string{"top\nname=inside art\nleft = right\nbottom"},
		},
		{
			name: "unquoted variants",
			content: "ascii_1=\ntop\nleft = right\nname = art\nbottom\n" +
				"ascii_2=\nsecond\ncolors=#fff\nroasts=one│two\n",
			want: []string{"top\nleft = right\nname = art\nbottom", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ascii.conf")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}

			config, err := loadASCIIConfig(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(config.ASCIIVariants) != len(tt.want) {
				t.Fatalf("got %d variants %q, want %q", len(config.ASCIIVariants), config.ASCIIVariants, tt.want)
			}
			for i, want := range tt.want {
				if config.ASCIIVariants[i] != want {
					t.Errorf("variant %d = %q, want %q", i+1, config.ASCIIVariants[i], want)
				}
			}
			if config.Roasts != "one│two" {
				t.Errorf("roasts = %q, want %q", config.Roasts, "one│two")
			}
		})
	}
}
