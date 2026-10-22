package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
				"top\nleft = right\nbottom\n" +
				"\"\"\"\nroasts=one│two\n",
			want: []string{"top\nleft = right\nbottom"},
		},
		{
			name: "unquoted variants",
			content: "ascii_1=\ntop\nleft = right\nbottom\n" +
				"ascii_2=\nsecond\ncolors=#fff\nroasts=one│two\n",
			want: []string{"top\nleft = right\nbottom", "second"},
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
