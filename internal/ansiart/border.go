package ansiart

import (
	"errors"
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
	TitleRow                 int // row in the top band hosting the title cap; -1 = none
	BlackIsClear             bool
	Ramp                     []int8
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

func parseNineSlice(s *Sauce, b *Border) bool {
	if s == nil {
		return false
	}
	found := false
	for _, c := range s.Comments {
		if !strings.HasPrefix(strings.TrimSpace(c), "nineslice") {
			continue
		}
		found = true
		for _, kv := range strings.Fields(c)[1:] {
			k, v, _ := strings.Cut(kv, "=")
			switch k {
			case "l":
				b.Left, _ = strconv.Atoi(v)
			case "t":
				b.Top, _ = strconv.Atoi(v)
			case "r":
				b.Right, _ = strconv.Atoi(v)
			case "b":
				b.Bottom, _ = strconv.Atoi(v)
			case "anchor":
				b.CenterAnchor = v == "center"
			case "title":
				b.TitleRow, _ = strconv.Atoi(v)
			case "shimmer":
				b.ShimmerSpeed, _ = strconv.ParseFloat(v, 64)
			case "scale":
				b.Scale, _ = strconv.Atoi(v)
			case "name":
				b.Name = v
			case "label":
				b.Label = v
			case "ramp":
				b.Ramp = nil
				for _, p := range strings.Split(v, ",") {
					n, _ := strconv.Atoi(p)
					b.Ramp = append(b.Ramp, int8(n))
				}
			}
		}
		break
	}
	return found
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
	if int(sauce.TInfo2) > maxGridHeight {
		return nil, errors.New("ansiart: bad height in SAUCE TInfo2")
	}
	br := &Border{TitleRow: -1, BlackIsClear: true, Scale: 1}
	if !parseNineSlice(sauce, br) {
		return nil, errors.New("ansiart: missing nineslice metadata")
	}
	br.Art = DecodeANSI(body, width, sauce.ICE())
	if br.Art.W <= br.Left+br.Right || br.Art.H <= br.Top+br.Bottom {
		return nil, errors.New("ansiart: insets leave no interior")
	}
	if br.TitleRow >= br.Top {
		br.TitleRow = -1
	}
	if br.Scale < 1 {
		br.Scale = 1
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

// Render stretches the border to w x h around a transparent interior and
// draws the title cap on TitleRow when title is non-empty.
func (b *Border) Render(w, h int, title string) *Grid {
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
	if title != "" && b.TitleRow >= 0 {
		b.drawTitle(g, title)
	}
	return g
}

func (b *Border) drawTitle(g *Grid, title string) {
	const capLeft = "▐▓▒░ "
	const capRight = " ░▒▓▌"
	interior := g.W - b.Left - b.Right
	avail := interior - ansi.StringWidth(capLeft+capRight)
	if avail < 1 {
		return
	}
	if ansi.StringWidth(title) > avail {
		r := []rune(title)
		for len(r) > 0 && ansi.StringWidth(string(r))+1 > avail {
			r = r[:len(r)-1]
		}
		title = string(r) + "…"
	}
	runes := []rune(capLeft + title + capRight)
	x0 := (g.W - ansi.StringWidth(string(runes))) / 2
	capEnd := len([]rune(capLeft))
	capTail := len(runes) - len([]rune(capRight))
	for i, r := range runes {
		if c := g.At(x0+i, b.TitleRow); c != nil {
			fg := int8(15)
			if i < capEnd || i >= capTail {
				fg = 6
			}
			*c = Cell{Ch: r, Fg: fg, Bg: 0}
		}
	}
}
