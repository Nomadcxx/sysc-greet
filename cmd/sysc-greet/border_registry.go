package main

import (
	"embed"
	"image"
	"image/color"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nomadcxx/sysc-greet/internal/ansiart"
)

//go:embed assets/borders/*.ans
var builtInBorderAssets embed.FS

type borderRenderer interface {
	ID() string
	Label() string
	Frame(m *model, termWidth, termHeight, formWidth, formHeight int) (borderFrame, bool)
	SupportsEffect(effect string) bool
}

type borderFrame struct {
	Grid                     *ansiart.Grid
	X, Y                     int
	FormX, FormY             int
	HeaderY, HelpY           int
	Left, Top, Right, Bottom int
	Speed                    float64
	ShimmerPrimary           bool
}

type borderRegistry struct {
	byID map[string]borderRenderer
	ids  []string
}

type nineSliceRenderer struct {
	id, label string
	assets    []*ansiart.Border
	innerOnly *ansiart.Border

	selectedScale      int
	selectedInnerOnly  bool
	selectedFrameWidth int
	selected           *ansiart.Border
	gridW, gridH       int
	grid               *ansiart.Grid
}

func newBorderRegistry(systemDir string) *borderRegistry {
	assets := make(map[string]*nineSliceRenderer)
	readAsset := func(name string, data []byte) {
		border, err := ansiart.LoadBorder(data)
		if err != nil {
			logDebug("Skipping border asset %s: %v", name, err)
			return
		}
		id := border.Name
		if id == "" {
			id = strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
			id = strings.TrimSuffix(id, "@2x")
			id = strings.TrimSuffix(id, "@3x")
		}
		if !validRegistryID(id) {
			logDebug("Skipping border asset %s: invalid name %q", name, id)
			return
		}
		r := assets[id]
		if r == nil {
			r = &nineSliceRenderer{id: id, label: borderLabel(id)}
			assets[id] = r
		}
		for _, existing := range r.assets {
			if existing.Scale == border.Scale {
				return
			}
		}
		if border.Label != "" {
			r.label = border.Label
		}
		r.assets = append(r.assets, border)
		sort.Slice(r.assets, func(i, j int) bool { return r.assets[i].Scale < r.assets[j].Scale })
	}

	if files, err := fs.Glob(builtInBorderAssets, "assets/borders/*.ans"); err == nil {
		for _, name := range files {
			data, err := builtInBorderAssets.ReadFile(name)
			if err == nil {
				readAsset(name, data)
			}
		}
	}
	if entries, err := os.ReadDir(systemDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".ans") {
				continue
			}
			path := filepath.Join(systemDir, entry.Name())
			f, err := os.Open(path)
			if err != nil {
				logDebug("Skipping border asset %s: %v", path, err)
				continue
			}
			data, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
			closeErr := f.Close()
			if readErr != nil || closeErr != nil || len(data) > 1<<20 {
				logDebug("Skipping border asset %s: invalid or oversized file", path)
				continue
			}
			readAsset(path, data)
		}
	}

	r := &borderRegistry{byID: make(map[string]borderRenderer)}
	r.register(newClassicBorderRenderer())
	r.register(newNineSliceRenderer("modern", "Modern", lipgloss.ThickBorder(), 14))
	r.register(newNineSliceRenderer("minimal", "Minimal", lipgloss.Border{
		Top: " ", Bottom: " ", Left: " ", Right: " ",
	}, 14))
	r.register(newNineSliceRenderer("ascii1", "ASCII-1", lipgloss.Border{
		TopLeft: "█", Top: "▀", TopRight: "█",
		Left: "█", Right: "█",
		BottomLeft: "█", Bottom: "▄", BottomRight: "█",
	}, 8))

	assetIDs := make([]string, 0, len(assets))
	for id := range assets {
		assetIDs = append(assetIDs, id)
	}
	sort.Strings(assetIDs)
	for _, id := range assetIDs {
		if _, exists := r.byID[id]; !exists {
			r.register(assets[id])
		}
	}
	r.register(&subcellBorderRenderer{id: "braille-hairline", label: "Braille Hairline", subW: 2, subH: 4})
	if sextantsSupported(os.Getenv("TERM")) {
		r.register(&subcellBorderRenderer{id: "sextant-bevel", label: "Sextant Bevel", subW: 2, subH: 3})
	}
	ensureUniqueBorderLabels(r)
	return r
}

func ensureUniqueBorderLabels(r *borderRegistry) {
	styles := r.styles()
	counts := make(map[string]int, len(styles))
	for _, style := range styles {
		counts[style.Label()]++
	}
	used := make(map[string]bool, len(styles))
	for _, style := range styles {
		base, label := style.Label(), style.Label()
		if counts[base] > 1 {
			label += " [" + style.ID() + "]"
		}
		for used[label] {
			label += " [" + style.ID() + "]"
		}
		used[label] = true
		if label != base {
			switch style := style.(type) {
			case *nineSliceRenderer:
				style.label = label
			case *subcellBorderRenderer:
				style.label = label
			}
		}
	}
}

func (r *borderRegistry) register(renderer borderRenderer) {
	if renderer == nil || r.byID[renderer.ID()] != nil {
		return
	}
	r.byID[renderer.ID()] = renderer
	r.ids = append(r.ids, renderer.ID())
}

func (r *borderRegistry) get(id string) borderRenderer {
	if r == nil {
		return nil
	}
	return r.byID[id]
}

func (r *borderRegistry) has(id string) bool { return r.get(id) != nil }

func (r *borderRegistry) styles() []borderRenderer {
	styles := make([]borderRenderer, 0, len(r.ids))
	for _, id := range r.ids {
		styles = append(styles, r.byID[id])
	}
	return styles
}

func newNineSliceRenderer(id, label string, border lipgloss.Border, fg int8) *nineSliceRenderer {
	grid := ansiart.NewGrid(3, 3)
	chars := [3][3]string{
		{border.TopLeft, border.Top, border.TopRight},
		{border.Left, " ", border.Right},
		{border.BottomLeft, border.Bottom, border.BottomRight},
	}
	for y := range chars {
		for x := range chars[y] {
			runes := []rune(chars[y][x])
			if len(runes) == 0 {
				runes = []rune{' '}
			}
			*grid.At(x, y) = ansiart.Cell{Ch: runes[0], Fg: fg, Bg: -1}
		}
	}
	return &nineSliceRenderer{
		id: id, label: label,
		assets: []*ansiart.Border{newGridBorder(id, label, grid, 1, 1, 1, 1)},
	}
}

func newGridBorder(id, label string, grid *ansiart.Grid, left, top, right, bottom int) *ansiart.Border {
	return &ansiart.Border{
		Name: id, Label: label, Art: grid,
		Left: left, Top: top, Right: right, Bottom: bottom,
		CenterAnchor: true, TitleRow: -1, BlackIsClear: true, ShimmerSpeed: 24, Scale: 1,
	}
}

func newClassicBorderRenderer() *nineSliceRenderer {
	outer, inner := lipgloss.DoubleBorder(), lipgloss.RoundedBorder()
	grid := ansiart.NewGrid(5, 5)
	set := func(x, y int, ch string, fg int8) {
		r := []rune(ch)
		if len(r) != 0 {
			*grid.At(x, y) = ansiart.Cell{Ch: r[0], Fg: fg, Bg: -1}
		}
	}
	for p := 1; p < 4; p++ {
		set(p, 0, outer.Top, 8)
		set(p, 4, outer.Bottom, 8)
		set(0, p, outer.Left, 8)
		set(4, p, outer.Right, 8)
	}
	set(0, 0, outer.TopLeft, 8)
	set(4, 0, outer.TopRight, 8)
	set(0, 4, outer.BottomLeft, 8)
	set(4, 4, outer.BottomRight, 8)
	set(1, 1, inner.TopLeft, 14)
	set(2, 1, inner.Top, 14)
	set(3, 1, inner.TopRight, 14)
	set(1, 2, inner.Left, 14)
	set(3, 2, inner.Right, 14)
	set(1, 3, inner.BottomLeft, 14)
	set(2, 3, inner.Bottom, 14)
	set(3, 3, inner.BottomRight, 14)
	r := &nineSliceRenderer{
		id: "classic", label: "Classic",
		assets: []*ansiart.Border{newGridBorder("classic", "Classic", grid, 2, 2, 2, 2)},
	}
	innerOnly := newNineSliceRenderer("classic", "Classic", lipgloss.RoundedBorder(), 14)
	r.innerOnly = innerOnly.assets[0]
	return r
}

func (r *nineSliceRenderer) ID() string    { return r.id }
func (r *nineSliceRenderer) Label() string { return r.label }
func (r *nineSliceRenderer) SupportsEffect(effect string) bool {
	return effect == "shimmer" && r.id != "minimal"
}

func suppressOuterBorder(background string) bool {
	switch background {
	case "fire", "ascii-rain", "matrix", "ticker", "fire+rain", "fireworks", "sonar", "cracktro", "aquarium", "plasma", "ak-47", "l2a1-sar":
		return true
	default:
		return isMeshBackground(background)
	}
}

func (r *nineSliceRenderer) borderFor(m *model, termWidth, termHeight, formWidth int) *ansiart.Border {
	target := borderScaleFor(termWidth, termHeight)
	innerOnly := r.innerOnly != nil && m != nil && suppressOuterBorder(m.selectedBackground)
	frameWidth := formWidth
	if m != nil && m.config.BorderLayout == "fill" {
		frameWidth = termWidth - 8
	}
	for target > 1 && !innerOnly {
		candidate := r.assetAtScale(target)
		candidateFrameWidth := frameWidth
		if m == nil || m.config.BorderLayout != "fill" {
			candidateFrameWidth += candidate.Left + candidate.Right + 4
		}
		if candidateFrameWidth >= 4*(candidate.Left+candidate.Right) {
			break
		}
		target--
	}
	if r.selected != nil && r.selectedScale == target && r.selectedInnerOnly == innerOnly && r.selectedFrameWidth == frameWidth {
		return r.selected
	}
	if innerOnly {
		r.selected = r.innerOnly
	} else {
		r.selected = r.assetAtScale(target)
	}
	r.selectedScale, r.selectedInnerOnly = target, innerOnly
	r.selectedFrameWidth = frameWidth
	r.gridW, r.gridH, r.grid = 0, 0, nil
	return r.selected
}

func (r *nineSliceRenderer) assetAtScale(target int) *ansiart.Border {
	base := r.assets[0]
	for _, candidate := range r.assets {
		if candidate.Scale <= target && candidate.Scale >= base.Scale {
			base = candidate
		}
	}
	if base.Scale < target && target%base.Scale == 0 {
		return base.Upscale(target / base.Scale)
	}
	return base
}

func (r *nineSliceRenderer) Frame(m *model, tw, th, formW, formH int) (borderFrame, bool) {
	b := r.borderFor(m, tw, th, formW)
	fw, fh, x, y, fx, fy, headerY, helpY, ok := frameLayout(m, tw, th, formW, formH, b.Left, b.Top, b.Right, b.Bottom)
	if !ok {
		return borderFrame{}, false
	}
	if r.grid == nil || r.gridW != fw || r.gridH != fh {
		r.grid = b.Render(fw, fh)
		r.gridW, r.gridH = fw, fh
	}
	speed := b.ShimmerSpeed
	if speed <= 0 {
		speed = 24
	}
	return borderFrame{
		Grid: r.grid, X: x, Y: y, FormX: fx, FormY: fy, HeaderY: headerY, HelpY: helpY,
		Left: b.Left, Top: b.Top, Right: b.Right, Bottom: b.Bottom, Speed: speed,
		ShimmerPrimary: r.id == "classic" || r.id == "modern" || r.id == "ascii1",
	}, true
}

func frameLayout(m *model, tw, th, formW, formH, left, top, right, bottom int) (fw, fh, x, y, fx, fy, headerY, helpY int, ok bool) {
	const padX, padY = 2, 1
	fw = max(formW+left+right+2*padX, left+right+1)
	fh = max(formH+top+bottom+2*padY, top+bottom+1)
	if m.config.BorderLayout == "fill" {
		if tw-8 < fw || th-8 < fh {
			return 0, 0, 0, 0, 0, 0, 0, 0, false
		}
		fw, fh = tw-8, th-8
	}
	if fw > tw || fh+4 > th {
		return 0, 0, 0, 0, 0, 0, 0, 0, false
	}
	x = max(0, (tw-fw)/2)
	groupY := max(0, (th-(fh+4))/2)
	headerY, y, helpY = groupY, groupY+2, groupY+fh+3
	innerW, innerH := fw-left-right-2*padX, fh-top-bottom-2*padY
	if formW > innerW || formH > innerH {
		return 0, 0, 0, 0, 0, 0, 0, 0, false
	}
	fx = x + left + padX + (innerW-formW)/2
	fy = y + top + padY + (innerH-formH)/2
	return fw, fh, x, y, fx, fy, headerY, helpY, true
}

func borderScaleFor(w, h int) int {
	if w >= 480 && h >= 120 {
		return 3
	}
	if w >= 320 && h >= 80 {
		return 2
	}
	return 1
}

func borderShimmerFrame(idle time.Duration) (frame int, active bool) {
	const (
		start = 10 * time.Second
		stop  = time.Minute
		step  = 50 * time.Millisecond
	)
	if idle < start {
		return 0, false
	}
	elapsed := idle - start
	if elapsed > stop-start {
		elapsed = stop - start
	}
	return int(elapsed / step), idle < stop
}

func borderPalette(mode string) (p [16]color.Color) {
	for i := range p {
		switch mode {
		case "vga":
			p[i] = vgaPalette[i]
		case "terminal":
			p[i] = ansi.BasicColor(i)
		default:
			p[i] = asciiANSIColor(i)
		}
	}
	return p
}

func validBorderPalette(mode string) bool {
	return mode == "theme" || mode == "vga" || mode == "terminal"
}

var vgaPalette = [16]color.Color{
	lipgloss.Color("#000000"), lipgloss.Color("#aa0000"), lipgloss.Color("#00aa00"), lipgloss.Color("#aa5500"),
	lipgloss.Color("#0000aa"), lipgloss.Color("#aa00aa"), lipgloss.Color("#00aaaa"), lipgloss.Color("#aaaaaa"),
	lipgloss.Color("#555555"), lipgloss.Color("#ff5555"), lipgloss.Color("#55ff55"), lipgloss.Color("#ffff55"),
	lipgloss.Color("#5555ff"), lipgloss.Color("#ff55ff"), lipgloss.Color("#55ffff"), lipgloss.Color("#ffffff"),
}

type nfoGridLayer struct {
	Grid                     *ansiart.Grid
	Palette                  [16]color.Color
	PaletteMode              string
	Profile                  colorprofile.Profile
	Left, Top, Right, Bottom int
	Shimmer                  bool
	ShimmerPrimary           bool
	Frame                    int
	Speed                    float64
	Rest                     float64
	FailureAge               time.Duration
	HasFailure               bool
}

var borderGlyphStrings = func() map[rune]string {
	stringsByRune := make(map[rune]string, len(ansiart.CP437)+256+60)
	for _, r := range ansiart.CP437 {
		stringsByRune[r] = string(r)
	}
	for r := rune(0x2800); r <= 0x28ff; r++ {
		stringsByRune[r] = string(r)
	}
	for r := rune(0x1fb00); r <= 0x1fb3b; r++ {
		stringsByRune[r] = string(r)
	}
	return stringsByRune
}()

func borderGlyphString(r rune) string {
	if s, ok := borderGlyphStrings[r]; ok {
		return s
	}
	return string(r)
}

func (l nfoGridLayer) Bounds() image.Rectangle {
	if l.Grid == nil {
		return image.Rectangle{}
	}
	return image.Rect(0, 0, l.Grid.W, l.Grid.H)
}

func (l nfoGridLayer) Draw(scr uv.Screen, area image.Rectangle) {
	if l.Grid == nil {
		return
	}
	var wave ansiart.ShimmerWave
	if l.Shimmer {
		_, p := ansiart.PerimeterDist(0, 0, l.Grid.W, l.Grid.H)
		wave = ansiart.NewShimmerWave(l.Frame, p, l.Speed, 6, l.Rest)
	}
	var cell uv.Cell
	for y := 0; y < l.Grid.H; y++ {
		for x := 0; x < l.Grid.W; x++ {
			if x >= l.Left && x < l.Grid.W-l.Right && y >= l.Top && y < l.Grid.H-l.Bottom {
				continue
			}
			c := l.Grid.Cells[y*l.Grid.W+x]
			if c.Bg < 0 && c.Ch == ' ' {
				continue
			}
			px, py := area.Min.X+x, area.Min.Y+y
			if !image.Pt(px, py).In(area) {
				continue
			}
			fgIndex := max(0, min(15, int(c.Fg)))
			fg := l.Palette[fgIndex]
			primary := l.ShimmerPrimary && (fgIndex == 8 || fgIndex == 14)
			if l.Shimmer && (fgIndex < 8 || primary) {
				d, _ := ansiart.PerimeterDist(x, y, l.Grid.W, l.Grid.H)
				gain := wave.Gain(d)
				if gain > 0 {
					if l.Profile == colorprofile.TrueColor && l.PaletteMode != "terminal" {
						highlight := l.Palette[fgIndex|8]
						if primary {
							if l.PaletteMode == "vga" {
								highlight = l.Palette[15]
							} else {
								highlight = FgPrimary
							}
						} else if l.PaletteMode != "vga" {
							highlight = Primary
						}
						fg = mixColor(fg, highlight, .6*gain)
					} else if gain > .5 {
						brightIndex := fgIndex | 8
						if primary {
							brightIndex = 15
						}
						fg = l.Palette[brightIndex]
					}
				}
			}
			if l.HasFailure && l.FailureAge >= 0 && l.FailureAge < 400*time.Millisecond {
				alpha := .7 * math.Exp(-float64(l.FailureAge)/float64(250*time.Millisecond))
				fg = mixColor(fg, Danger, alpha)
			}
			style := uv.Style{Fg: fg}
			if c.Bg >= 0 {
				style.Bg = l.Palette[max(0, min(15, int(c.Bg)))]
			} else if under := scr.CellAt(px, py); under != nil {
				style.Bg = under.Style.Bg
			}
			cell = uv.Cell{Content: borderGlyphString(c.Ch), Width: 1, Style: style}
			scr.SetCell(px, py, &cell)
		}
	}
}

type subcellBorderRenderer struct {
	id, label  string
	subW, subH int
	w, h       int
	grid       *ansiart.Grid
}

func (r *subcellBorderRenderer) ID() string                        { return r.id }
func (r *subcellBorderRenderer) Label() string                     { return r.label }
func (r *subcellBorderRenderer) SupportsEffect(effect string) bool { return effect == "shimmer" }
func (r *subcellBorderRenderer) Frame(m *model, tw, th, formW, formH int) (borderFrame, bool) {
	fw, fh, x, y, fx, fy, headerY, helpY, ok := frameLayout(m, tw, th, formW, formH, 1, 1, 1, 1)
	if !ok {
		return borderFrame{}, false
	}
	if r.grid == nil || r.w != fw || r.h != fh {
		r.grid = makeSubcellBorder(fw, fh, r.subW, r.subH, 14)
		r.w, r.h = fw, fh
	}
	return borderFrame{Grid: r.grid, X: x, Y: y, FormX: fx, FormY: fy, HeaderY: headerY, HelpY: helpY, Left: 1, Top: 1, Right: 1, Bottom: 1, Speed: 24, ShimmerPrimary: true}, true
}

func makeSubcellBorder(w, h, subW, subH int, fg int8) *ansiart.Grid {
	g := ansiart.NewGrid(w, h)
	dw, dh := w*subW, h*subH
	cellMasks := make([]uint8, w*h)
	set := func(dx, dy int) {
		if dx < 0 || dy < 0 || dx >= dw || dy >= dh {
			return
		}
		x, y := dx/subW, dy/subH
		if subW == 2 && subH == 4 {
			cellMasks[y*w+x] |= brailleDotBit(dx%subW, dy%subH)
			return
		}
		bit := dy%subH*subW + dx%subW
		cellMasks[y*w+x] |= 1 << bit
	}
	minDim := min(w, h)
	radius := max(2, min(8, minDim/6))
	rx, ry := min(dw/2, radius*subW), min(dh/2, radius*subH)
	drawContour := func(inset int) {
		x0, y0, x1, y1 := inset, inset, dw-1-inset, dh-1-inset
		if x0 >= x1 || y0 >= y1 {
			return
		}
		cornerRX, cornerRY := max(0, rx-inset), max(0, ry-inset)
		for dx := x0 + cornerRX; dx <= x1-cornerRX; dx++ {
			set(dx, y0)
			set(dx, y1)
		}
		for dy := y0 + cornerRY; dy <= y1-cornerRY; dy++ {
			set(x0, dy)
			set(x1, dy)
		}
		for cy := 0; cy <= cornerRY; cy++ {
			delta := float64(cornerRY-cy) / float64(max(1, cornerRY))
			cx := cornerRX - int(math.Round(float64(cornerRX)*math.Sqrt(max(0, 1-delta*delta))))
			set(x0+cx, y0+cy)
			set(x1-cx, y0+cy)
			set(x0+cx, y1-cy)
			set(x1-cx, y1-cy)
		}
	}
	drawContour(0)
	if subW == 2 && subH == 4 {
		// ponytail: two one-dot strokes leave one dot between the braille hairlines.
		drawContour(2)
	}
	for i, mask := range cellMasks {
		if mask == 0 {
			continue
		}
		r := rune(0x2800) + rune(mask)
		if subW == 2 && subH == 3 {
			r = ansiart.SextantRune(mask)
		}
		g.Cells[i] = ansiart.Cell{Ch: r, Fg: fg, Bg: -1}
	}
	return g
}

func brailleDotBit(x, y int) uint8 {
	if x < 0 || x > 1 || y < 0 || y > 3 {
		return 0
	}
	return [4][2]uint8{{1, 8}, {2, 16}, {4, 32}, {64, 128}}[y][x]
}

func borderLabel(id string) string {
	switch id {
	case "ascii2":
		return "ASCII-2"
	case "nfo01":
		return "NFO 01: Cyan Bevel"
	case "nfo02":
		return "NFO 02: Shaded Double-Line"
	case "nfo03":
		return "NFO 03: Gradient Bar"
	default:
		words := strings.ReplaceAll(id, "-", " ")
		if words == "" {
			return id
		}
		return strings.ToUpper(words[:1]) + words[1:]
	}
}

func validRegistryID(id string) bool {
	if id == "" || id[0] < 'a' || id[0] > 'z' {
		return false
	}
	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

func sextantsSupported(term string) bool {
	return strings.EqualFold(term, "xterm-kitty")
}
