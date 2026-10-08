package main

import "testing"

func TestParseOSRelease(t *testing.T) {
	tests := []struct {
		name, data, id, version string
	}{
		{"ubuntu", "ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"24.04\"\nVERSION_CODENAME=noble\nUBUNTU_CODENAME=noble\n", "ubuntu", "24.04"},
		{"debian", "ID=debian\nVERSION_ID=\"13\"\nVERSION_CODENAME=trixie\n", "debian", "13"},
		{"fedora", "ID=fedora\nVERSION_ID=43\n", "fedora", "43"},
		{"linux mint 22", "ID=linuxmint\nID_LIKE=\"ubuntu debian\"\nVERSION_ID=\"22\"\nVERSION_CODENAME=wilma\nUBUNTU_CODENAME=noble\n", "ubuntu", "24.04"},
		{"pop os 24.04", "ID=pop\nID_LIKE=\"ubuntu debian\"\nVERSION_ID=\"24.04\"\nVERSION_CODENAME=noble\nUBUNTU_CODENAME=noble\n", "ubuntu", "24.04"},
		{"lmde 7", "ID=linuxmint\nID_LIKE=debian\nVERSION_ID=\"7\"\nVERSION_CODENAME=gigi\nDEBIAN_CODENAME=trixie\n", "debian", "13"},
		{"debian derivative by version codename", "ID=mx\nID_LIKE=debian\nVERSION_ID=\"25\"\nVERSION_CODENAME=trixie\n", "debian", "13"},
		{"unknown ubuntu base keeps own id", "ID=linuxmint\nID_LIKE=\"ubuntu debian\"\nVERSION_ID=\"21\"\nUBUNTU_CODENAME=jammy\n", "linuxmint", "21"},
		{"unknown debian base keeps own id", "ID=kali\nID_LIKE=debian\nVERSION_ID=\"2026.3\"\nVERSION_CODENAME=kali-rolling\n", "kali", "2026.3"},
		{"empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, version := parseOSRelease(tt.data)
			if id != tt.id || version != tt.version {
				t.Fatalf("got (%q, %q), want (%q, %q)", id, version, tt.id, tt.version)
			}
		})
	}
}
