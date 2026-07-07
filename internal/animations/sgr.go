package animations

import (
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// sgrLineWriter builds effect frames by writing truecolor SGR sequences
// directly, emitting a color code only when the foreground changes along a
// row. This replaces per-cell lipgloss.NewStyle().Render() calls, which
// allocate a style and emit a full sequence + reset for every character
// (measured 39x slower on a fullscreen fire frame).
//
// The output is parsed into cells by the compositor (ultraviolet), which
// re-emits colors according to the detected terminal profile, so truecolor
// sequences here do not bypass TTY color fallback.
//
// Every row is closed with a reset before its newline so colors never bleed
// into layer padding (ghosting protection, see e265c41).
type sgrLineWriter struct {
	sb      strings.Builder
	current string // hex color of the active SGR state; "" = terminal default
}

func newSGRLineWriter(capacity int) *sgrLineWriter {
	w := &sgrLineWriter{}
	w.sb.Grow(capacity)
	return w
}

// WriteCell writes one rune with the given "#rrggbb" foreground.
// An empty color writes the rune in the terminal default.
func (w *sgrLineWriter) WriteCell(hex string, r rune) {
	if hex != w.current {
		if hex == "" {
			w.sb.WriteString("\x1b[0m")
		} else if len(hex) == 7 && hex[0] == '#' {
			w.sb.WriteString("\x1b[38;2;")
			writeByteDecimal(&w.sb, hexPair(hex[1], hex[2]))
			w.sb.WriteByte(';')
			writeByteDecimal(&w.sb, hexPair(hex[3], hex[4]))
			w.sb.WriteByte(';')
			writeByteDecimal(&w.sb, hexPair(hex[5], hex[6]))
			w.sb.WriteByte('m')
		} else {
			// Non-hex color (should not happen with palette colors):
			// fall back to lipgloss and clear run-length state
			w.sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(string(r)))
			w.current = ""
			return
		}
		w.current = hex
	}
	w.sb.WriteRune(r)
}

// EndRow resets the SGR state and terminates the row.
func (w *sgrLineWriter) EndRow() {
	if w.current != "" {
		w.sb.WriteString("\x1b[0m")
		w.current = ""
	}
	w.sb.WriteByte('\n')
}

// String returns the frame without the trailing newline of the last row.
func (w *sgrLineWriter) String() string {
	return strings.TrimSuffix(w.sb.String(), "\n")
}

func hexPair(hi, lo byte) int {
	return hexNibble(hi)<<4 | hexNibble(lo)
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}

// writeRGBSGR writes a truecolor foreground sequence for numeric r,g,b.
func writeRGBSGR(sb *strings.Builder, r, g, b int) {
	sb.WriteString("\x1b[38;2;")
	writeByteDecimal(sb, r)
	sb.WriteByte(';')
	writeByteDecimal(sb, g)
	sb.WriteByte(';')
	writeByteDecimal(sb, b)
	sb.WriteByte('m')
}

func writeByteDecimal(sb *strings.Builder, n int) {
	if n < 0 || n > 255 {
		n = 0
	}
	var buf [3]byte
	i := 3
	for {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
		if n == 0 {
			break
		}
	}
	sb.Write(buf[i:])
}
