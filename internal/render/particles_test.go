package render

import (
	"reflect"
	"testing"
)

func TestBurstsArePrimitiveAndDistinct(t *testing.T) {
	bursts := map[string][]Cmd{
		"comb":       BuildCombBurst(),
		"landing":    BuildLandingBurst(),
		"snip":       BuildSnipBurst(),
		"stomp":      BuildStompBurst(),
		"checkpoint": BuildCheckpointShine(),
		"boss":       BuildBossHitBurst(),
		"slide":      BuildSlideSparks(),
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

	names := []string{"comb", "landing", "snip", "stomp", "checkpoint", "boss", "slide"}
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
	if !cmdsHaveColor(BuildStompBurst(), stompParticle) {
		t.Fatal("stomp burst has no stomp particle color")
	}
	if !cmdsHaveColor(BuildCheckpointShine(), checkpointShine) {
		t.Fatal("checkpoint shine has no shine color")
	}
	if !cmdsHaveColor(BuildBossHitBurst(), bossHitParticle) {
		t.Fatal("boss hit burst has no hit color")
	}
	if !cmdsHaveColor(BuildSlideSparks(), slideParticle) {
		t.Fatal("slide sparks have no spark color")
	}
}

func TestNewBurstPositionsMatchKind(t *testing.T) {
	cases := []struct {
		name     string
		cs       []Cmd
		yLo, yHi float64
		zLo, zHi float64
	}{
		{name: "stomp", cs: BuildStompBurst(), yLo: 0, yHi: 0.25},
		{name: "checkpoint", cs: BuildCheckpointShine(), yLo: 1.0, yHi: 2.0},
		{name: "boss", cs: BuildBossHitBurst(), yLo: 0.8, yHi: 2.0},
		{name: "slide", cs: BuildSlideSparks(), yLo: 0, yHi: 0.4, zLo: -2, zHi: -0.1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !cmdsHaveY(tc.cs, tc.yLo, tc.yHi) {
				t.Fatalf("%s burst has no command with Y in [%v, %v]", tc.name, tc.yLo, tc.yHi)
			}
			if tc.zLo != 0 || tc.zHi != 0 {
				if !cmdsHaveZ(tc.cs, tc.zLo, tc.zHi) {
					t.Fatalf("%s burst has no command with Z in [%v, %v]", tc.name, tc.zLo, tc.zHi)
				}
			}
		})
	}
}

func cmdsHaveY(cs []Cmd, lo, hi float64) bool {
	for _, c := range cs {
		if float64(c.Pos.Y) >= lo && float64(c.Pos.Y) <= hi {
			return true
		}
	}
	return false
}

func cmdsHaveZ(cs []Cmd, lo, hi float64) bool {
	for _, c := range cs {
		if float64(c.Pos.Z) >= lo && float64(c.Pos.Z) <= hi {
			return true
		}
	}
	return false
}
