package wallpaper

import "testing"

func TestIsOKResponse(t *testing.T) {
	cases := []struct {
		resp string
		want bool
	}{
		{"OK", true},
		// change on an image with a transition configured
		{"OK: transition started", true},
		{"OK: transitions disabled", true},
		{"OK done", true},
		{"ERROR: cannot update path (use --auto-stop for video changes)", false},
		{"ERROR: no pipeline", false},
		{"OKAY", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isOKResponse(c.resp); got != c.want {
			t.Errorf("isOKResponse(%q) = %v, want %v", c.resp, got, c.want)
		}
	}
}
