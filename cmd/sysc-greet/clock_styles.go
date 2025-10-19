package main

// Clock style digit maps for screensaver
// Individual digit maps are in separate files:
// - clock_kompaktblk.go (3 rows)
// - clock_phmvga.go (2 rows)
// - clock_delta_corp.go (8 rows)
// - clock_dos_rebel.go (8 rows)

// getClockStyleDigits returns the digit map for the specified style
func getClockStyleDigits(style string) map[rune][]string {
	switch style {
	case "phmvga":
		return phmvgaDigits
	case "delta_corp":
		return deltaCorpDigits
	case "dos_rebel":
		return dosRebelDigits
	case "plain":
		return nil // Will return plain text
	default:
		return kompaktblkDigits
	}
}
