package main

import "time"

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

// secondaryLaunchLimit is how many times in a row a helper may fail on one
// output before the manager stops relaunching it. Compositors that cannot place
// secondary surfaces then keep their wallpaper instead of flickering through
// relaunches every reconcile tick.
const secondaryLaunchLimit = 3

// A helper that ran this long before exiting counts as a fresh failure.
const secondaryStableRun = time.Minute

type secondaryLaunchFailure struct {
	key   string
	count int
}

// secondaryLaunchFailures tracks consecutive helper failures per output. A new
// key (effect, theme, or screensaver activation) gets a fresh set of attempts.
type secondaryLaunchFailures map[string]secondaryLaunchFailure

func (f secondaryLaunchFailures) blocked(output, key string) bool {
	failure, ok := f[output]
	return ok && failure.key == key && failure.count >= secondaryLaunchLimit
}

func (f secondaryLaunchFailures) record(output, key string) {
	failure := f[output]
	if failure.key != key {
		failure = secondaryLaunchFailure{key: key}
	}
	failure.count++
	f[output] = failure
}

// forgetMissing drops outputs that are no longer connected, so a reconnected
// output is tried again.
func (f secondaryLaunchFailures) forgetMissing(outputs []secondaryOutput) {
	for name := range f {
		if !hasOutput(outputs, name) {
			delete(f, name)
		}
	}
}

func secondaryLaunchKey(saver bool, want secondaryRuntimeState) string {
	if saver {
		return "saver\x00" + want.Theme + "\x00" + want.SaverStart.Format(time.RFC3339Nano)
	}
	return "background\x00" + want.Background + "\x00" + want.Theme
}
