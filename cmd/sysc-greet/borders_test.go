package main

import "testing"

func TestBorderStyleMigration(t *testing.T) {
	cases := map[string]string{
		"ascii3":  "ascii2",
		"ascii4":  "ascii2",
		"ascii2":  "ascii2",
		"ascii1":  "ascii1",
		"classic": "classic",
		"":        "",
	}
	for in, want := range cases {
		if got := migrateBorderStyle(in); got != want {
			t.Errorf("migrateBorderStyle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMigratedBorderStyleRendersWithoutPanic(t *testing.T) {
	for _, size := range [][2]int{{40, 12}, {60, 20}, {80, 24}} {
		for _, removed := range []string{"ascii3", "ascii4"} {
			t.Run(removed, func(t *testing.T) {
				m := layoutModel(size[0], size[1])
				m.selectedBorderStyle = migrateBorderStyle(removed)
				if m.selectedBorderStyle != "ascii2" {
					t.Fatalf("style = %q, want ascii2", m.selectedBorderStyle)
				}
				view := m.renderMainView(m.width, m.height)
				if view == "" {
					t.Fatal("empty render")
				}
			})
		}
	}
}
