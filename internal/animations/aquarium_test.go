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
// spawn helpers must clamp or skip empty ranges instead of passing them to
// rand.Intn. The large fish spawns at frame 700 and the medium fish before
// that, so the run lasts long enough to reach both.
func TestAquariumRendersEveryCanvasHeight(t *testing.T) {
	heights := make([]int, 0, 27)
	for height := 24; height >= 1; height-- {
		heights = append(heights, height)
	}
	heights = append(heights, 30, 40, 54)

	for _, height := range heights {
		for _, width := range []int{1, 40, 80, 213} {
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

func TestAquariumResizeFromNormalToTiny(t *testing.T) {
	a := NewAquariumEffect(aquariumPalette(80, 24))
	a.Resize(40, 3)

	for frame := 1; frame <= 760; frame++ {
		a.Update()
		if a.Render() == "" {
			t.Fatalf("frame %d rendered nothing after resizing to a 40x3 canvas", frame)
		}
	}
}
