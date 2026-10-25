package animations

import (
	"fmt"
	"testing"
)

// aquariumPalette is the palette shape the greeter hands the scene.
func aquariumPalette(width, height int) AquariumConfig {
	return AquariumConfig{
		Width:         width,
		Height:        height,
		FishColors:    []string{"#ff6666", "#66ff66", "#6666ff"},
		WaterColors:   []string{"#4a9eff", "#c2b280"},
		SeaweedColors: []string{"#336633", "#558855"},
		BubbleColor:   "#ffffff",
		DiverColor:    "#ffcc00",
		BoatColor:     "#cc8844",
		MermaidColor:  "#ff66cc",
		AnchorColor:   "#cccccc",
	}
}

// TestAquariumRendersEveryCanvasHeight runs the scene the way the greeter does,
// a frame at a time, on canvases a console can really hand it: a small display
// or a large font leaves few rows, and the canvas is whatever the console says.
// A short canvas has no band between the ocean surface and the sea floor, so
// every spawn helper must clamp its row range instead of asking for an empty
// one, which panics inside rand.Intn. The large fish spawns at frame 700 and
// the medium fish before that, so the run lasts long enough to reach both.
func TestAquariumRendersEveryCanvasHeight(t *testing.T) {
	for _, height := range []int{54, 40, 30, 24, 23, 20, 18, 12, 8} {
		for _, width := range []int{213, 80, 40} {
			t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
				a := NewAquariumEffect(aquariumPalette(width, height))
				for frame := 1; frame <= 760; frame++ {
					a.Update()
					if a.Render() == "" {
						t.Fatalf("frame %d rendered nothing on a %dx%d canvas", frame, width, height)
					}
				}
			})
		}
	}
}
