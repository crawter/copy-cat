package render

import (
	"reflect"
	"testing"

	"copycat/internal/campaign"
)

var heroPoses = []HeroPose{
	PoseIdle, PoseRun, PoseJump, PoseSlide, PoseSlam,
	PoseLanding, PoseStomp, PoseSprint,
}

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

func TestBuildHeroUsesRoundedParts(t *testing.T) {
	for _, pose := range heroPoses {
		cs := BuildHero(pose, false)
		if cmdsHaveShape(cs, ShapeBox) {
			t.Fatalf("pose %v: rounded hero still uses a box", pose)
		}
		for _, want := range []Shape{ShapeCylinder, ShapeSphere, ShapeCone} {
			if !cmdsHaveShape(cs, want) {
				t.Fatalf("pose %v: rounded hero missing shape %v", pose, want)
			}
		}
	}
}

func TestBuildHeroAtAnimatesRun(t *testing.T) {
	a := BuildHeroAt(PoseRun, false, 0)
	b := BuildHeroAt(PoseRun, false, 0.125)
	c := BuildHeroAt(PoseRun, false, 0.25)
	if reflect.DeepEqual(a, b) || reflect.DeepEqual(a, c) || reflect.DeepEqual(b, c) {
		t.Fatal("run pose does not animate over time")
	}
	if !reflect.DeepEqual(a, BuildHero(PoseRun, false)) {
		t.Fatal("BuildHero does not match the first animation frame")
	}
	// Animation time wraps like a phase: a whole cycle repeats.
	if !reflect.DeepEqual(b, BuildHeroAt(PoseRun, false, 1.125)) {
		t.Fatal("animation time does not wrap to the phase")
	}
}

func TestNewHeroPosesDifferAndKeepParts(t *testing.T) {
	newPoses := []HeroPose{PoseLanding, PoseStomp, PoseSprint}
	models := make(map[HeroPose][]Cmd, len(newPoses))
	for _, pose := range newPoses {
		cs := BuildHero(pose, false)
		models[pose] = cs
		for _, part := range heroParts {
			if !cmdsHaveColor(cs, part.color) {
				t.Fatalf("pose %v: %s part missing", pose, part.name)
			}
		}
	}
	for i := 0; i < len(newPoses); i++ {
		for j := i + 1; j < len(newPoses); j++ {
			if reflect.DeepEqual(models[newPoses[i]], models[newPoses[j]]) {
				t.Fatalf("poses %v and %v produce identical models", newPoses[i], newPoses[j])
			}
		}
	}
	for _, pose := range newPoses {
		for _, existing := range heroPoses {
			if existing == pose {
				continue
			}
			if reflect.DeepEqual(models[pose], BuildHero(existing, false)) {
				t.Fatalf("pose %v produces the same model as pose %v", pose, existing)
			}
		}
	}
}

func TestBuildHeroAtAnimatesIdleAndSprint(t *testing.T) {
	for _, pose := range []HeroPose{PoseIdle, PoseSprint} {
		a := BuildHeroAt(pose, false, 0)
		b := BuildHeroAt(pose, false, 0.125)
		c := BuildHeroAt(pose, false, 0.25)
		if reflect.DeepEqual(a, b) || reflect.DeepEqual(a, c) || reflect.DeepEqual(b, c) {
			t.Fatalf("pose %v does not animate over time", pose)
		}
		if !reflect.DeepEqual(a, BuildHero(pose, false)) {
			t.Fatalf("BuildHero does not match the first animation frame for %v", pose)
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

func TestBuildCopycatWindupDiffersFromIdleAndAnimates(t *testing.T) {
	idle := BuildCopycat(2)
	if !reflect.DeepEqual(idle, BuildCopycatAt(2, CopycatPoseIdle, 0)) {
		t.Fatal("idle CopycatAt does not match BuildCopycat")
	}

	windup := BuildCopycatAt(2, CopycatPoseWindup, 0)
	if reflect.DeepEqual(idle, windup) {
		t.Fatal("wind-up Copycat stance matches idle")
	}
	peak := BuildCopycatAt(2, CopycatPoseWindup, 0.5)
	if reflect.DeepEqual(windup, peak) {
		t.Fatal("wind-up Copycat does not animate over time")
	}
	if !reflect.DeepEqual(BuildCopycatAt(2, CopycatPoseWindup, 1), windup) {
		t.Fatal("wind-up phase does not wrap")
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

func TestBuildBossPosesDifferFromIdleAndAnimate(t *testing.T) {
	idle := BuildBoss(1, false)
	if !reflect.DeepEqual(idle, BuildBossAt(1, false, BossStanceIdle, 0)) {
		t.Fatal("idle BuildBossAt does not match BuildBoss")
	}

	windup := BuildBossAt(1, false, BossStanceWindup, 0)
	jump := BuildBossAt(1, false, BossStanceJump, 0)
	if reflect.DeepEqual(idle, windup) {
		t.Fatal("boss wind-up stance matches idle")
	}
	if reflect.DeepEqual(idle, jump) {
		t.Fatal("boss jump stance matches idle")
	}
	if reflect.DeepEqual(windup, jump) {
		t.Fatal("boss wind-up and jump stances are identical")
	}

	windupLater := BuildBossAt(1, false, BossStanceWindup, 0.25)
	jumpLater := BuildBossAt(1, false, BossStanceJump, 0.25)
	if reflect.DeepEqual(windup, windupLater) {
		t.Fatal("boss wind-up stance does not animate over time")
	}
	if reflect.DeepEqual(jump, jumpLater) {
		t.Fatal("boss jump stance does not animate over time")
	}
}

func TestModelsUseOnlyPrimitiveShapes(t *testing.T) {
	lists := [][]Cmd{
		BuildHero(PoseRun, false),
		BuildHero(PoseSlide, true),
		BuildHeroAt(PoseRun, false, 0.25),
		BuildHeroAt(PoseIdle, false, 0.25),
		BuildHeroAt(PoseSprint, false, 0.25),
		BuildEnemy(campaign.ThemeMangroves, Vec3{Z: 0}),
		BuildEnemy(campaign.ThemeHarbor, Vec3{Z: 0}),
		BuildEnemy(campaign.ThemeHarbor, Vec3{Z: 1}),
		BuildEnemy(campaign.ThemeNeon, Vec3{Z: 0}),
		BuildEnemy(campaign.ThemeNeon, Vec3{Z: 1}),
		BuildCopycat(3),
		BuildCopycatAt(3, CopycatPoseWindup, 0.25),
		BuildBoss(1, false),
		BuildBoss(3, true),
		BuildBossAt(1, false, BossStanceWindup, 0.25),
		BuildBossAt(1, false, BossStanceJump, 0.25),
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

func cmdsHaveShape(cs []Cmd, shape Shape) bool {
	for _, c := range cs {
		if c.Shape == shape {
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
