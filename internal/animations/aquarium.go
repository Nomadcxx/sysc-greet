package animations

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// AquariumEffect implements an animated aquarium scene
type AquariumEffect struct {
	width  int
	height int

	// Entities
	fish    []Fish
	seaweed []Seaweed
	bubbles []Bubble
	diver   *Diver
	boat    *Boat
	mermaid *Mermaid
	anchor  *Anchor

	// Spawn timers (in frames, 20fps)
	lastMediumFishSpawn int
	lastLargeFishSpawn  int
	lastMermaidSpawn    int

	// Theme colors
	waterColors   []string
	fishColors    []string
	seaweedColors []string
	bubbleColor   string
	diverColor    string
	boatColor     string
	mermaidColor  string

	frameCount int
	rng        *rand.Rand
}

// Fish represents a swimming fish
type Fish struct {
	x         float64
	y         float64
	speed     float64
	size      int // 0=tiny, 1=small, 2=medium, 3=large
	direction int // 1=right, -1=left
	pattern   []string // Multi-line pattern
	color     string
	swimPhase float64
}

// Seaweed represents swaying underwater plants
type Seaweed struct {
	x            int
	height       int
	swayPhase    float64
	swaySpeed    float64
	swayAmount   float64
	colors       []string
	variant      int // 0=straight, 1=wavy
}

// Bubble represents a rising bubble
type Bubble struct {
	x         float64
	y         float64
	speed     float64
	wobble    float64
	wobbleAmt float64
	size      int
}

// Diver represents a scuba diver
type Diver struct {
	x         float64
	y         float64
	speed     float64
	direction int
	pattern   []string
	swimPhase float64
}

// Boat represents a boat on the surface
type Boat struct {
	x         float64
	y         float64
	speed     float64
	direction int
	pattern   []string
	boatType  int // 0=small boat (can go either way), 1=large ship (left only)
}

// Mermaid represents a swimming mermaid
type Mermaid struct {
	x         float64
	y         float64
	speed     float64
	direction int
	pattern   []string
	swimPhase float64
}

// Anchor represents a static anchor on the ocean floor
type Anchor struct {
	x       int
	y       int
	pattern []string
}

// AquariumConfig holds configuration for the aquarium effect
type AquariumConfig struct {
	Width         int
	Height        int
	FishColors    []string
	WaterColors   []string
	SeaweedColors []string
	BubbleColor   string
	DiverColor    string
	BoatColor     string
	MermaidColor  string
	AnchorColor   string
}

// NewAquariumEffect creates a new aquarium effect
func NewAquariumEffect(config AquariumConfig) *AquariumEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	a := &AquariumEffect{
		width:         config.Width,
		height:        config.Height,
		fishColors:    config.FishColors,
		waterColors:   config.WaterColors,
		seaweedColors: config.SeaweedColors,
		bubbleColor:   config.BubbleColor,
		diverColor:    config.DiverColor,
		boatColor:     config.BoatColor,
		mermaidColor:  config.MermaidColor,
		frameCount:    0,
		rng:           rng,
	}

	a.init()
	return a
}

// init initializes the aquarium entities
func (a *AquariumEffect) init() {
	// Create seaweed (bottom decoration)
	seaweedCount := a.width / 8
	for i := 0; i < seaweedCount; i++ {
		x := a.rng.Intn(a.width)
		height := 3 + a.rng.Intn(a.height/3)
		variant := a.rng.Intn(2) // 0=straight, 1=wavy

		a.seaweed = append(a.seaweed, Seaweed{
			x:          x,
			height:     height,
			swayPhase:  a.rng.Float64() * math.Pi * 2,
			swaySpeed:  0.05 + a.rng.Float64()*0.05,
			swayAmount: 1.0 + a.rng.Float64()*0.5,
			colors:     a.seaweedColors,
			variant:    variant,
		})
	}

	// Create diver - position so full diver is visible above bottom
	diverPattern := a.getDiverPattern()
	diverHeight := len(diverPattern)
	a.diver = &Diver{
		x:         -20,
		y:         float64(a.height - diverHeight - 2), // Place above bottom with margin
		speed:     0.3,
		direction: 1,
		pattern:   diverPattern,
		swimPhase: 0,
	}

	// Create initial fish (mostly small/tiny)
	fishCount := 8 + a.rng.Intn(15)
	for i := 0; i < fishCount; i++ {
		a.spawnFish()
	}

	// Create initial bubbles (increased count)
	bubbleCount := 15 + a.rng.Intn(10)
	for i := 0; i < bubbleCount; i++ {
		a.spawnBubble()
	}

	// Create boat on surface (above the waves)
	boatType := a.rng.Intn(2) // 0 = small boat, 1 = large ship
	boatPattern := a.getBoatPatternByType(boatType)
	boatHeight := len(boatPattern)
	oceanY := int(float64(a.height) * 0.15)

	// Large ship (type 1) only travels left, small boat (type 0) can go either way
	var boatDirection int
	if boatType == 1 {
		boatDirection = -1 // Large ship always left
	} else {
		// Small boat randomly left or right
		if a.rng.Float64() < 0.5 {
			boatDirection = -1
		} else {
			boatDirection = 1
		}
	}

	a.boat = &Boat{
		x:         float64(a.rng.Intn(a.width)),
		y:         float64(oceanY - boatHeight), // Above ocean surface
		speed:     0.4,
		direction: boatDirection,
		pattern:   boatPattern,
		boatType:  boatType,
	}

	// Create anchor on ocean floor
	anchorPattern := a.getAnchorPattern()
	anchorHeight := len(anchorPattern)
	a.anchor = &Anchor{
		x:       a.width/2 - 5,
		y:       a.height - anchorHeight - 1,
		pattern: anchorPattern,
	}

	// Initialize spawn timers
	a.lastMediumFishSpawn = -1000 // Allow immediate spawn
	a.lastLargeFishSpawn = -1000  // Allow immediate spawn
	a.lastMermaidSpawn = -1000    // Allow immediate spawn
}

// spawnFish creates a new fish at a random or edge position (tiny/small only)
func (a *AquariumEffect) spawnFish() {
	// Randomly choose direction
	direction := -1
	if a.rng.Float64() < 0.5 {
		direction = 1
	}

	// Starting position based on direction
	var x float64
	if direction == 1 {
		x = -10 // Start off left side
	} else {
		x = float64(a.width + 10) // Start off right side
	}

	// Only spawn tiny/small fish here (size 0-1)
	var size int
	if a.rng.Float64() < 0.7 {
		size = 0 // 70% tiny
	} else {
		size = 1 // 30% small
	}

	speed := 0.5 + a.rng.Float64()*1.5
	if size == 0 {
		speed *= 1.8 // Tiny fish fastest
	} else if size == 1 {
		speed *= 1.5 // Small fish faster
	}

	color := a.fishColors[a.rng.Intn(len(a.fishColors))]

	oceanY := int(float64(a.height) * 0.15)
	minY := oceanY + 2
	maxY := a.height - 10
	if maxY <= minY {
		maxY = a.height - 2
	}

	fish := Fish{
		x:         x,
		y:         float64(minY + a.rng.Intn(maxY-minY)),
		speed:     speed,
		size:      size,
		direction: direction,
		color:     color,
		swimPhase: a.rng.Float64() * math.Pi * 2,
		pattern:   a.getFishPattern(size, direction),
	}

	a.fish = append(a.fish, fish)
}

// spawnMediumFish creates a medium-sized fish
func (a *AquariumEffect) spawnMediumFish() {
	direction := -1
	if a.rng.Float64() < 0.5 {
		direction = 1
	}

	var x float64
	if direction == 1 {
		x = -15
	} else {
		x = float64(a.width + 15)
	}

	speed := 0.4 + a.rng.Float64()*0.8
	color := a.fishColors[a.rng.Intn(len(a.fishColors))]

	oceanY := int(float64(a.height) * 0.15)
	minY := oceanY + 2
	maxY := a.height - 10

	fish := Fish{
		x:         x,
		y:         float64(minY + a.rng.Intn(maxY-minY)),
		speed:     speed,
		size:      2, // Medium
		direction: direction,
		color:     color,
		swimPhase: a.rng.Float64() * math.Pi * 2,
		pattern:   a.getFishPattern(2, direction),
	}

	a.fish = append(a.fish, fish)
}

// spawnLargeFish creates a large fish
func (a *AquariumEffect) spawnLargeFish() {
	direction := -1
	if a.rng.Float64() < 0.5 {
		direction = 1
	}

	var x float64
	if direction == 1 {
		x = -20
	} else {
		x = float64(a.width + 20)
	}

	speed := 0.3 + a.rng.Float64()*0.5
	color := a.fishColors[a.rng.Intn(len(a.fishColors))]

	oceanY := int(float64(a.height) * 0.15)
	minY := oceanY + 5
	maxY := a.height - 15

	fish := Fish{
		x:         x,
		y:         float64(minY + a.rng.Intn(maxY-minY)),
		speed:     speed,
		size:      3, // Large
		direction: direction,
		color:     color,
		swimPhase: a.rng.Float64() * math.Pi * 2,
		pattern:   a.getFishPattern(3, direction),
	}

	a.fish = append(a.fish, fish)
}

// getFishPattern returns ASCII art for a fish based on size and direction
func (a *AquariumEffect) getFishPattern(size int, direction int) []string {
	var pattern []string

	switch size {
	case 0: // Tiny fish
		if direction == -1 { // Left-facing
			pattern = []string{"<°)))><"}
		} else { // Right-facing
			pattern = []string{"><(((('>"}
		}

	case 1: // Small fish
		if direction == -1 { // Left-facing
			pattern = []string{
				"  _///_",
				" /o    \\/",
				" > ))_./\\",
				"    <",
			}
		} else { // Right-facing
			pattern = []string{
				"     |\\    o",
				"    |  \\    o",
				"|\\ /    .\\ o",
				"| |       (",
				"|/ \\     /",
				"    |  /",
				"     |/",
			}
		}

	case 2: // Medium fish
		if direction == -1 { // Left-facing
			mediumPatterns := [][]string{
				{
					"          ,,////,",
					"        _////////_",
					"      .' -,  / / /`'-._     _.-'|",
					"     / _  \\\\/ / / / /  ',.='_.'/",
					"    / (o)  ||/_/_/_/_/_/_.-'_.'",
					"  .'       ||\\ \\ \\ \\ \\ \\ '-._'.",
					" '.--.    //\\ \\ \\ \\ \\  .'\"-._ '.",
					"   `'-.\\ \\   \\ \\ \\__.-'\\)    '-.|",
					"       \\\\)`\"\"\"\"\"` ",
					"        `",
				},
				{
					"                ,      /",
					"             . ~ ~ . ,/{",
					"           .'@ ))ejm'~.~",
					"           = - ~``   ",
				},
			}
			pattern = mediumPatterns[a.rng.Intn(len(mediumPatterns))]
		} else { // Right-facing
			// Simple right-facing medium fish
			pattern = []string{
				"\\o    o",
				" \\     \\",
				"  )=====>",
				" /     /",
				"/o    o",
			}
		}

	case 3: // Large fish
		if direction == -1 { // Left-facing
			largePatterns := [][]string{
				{
					"                 __,",
					"               .-'_-'`",
					"             .' {`",
					"         .-'````'-.    .-'``'.",
					"       .'(0)       '._/ _.-.  `\\",
					"      }     '. ))    _<`    )`  |",
					"       `-.,\\'.\\_, -\\` \\`---; .' /",
					"            )  )       '-.  '--:",
					"           ( ' (          ) '.  \\",
					"            '.  )      .'(   /   )",
					"              )/      (   '.    /",
					"                       '._( ) .'",
					"                           ( (",
					"                            `-.",
				},
				{
					"    o   o",
					"                  /^^^^^7",
					"    '  '     ,oO))))))))Oo,",
					"           ,'))))))))))))))), /{",
					"      '  ,'o  ))))))))))))))))={",
					"         >    ))))))))))))))))={",
					"         `,   ))))))\\\\\\)))))))={ ",
					"           ',))))))))\\/)))))' \\{",
					"             '*O))))))))O*'",
				},
			}
			pattern = largePatterns[a.rng.Intn(len(largePatterns))]
		} else { // Right-facing (no large right-facing in fish.txt, use reversed positions visually)
			// For now, just use a simple right-facing large fish
			pattern = []string{
				"    __,",
				"   / - \\",
				"  (  O  )======>",
				"   \\ - /",
				"    `-'",
			}
		}
	}

	return pattern
}

// getDiverPattern returns ASCII art for a scuba diver
func (a *AquariumEffect) getDiverPattern() []string {
	return []string{
		"              _______ ______",
		"              |     / |    /",
		"   O          |    /  |   /",
		"              |   /   |  /",
		"o  O 0         \\  \\   \\  \\",
		"o               \\  \\   \\  \\",
		"   o            /  /   /  /",
		"    o     /\\_  /\\\\\\   /  /",
		"     O  /    /    /     /",
		"..       /    /    /\\=    /",
		" ))))))) = /====/    \\",
		"(((((((( /    /\\=  _ }",
		"|-----_|_+( /   \\}",
		"\\_<\\_//|  \\  \\ }",
		"  =Q=  |==)\\  \\",
		"\\----/     ) )",
		"         / /",
		"        /=/",
		"      \\|/",
		"      o}",
	}
}

// getBoatPattern returns ASCII art for a random boat
func (a *AquariumEffect) getBoatPattern() []string {
	return a.getBoatPatternByType(a.rng.Intn(2))
}

// getBoatPatternByType returns ASCII art for a specific boat type
func (a *AquariumEffect) getBoatPatternByType(boatType int) []string {
	boats := [][]string{
		{
			"     _",
			"    /|\\",
			"   /_|_\\",
			" ____|____",
			" \\_o_o_o_/",
		},
		{
			"                __/___            ",
			"          _____/______|           ",
			"  _______/_____\\_______\\_____     ",
			"  \\              < < <       |",
		},
	}
	return boats[boatType]
}

// getAnchorPattern returns ASCII art for an anchor
func (a *AquariumEffect) getAnchorPattern() []string {
	return []string{
		"        _-_",
		"       |(_)|",
		"        |||",
		"        |||",
		"        |||",
		"        |||",
		"        |||",
		"  ^     |^|     ^",
		"< ^ >   <+>   < ^ >",
		" | |    |||    | |",
		"  \\ \\__/ | \\__/ /",
		"    \\,__.|.__,/",
		"        (_)",
	}
}

// getMermaidPattern returns ASCII art for a mermaid
func (a *AquariumEffect) getMermaidPattern() []string {
	return []string{
		"                           .-\"\"-.",
		"                          (___/\\ \\",
		"        ,                 (|^ ^ ) )",
		"       /(                _)_\\=_/  (",
		" ,..__/ `\\          ____(_/_ ` \\   )",
		"  `\\    _/        _/---._/(_)_  `\\ (",
		"    '--\\ `-.__..-'    /.    (_), |  )",
		"        `._        ___\\_____.'_| |__/",
		"           `~----\"`   `-.........' ",
	}
}

// spawnBubble creates a new bubble
func (a *AquariumEffect) spawnBubble() {
	oceanY := int(float64(a.height) * 0.15)
	minY := oceanY + 2
	maxY := a.height - 1

	a.bubbles = append(a.bubbles, Bubble{
		x:         float64(a.rng.Intn(a.width)),
		y:         float64(minY + a.rng.Intn(maxY-minY)),
		speed:     0.2 + a.rng.Float64()*0.3,
		wobble:    a.rng.Float64() * math.Pi * 2,
		wobbleAmt: 0.3 + a.rng.Float64()*0.3,
		size:      1,
	})
}

// spawnMermaid creates a mermaid
func (a *AquariumEffect) spawnMermaid() {
	direction := -1
	if a.rng.Float64() < 0.5 {
		direction = 1
	}

	var x float64
	if direction == 1 {
		x = -50
	} else {
		x = float64(a.width + 50)
	}

	// Mermaids swim in bottom region
	mermaidPattern := a.getMermaidPattern()
	mermaidHeight := len(mermaidPattern)
	minY := a.height - mermaidHeight - 15
	maxY := a.height - mermaidHeight - 5

	a.mermaid = &Mermaid{
		x:         x,
		y:         float64(minY + a.rng.Intn(maxY-minY+1)),
		speed:     0.2 + a.rng.Float64()*0.3,
		direction: direction,
		pattern:   mermaidPattern,
		swimPhase: 0,
	}
}

// Update advances the aquarium animation
func (a *AquariumEffect) Update() {
	a.frameCount++

	// Update seaweed sway
	for i := range a.seaweed {
		a.seaweed[i].swayPhase += a.seaweed[i].swaySpeed
	}

	// Update fish
	for i := len(a.fish) - 1; i >= 0; i-- {
		fish := &a.fish[i]

		// Move fish
		fish.x += fish.speed * float64(fish.direction)
		fish.swimPhase += 0.2

		// Add slight vertical bobbing
		fish.y += math.Sin(fish.swimPhase) * 0.1

		// Remove fish that swim off screen
		if (fish.direction == 1 && fish.x > float64(a.width+30)) ||
			(fish.direction == -1 && fish.x < -30) {
			a.fish = append(a.fish[:i], a.fish[i+1:]...)
		}
	}

	// Update bubbles
	oceanY := int(float64(a.height) * 0.15)
	for i := len(a.bubbles) - 1; i >= 0; i-- {
		bubble := &a.bubbles[i]

		// Rise upward
		bubble.y -= bubble.speed

		// Wobble side to side
		bubble.wobble += 0.1
		bubble.x += math.Sin(bubble.wobble) * bubble.wobbleAmt

		// Remove bubbles that reach ocean surface
		if bubble.y < float64(oceanY) {
			a.bubbles = append(a.bubbles[:i], a.bubbles[i+1:]...)
		}
	}

	// Update diver
	if a.diver != nil {
		a.diver.x += a.diver.speed * float64(a.diver.direction)
		a.diver.swimPhase += 0.1

		// Add slight vertical bobbing
		a.diver.y += math.Sin(a.diver.swimPhase) * 0.05

		// Reset when off screen
		if a.diver.direction == 1 && a.diver.x > float64(a.width+30) {
			a.diver.x = -30
			a.diver.direction = 1
		} else if a.diver.direction == -1 && a.diver.x < -30 {
			a.diver.x = float64(a.width + 30)
			a.diver.direction = -1
		}
	}

	// Update boat (always moving)
	if a.boat != nil {
		a.boat.x += a.boat.speed * float64(a.boat.direction)

		// Wrap around when off screen
		if a.boat.direction == 1 && a.boat.x > float64(a.width+15) {
			a.boat.x = -15
		} else if a.boat.direction == -1 && a.boat.x < -15 {
			a.boat.x = float64(a.width + 15)
		}
	}

	// Update mermaid
	if a.mermaid != nil {
		a.mermaid.x += a.mermaid.speed * float64(a.mermaid.direction)
		a.mermaid.swimPhase += 0.1

		// Add slight vertical bobbing
		a.mermaid.y += math.Sin(a.mermaid.swimPhase) * 0.08

		// Remove when off screen and bring back diver
		if (a.mermaid.direction == 1 && a.mermaid.x > float64(a.width+50)) ||
			(a.mermaid.direction == -1 && a.mermaid.x < -50) {
			a.mermaid = nil

			// Bring diver back when mermaid leaves
			if a.diver == nil {
				diverPattern := a.getDiverPattern()
				diverHeight := len(diverPattern)
				a.diver = &Diver{
					x:         -20,
					y:         float64(a.height - diverHeight - 2),
					speed:     0.3,
					direction: 1,
					pattern:   diverPattern,
					swimPhase: 0,
				}
			}
		}
	}

	// Count fish by size
	mediumCount := 0
	largeCount := 0
	for _, fish := range a.fish {
		if fish.size == 2 {
			mediumCount++
		} else if fish.size == 3 {
			largeCount++
		}
	}

	// Spawn new tiny/small fish regularly
	if a.frameCount%25 == 0 && len(a.fish) < 30 {
		a.spawnFish()
	}

	// Spawn medium fish (max 1, every 15-20 seconds)
	// 15-20 seconds at 20fps = 300-400 frames
	if mediumCount == 0 && a.frameCount-a.lastMediumFishSpawn >= 300+a.rng.Intn(100) {
		a.spawnMediumFish()
		a.lastMediumFishSpawn = a.frameCount
	}

	// Spawn large fish (max 1, every 35 seconds)
	// 35 seconds at 20fps = 700 frames
	if largeCount == 0 && a.frameCount-a.lastLargeFishSpawn >= 700 {
		a.spawnLargeFish()
		a.lastLargeFishSpawn = a.frameCount
	}

	// Spawn mermaid (every 2-3 minutes if not present)
	// 2-3 minutes at 20fps = 2400-3600 frames
	// Mermaid and diver are mutually exclusive
	if a.mermaid == nil && a.frameCount-a.lastMermaidSpawn >= 2400+a.rng.Intn(1200) {
		a.spawnMermaid()
		a.lastMermaidSpawn = a.frameCount
		// Remove diver when mermaid appears
		a.diver = nil
	}

	// Spawn bubbles more frequently (increased count)
	if a.frameCount%15 == 0 && len(a.bubbles) < 40 {
		a.spawnBubble()
	}
}

// Render converts the aquarium to colored text output
func (a *AquariumEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, a.height)
	colors := make([][]string, a.height)
	for i := range canvas {
		canvas[i] = make([]rune, a.width)
		colors[i] = make([]string, a.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Draw ocean surface at 15% from top
	waterColor := "#4a9eff"
	if len(a.waterColors) > 0 {
		waterColor = a.waterColors[0]
	}
	oceanY := int(float64(a.height) * 0.15) // 15% from top
	if oceanY < 2 {
		oceanY = 2
	}
	for x := 0; x < a.width; x++ {
		if (a.frameCount/2+x)%3 == 0 {
			canvas[oceanY][x] = '~'
			colors[oceanY][x] = waterColor
		}
	}

	// Draw ocean floor (last 2 rows)
	sandColor := "#c2b280"
	if len(a.waterColors) > 1 {
		sandColor = a.waterColors[1]
	}
	for y := a.height - 2; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			if y == a.height-2 {
				// Top of ocean floor with variation
				if (x+a.frameCount/5)%7 == 0 {
					canvas[y][x] = '^'
				} else if (x+a.frameCount/5)%5 == 0 {
					canvas[y][x] = '.'
				} else {
					canvas[y][x] = '_'
				}
			} else {
				// Bottom of ocean floor
				if (x+y)%3 == 0 {
					canvas[y][x] = '.'
				} else {
					canvas[y][x] = ' '
				}
			}
			colors[y][x] = sandColor
		}
	}

	// Draw seaweed (above ocean floor)
	for _, seaweed := range a.seaweed {
		sway := int(math.Sin(seaweed.swayPhase) * seaweed.swayAmount)

		for h := 0; h < seaweed.height; h++ {
			y := a.height - 3 - h // Start above ocean floor
			x := seaweed.x + sway

			if y >= oceanY && y < a.height-2 && x >= 0 && x < a.width {
				// Different variants
				if seaweed.variant == 0 {
					canvas[y][x] = '|'
				} else {
					// Wavy seaweed alternates
					if (h+seaweed.x)%2 == 0 {
						canvas[y][x] = '('
					} else {
						canvas[y][x] = ')'
					}
				}

				// Gradient from bottom to top
				colorIdx := int(float64(h) / float64(seaweed.height) * float64(len(seaweed.colors)))
				if colorIdx >= len(seaweed.colors) {
					colorIdx = len(seaweed.colors) - 1
				}
				colors[y][x] = seaweed.colors[colorIdx]
			}
		}
	}

	// Draw anchor (static on ocean floor)
	if a.anchor != nil {
		anchorColor := "#888888"
		startX := a.anchor.x
		startY := a.anchor.y

		for lineIdx, line := range a.anchor.pattern {
			y := startY + lineIdx
			if y >= 0 && y < a.height {
				for charIdx, char := range line {
					x := startX + charIdx
					if x >= 0 && x < a.width && char != ' ' {
						canvas[y][x] = char
						colors[y][x] = anchorColor
					}
				}
			}
		}
	}

	// Draw bubbles
	for _, bubble := range a.bubbles {
		x := int(bubble.x)
		y := int(bubble.y)

		if y >= 0 && y < a.height && x >= 0 && x < a.width {
			canvas[y][x] = 'o'
			colors[y][x] = a.bubbleColor
		}
	}

	// Draw diver
	if a.diver != nil {
		startX := int(a.diver.x)
		startY := int(a.diver.y)

		for lineIdx, line := range a.diver.pattern {
			y := startY + lineIdx
			if y >= 0 && y < a.height {
				for charIdx, char := range line {
					x := startX + charIdx
					if x >= 0 && x < a.width && char != ' ' {
						canvas[y][x] = char
						colors[y][x] = a.diverColor
					}
				}
			}
		}
	}

	// Draw boat (on surface)
	if a.boat != nil {
		startX := int(a.boat.x)
		startY := int(a.boat.y)

		for lineIdx, line := range a.boat.pattern {
			y := startY + lineIdx
			if y >= 0 && y < a.height {
				for charIdx, char := range line {
					x := startX + charIdx
					if x >= 0 && x < a.width && char != ' ' {
						canvas[y][x] = char
						colors[y][x] = a.boatColor
					}
				}
			}
		}
	}

	// Draw mermaid
	if a.mermaid != nil {
		startX := int(a.mermaid.x)
		startY := int(a.mermaid.y)

		for lineIdx, line := range a.mermaid.pattern {
			y := startY + lineIdx
			if y >= 0 && y < a.height {
				for charIdx, char := range line {
					x := startX + charIdx
					if x >= 0 && x < a.width && char != ' ' {
						canvas[y][x] = char
						colors[y][x] = a.mermaidColor
					}
				}
			}
		}
	}

	// Draw fish (on top of everything else)
	for _, fish := range a.fish {
		startX := int(fish.x)
		startY := int(fish.y)

		for lineIdx, line := range fish.pattern {
			y := startY + lineIdx
			if y >= 0 && y < a.height {
				for charIdx, char := range line {
					x := startX + charIdx
					if x >= 0 && x < a.width && char != ' ' {
						canvas[y][x] = char
						colors[y][x] = fish.color
					}
				}
			}
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < a.height; y++ {
		var line strings.Builder
		for x := 0; x < a.width; x++ {
			char := canvas[y][x]
			if char != ' ' && colors[y][x] != "" {
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(colors[y][x])).
					Render(string(char))
				line.WriteString(styled)
			} else {
				line.WriteRune(char)
			}
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// Reset restarts the animation
func (a *AquariumEffect) Reset() {
	a.fish = a.fish[:0]
	a.bubbles = a.bubbles[:0]
	a.seaweed = a.seaweed[:0]
	a.frameCount = 0
	a.init()
}

// Resize updates the aquarium dimensions
func (a *AquariumEffect) Resize(width, height int) {
	a.width = width
	a.height = height
	a.Reset()
}

// Helper function to reverse a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
