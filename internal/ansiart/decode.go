package ansiart

import (
	"strconv"
	"strings"
)

// DecodeANSI interprets an ANSI.SYS stream (CP437 bytes) onto a grid of
// the given width: CUP/CUU/CUD/CUF/CUB, save/restore (SCP/RCP), ED, EL,
// SGR, CR/LF/TAB and DEC-style deferred wrap. With ice set, SGR 5 selects
// bright backgrounds instead of being ignored. Unwritten cells read as
// space on black.
func DecodeANSI(body []byte, width int, ice bool) *Grid {
	if width < 1 {
		width = 1
	}
	type pos struct{ x, y int }
	cells := map[pos]Cell{}
	x, y, maxY := 0, 0, 0
	var fg, bg int8 = 7, 0
	bold, blink, reverse := false, false, false
	pendingWrap := false
	saved := pos{}

	put := func(p pos, c Cell) {
		cells[p] = c
		if p.y > maxY {
			maxY = p.y
		}
	}
	blank := func() Cell {
		f, b := fg, bg
		if bold {
			f += 8
		}
		if blink && ice {
			b += 8
		}
		if reverse {
			f, b = b, f
		}
		return Cell{Ch: ' ', Fg: f, Bg: b}
	}

	for i := 0; i < len(body); i++ {
		c := body[i]
		if c == 0x1b && i+1 < len(body) && body[i+1] == '[' {
			j := i + 2
			for j < len(body) && (body[j] == ';' || body[j] == '?' || (body[j] >= '0' && body[j] <= '9')) {
				j++
			}
			if j >= len(body) {
				break
			}
			var ps []int
			for _, p := range strings.Split(strings.TrimPrefix(string(body[i+2:j]), "?"), ";") {
				n, _ := strconv.Atoi(p)
				ps = append(ps, n)
			}
			p0 := 1
			if len(ps) > 0 && ps[0] > 0 {
				p0 = ps[0]
			}
			switch body[j] {
			case 'm':
				for _, p := range ps {
					switch {
					case p == 0:
						fg, bg, bold, blink, reverse = 7, 0, false, false, false
					case p == 1:
						bold = true
					case p == 5:
						blink = true
					case p == 7:
						reverse = true
					case p == 22:
						bold = false
					case p == 25:
						blink = false
					case p == 27:
						reverse = false
					case p == 39:
						fg = 7
					case p == 49:
						bg = 0
					case p >= 30 && p <= 37:
						fg = int8(p - 30)
					case p >= 40 && p <= 47:
						bg = int8(p - 40)
					case p >= 90 && p <= 97:
						fg, bold = int8(p-90), true
					case p >= 100 && p <= 107:
						bg, blink = int8(p-100), true
					}
				}
			case 'A':
				y = max(0, y-p0)
			case 'B':
				y += p0
			case 'C':
				x = min(width-1, x+p0)
			case 'D':
				x = max(0, x-p0)
			case 'H', 'f':
				y, x = 0, 0
				if len(ps) > 0 && ps[0] > 0 {
					y = ps[0] - 1
				}
				if len(ps) > 1 && ps[1] > 0 {
					x = ps[1] - 1
				}
			case 's':
				saved = pos{x, y}
			case 'u':
				x, y = saved.x, saved.y
			case 'J':
				switch {
				case len(ps) > 0 && ps[0] == 2:
					cells, x, y = map[pos]Cell{}, 0, 0
				case len(ps) == 0 || ps[0] == 0:
					for xx := x; xx < width; xx++ {
						put(pos{xx, y}, blank())
					}
					for yy := y + 1; yy <= maxY; yy++ {
						for xx := 0; xx < width; xx++ {
							put(pos{xx, yy}, blank())
						}
					}
				}
			case 'K':
				switch {
				case len(ps) > 0 && ps[0] == 2:
					for xx := 0; xx < width; xx++ {
						put(pos{xx, y}, blank())
					}
				case len(ps) > 0 && ps[0] == 1:
					for xx := 0; xx <= x && xx < width; xx++ {
						put(pos{xx, y}, blank())
					}
				default:
					for xx := x; xx < width; xx++ {
						put(pos{xx, y}, blank())
					}
				}
			}
			pendingWrap = pendingWrap && body[j] == 'm'
			i = j
			continue
		}
		switch c {
		case '\r':
			x, pendingWrap = 0, false
			continue
		case '\n':
			x, y, pendingWrap = 0, y+1, false
			continue
		case '\t':
			x = min(width-1, (x/8+1)*8)
			continue
		case 0x1a, 0x00:
			continue
		}
		if pendingWrap {
			x, y, pendingWrap = 0, y+1, false
		}
		f, b := fg, bg
		if bold {
			f += 8
		}
		if blink && ice {
			b += 8
		}
		if reverse {
			f, b = b, f
		}
		put(pos{x, y}, Cell{Ch: CP437[c], Fg: f, Bg: b})
		if x++; x >= width {
			x, pendingWrap = width-1, true
		}
	}

	g := NewGrid(width, maxY+1)
	for p, c := range cells {
		if cell := g.At(p.x, p.y); cell != nil {
			*cell = c
		}
	}
	for i := range g.Cells {
		if g.Cells[i].Bg < 0 {
			g.Cells[i].Bg = 0
		}
	}
	return g
}

// ToCP437ANS writes the grid as a classic 16-colour .ans (CP437, CRLF)
// with a SAUCE record, so Moebius/PabloDraw/ansilove can open the output.
// Runes without a CP437 slot are written as '?' and control-range glyphs
// are never emitted as raw bytes.
func (g *Grid) ToCP437ANS(s Sauce) []byte {
	var buf []byte
	buf = append(buf, "\x1b[0m"...)
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.Cells[y*g.W+x]
			bg := c.Bg
			if bg < 0 {
				bg = 0
			}
			attr := "0"
			if c.Fg >= 8 {
				attr += ";1"
			}
			if bg >= 8 {
				attr += ";5"
			}
			esc := "\x1b[" + attr + ";" + strconv.Itoa(30+int(c.Fg&7)) + ";" + strconv.Itoa(40+int(bg&7)) + "m"
			buf = append(buf, esc...)
			ch, ok := RuneToCP437(c.Ch)
			if !ok || ch < 32 {
				ch = '?'
			}
			buf = append(buf, ch)
		}
		if g.W < 80 {
			buf = append(buf, "\r\n"...)
		}
	}
	buf = append(buf, "\x1b[0m"...)
	s.TInfo1, s.TInfo2 = uint16(g.W), uint16(g.H)
	if s.TInfoS == "" {
		s.TInfoS = "IBM VGA"
	}
	if s.DataType == 0 {
		s.DataType = 1
		s.FileType = 1
	}
	buf = append(buf, EncodeSAUCE(s)...)
	return buf
}
