package ansiart

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Border is a 9-slice border asset. Art holds the source frame; the
// insets mark fixed corner/edge thickness. The centre band of each edge
// tiles. Interior cells stay transparent.
type Border struct {
	Name, Label              string
	Art                      *Grid
	Left, Top, Right, Bottom int
	CenterAnchor             bool
	TitleRow                 int
	BlackIsClear             bool
	ShimmerSpeed             float64
	Scale                    int
}

const maxANSIFile = 1 << 20
const maxGridWidth = 512
const maxGridHeight = 256

// MinSize returns the smallest frame this border can stretch to.
func (b *Border) MinSize() (w, h int) {
	return b.Left + b.Right + 1, b.Top + b.Bottom + 1
}

func parseNineSlice(s *Sauce, b *Border) error {
	if s == nil {
		return errors.New("ansiart: missing nineslice metadata")
	}
	found := false
	var gotLeft, gotTop, gotRight, gotBottom, gotAnchor bool
	seen := make(map[string]bool)
	for _, c := range s.Comments {
		fields := strings.Fields(strings.TrimSpace(c))
		if len(fields) == 0 || fields[0] != "nineslice" {
			continue
		}
		if found {
			return errors.New("ansiart: duplicate nineslice metadata")
		}
		found = true
		for _, kv := range fields[1:] {
			k, v, ok := strings.Cut(kv, "=")
			if !ok || v == "" {
				return errors.New("ansiart: malformed nineslice metadata")
			}
			if seen[k] {
				return errors.New("ansiart: duplicate nineslice field")
			}
			seen[k] = true
			setInt := func(dst *int, seen *bool) error {
				n, err := strconv.Atoi(v)
				if err != nil || n < 0 || n > maxGridWidth {
					return errors.New("ansiart: invalid nineslice inset")
				}
				*dst, *seen = n, true
				return nil
			}
			switch k {
			case "l":
				if err := setInt(&b.Left, &gotLeft); err != nil {
					return err
				}
			case "t":
				if err := setInt(&b.Top, &gotTop); err != nil {
					return err
				}
			case "r":
				if err := setInt(&b.Right, &gotRight); err != nil {
					return err
				}
			case "b":
				if err := setInt(&b.Bottom, &gotBottom); err != nil {
					return err
				}
			case "anchor":
				if v != "center" && v != "start" {
					return errors.New("ansiart: invalid nineslice anchor")
				}
				b.CenterAnchor, gotAnchor = v == "center", true
			case "title":
				n, err := strconv.Atoi(v)
				if err != nil || n < -1 || n >= maxGridHeight {
					return errors.New("ansiart: invalid nineslice title row")
				}
				b.TitleRow = n
			case "shimmer":
				n, err := strconv.ParseFloat(v, 64)
				if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 || n > 512 {
					return errors.New("ansiart: invalid nineslice shimmer speed")
				}
				b.ShimmerSpeed = n
			case "scale":
				n, err := strconv.Atoi(v)
				if err != nil || n < 1 || n > 3 {
					return errors.New("ansiart: invalid nineslice scale")
				}
				b.Scale = n
			case "name":
				if !validBorderName(v) {
					return errors.New("ansiart: invalid nineslice name")
				}
				b.Name = v
			case "label":
				if !validBorderLabel(v) {
					return errors.New("ansiart: invalid nineslice label")
				}
				b.Label = v
			}
		}
	}
	if !found {
		return errors.New("ansiart: missing nineslice metadata")
	}
	if !gotLeft || !gotTop || !gotRight || !gotBottom || !gotAnchor {
		return errors.New("ansiart: incomplete nineslice metadata")
	}
	if b.Scale == 0 {
		b.Scale = 1
	}
	return nil
}

func validBorderLabel(s string) bool {
	if s == "" || len(s) > 48 {
		return false
	}
	for _, b := range []byte(s) {
		if b < 0x20 || b > 0x7e {
			return false
		}
	}
	return true
}

func validBorderName(s string) bool {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

// LoadBorder parses a CP437 .ans file with a SAUCE record into a border
// asset. It rejects oversized files, missing SAUCE or nineslice metadata,
// and insets that leave no interior.
func LoadBorder(b []byte) (*Border, error) {
	if len(b) > maxANSIFile {
		return nil, errors.New("ansiart: file exceeds 1 MiB")
	}
	body, sauce := ParseSAUCE(b)
	if sauce == nil {
		return nil, errors.New("ansiart: missing SAUCE record")
	}
	width := int(sauce.TInfo1)
	if width < 1 || width > maxGridWidth {
		return nil, errors.New("ansiart: bad width in SAUCE TInfo1")
	}
	height := int(sauce.TInfo2)
	if height < 1 || height > maxGridHeight {
		return nil, errors.New("ansiart: bad height in SAUCE TInfo2")
	}
	if sauce.DataType != 1 || sauce.FileType != 1 {
		return nil, errors.New("ansiart: SAUCE record is not ANSi type 1")
	}
	if sauce.ICE() {
		return nil, errors.New("ansiart: iCE colours are unsupported")
	}
	br := &Border{BlackIsClear: true, TitleRow: -1}
	if err := parseNineSlice(sauce, br); err != nil {
		return nil, err
	}
	grid, err := decodeANSI(body, width, height, false)
	if err != nil {
		return nil, err
	}
	br.Art = grid
	if br.Art.H < height {
		padded := NewGrid(width, height)
		copy(padded.Cells, br.Art.Cells)
		br.Art = padded
	}
	if br.Art.W <= br.Left+br.Right || br.Art.H <= br.Top+br.Bottom {
		return nil, errors.New("ansiart: insets leave no interior")
	}
	if br.Left+br.Right >= width || br.Top+br.Bottom >= height {
		return nil, errors.New("ansiart: insets leave no interior")
	}
	if br.TitleRow >= 0 && br.TitleRow >= br.Top {
		return nil, errors.New("ansiart: title row must be inside the top inset")
	}
	return br, nil
}

// sliceIndex maps an output coordinate to a source coordinate on one
// axis: fixed insets keep their corners, the middle band tiles.
func sliceIndex(o, out, src, lo, hi int, center bool) int {
	if o < lo {
		return o
	}
	if o >= out-hi {
		return src - (out - o)
	}
	mid := src - lo - hi
	if mid <= 0 {
		return lo
	}
	k := o - lo
	if center {
		run := out - lo - hi
		k = ((k-(run-mid)/2)%mid + mid) % mid
	} else {
		k %= mid
	}
	return lo + k
}

// Render stretches the border to w x h around a transparent interior.
func (b *Border) Render(w, h int) *Grid {
	return b.RenderWithTitle(w, h, "")
}

// RenderWithTitle draws an optional title cap when the asset declares a title row.
// ponytail: Grid stores one rune per cell, so wide or zero-width title runes omit the cap.
func (b *Border) RenderWithTitle(w, h int, title string) *Grid {
	minW, minH := b.MinSize()
	w, h = max(w, minW), max(h, minH)
	g := NewGrid(w, h)
	for y := 0; y < h; y++ {
		sy := sliceIndex(y, h, b.Art.H, b.Top, b.Bottom, b.CenterAnchor)
		for x := 0; x < w; x++ {
			if x >= b.Left && x < w-b.Right && y >= b.Top && y < h-b.Bottom {
				continue
			}
			sx := sliceIndex(x, w, b.Art.W, b.Left, b.Right, b.CenterAnchor)
			c := *b.Art.At(sx, sy)
			if b.BlackIsClear && c.Bg == 0 && (c.Ch == ' ' || c.Ch == '\u00a0') {
				c.Bg = -1
			}
			*g.At(x, y) = c
		}
	}
	if title != "" && b.TitleRow >= 0 && b.TitleRow < g.H {
		b.drawTitle(g, title)
	}
	return g
}

func (b *Border) drawTitle(g *Grid, title string) {
	const capLeft, capRight = "▐▓▒░ ", " ░▒▓▌"
	title = ansi.Strip(title)
	for _, r := range title {
		if r < ' ' || r == 0x7f || ansi.StringWidth(string(r)) != 1 {
			return
		}
	}
	available := g.W - b.Left - b.Right - ansi.StringWidth(capLeft+capRight)
	if available < 1 {
		return
	}
	if ansi.StringWidth(title) > available {
		title = ansi.Truncate(title, available, "…")
	}
	if ansi.StringWidth(title) != len([]rune(title)) || ansi.StringWidth(title) > available {
		return
	}
	cap := capLeft + title + capRight
	runes := []rune(cap)
	x0 := (g.W - ansi.StringWidth(cap)) / 2
	leftEnd := len([]rune(capLeft))
	rightStart := len(runes) - len([]rune(capRight))
	for i, r := range runes {
		if c := g.At(x0+i, b.TitleRow); c != nil {
			fg := int8(15)
			if i < leftEnd || i >= rightStart {
				fg = 6
			}
			*c = Cell{Ch: r, Fg: fg, Bg: 0}
		}
	}
}
