package gameplay_test

import (
	"math"
	"testing"

	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

const stepDT = 1.0 / 60.0

func TestFirstStumbleStartsChase(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && (e.Kind == levelgen.LowBarrier || e.Kind == levelgen.HighBarrier) && e.Lane == 1
	})
	bar, _ := firstForced(l)

	r := gameplay.NewRun(runner.New(l, cfg), l, 3)
	stumbled := false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			if o == gameplay.OutcomeStumble {
				stumbled = true
			}
		}
		if stumbled {
			break
		}
	}
	if !stumbled {
		t.Fatalf("never stumbled on barrier %+v", bar)
	}
	if r.Lives != 3 {
		t.Fatalf("Lives = %d after first stumble, want 3", r.Lives)
	}
	if r.Chase != 1 {
		t.Fatalf("Chase = %v after first stumble, want 1", r.Chase)
	}
	if r.Bald {
		t.Fatalf("Bald = true after first stumble, want false")
	}
	if r.State != gameplay.StateRunning {
		t.Fatalf("State = %v after first stumble, want StateRunning", r.State)
	}
	if r.Cause != gameplay.CauseBarrier {
		t.Fatalf("Cause = %v after first stumble, want CauseBarrier", r.Cause)
	}

	before := r.Runner.Z()
	r.Step(runner.Input{}, stepDT)
	if r.Runner.Z() <= before {
		t.Fatalf("run did not keep moving after a stumble")
	}
}

func TestSecondStumbleWhileChasedSnips(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l, a, b := twoCloseLaneOneObstacles(t, p)
	r := gameplay.NewRun(runner.New(l, cfg), l, 3)

	stumbled, snipped := false, false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning && !snipped; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			switch o {
			case gameplay.OutcomeStumble:
				stumbled = true
			case gameplay.OutcomeSnip:
				snipped = true
			}
		}
	}
	if !stumbled {
		t.Fatalf("never had the first stumble on obstacle %+v", a)
	}
	if !snipped {
		t.Fatalf("never snipped on the second obstacle %+v", b)
	}
	if r.Lives != 2 {
		t.Fatalf("Lives = %d after snip, want 2", r.Lives)
	}
	if !r.Bald {
		t.Fatalf("Bald = false after snip, want true")
	}
	if r.Chase != 0 {
		t.Fatalf("Chase = %v after snip, want 0", r.Chase)
	}
	if r.Cause != gameplay.CauseSnip {
		t.Fatalf("Cause = %v after snip, want CauseSnip", r.Cause)
	}
	if r.Runner.Z() != 0 {
		t.Fatalf("runner Z = %v after snip, want CheckpointZ 0", r.Runner.Z())
	}
	if r.State != gameplay.StateRunning {
		t.Fatalf("State = %v after snip with lives left, want StateRunning", r.State)
	}
}

func TestShieldAbsorbsOneSnip(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && (e.Kind == levelgen.LowBarrier || e.Kind == levelgen.HighBarrier) && e.Lane == 1
	})
	bar, _ := firstForced(l)

	r := gameplay.NewRun(runner.New(l, cfg), l, 3)
	r.ShieldHits = 1
	r.CheckpointZ = bar.Z - 4
	r.Runner.Restart(r.CheckpointZ)
	r.Chase = 1

	saved := false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning && !saved; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			if o == gameplay.OutcomeShieldSaved {
				saved = true
			}
		}
	}
	if !saved {
		t.Fatalf("shield never saved the run on barrier %+v", bar)
	}
	if r.Lives != 3 {
		t.Fatalf("Lives = %d after shield save, want 3", r.Lives)
	}
	if r.ShieldHits != 0 {
		t.Fatalf("ShieldHits = %d after shield save, want 0", r.ShieldHits)
	}
	if r.Bald {
		t.Fatalf("Bald = true after shield save, want false")
	}
	if r.Chase != 0 {
		t.Fatalf("Chase = %v after shield save, want 0", r.Chase)
	}
	if r.State != gameplay.StateRunning {
		t.Fatalf("State = %v after shield save, want StateRunning", r.State)
	}
}

func TestPitLosesLifeRespawnsAndResetsChase(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Pit
	})
	pit, _ := firstForced(l)

	r := gameplay.NewRun(runner.New(l, cfg), l, 3)
	r.ShieldHits = 1
	r.CheckpointZ = pit.Z - 4
	r.Runner.Restart(r.CheckpointZ)
	r.Chase = 1

	fell := false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning && !fell; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			if o == gameplay.OutcomePit {
				fell = true
			}
		}
	}
	if !fell {
		t.Fatalf("never fell in pit %+v", pit)
	}
	if r.Lives != 2 {
		t.Fatalf("Lives = %d after pit, want 2", r.Lives)
	}
	if r.ShieldHits != 1 {
		t.Fatalf("ShieldHits = %d after pit, want 1 (the shield never stops a pit)", r.ShieldHits)
	}
	if r.Chase != 0 {
		t.Fatalf("Chase = %v after pit, want 0", r.Chase)
	}
	if r.Cause != gameplay.CausePit {
		t.Fatalf("Cause = %v after pit, want CausePit", r.Cause)
	}
	if r.Runner.Z() != r.CheckpointZ {
		t.Fatalf("runner Z = %v after pit, want CheckpointZ %v", r.Runner.Z(), r.CheckpointZ)
	}
}

func TestStompLosesNothing(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Enemy && e.Lane == 1
	})
	enemy, _ := firstForced(l)

	r := gameplay.NewRun(runner.New(l, cfg), l, 3)
	stomped := false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning && !stomped; i++ {
		var in runner.Input
		if r.Runner.Z() >= enemy.Z-cfg.BaseSpeed*0.32 && r.Runner.Z() < enemy.Z {
			in.Jump = true
		}
		for _, o := range r.Step(in, stepDT) {
			if o == gameplay.OutcomeStomp {
				stomped = true
			}
		}
	}
	if !stomped {
		t.Fatalf("never stomped enemy %+v", enemy)
	}
	if r.Lives != 3 {
		t.Fatalf("Lives = %d after stomp, want 3", r.Lives)
	}
	if r.Chase != 0 {
		t.Fatalf("Chase = %v after stomp, want 0", r.Chase)
	}
	if r.Bald {
		t.Fatalf("Bald = true after stomp, want false")
	}
}

func TestPitDeathSetsStateDeadWithCausePit(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Pit
	})
	pit, _ := firstForced(l)

	r := gameplay.NewRun(runner.New(l, cfg), l, 1)
	r.CheckpointZ = pit.Z - 4
	r.Runner.Restart(r.CheckpointZ)

	pitted, died := false, false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			switch o {
			case gameplay.OutcomePit:
				pitted = true
			case gameplay.OutcomeDied:
				died = true
			}
		}
		if died {
			break
		}
	}
	if !pitted {
		t.Fatalf("never fell in pit %+v", pit)
	}
	if !died {
		t.Fatalf("OutcomeDied never fired after the last life")
	}
	if r.Lives != 0 {
		t.Fatalf("Lives = %d after death, want 0", r.Lives)
	}
	if r.State != gameplay.StateDead {
		t.Fatalf("State = %v after death, want StateDead", r.State)
	}
	if r.Cause != gameplay.CausePit {
		t.Fatalf("Cause = %v after pit death, want CausePit", r.Cause)
	}
}

func TestSnipDeathSetsStateDeadWithCauseSnip(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && (e.Kind == levelgen.LowBarrier || e.Kind == levelgen.HighBarrier) && e.Lane == 1
	})
	bar, _ := firstForced(l)

	r := gameplay.NewRun(runner.New(l, cfg), l, 1)
	r.CheckpointZ = bar.Z - 4
	r.Runner.Restart(r.CheckpointZ)
	r.Chase = 1

	snipped, died := false, false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			switch o {
			case gameplay.OutcomeSnip:
				snipped = true
			case gameplay.OutcomeDied:
				died = true
			}
		}
		if died {
			break
		}
	}
	if !snipped {
		t.Fatalf("never snipped on barrier %+v", bar)
	}
	if !died {
		t.Fatalf("OutcomeDied never fired after the last life")
	}
	if r.Lives != 0 {
		t.Fatalf("Lives = %d after death, want 0", r.Lives)
	}
	if !r.Bald {
		t.Fatalf("Bald = false after snip death, want true")
	}
	if r.State != gameplay.StateDead {
		t.Fatalf("State = %v after death, want StateDead", r.State)
	}
	if r.Cause != gameplay.CauseSnip {
		t.Fatalf("Cause = %v after snip death, want CauseSnip", r.Cause)
	}
}

func TestChaseDecaysWithoutStumbles(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := levelgen.New(7, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 80, Density: 0})
	r := gameplay.NewRun(runner.New(l, cfg), l, 3)
	r.Chase = 1

	for i := 0; i < 120; i++ {
		r.Step(runner.Input{}, stepDT)
	}
	if got := r.Chase; math.Abs(got-0.5) > 0.02 {
		t.Fatalf("Chase = %v after 2s of clean running, want about 0.5", got)
	}

	for i := 0; i < 120; i++ {
		r.Step(runner.Input{}, stepDT)
	}
	if r.Chase > 1e-9 {
		t.Fatalf("Chase = %v after 4s of clean running, want 0", r.Chase)
	}
}

func TestFinishingSetsStateFinished(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := levelgen.New(8, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 40, Density: 0})
	r := gameplay.NewRun(runner.New(l, cfg), l, 3)

	finished := false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			if o == gameplay.OutcomeFinished {
				finished = true
			}
		}
		if finished {
			break
		}
	}
	if !finished {
		t.Fatalf("OutcomeFinished never fired")
	}
	if r.State != gameplay.StateFinished {
		t.Fatalf("State = %v after finishing, want StateFinished", r.State)
	}

	before := r.Runner.Z()
	if outs := r.Step(runner.Input{Jump: true}, stepDT); len(outs) != 0 {
		t.Fatalf("Step after finishing returned outcomes: %v", outs)
	}
	if r.Runner.Z() != before {
		t.Fatalf("run changed after finishing")
	}
}

func TestRespawnRestartsAtCheckpointAndKeepsPlaying(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Pit && e.Z > 12
	})
	pit, _ := firstForced(l)

	checkpoint := pit.Z - 8
	r := gameplay.NewRun(runner.New(l, cfg), l, 3)
	r.CheckpointZ = checkpoint
	r.Runner.Restart(checkpoint)

	fell := false
	for i := 0; i < 4000 && r.State == gameplay.StateRunning && !fell; i++ {
		for _, o := range r.Step(runner.Input{}, stepDT) {
			if o == gameplay.OutcomePit {
				fell = true
			}
		}
	}
	if !fell {
		t.Fatalf("never fell in pit %+v", pit)
	}
	if r.Lives != 2 {
		t.Fatalf("Lives = %d after pit, want 2", r.Lives)
	}
	if r.Runner.Z() != checkpoint {
		t.Fatalf("runner Z = %v after pit, want checkpoint %v", r.Runner.Z(), checkpoint)
	}

	before := r.Runner.Z()
	advanced := false
	for i := 0; i < 60; i++ {
		r.Step(runner.Input{}, stepDT)
		if r.Runner.Z() > before {
			advanced = true
			break
		}
	}
	if !advanced {
		t.Fatalf("run did not keep playing after respawn")
	}
}

func findLevel(t *testing.T, p levelgen.Params, match func(*levelgen.Level) bool) *levelgen.Level {
	t.Helper()
	for seed := uint64(0); seed < 2000; seed++ {
		l := levelgen.New(seed, p)
		if match(l) {
			return l
		}
	}
	t.Fatalf("no matching level within 2000 seeds")
	return nil
}

// twoCloseLaneOneObstacles finds a level whose first two forced elements are
// barriers or enemies on the center lane, close enough that the chase stays
// above the snip threshold at the second hit.
func twoCloseLaneOneObstacles(t *testing.T, p levelgen.Params) (*levelgen.Level, levelgen.Element, levelgen.Element) {
	t.Helper()
	for seed := uint64(0); seed < 5000; seed++ {
		l := levelgen.New(seed, p)
		forced := forcedOf(l)
		if len(forced) < 2 {
			continue
		}
		a, b := forced[0], forced[1]
		if isBarrierOrEnemy(a.Kind) && isBarrierOrEnemy(b.Kind) && a.Lane == 1 && b.Lane == 1 {
			if b.Z-a.Z < p.BaseSpeed*2 {
				return l, a, b
			}
		}
	}
	t.Fatalf("no level with two close lane-1 obstacles within 5000 seeds")
	return nil, levelgen.Element{}, levelgen.Element{}
}

func firstForced(l *levelgen.Level) (levelgen.Element, bool) {
	for _, e := range l.Elements(0, l.Length()) {
		if isForced(e.Kind) {
			return e, true
		}
	}
	return levelgen.Element{}, false
}

func forcedOf(l *levelgen.Level) []levelgen.Element {
	var out []levelgen.Element
	for _, e := range l.Elements(0, l.Length()) {
		if isForced(e.Kind) {
			out = append(out, e)
		}
	}
	return out
}

func isForced(k levelgen.Kind) bool {
	switch k {
	case levelgen.Pit, levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
		return true
	}
	return false
}

func isBarrierOrEnemy(k levelgen.Kind) bool {
	return k == levelgen.LowBarrier || k == levelgen.HighBarrier || k == levelgen.Enemy
}
