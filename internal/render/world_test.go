package render

import (
	"math"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/levelgen"
)

func TestPalettesDiffer(t *testing.T) {
	mangroves := Palette(campaign.ThemeMangroves)
	harbor := Palette(campaign.ThemeHarbor)
	neon := Palette(campaign.ThemeNeon)
	if mangroves == harbor || mangroves == neon || harbor == neon {
		t.Fatal("theme palettes are not distinct")
	}
}

func TestBuildWorldIncludesCorridor(t *testing.T) {
	l := levelgen.New(1, levelgen.Mangroves())
	cs := BuildWorld(l, campaign.ThemeMangroves, Cam{Lane: 1, Z: 40, Height: 0})

	seenLanes := map[int]bool{}
	walls := 0
	for _, c := range cs {
		if c.Shape == ShapeBox && math.Abs(c.Size.Y-floorThickness) < 0.01 {
			for lane := 0; lane < 3; lane++ {
				if math.Abs(float64(c.Pos.X)-laneX(float64(lane))) < 0.01 {
					seenLanes[lane] = true
				}
			}
		}
		if c.Shape == ShapeBox && math.Abs(c.Size.Y-wallHeight) < 0.01 {
			walls++
		}
	}
	for lane := 0; lane < 3; lane++ {
		if !seenLanes[lane] {
			t.Fatalf("lane %d has no floor strip", lane)
		}
	}
	if walls < 2 {
		t.Fatalf("want 2 walls, got %d", walls)
	}
}

func TestElementKindsProduceCommands(t *testing.T) {
	pal := Palette(campaign.ThemeMangroves)
	cases := []struct {
		name  string
		kind  levelgen.Kind
		check func(t *testing.T, cs []Cmd, e levelgen.Element)
	}{
		{
			name: "pit",
			kind: levelgen.Pit,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasHazardCmd(cs, pal, e) {
					t.Fatalf("no hazard command for pit %+v", e)
				}
			},
		},
		{
			name: "low barrier",
			kind: levelgen.LowBarrier,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasLowBarrierCmd(cs, e) {
					t.Fatalf("no command for low barrier %+v", e)
				}
			},
		},
		{
			name: "high barrier",
			kind: levelgen.HighBarrier,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasHighBarrierCmd(cs, e) {
					t.Fatalf("no command for high barrier %+v", e)
				}
			},
		},
		{
			name: "ramp",
			kind: levelgen.Ramp,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasContainerCmd(cs, pal, e) {
					t.Fatalf("no container command for ramp %+v", e)
				}
			},
		},
		{
			name: "enemy",
			kind: levelgen.Enemy,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasEnemyCmd(cs, e) {
					t.Fatalf("no command for enemy %+v", e)
				}
			},
		},
		{
			name: "comb",
			kind: levelgen.Comb,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasCombCmd(cs, e) {
					t.Fatalf("no command for comb %+v", e)
				}
			},
		},
		{
			name: "hairspray",
			kind: levelgen.Hairspray,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasHairsprayCmd(cs, e) {
					t.Fatalf("no command for hairspray %+v", e)
				}
			},
		},
		{
			name: "mirror",
			kind: levelgen.Mirror,
			check: func(t *testing.T, cs []Cmd, e levelgen.Element) {
				if !hasMirrorCmd(cs, e) {
					t.Fatalf("no command for mirror %+v", e)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, e := findElement(tc.kind)
			if l == nil {
				t.Fatalf("no %s element found in any seeded level", tc.name)
			}
			cs := BuildWorld(l, campaign.ThemeMangroves, Cam{Lane: float64(e.Lane), Z: e.Z, Height: 0})
			tc.check(t, cs, e)
		})
	}
}

func TestPitLeavesFloorGap(t *testing.T) {
	l, pit := findElementBetween(levelgen.Pit, 20, math.Inf(1))
	if l == nil {
		t.Fatal("no pit found")
	}
	pal := Palette(campaign.ThemeMangroves)
	cs := BuildWorld(l, campaign.ThemeMangroves, Cam{Lane: 1, Z: pit.Z, Height: 0})

	var before, after bool
	for _, c := range floorCmds(cs) {
		z0 := float64(c.Pos.Z) - c.Size.Z/2
		z1 := float64(c.Pos.Z) + c.Size.Z/2
		if overlapsInterval(z0, z1, pit.Z, pit.Z+pit.Span) {
			t.Fatalf("floor command %+v overlaps pit [%v, %v)", c, pit.Z, pit.Z+pit.Span)
		}
		if z1 <= pit.Z+0.01 {
			before = true
		}
		if z0 >= pit.Z+pit.Span-0.01 {
			after = true
		}
	}
	if !before || !after {
		t.Fatalf("pit %+v does not leave a floor gap: before=%v after=%v", pit, before, after)
	}
	if !hasHazardCmd(cs, pal, pit) {
		t.Fatalf("pit %+v has no themed hazard", pit)
	}
}

func TestRampAddsContainer(t *testing.T) {
	l, ramp := findElement(levelgen.Ramp)
	if l == nil {
		t.Fatal("no ramp found")
	}
	pal := Palette(campaign.ThemeMangroves)
	cs := BuildWorld(l, campaign.ThemeMangroves, Cam{Lane: float64(ramp.Lane), Z: ramp.Z, Height: 0})
	if !hasContainerCmd(cs, pal, ramp) {
		t.Fatalf("ramp %+v did not add its container", ramp)
	}
}

func TestPickupsKeepDistinctShapes(t *testing.T) {
	sets := map[levelgen.Kind]map[Shape]bool{}
	for _, kind := range []levelgen.Kind{levelgen.Comb, levelgen.Hairspray, levelgen.Mirror} {
		l, e := findElement(kind)
		if l == nil {
			t.Fatalf("no %v element found", kind)
		}
		cs := BuildWorld(l, campaign.ThemeMangroves, Cam{Lane: float64(e.Lane), Z: e.Z, Height: 0})
		sets[kind] = shapesNearZ(cs, e)
	}
	comb, hair, mirror := sets[levelgen.Comb], sets[levelgen.Hairspray], sets[levelgen.Mirror]
	if len(comb) == 0 || len(hair) == 0 || len(mirror) == 0 {
		t.Fatalf("pickup produced no shapes: comb=%v hairspray=%v mirror=%v", comb, hair, mirror)
	}
	if shapeSetEqual(comb, hair) || shapeSetEqual(comb, mirror) || shapeSetEqual(hair, mirror) {
		t.Fatalf("pickup shape sets are not distinct: comb=%v hairspray=%v mirror=%v", comb, hair, mirror)
	}
}

func TestBuildWorldCullsByDistance(t *testing.T) {
	l, near, farAhead, farBehind := findBarrierSpread()
	if l == nil {
		t.Skip("no seed produced near, far-ahead and far-behind low barriers")
	}
	camZ := l.Length() / 2
	cs := BuildWorld(l, campaign.ThemeMangroves, Cam{Lane: 1, Z: camZ, Height: 0})
	if !hasLowBarrierCmd(cs, near) {
		t.Fatalf("near barrier %+v was culled", near)
	}
	if hasLowBarrierCmd(cs, farAhead) {
		t.Fatalf("far-ahead barrier %+v was not culled", farAhead)
	}
	if hasLowBarrierCmd(cs, farBehind) {
		t.Fatalf("far-behind barrier %+v was not culled", farBehind)
	}
}

func TestBuildBossArena(t *testing.T) {
	cs := BuildBossArena(Cam{Lane: 1, Z: 0, Height: 0})
	if len(cs) == 0 {
		t.Fatal("arena returned no commands")
	}
	var hasChecker, hasChair, hasMirror bool
	for _, c := range cs {
		if c.Pos.Y < 0 {
			t.Fatalf("arena touches the corridor floor: %+v", c)
		}
		if c.Shape == ShapePlane {
			hasChecker = true
		}
		if c.Shape == ShapeBox && c.Color == bossChair {
			hasChair = true
		}
		if c.Shape == ShapeBox && c.Color == bossMirror {
			hasMirror = true
		}
	}
	if !hasChecker {
		t.Fatal("arena has no checkered floor")
	}
	if !hasChair {
		t.Fatal("arena has no barber chairs")
	}
	if !hasMirror {
		t.Fatal("arena has no mirrors")
	}
}

func findElement(kind levelgen.Kind) (*levelgen.Level, levelgen.Element) {
	return findElementBetween(kind, 0, math.Inf(1))
}

func findElementBetween(kind levelgen.Kind, minZ, maxZ float64) (*levelgen.Level, levelgen.Element) {
	for seed := uint64(1); seed <= 500; seed++ {
		l := levelgen.New(seed, levelgen.Mangroves())
		for _, e := range l.Elements(0, l.Length()) {
			if e.Kind == kind && e.Z >= minZ && e.Z < maxZ {
				return l, e
			}
		}
	}
	return nil, levelgen.Element{}
}

func findBarrierSpread() (*levelgen.Level, levelgen.Element, levelgen.Element, levelgen.Element) {
	for seed := uint64(1); seed <= 500; seed++ {
		l := levelgen.New(seed, levelgen.Mangroves())
		camZ := l.Length() / 2
		start, end := camWindow(camZ)
		var near, farAhead, farBehind levelgen.Element
		var foundNear, foundAhead, foundBehind bool
		for _, e := range l.Elements(0, l.Length()) {
			if e.Kind != levelgen.LowBarrier {
				continue
			}
			s, t := spanOf(e)
			switch {
			case !foundNear && s >= start+2 && t <= end-2:
				near, foundNear = e, true
			case !foundAhead && s > end+2:
				farAhead, foundAhead = e, true
			case !foundBehind && t < start-2:
				farBehind, foundBehind = e, true
			}
		}
		if foundNear && foundAhead && foundBehind {
			return l, near, farAhead, farBehind
		}
	}
	return nil, levelgen.Element{}, levelgen.Element{}, levelgen.Element{}
}

func floorCmds(cs []Cmd) []Cmd {
	var out []Cmd
	for _, c := range cs {
		if c.Shape == ShapeBox && math.Abs(c.Size.Y-floorThickness) < 0.01 {
			out = append(out, c)
		}
	}
	return out
}

func shapesNearZ(cs []Cmd, e levelgen.Element) map[Shape]bool {
	x := laneX(float64(e.Lane))
	out := map[Shape]bool{}
	for _, c := range cs {
		if c.Shape == ShapeBox && math.Abs(c.Size.Y-floorThickness) < 0.01 {
			continue
		}
		if math.Abs(float64(c.Pos.Z)-e.Z) > 0.25 {
			continue
		}
		if math.Abs(float64(c.Pos.X)-x) > 0.8 {
			continue
		}
		out[c.Shape] = true
	}
	return out
}

func shapeSetEqual(a, b map[Shape]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func overlapsInterval(a0, a1, b0, b1 float64) bool {
	return a0 < b1 && b0 < a1
}

func hasHazardCmd(cs []Cmd, pal ThemePalette, e levelgen.Element) bool {
	z := e.Z + e.Span/2
	for _, c := range cs {
		if c.Shape != ShapeBox || c.Color != pal.Hazard {
			continue
		}
		if math.Abs(c.Size.Y-pitDepth) > 0.01 {
			continue
		}
		if math.Abs(float64(c.Pos.Z)-z) > 0.6 {
			continue
		}
		return true
	}
	return false
}

func hasLowBarrierCmd(cs []Cmd, e levelgen.Element) bool {
	z := e.Z + barrierDepth/2
	x := laneX(float64(e.Lane))
	for _, c := range cs {
		if c.Shape != ShapeBox || math.Abs(c.Size.Y-lowBarrierH) > 0.01 {
			continue
		}
		if math.Abs(float64(c.Pos.Z)-z) > 0.6 || math.Abs(float64(c.Pos.X)-x) > 0.6 {
			continue
		}
		return true
	}
	return false
}

func hasHighBarrierCmd(cs []Cmd, e levelgen.Element) bool {
	z := e.Z + barrierDepth/2
	x := laneX(float64(e.Lane))
	for _, c := range cs {
		if c.Shape != ShapeBox || c.Size.Y <= 0.5 || float64(c.Pos.Y) < highBarrierGap {
			continue
		}
		if math.Abs(float64(c.Pos.Z)-z) > 0.6 || math.Abs(float64(c.Pos.X)-x) > 0.6 {
			continue
		}
		return true
	}
	return false
}

func hasContainerCmd(cs []Cmd, pal ThemePalette, e levelgen.Element) bool {
	z := e.Z + e.Span/2
	x := laneX(float64(e.Lane))
	for _, c := range cs {
		if c.Shape != ShapeBox || c.Color != pal.Ramp {
			continue
		}
		if math.Abs(c.Size.Y-containerH) > 0.01 || math.Abs(c.Size.Z-e.Span) > 0.01 {
			continue
		}
		if math.Abs(float64(c.Pos.Z)-z) > 0.6 || math.Abs(float64(c.Pos.X)-x) > 0.6 {
			continue
		}
		return true
	}
	return false
}

func hasEnemyCmd(cs []Cmd, e levelgen.Element) bool {
	z := e.Z + barrierDepth/2
	x := laneX(float64(e.Lane))
	for _, c := range cs {
		if c.Shape != ShapeSphere {
			continue
		}
		if math.Abs(float64(c.Pos.Z)-z) > 0.6 || math.Abs(float64(c.Pos.X)-x) > 0.6 {
			continue
		}
		return true
	}
	return false
}

func hasCombCmd(cs []Cmd, e levelgen.Element) bool {
	return shapeSetEqual(shapesNearZ(cs, e), map[Shape]bool{ShapeBox: true})
}

func hasHairsprayCmd(cs []Cmd, e levelgen.Element) bool {
	return shapeSetEqual(shapesNearZ(cs, e), map[Shape]bool{ShapeCylinder: true, ShapeSphere: true})
}

func hasMirrorCmd(cs []Cmd, e levelgen.Element) bool {
	return shapeSetEqual(shapesNearZ(cs, e), map[Shape]bool{ShapeBox: true, ShapeCylinder: true})
}
