package ansiart

// Cell stores a glyph and 16-colour attributes in SGR order
// (0 black, 1 red, 2 green, 3 yellow, 4 blue, 5 magenta, 6 cyan, 7 white;
// +8 bright). Bg == -1 means transparent.
type Cell struct {
	Ch     rune
	Fg, Bg int8
}

// Grid is a rectangular cell grid.
type Grid struct {
	W, H  int
	Cells []Cell
}

// NewGrid returns a w x h grid of blank cells (space, fg 7, transparent bg).
func NewGrid(w, h int) *Grid {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	g := &Grid{W: w, H: h, Cells: make([]Cell, w*h)}
	for i := range g.Cells {
		g.Cells[i] = Cell{Ch: ' ', Fg: 7, Bg: -1}
	}
	return g
}

// At returns the cell at x,y or nil when out of range.
func (g *Grid) At(x, y int) *Cell {
	if x < 0 || y < 0 || x >= g.W || y >= g.H {
		return nil
	}
	return &g.Cells[y*g.W+x]
}

// CP437 maps byte values to their IBM CP437 glyphs. Index 0 renders as
// space so it never emits a NUL into terminal output.
var CP437 = [256]rune{
	' ', '☺', '☻', '♥', '♦', '♣', '♠', '•', '◘', '○', '◙', '♂', '♀', '♪', '♫', '☼',
	'►', '◄', '↕', '‼', '¶', '§', '▬', '↨', '↑', '↓', '→', '←', '∟', '↔', '▲', '▼',
	' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?',
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
	'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', '⌂',
	'Ç', 'ü', 'é', 'â', 'ä', 'à', 'å', 'ç', 'ê', 'ë', 'è', 'ï', 'î', 'ì', 'Ä', 'Å',
	'É', 'æ', 'Æ', 'ô', 'ö', 'ò', 'û', 'ù', 'ÿ', 'Ö', 'Ü', '¢', '£', '¥', '₧', 'ƒ',
	'á', 'í', 'ó', 'ú', 'ñ', 'Ñ', 'ª', 'º', '¿', '⌐', '¬', '½', '¼', '¡', '«', '»',
	'░', '▒', '▓', '│', '┤', '╡', '╢', '╖', '╕', '╣', '║', '╗', '╝', '╜', '╛', '┐',
	'└', '┴', '┬', '├', '─', '┼', '╞', '╟', '╚', '╔', '╩', '╦', '╠', '═', '╬', '╧',
	'╨', '╤', '╥', '╙', '╘', '╒', '╓', '╫', '╪', '┘', '┌', '█', '▄', '▌', '▐', '▀',
	'α', 'ß', 'Γ', 'π', 'Σ', 'σ', 'µ', 'τ', 'Φ', 'Θ', 'Ω', 'δ', '∞', 'φ', 'ε', '∩',
	'≡', '±', '≥', '≤', '⌠', '⌡', '÷', '≈', '°', '∙', '·', '√', 'ⁿ', '²', '■', '\u00a0',
}

var runeToCP437 = func() map[rune]byte {
	m := make(map[rune]byte, 256)
	for i := 255; i >= 32; i-- {
		m[CP437[i]] = byte(i)
	}
	for i := 31; i >= 1; i-- {
		if _, ok := m[CP437[i]]; !ok {
			m[CP437[i]] = byte(i)
		}
	}
	return m
}()

// RuneToCP437 maps a rune to its CP437 byte. Printable slots win over the
// 0x01-0x1F glyph slots for duplicated runes (so 0x20 maps to ' ').
func RuneToCP437(r rune) (byte, bool) {
	b, ok := runeToCP437[r]
	return b, ok
}
