package main

type secondaryContentMode uint8

const (
	secondaryContentNone secondaryContentMode = iota
	secondaryContentBackground
	secondaryContentScreensaver
)

type secondaryContent struct {
	Mode   secondaryContentMode
	Effect string
}

func secondaryContentFor(background, wallpaper string, screensaver, enabled bool, override string) secondaryContent {
	if screensaver {
		return secondaryContent{Mode: secondaryContentScreensaver}
	}
	if !enabled || wallpaper != "" {
		return secondaryContent{Mode: secondaryContentNone}
	}
	if background == "" || background == "none" {
		background = override
	}
	if background == "" || background == "none" {
		return secondaryContent{Mode: secondaryContentNone}
	}
	return secondaryContent{Mode: secondaryContentBackground, Effect: background}
}
