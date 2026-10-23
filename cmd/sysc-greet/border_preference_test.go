package main

import "testing"

func TestBorderPreferenceFallbackPreservesValidSavedStyle(t *testing.T) {
	registry := newBorderRegistry("")
	for _, tc := range []struct {
		name  string
		saved string
		want  string
	}{
		{name: "missing preference uses minimal", want: "minimal"},
		{name: "invalid preference uses minimal", saved: "removed-style", want: "minimal"},
		{name: "valid saved style is preserved", saved: "classic", want: "classic"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := preferredBorderStyle(tc.saved, registry); got != tc.want {
				t.Fatalf("preferredBorderStyle(%q) = %q, want %q", tc.saved, got, tc.want)
			}
		})
	}
}
