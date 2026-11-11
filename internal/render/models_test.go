package render

import (
	"reflect"
	"testing"

	"copycat/internal/campaign"
)

var heroPoses = []HeroPose{PoseIdle, PoseRun, PoseJump, PoseSlide, PoseSlam}

var heroParts = []struct {
	name  string
	color Color
}{
	{"body", heroBody},
	{"belly", heroBelly},
	{"limbs", heroLimb},
	{"tail", heroTail},
	{"nose", heroNose},
	{"hair", heroHair},
}

func TestBuildHeroHoldsAllParts(t *testing.T) {
	for _, pose := range heroPoses {
		cs := BuildHero(pose, false)
		for _, part := range heroParts {
			if !cmdsHaveColor(cs, part.color) {
				t.Fatalf("pose %v: %s part missing", pose, part.name)
			}
		}
	}
}

func TestBuildHeroBaldDropsOnlyHair(t *testing.T) {
	for _, pose := range heroPoses {
		with := BuildHero(pose, false)
		bald := BuildHero(pose, true)
		kept := dropColor(with, heroHair)
		if !reflect.DeepEqual(kept, bald) {
			t.Fatalf("pose %v: bald hero changes more than the hair", pose)
		}
		if cmdsHaveColor(bald, heroHair) {
			t.Fatalf("pose %v: bald hero still has hair", pose)
		}
		for _, part := range heroParts {
			if part.color == heroHair {
				continue
			}
			if !cmdsHaveColor(bald, part.color) {
				t.Fatalf("pose %v: bald hero lost %s part", pose, part.name)
			}
		}
	}
}

func TestHeroPosesDiffer(t *testing.T) {
	lists := make(map[HeroPose][]Cmd, len(heroPoses))
	for _, pose := range heroPoses {
		lists[pose] = BuildHero(pose, false)
	}
	for i := 0; i < len(heroPoses); i++ {
		for j := i + 1; j < len(heroPoses); j++ {
			a, b := heroPoses[i], heroPoses[j]
			if reflect.DeepEqual(lists[a], lists[b]) {
				t.Fatalf("poses %v and %v produce identical models", a, b)
			}
		}
	}
}

func TestBuildHeroIsBuildHeroAtRest(t *testing.T) {
	for _, pose := range heroPoses {
		for _, bald := range []bool{false, true} {
			if got, want := BuildHero(pose, bald), BuildHeroAt(pose, bald, 0); !reflect.DeepEqual(got, want) {
				t.Fatalf("pose %v bald %v: BuildHero != BuildHeroAt at t=0", pose, bald)
			}
		}
	}
}

func TestBuildHeroAtAnimatesWithTime(t *testing.T) {
	for _, pose := range heroPoses {
		a := BuildHeroAt(pose, false, 0)
		b := BuildHeroAt(pose, false, 0.31)
		if reflect.DeepEqual(a, b) {
			t.Fatalf("pose %v: t does not animate the model", pose)
		}
	}

	// Pin each animation seam: the run-cycle leg swing, the nose wobble and
	// the hair flutter must each move with t, not merely some other part.
	restLegs := legZs(BuildHeroAt(PoseRun, false, 0))
	movedLegs := legZs(BuildHeroAt(PoseRun, false, 0.15))
	if reflect.DeepEqual(restLegs, movedLegs) {
		t.Fatal("run cycle does not swing the legs with t")
	}
	if minF32(movedLegs) >= -0.1 || maxF32(movedLegs) <= 0.1 {
		t.Fatalf("run-cycle leg swing is too small: %v", movedLegs)
	}

	restNose := noseXs(BuildHeroAt(PoseIdle, false, 0))
	movedNose := noseXs(BuildHeroAt(PoseIdle, false, 0.18))
	if reflect.DeepEqual(restNose, movedNose) {
		t.Fatal("nose does not wobble with t")
	}

	restHair := hairYs(BuildHeroAt(PoseIdle, false, 0))
	movedHair := hairYs(BuildHeroAt(PoseIdle, false, 0.15))
	if reflect.DeepEqual(restHair, movedHair) {
		t.Fatal("hair does not flutter with t")
	}
}

func TestBuildHeroAtBaldDropsOnlyHair(t *testing.T) {
	for _, pose := range heroPoses {
		with := BuildHeroAt(pose, false, 0.42)
		bald := BuildHeroAt(pose, true, 0.42)
		if got, want := dropColor(with, heroHair), bald; !reflect.DeepEqual(got, want) {
			t.Fatalf("pose %v: bald hero changes more than the hair", pose)
		}
		if cmdsHaveColor(bald, heroHair) {
			t.Fatalf("pose %v: bald hero still has hair", pose)
		}
		for _, part := range heroParts {
			if part.color == heroHair {
				continue
			}
			if !cmdsHaveColor(bald, part.color) {
				t.Fatalf("pose %v: bald hero lost %s part", pose, part.name)
			}
		}
	}
}

func TestHeroNoseAndQuiffReadDirections(t *testing.T) {
	cs := BuildHeroAt(PoseIdle, false, 0)
	if got, want := maxZOfColor(cs, heroNose), maxZOfColor(cs, heroBody); got <= want {
		t.Fatalf("nose z = %v, not in front of body z = %v", got, want)
	}
	if got, want := minZOfColor(cs, heroHair), minZOfColor(cs, heroBody); got >= want {
		t.Fatalf("quiff z = %v, not behind body z = %v", got, want)
	}
}

func TestBuildHeroAtFacingNegatesZ(t *testing.T) {
	for _, pose := range heroPoses {
		orig := BuildHeroAt(pose, false, 0.37)
		facing := BuildHeroAtFacing(pose, false, 0.37)
		if got, want := facing, negateAllZ(orig); !reflect.DeepEqual(got, want) {
			t.Fatalf("pose %v: facing variant is not the Z-negated model", pose)
		}
	}
}

func TestHeroUsesOnlyRoundedShapes(t *testing.T) {
	for _, pose := range heroPoses {
		for _, c := range BuildHeroAt(pose, false, 0.7) {
			if c.Shape == ShapeBox || c.Shape == ShapePlane {
				t.Fatalf("pose %v: rounded hero uses shape %v", pose, c.Shape)
			}
		}
	}
}

func TestBuildEnemyCoversAllSpecies(t *testing.T) {
	cases := []struct {
		theme campaign.Theme
		pos   Vec3
		color Color
		name  string
	}{
		{campaign.ThemeMangroves, Vec3{Z: 0}, mudskipperBody, "mudskipper"},
		{campaign.ThemeMangroves, Vec3{Z: 5}, mudskipperBody, "mudskipper"},
		{campaign.ThemeHarbor, Vec3{Z: 0}, gullBody, "gull"},
		{campaign.ThemeHarbor, Vec3{Z: 1}, ratFur, "rat"},
		{campaign.ThemeNeon, Vec3{Z: 0}, chairSeat, "rolling barber chair"},
		{campaign.ThemeNeon, Vec3{Z: 1}, vacuumBody, "robot vacuum"},
	}
	for _, tc := range cases {
		cs := BuildEnemy(tc.theme, tc.pos)
		if len(cs) == 0 {
			t.Fatalf("%s: no commands", tc.name)
		}
		if !cmdsHaveColor(cs, tc.color) {
			t.Fatalf("%s: species color missing", tc.name)
		}
	}
}

func TestBuildEnemyUnknownThemeFallsBack(t *testing.T) {
	cs := BuildEnemy(campaign.Theme(99), Vec3{Z: 3})
	if !cmdsHaveColor(cs, mudskipperBody) {
		t.Fatal("unknown theme did not fall back to the mudskipper")
	}
}

func TestBuildCopycatIncludesScissors(t *testing.T) {
	cs := BuildCopycat(2)
	if !cmdsHaveColor(cs, catScissor) {
		t.Fatal("Copycat has no scissors")
	}
}

func TestBuildCopycatChaseDistanceMovesBehind(t *testing.T) {
	near := BuildCopycat(1)
	far := BuildCopycat(5)
	if meanZ(near) <= meanZ(far) {
		t.Fatalf("larger chase distance did not move Copycat behind: near=%v far=%v", meanZ(near), meanZ(far))
	}
}

func TestBuildBossDiffersAcrossPhasesAndVulnerability(t *testing.T) {
	p1 := BuildBoss(1, false)
	p2 := BuildBoss(2, false)
	p3 := BuildBoss(3, false)
	v1 := BuildBoss(1, true)
	if reflect.DeepEqual(p1, p2) || reflect.DeepEqual(p1, p3) || reflect.DeepEqual(p2, p3) {
		t.Fatal("boss stances do not change across phases")
	}
	if reflect.DeepEqual(p1, v1) {
		t.Fatal("vulnerable boss stance does not differ")
	}
	if minCmdY(v1) >= minCmdY(p1) {
		t.Fatal("vulnerable boss does not crouch")
	}
	if !cmdsHaveColor(p1, bossAnger) {
		t.Fatal("boss has no anger accent")
	}
	if !cmdsHaveColor(p1, catScissor) {
		t.Fatal("boss has no scissors")
	}
}

func TestModelsUseOnlyPrimitiveShapes(t *testing.T) {
	lists := [][]Cmd{
		BuildHero(PoseRun, false),
		BuildHero(PoseSlide, true),
		BuildEnemy(campaign.ThemeMangroves, Vec3{Z: 0}),
		BuildEnemy(campaign.ThemeHarbor, Vec3{Z: 0}),
		BuildEnemy(campaign.ThemeHarbor, Vec3{Z: 1}),
		BuildEnemy(campaign.ThemeNeon, Vec3{Z: 0}),
		BuildEnemy(campaign.ThemeNeon, Vec3{Z: 1}),
		BuildCopycat(3),
		BuildBoss(1, false),
		BuildBoss(3, true),
	}
	for i, cs := range lists {
		if len(cs) == 0 {
			t.Fatalf("model %d returned no commands", i)
		}
		for _, c := range cs {
			if c.Shape < ShapeBox || c.Shape > ShapePlane {
				t.Fatalf("model %d uses a non-primitive shape %v", i, c.Shape)
			}
		}
	}
}

func cmdsHaveColor(cs []Cmd, col Color) bool {
	for _, c := range cs {
		if c.Color == col {
			return true
		}
	}
	return false
}

func dropColor(cs []Cmd, col Color) []Cmd {
	out := make([]Cmd, 0, len(cs))
	for _, c := range cs {
		if c.Color != col {
			out = append(out, c)
		}
	}
	return out
}

func meanZ(cs []Cmd) float64 {
	if len(cs) == 0 {
		return 0
	}
	var sum float64
	for _, c := range cs {
		sum += float64(c.Pos.Z)
	}
	return sum / float64(len(cs))
}

func minCmdY(cs []Cmd) float32 {
	if len(cs) == 0 {
		return 0
	}
	m := cs[0].Pos.Y
	for _, c := range cs[1:] {
		if c.Pos.Y < m {
			m = c.Pos.Y
		}
	}
	return m
}

func maxZOfColor(cs []Cmd, col Color) float32 {
	m := float32(0)
	found := false
	for _, c := range cs {
		if c.Color != col {
			continue
		}
		if !found || c.Pos.Z > m {
			m = c.Pos.Z
			found = true
		}
	}
	return m
}

func minZOfColor(cs []Cmd, col Color) float32 {
	m := float32(0)
	found := false
	for _, c := range cs {
		if c.Color != col {
			continue
		}
		if !found || c.Pos.Z < m {
			m = c.Pos.Z
			found = true
		}
	}
	return m
}

func negateAllZ(cs []Cmd) []Cmd {
	out := make([]Cmd, len(cs))
	copy(out, cs)
	for i := range out {
		out[i].Pos.Z = -out[i].Pos.Z
	}
	return out
}

// legZs returns the Z positions of the hero's leg commands: the gray limb
// commands below the shoulder line. Arms hang at and above heroArmY, so the
// y < 0.5 threshold picks out the legs alone.
func legZs(cs []Cmd) []float32 {
	var zs []float32
	for _, c := range cs {
		if c.Color == heroLimb && c.Pos.Y < 0.5 {
			zs = append(zs, c.Pos.Z)
		}
	}
	return zs
}

// noseXs returns the X positions of the nose commands, whose side-to-side
// wobble moves them along X.
func noseXs(cs []Cmd) []float32 {
	var xs []float32
	for _, c := range cs {
		if c.Color == heroNose {
			xs = append(xs, c.Pos.X)
		}
	}
	return xs
}

// hairYs returns the Y positions of the hair commands, whose flutter moves
// them up and down.
func hairYs(cs []Cmd) []float32 {
	var ys []float32
	for _, c := range cs {
		if c.Color == heroHair {
			ys = append(ys, c.Pos.Y)
		}
	}
	return ys
}

func minF32(xs []float32) float32 {
	if len(xs) == 0 {
		return 0
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

func maxF32(xs []float32) float32 {
	if len(xs) == 0 {
		return 0
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}
