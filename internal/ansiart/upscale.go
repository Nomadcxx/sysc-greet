package ansiart

// Upscale returns a k-by-k integer enlargement of a border asset.
func (b *Border) Upscale(k int) *Border {
	if k < 2 || k > 3 {
		return b
	}
	g := NewGrid(b.Art.W*k, b.Art.H*k)
	for y := 0; y < b.Art.H; y++ {
		for x := 0; x < b.Art.W; x++ {
			src := *b.Art.At(x, y)
			for dy := 0; dy < k; dy++ {
				for dx := 0; dx < k; dx++ {
					*g.At(x*k+dx, y*k+dy) = upscaleCell(src, dx, dy, k)
				}
			}
		}
	}
	out := *b
	out.Art = g
	out.Left *= k
	out.Top *= k
	out.Right *= k
	out.Bottom *= k
	out.Scale *= k
	return &out
}

func upscaleCell(c Cell, x, y, k int) Cell {
	blank := Cell{Ch: ' ', Fg: c.Fg, Bg: c.Bg}
	switch c.Ch {
	case '▀':
		if y < k/2 || k%2 == 1 && y == k/2 {
			if y == k/2 && k%2 == 1 {
				c.Ch = '▀'
			} else {
				c.Ch = '█'
			}
			return c
		}
		return blank
	case '▄':
		if k%2 == 1 && y == k/2 {
			c.Ch = '▄'
			return c
		}
		if y >= (k+1)/2 {
			c.Ch = '█'
			return c
		}
		return blank
	case '▌':
		if x < k/2 || k%2 == 1 && x == k/2 {
			if x == k/2 && k%2 == 1 {
				c.Ch = '▌'
			} else {
				c.Ch = '█'
			}
			return c
		}
		return blank
	case '▐':
		if k%2 == 1 && x == k/2 {
			c.Ch = '▐'
			return c
		}
		if x >= (k+1)/2 {
			c.Ch = '█'
			return c
		}
		return blank
	case '░', '▒', '▓', '█', ' ':
		return c
	}

	if s, ok := boxStroke(c.Ch); ok {
		cx, cy := k/2, k/2
		h, v := false, false
		if y == cy {
			h = x == cx || x < cx && s.left || x > cx && s.right
		}
		if x == cx {
			v = y == cy || y < cy && s.up || y > cy && s.down
		}
		if x == cx && y == cy {
			return c
		}
		if h {
			c.Ch = s.h
			return c
		}
		if v {
			c.Ch = s.v
			return c
		}
		return blank
	}

	// ponytail: uncommon glyphs stay centered when enlarged; add stroke mappings when shipped assets use them.
	if x == k/2 && y == k/2 {
		return c
	}
	return blank
}

type boxStrokeInfo struct {
	left, right, up, down bool
	h, v                  rune
}

var boxStrokes = map[rune]boxStrokeInfo{
	'─': {left: true, right: true, h: '─', v: '│'},
	'│': {up: true, down: true, h: '─', v: '│'},
	'┌': {right: true, down: true, h: '─', v: '│'},
	'┐': {left: true, down: true, h: '─', v: '│'},
	'└': {right: true, up: true, h: '─', v: '│'},
	'┘': {left: true, up: true, h: '─', v: '│'},
	'├': {right: true, up: true, down: true, h: '─', v: '│'},
	'┤': {left: true, up: true, down: true, h: '─', v: '│'},
	'┬': {left: true, right: true, down: true, h: '─', v: '│'},
	'┴': {left: true, right: true, up: true, h: '─', v: '│'},
	'┼': {left: true, right: true, up: true, down: true, h: '─', v: '│'},
	'═': {left: true, right: true, h: '═', v: '║'},
	'║': {up: true, down: true, h: '═', v: '║'},
	'╔': {right: true, down: true, h: '═', v: '║'},
	'╗': {left: true, down: true, h: '═', v: '║'},
	'╚': {right: true, up: true, h: '═', v: '║'},
	'╝': {left: true, up: true, h: '═', v: '║'},
	'╠': {right: true, up: true, down: true, h: '═', v: '║'},
	'╣': {left: true, up: true, down: true, h: '═', v: '║'},
	'╦': {left: true, right: true, down: true, h: '═', v: '║'},
	'╩': {left: true, right: true, up: true, h: '═', v: '║'},
	'╬': {left: true, right: true, up: true, down: true, h: '═', v: '║'},
}

func boxStroke(r rune) (boxStrokeInfo, bool) {
	s, ok := boxStrokes[r]
	return s, ok
}

// SextantRune maps a six-dot mask in TL,TR,ML,MR,BL,BR order to Unicode.
func SextantRune(mask uint8) rune {
	if mask == 0 {
		return ' '
	}
	if mask == 63 {
		return '█'
	}
	if mask == 21 {
		return '▌'
	}
	if mask == 42 {
		return '▐'
	}
	index := int(mask) - 1
	if mask > 21 {
		index--
	}
	if mask > 42 {
		index--
	}
	return rune(0x1fb00 + index)
}
