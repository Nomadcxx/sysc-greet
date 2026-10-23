package main

import (
	"reflect"
	"testing"
)

func TestSecondaryEffectCatalogAndIDs(t *testing.T) {
	effects := secondaryEffectCatalog("effect fire 0\neffect rain 0\neffect l2a1-sar 0\ntheme dracula\n")
	if !reflect.DeepEqual(effects, map[string]bool{"fire": true, "rain": true, "l2a1-sar": true}) {
		t.Fatalf("secondaryEffectCatalog() = %v", effects)
	}
	for _, tc := range []struct {
		background string
		want       string
		ok         bool
	}{
		{background: "fire", want: "fire", ok: true},
		{background: "ascii-rain", want: "rain", ok: true},
		{background: "l2a1-sar", want: "l2a1-sar", ok: true},
		{background: "plasma", ok: false},
	} {
		got, ok := secondaryEffectID(tc.background, effects)
		if got != tc.want || ok != tc.ok {
			t.Errorf("secondaryEffectID(%q) = %q, %v; want %q, %v", tc.background, got, ok, tc.want, tc.ok)
		}
	}
}
