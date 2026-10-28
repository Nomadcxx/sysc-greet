package main

import (
	"image/color"
	"strings"
	"time"
)

const asciiConfigCacheTTL = time.Minute

type asciiConfigCacheEntry struct {
	config   ASCIIConfig
	err      error
	loadedAt time.Time
	version  uint64
}

type asciiRenderKey struct {
	path       string
	art        string
	variant    int
	ansi       bool
	foreground uint64
	background uint64
}

type asciiRenderResult struct {
	text   string
	inline bool
}

type asciiConfigCache struct {
	configs  map[string]asciiConfigCacheEntry
	rendered map[asciiRenderKey]asciiRenderResult
	load     func(string) (ASCIIConfig, error)
	now      func() time.Time
}

func newASCIIConfigCache() *asciiConfigCache {
	return &asciiConfigCache{
		configs:  make(map[string]asciiConfigCacheEntry),
		rendered: make(map[asciiRenderKey]asciiRenderResult),
		load:     loadASCIIConfig,
		now:      time.Now,
	}
}

func (c *asciiConfigCache) loadConfig(path string) (ASCIIConfig, error) {
	if c == nil {
		return loadASCIIConfig(path)
	}
	now := time.Now
	if c.now != nil {
		now = c.now
	}
	loadedAt := now()
	if entry, ok := c.configs[path]; ok && loadedAt.Sub(entry.loadedAt) < asciiConfigCacheTTL {
		return entry.config, entry.err
	}
	load := c.load
	if load == nil {
		load = loadASCIIConfig
	}
	config, err := load(path)
	version := c.configs[path].version + 1
	c.configs[path] = asciiConfigCacheEntry{config: config, err: err, loadedAt: now(), version: version}
	for key := range c.rendered {
		if key.path == path {
			delete(c.rendered, key)
		}
	}
	return config, err
}

func (c *asciiConfigCache) version(path string) uint64 {
	if c == nil {
		return 0
	}
	return c.configs[path].version
}

func (c *asciiConfigCache) render(key asciiRenderKey, render func() asciiRenderResult) asciiRenderResult {
	if c == nil {
		return render()
	}
	if result, ok := c.rendered[key]; ok {
		return result
	}
	result := render()
	c.rendered[key] = result
	return result
}

type loginViewKey struct {
	width, height              int
	mode                       ViewMode
	focus                      FocusState
	sessionName, sessionType   string
	sessionExec                string
	sessionIndex, asciiIndex   int
	selectedBorderStyle        string
	selectedBackground         string
	currentTheme               string
	borderLayout               string
	showTime                   bool
	asciiVersion               uint64
	errorMessage, authNotice   string
	authPromptText             string
	authPromptType             string
	authPromptPending          bool
	failedAttempts             int
	capsLock, dropdownOpen     bool
	hideAmbient                bool
	usernameView, passwordView string
	ambientDual, ambientShort  string
	clockSecond                int64
	themeColors                [12]uint64
}

type loginViewCache struct {
	valid     bool
	key       loginViewKey
	content   string
	formValid bool
	formKey   loginViewKey
	form      string
}

func colorCacheKey(c color.Color) uint64 {
	if c == nil {
		return 0
	}
	r, g, b, a := c.RGBA()
	return uint64(r)<<48 | uint64(g)<<32 | uint64(b)<<16 | uint64(a)
}

func (m model) loginViewKey(width, height int) (loginViewKey, bool) {
	if m.loginViewCache == nil || (m.mode != ModeLogin && m.mode != ModePassword) {
		return loginViewKey{}, false
	}
	switch m.selectedBackground {
	case "print", "beams", "pour", "ticker":
		return loginViewKey{}, false
	}
	var sessionName, sessionType, sessionExec string
	if m.selectedSession != nil {
		sessionName = m.selectedSession.Name
		sessionType = m.selectedSession.Type
		sessionExec = m.selectedSession.Exec
	}
	innerWidth := min(100, width-8)
	compactWidth := min(72, max(15, width-6))
	key := loginViewKey{
		width:               width,
		height:              height,
		mode:                m.mode,
		focus:               m.focusState,
		sessionName:         sessionName,
		sessionType:         sessionType,
		sessionExec:         sessionExec,
		sessionIndex:        m.sessionIndex,
		asciiIndex:          m.asciiArtIndex,
		selectedBorderStyle: m.selectedBorderStyle,
		selectedBackground:  m.selectedBackground,
		currentTheme:        m.currentTheme,
		borderLayout:        m.config.BorderLayout,
		showTime:            m.config.ShowTime,
		errorMessage:        m.errorMessage,
		authNotice:          m.authNotice,
		authPromptText:      m.authPromptText,
		authPromptType:      string(m.authPromptType),
		authPromptPending:   m.authPromptReply != nil,
		failedAttempts:      m.failedAttempts,
		capsLock:            m.capsLockOn,
		dropdownOpen:        m.sessionDropdownOpen,
		hideAmbient:         m.hideAmbient,
		usernameView:        m.usernameInput.View(),
		passwordView:        m.passwordInput.View(),
		themeColors: [12]uint64{
			colorCacheKey(Primary), colorCacheKey(Secondary), colorCacheKey(Accent),
			colorCacheKey(Danger), colorCacheKey(Warning), colorCacheKey(BgBase),
			colorCacheKey(BorderDefault), colorCacheKey(BorderFocus), colorCacheKey(FgPrimary),
			colorCacheKey(FgSecondary), colorCacheKey(FgMuted), colorCacheKey(FgSubtle),
		},
	}
	if m.config.ShowTime {
		key.clockSecond = time.Now().Unix()
	}
	if sessionName != "" {
		configPath := dataDir + "/ascii_configs/" + asciiConfigFileName(sessionName) + ".conf"
		_, _ = m.cachedASCIIConfig(configPath)
		key.asciiVersion = m.asciiConfigCache.version(configPath)
	}
	if m.config.Ambient.Metrics || strings.TrimSpace(m.config.Ambient.WeatherLocation) != "" {
		key.ambientDual = m.renderAmbientRow(innerWidth - 20)
		key.ambientShort = m.renderAmbientRow(compactWidth)
	}
	return key, true
}
