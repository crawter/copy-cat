package render

import (
	"math"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

func TestHeroPoseForJumpSlideSlamRun(t *testing.T) {
	t.Run("jump", func(t *testing.T) {
		r := newPoseRun(t)
		r.Step(runner.Input{Jump: true}, testDT)
		if got := heroPoseFor(r); got != PoseJump {
			t.Fatalf("jump pose = %v, want PoseJump", got)
		}
	})
	t.Run("slide", func(t *testing.T) {
		r := newPoseRun(t)
		r.Step(runner.Input{Slide: true}, testDT)
		if got := heroPoseFor(r); got != PoseSlide {
			t.Fatalf("slide pose = %v, want PoseSlide", got)
		}
	})
	t.Run("slam", func(t *testing.T) {
		r := newPoseRun(t)
		r.Step(runner.Input{Jump: true}, testDT)

		// Rise to near the jump apex before slamming: a slam from ground
		// height lands within a single step, so it would be over before the
		// pose is read.
		cfg := runner.DefaultConfig()
		for i := 0; i < int(math.Ceil(cfg.JumpVelocity/cfg.Gravity/testDT)); i++ {
			r.Step(runner.Input{}, testDT)
		}

		r.Step(runner.Input{Slide: true}, testDT)
		if got := heroPoseFor(r); got != PoseSlam {
			t.Fatalf("slam pose = %v, want PoseSlam", got)
		}
	})
	t.Run("run", func(t *testing.T) {
		r := newPoseRun(t)
		if got := heroPoseFor(r); got != PoseRun {
			t.Fatalf("grounded pose = %v, want PoseRun", got)
		}
	})
}

func TestHeroPoseForContainerRideIsRun(t *testing.T) {
	l := findRampLevel(t)
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = 8
	r := gameplay.NewRun(runner.New(l, cfg), l, 3)

	for i := 0; i < 4000 && !r.Runner.OnContainer(); i++ {
		r.Step(runner.Input{}, testDT)
	}
	if !r.Runner.OnContainer() {
		t.Fatal("never reached the container ramp")
	}
	if r.Runner.Height() <= 0.05 {
		t.Fatalf("on container but Height() = %v, want raised", r.Runner.Height())
	}
	if r.Runner.Airborne() {
		t.Fatal("container ride reports airborne")
	}
	if got := heroPoseFor(r); got != PoseRun {
		t.Fatalf("container ride pose = %v, want PoseRun (not PoseJump)", got)
	}
}

func TestCopycatChaseDistance(t *testing.T) {
	if got := copycatChase(1); math.Abs(got-2) > 1e-9 {
		t.Fatalf("copycatChase(1) = %v, want about 2", got)
	}
	if got := copycatChase(0); math.Abs(got-9) > 1e-9 {
		t.Fatalf("copycatChase(0) = %v, want about 9", got)
	}
	if copycatChase(1) >= copycatChase(0) {
		t.Fatal("chase distance is not larger when Copycat is far")
	}
	if got := copycatChase(2); got != copycatChase(1) {
		t.Fatalf("chase above 1 is not clamped: %v", got)
	}
	if got := copycatChase(-1); got != copycatChase(0) {
		t.Fatalf("chase below 0 is not clamped: %v", got)
	}
}

func TestLerpColor(t *testing.T) {
	a := Color{R: 0, G: 100, B: 200, A: 255}
	b := Color{R: 200, G: 0, B: 0, A: 0}
	if got := LerpColor(a, b, 0); got != a {
		t.Fatalf("LerpColor t=0 = %+v, want %+v", got, a)
	}
	if got := LerpColor(a, b, 1); got != b {
		t.Fatalf("LerpColor t=1 = %+v, want %+v", got, b)
	}
	mid := LerpColor(a, b, 0.5)
	if mid.R != 100 || mid.G != 50 || mid.B != 100 || mid.A != 128 {
		t.Fatalf("LerpColor t=0.5 = %+v", mid)
	}
	if got := LerpColor(a, b, -1); got != a {
		t.Fatalf("LerpColor below 0 not clamped to a: %+v", got)
	}
	if got := LerpColor(a, b, 2); got != b {
		t.Fatalf("LerpColor above 1 not clamped to b: %+v", got)
	}
}

func TestFogCmdsLerpsTowardBackdrop(t *testing.T) {
	col := Color{R: 200, G: 100, B: 50, A: 255}
	back := Color{R: 10, G: 20, B: 30, A: 255}
	cs := []Cmd{
		{Shape: ShapeBox, Pos: v3(0, 0, 0), Color: col},
		{Shape: ShapeBox, Pos: v3(0, 0, fogEnd+10), Color: col},
	}
	out := FogCmds(cs, back, 0)
	if out[0].Color != col {
		t.Fatalf("near command color = %+v, want unchanged %+v", out[0].Color, col)
	}
	if out[1].Color != back {
		t.Fatalf("far command color = %+v, want full backdrop %+v", out[1].Color, back)
	}
	if cs[0].Color != col || cs[1].Color != col {
		t.Fatal("FogCmds mutated its input")
	}
}

func TestFogBoundaries(t *testing.T) {
	col := Color{R: 200, G: 100, B: 50, A: 255}
	back := Color{R: 10, G: 20, B: 30, A: 255}
	cs := []Cmd{
		{Shape: ShapeBox, Pos: v3(0, 0, fogStart), Color: col},
		{Shape: ShapeBox, Pos: v3(0, 0, fogEnd), Color: col},
	}
	out := FogCmds(cs, back, 0)
	if out[0].Color != col {
		t.Fatalf("at fogStart color = %+v, want unchanged", out[0].Color)
	}
	if out[1].Color != back {
		t.Fatalf("at fogEnd color = %+v, want full backdrop", out[1].Color)
	}
}

func TestTelegraphOverlayMarksLaneAndKind(t *testing.T) {
	b := gameplay.NewBossFight(3)
	b.Step(runner.Input{}, testDT) // announce the opening attack
	if b.Vulnerable() {
		t.Fatal("fight became vulnerable before any attacks")
	}
	cs := telegraphCmds(b)
	if len(cs) != 1 {
		t.Fatalf("telegraph commands = %d, want 1", len(cs))
	}
	if got := float64(cs[0].Pos.X); math.Abs(got-laneX(float64(b.AttackLane()))) > 1e-9 {
		t.Fatalf("telegraph x = %v, want lane %d at %v", cs[0].Pos.X, b.AttackLane(), laneX(float64(b.AttackLane())))
	}
	if b.AttackKind() == gameplay.AttackLow {
		if cs[0].Pos.Y > 0.5 {
			t.Fatalf("low attack telegraph raised at y=%v, want floor level", cs[0].Pos.Y)
		}
	} else {
		if cs[0].Pos.Y < 0.5 {
			t.Fatalf("high attack telegraph at y=%v, want raised", cs[0].Pos.Y)
		}
	}
}

func TestTelegraphHidesWhileVulnerable(t *testing.T) {
	b := gameplay.NewBossFight(3)
	for i := 0; i < 4000 && !b.Vulnerable() && b.State() == gameplay.BossFighting; i++ {
		b.Step(bossDodgeInput(b), testDT)
	}
	if !b.Vulnerable() {
		t.Fatal("fight never became vulnerable")
	}
	if cs := telegraphCmds(b); len(cs) != 0 {
		t.Fatalf("vulnerable telegraph commands = %d, want 0", len(cs))
	}
}

func TestBossHeroPose(t *testing.T) {
	b := gameplay.NewBossFight(3)
	if got := bossHeroPose(b); got != PoseRun {
		t.Fatalf("grounded boss pose = %v, want PoseRun", got)
	}
	b.Step(runner.Input{Jump: true}, testDT)
	if got := bossHeroPose(b); got != PoseJump {
		t.Fatalf("jumping boss pose = %v, want PoseJump", got)
	}

	s := gameplay.NewBossFight(3)
	s.Step(runner.Input{Slide: true}, testDT)
	if got := bossHeroPose(s); got != PoseSlide {
		t.Fatalf("sliding boss pose = %v, want PoseSlide", got)
	}
}

func TestFadeAdvancesOnSceneChange(t *testing.T) {
	u := &UI{}
	u.advanceFade(campaign.SceneMainMenu, 0)
	if u.fadeCover != 1 {
		t.Fatalf("first frame cover = %v, want 1", u.fadeCover)
	}
	u.advanceFade(campaign.SceneMainMenu, fadeDuration)
	if u.fadeCover != 0 {
		t.Fatalf("cover after fadeDuration = %v, want 0", u.fadeCover)
	}
	u.advanceFade(campaign.SceneMainMenu, 0.1)
	if u.fadeCover != 0 {
		t.Fatalf("cover clamped below zero = %v, want 0", u.fadeCover)
	}
	u.advanceFade(campaign.ScenePlay, 0)
	if u.fadeCover != 1 {
		t.Fatalf("cover after scene change = %v, want 1", u.fadeCover)
	}
}

func TestSceneBackdrop(t *testing.T) {
	menu := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	if got := sceneBackdrop(menu); got != uiBackground {
		t.Fatalf("menu backdrop = %+v, want %+v", got, uiBackground)
	}

	play := newGameFlow(t)
	if got, want := sceneBackdrop(play), Palette(runTheme(play)).Backdrop; got != want {
		t.Fatalf("run backdrop = %+v, want %+v", got, want)
	}

	boss := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.Progress{Unlocked: 10}, false)
	boss.StartBoss()
	if got := sceneBackdrop(boss); got != bossDark {
		t.Fatalf("boss backdrop = %+v, want %+v", got, bossDark)
	}
}

func newPoseRun(t *testing.T) *gameplay.Run {
	t.Helper()
	l := levelgen.New(1, levelgen.Params{BaseSpeed: 8, Length: 60, Density: 0})
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = 8
	return gameplay.NewRun(runner.New(l, cfg), l, 3)
}

func findRampLevel(t *testing.T) *levelgen.Level {
	t.Helper()
	p := levelgen.Params{BaseSpeed: 8, Length: 60, Density: 0}
	for seed := uint64(0); seed < 200; seed++ {
		l := levelgen.New(seed, p)
		for _, e := range l.Elements(0, l.Length()) {
			if e.Kind == levelgen.Ramp && e.Lane == 1 {
				return l
			}
		}
	}
	t.Fatal("no center-lane ramp within 200 seeds")
	return nil
}
