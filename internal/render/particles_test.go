package render

import (
	"reflect"
	"testing"
)

func TestBurstsArePrimitiveAndDistinct(t *testing.T) {
	bursts := map[string][]Cmd{
		"comb":    BuildCombBurst(),
		"landing": BuildLandingBurst(),
		"snip":    BuildSnipBurst(),
	}
	for name, cs := range bursts {
		if len(cs) == 0 {
			t.Fatalf("%s burst returned no commands", name)
		}
		for _, c := range cs {
			if c.Shape < ShapeBox || c.Shape > ShapePlane {
				t.Fatalf("%s burst uses non-primitive shape %v", name, c.Shape)
			}
		}
	}

	names := []string{"comb", "landing", "snip"}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if reflect.DeepEqual(bursts[names[i]], bursts[names[j]]) {
				t.Fatalf("%s and %s bursts are identical", names[i], names[j])
			}
		}
	}
}

func TestBurstColorsMatchKind(t *testing.T) {
	if !cmdsHaveColor(BuildCombBurst(), combParticle) {
		t.Fatal("comb burst has no comb particle color")
	}
	if !cmdsHaveColor(BuildLandingBurst(), landParticle) {
		t.Fatal("landing burst has no dust color")
	}
	if !cmdsHaveColor(BuildSnipBurst(), snipParticle) {
		t.Fatal("snip burst has no hair color")
	}
}
