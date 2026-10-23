package main

import (
	"sync/atomic"
	"time"
)

type secondaryRuntimeState struct {
	Theme       string
	Background  string
	Wallpaper   string
	Screensaver bool
	SaverStart  time.Time
}

type secondaryStateBus struct {
	value   atomic.Value
	changed chan struct{}
}

func newSecondaryStateBus() *secondaryStateBus {
	bus := &secondaryStateBus{changed: make(chan struct{}, 1)}
	bus.value.Store(secondaryRuntimeState{})
	return bus
}

func (b *secondaryStateBus) load() secondaryRuntimeState {
	return b.value.Load().(secondaryRuntimeState)
}

func (b *secondaryStateBus) publish(state secondaryRuntimeState) {
	if b.load() == state {
		return
	}
	b.value.Store(state)
	select {
	case b.changed <- struct{}{}:
	default:
	}
}

func (m model) publishSecondaryRuntime() {
	if m.secondaryRuntime == nil {
		return
	}
	m.secondaryRuntime.publish(secondaryRuntimeState{
		Theme:       m.currentTheme,
		Background:  m.selectedBackground,
		Wallpaper:   m.selectedWallpaper,
		Screensaver: m.mode == ModeScreensaver,
		SaverStart:  m.screensaverStart,
	})
}
