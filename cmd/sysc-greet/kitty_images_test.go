package main

import (
	"bytes"
	"fmt"
	"image"
	"regexp"
	"strings"
	"testing"
	"time"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func TestKittyPNGCommandsChunkAndReplacePreviousImage(t *testing.T) {
	sequence, err := kittyPNGCommands(bytes.Repeat([]byte{'x'}, 4097), 42, 80, 24, 41)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(sequence, "\x1b_G"); got != 3 {
		t.Fatalf("graphics commands = %d, want two chunks and one deletion", got)
	}
	for _, fragment := range []string{
		"\x1b[H",
		"a=T,f=100,i=42,p=1,c=80,r=24,z=-2147483648,C=1,q=2,m=1;",
		"\x1b_Gm=0,q=2;",
		"\x1b_Ga=d,d=I,i=41,q=2\x1b\\",
	} {
		if !strings.Contains(sequence, fragment) {
			t.Fatalf("graphics sequence missing %q", fragment)
		}
	}
}

func TestKittyImagePresenterAlternatesAndClearsImageIDs(t *testing.T) {
	presenter := kittyImagePresenter{ids: [2]uint32{10, 11}}
	frame := syscgoanimations.MeshFrame{Pixels: image.NewNRGBA(image.Rect(0, 0, 2, 4))}
	first, err := presenter.render(frame, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(first, "a=T,f=100,i=10,") || strings.Contains(first, "d=I,i=") {
		t.Fatalf("first image command = %q", first)
	}
	second, err := presenter.render(frame, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(second, "a=T,f=100,i=11,") || !strings.Contains(second, "d=I,i=10,") {
		t.Fatalf("replacement image command = %q", second)
	}
	// Both IDs go: at exit a queued frame may not have replaced the one on screen.
	if got := presenter.clear(); !strings.Contains(got, "d=I,i=10,") || !strings.Contains(got, "d=I,i=11,") {
		t.Fatalf("clear command = %q", got)
	}
	if got := presenter.clear(); got != "" {
		t.Fatalf("second clear = %q, want empty", got)
	}
}

func TestKittyGraphicsDetectionUsesKittyWindowID(t *testing.T) {
	t.Setenv("KITTY_WINDOW_ID", "")
	if kittyGraphicsAvailable() {
		t.Fatal("Kitty detected without a window ID")
	}
	t.Setenv("KITTY_WINDOW_ID", "23")
	if !kittyGraphicsAvailable() {
		t.Fatal("Kitty window ID not detected")
	}
}

func TestKittyMeshBackgroundLeavesTerminalDefaultBackgroundVisible(t *testing.T) {
	m := initialModel(Config{TestMode: true}, false)
	m.width, m.height = 40, 12
	m.mode = ModeLogin
	m.selectedBackground = "ak-47"
	m.kittyGraphics = true

	if view := m.View(); view.BackgroundColor != nil {
		t.Fatalf("Kitty mesh view background = %v, want terminal default", view.BackgroundColor)
	}
}

func TestKittyCreditBackgroundWrapsLongAttribution(t *testing.T) {
	credit := "The bust of Napoleon Bonaparte / Virtual Museums of Małopolska / CC BY 4.0"
	m := model{width: 40, height: 12, kittyCredit: credit}
	background := m.kittyCreditBackground()
	if got := strings.Join(strings.Fields(background), " "); !strings.Contains(got, credit) {
		t.Fatalf("wrapped credit %q is missing from background", credit)
	}
	if got := strings.Count(background, "\n") + 1; got != 12 {
		t.Fatalf("credit background height = %d, want 12", got)
	}
}

func TestNonKittyMeshBackgroundKeepsANSIBackground(t *testing.T) {
	m := initialModel(Config{TestMode: true}, false)
	m.width, m.height = 40, 12
	m.mode = ModeLogin
	m.selectedBackground = "ak-47"

	if view := m.View(); view.BackgroundColor != BgBase {
		t.Fatalf("ANSI mesh view background = %v, want %v", view.BackgroundColor, BgBase)
	}
}

func TestKittyMeshBackgroundClearsWhenLeavingLogin(t *testing.T) {
	previous := model{
		kittyGraphics:      true,
		mode:               ModeLogin,
		selectedBackground: "ak-47",
		kittyCredit:        "AK-47 credit",
		kittyImages:        kittyImagePresenter{ids: [2]uint32{10, 11}, currentID: 11},
	}
	next := previous
	next.mode = ModeMenu

	cmd := syncKittyMeshFrame(previous, &next, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("leaving the mesh background did not schedule image cleanup")
	}
	if next.kittyImages.currentID != 0 || next.kittyCredit != "" {
		t.Fatalf("mesh image remained active after transition: id=%d credit=%q", next.kittyImages.currentID, next.kittyCredit)
	}
}

func TestKittyMeshBackgroundRefreshesOnBackgroundTick(t *testing.T) {
	m := model{kittyGraphics: true, mode: ModeLogin, selectedBackground: "ak-47"}
	m.width, m.height = 40, 12
	m.ak47Effect = syscgoanimations.NewAK47Effect(m.width, m.height, "Dracula")
	cmd := syncKittyMeshFrame(m, &m, bgTickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("background tick did not schedule a Kitty image refresh")
	}
	if m.kittyImages.currentID == 0 || m.kittyCredit == "" {
		t.Fatalf("background tick did not retain its image and credit: id=%d credit=%q", m.kittyImages.currentID, m.kittyCredit)
	}
}

// visibleKittyImages applies Kitty transmit and delete commands in order and
// returns the image IDs still placed on screen.
func visibleKittyImages(writes ...string) map[string]bool {
	visible := map[string]bool{}
	command := regexp.MustCompile(`\x1b_Ga=([Td]),(?:f=100,|d=I,)i=(\d+)`)
	for _, write := range writes {
		for _, match := range command.FindAllStringSubmatch(write, -1) {
			if match[1] == "T" {
				visible[match[2]] = true
			} else {
				delete(visible, match[2])
			}
		}
	}
	return visible
}

func TestKittyMeshWritesStayOrderedWhenDeliveredOutOfOrder(t *testing.T) {
	m := model{kittyGraphics: true, mode: ModeLogin, selectedBackground: "ak-47"}
	m.width, m.height = 40, 12
	m.ak47Effect = syscgoanimations.NewAK47Effect(m.width, m.height, "Dracula")

	var raws []tea.Msg
	for _, step := range []func(*model) tea.Msg{
		func(*model) tea.Msg { return bgTickMsg(time.Now()) },
		func(*model) tea.Msg { return bgTickMsg(time.Now()) },
		func(next *model) tea.Msg { next.mode = ModeMenu; return tea.KeyPressMsg{Code: tea.KeyF1} },
	} {
		previous := m
		msg := step(&m)
		cmd := syncKittyMeshFrame(previous, &m, msg)
		if cmd == nil {
			t.Fatalf("no Kitty write for %T", msg)
		}
		raws = append(raws, cmd())
	}

	// Bubble Tea runs each command on its own goroutine, so the event loop can
	// receive these writes in any order. It prints each one with fmt.Sprint.
	for _, order := range [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		writes := make([]string, 0, len(order))
		for _, index := range order {
			writes = append(writes, fmt.Sprint(raws[index].(tea.RawMsg).Msg))
		}
		if visible := visibleKittyImages(writes...); len(visible) != 0 {
			t.Errorf("delivery order %v left images %v on screen", order, visible)
		}
	}
}
