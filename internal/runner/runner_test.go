package runner_test

import (
	"math"
	"reflect"
	"testing"

	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

const (
	barrierDepth    = 1.0
	containerHeight = 1.5
)

func TestJumpClearsWidestPit(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := levelgen.New(0, p)
	airtime := 2 * cfg.JumpVelocity / cfg.Gravity
	jumpDistance := cfg.BaseSpeed * airtime
	if jumpDistance <= l.Constraints().MaxPitWidth+1.0 {
		t.Fatalf("jump distance %v does not clear MaxPitWidth %v with margin", jumpDistance, l.Constraints().MaxPitWidth)
	}
	reactionTime := l.Constraints().ReactionGap / cfg.BaseSpeed
	if cfg.LaneSwitchDuration > reactionTime {
		t.Fatalf("lane switch duration %v exceeds reaction time %v", cfg.LaneSwitchDuration, reactionTime)
	}

	wide := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Pit && e.Span >= 2.8
	})
	pit, _ := firstForced(wide)

	r := runner.New(wide, cfg)
	dt := 1.0 / 60.0
	lead := cfg.BaseSpeed * 0.15
	jumped := false
	for i := 0; i < 4000 && r.Z() < pit.Z+pit.Span+1; i++ {
		var in runner.Input
		if r.Z() >= pit.Z-lead && r.Z() < pit.Z {
			in.Jump = true
		}
		for _, ev := range r.Step(in, dt) {
			switch ev {
			case runner.EventJump:
				jumped = true
			case runner.EventFellInPit:
				t.Fatalf("fell into pit %+v at z %v", pit, r.Z())
			case runner.EventHitBarrier:
				t.Fatalf("hit a barrier while clearing pit %+v at z %v", pit, r.Z())
			}
		}
	}
	if !jumped {
		t.Fatalf("never jumped for pit %+v", pit)
	}
	if r.Z() <= pit.Z+pit.Span {
		t.Fatalf("z %v did not clear pit ending at %v", r.Z(), pit.Z+pit.Span)
	}
}

func TestJumpClearsLowBarrier(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.LowBarrier && e.Lane == 1
	})
	bar, _ := firstForced(l)

	r := runner.New(l, cfg)
	dt := 1.0 / 60.0
	jumped := false
	for i := 0; i < 4000 && r.Z() < bar.Z+barrierDepth+1; i++ {
		var in runner.Input
		if r.Z() >= bar.Z-cfg.BaseSpeed*0.2 && r.Z() < bar.Z {
			in.Jump = true
		}
		for _, ev := range r.Step(in, dt) {
			switch ev {
			case runner.EventJump:
				jumped = true
			case runner.EventHitBarrier:
				t.Fatalf("hit low barrier %+v at z %v", bar, r.Z())
			}
		}
	}
	if !jumped {
		t.Fatalf("never jumped for low barrier %+v", bar)
	}
	if r.Z() <= bar.Z+barrierDepth {
		t.Fatalf("z %v did not pass low barrier ending at %v", r.Z(), bar.Z+barrierDepth)
	}
}

func TestSlidePassesUnderHighBarrier(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.HighBarrier && e.Lane == 1
	})
	bar, _ := firstForced(l)

	r := runner.New(l, cfg)
	dt := 1.0 / 60.0
	slid := false
	for i := 0; i < 4000 && r.Z() < bar.Z+barrierDepth+1; i++ {
		var in runner.Input
		if r.Z() >= bar.Z-cfg.BaseSpeed*0.3 && r.Z() < bar.Z+barrierDepth {
			in.Slide = true
		}
		for _, ev := range r.Step(in, dt) {
			switch ev {
			case runner.EventSlide:
				slid = true
			case runner.EventHitBarrier:
				t.Fatalf("hit high barrier %+v at z %v", bar, r.Z())
			}
		}
	}
	if !slid {
		t.Fatalf("never slid for high barrier %+v", bar)
	}
	if r.Z() <= bar.Z+barrierDepth {
		t.Fatalf("z %v did not pass high barrier ending at %v", r.Z(), bar.Z+barrierDepth)
	}
}

func TestLaneSwitchDodgesEnemy(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Enemy && e.Lane == 1
	})
	enemy, _ := firstForced(l)

	r := runner.New(l, cfg)
	dt := 1.0 / 60.0
	dodged := false
	for i := 0; i < 4000 && r.Z() < enemy.Z+barrierDepth+1; i++ {
		var in runner.Input
		if !dodged && r.Z() >= enemy.Z-(cfg.BaseSpeed*cfg.LaneSwitchDuration+1.0) {
			in.Left = true
			dodged = true
		}
		for _, ev := range r.Step(in, dt) {
			switch ev {
			case runner.EventHitEnemy:
				t.Fatalf("hit enemy %+v at z %v", enemy, r.Z())
			case runner.EventStomp:
				t.Fatalf("stomped enemy %+v instead of dodging", enemy)
			}
		}
	}
	if !dodged {
		t.Fatalf("never started the dodge for enemy %+v", enemy)
	}
	if r.Z() <= enemy.Z+barrierDepth {
		t.Fatalf("z %v did not pass enemy ending at %v", r.Z(), enemy.Z+barrierDepth)
	}
}

func TestSlamDescendsFasterThanFall(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := levelgen.New(1, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 60, Density: 0})
	dt := 1.0 / 60.0

	fall := runner.New(l, cfg)
	fall.Step(runner.Input{Jump: true}, dt)
	fallLand := stepsToLand(t, fall, dt)

	slam := runner.New(l, cfg)
	slam.Step(runner.Input{Jump: true}, dt)
	apexSteps := int(math.Ceil(cfg.JumpVelocity / cfg.Gravity / dt))
	for i := 0; i < apexSteps; i++ {
		slam.Step(runner.Input{}, dt)
	}
	slamEvents := slam.Step(runner.Input{Slide: true}, dt)
	slammed := false
	for _, ev := range slamEvents {
		if ev == runner.EventSlam {
			slammed = true
		}
	}
	if !slammed {
		t.Fatalf("airborne Slide did not report EventSlam; events: %v", slamEvents)
	}
	slamLand := stepsToLand(t, slam, dt)

	if slamLand >= fallLand {
		t.Fatalf("slam landed in %d steps, free fall landed in %d steps", slamLand, fallLand)
	}
}

func TestStompReportsEventStomp(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		return ok && e.Kind == levelgen.Enemy && e.Lane == 1
	})
	enemy, _ := firstForced(l)

	r := runner.New(l, cfg)
	dt := 1.0 / 60.0
	stomped := false
	for i := 0; i < 4000 && r.Z() < enemy.Z+barrierDepth+1; i++ {
		var in runner.Input
		if r.Z() >= enemy.Z-cfg.BaseSpeed*0.32 && r.Z() < enemy.Z {
			in.Jump = true
		}
		for _, ev := range r.Step(in, dt) {
			switch ev {
			case runner.EventStomp:
				stomped = true
			case runner.EventHitEnemy:
				t.Fatalf("hit enemy %+v instead of stomping at z %v", enemy, r.Z())
			}
		}
	}
	if !stomped {
		t.Fatalf("never stomped enemy %+v", enemy)
	}
}

func TestSpeedRampsTowardCap(t *testing.T) {
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = 8
	cfg.MaxSpeed = 16
	cfg.SpeedRampDistance = 100

	l := levelgen.New(2, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 400, Density: 0})
	r := runner.New(l, cfg)
	if got := r.Speed(); got != cfg.BaseSpeed {
		t.Fatalf("initial Speed() = %v, want %v", got, cfg.BaseSpeed)
	}

	dt := 1.0 / 60.0
	maxObserved := 0.0
	seenCap := false
	for i := 0; i < 20000; i++ {
		if r.Z() >= l.FinishZ() {
			break
		}
		s := r.Speed()
		if s > cfg.MaxSpeed {
			t.Fatalf("Speed() = %v exceeds cap %v at z %v", s, cfg.MaxSpeed, r.Z())
		}
		if s > maxObserved {
			maxObserved = s
		}
		if s == cfg.MaxSpeed {
			seenCap = true
		}
		r.Step(runner.Input{}, dt)
	}
	if maxObserved <= cfg.BaseSpeed {
		t.Fatalf("speed never rose above base %v", cfg.BaseSpeed)
	}
	if !seenCap {
		t.Fatalf("speed never reached cap %v", cfg.MaxSpeed)
	}
}

func TestSameInputScriptGivesIdenticalStates(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed
	l := levelgen.New(42, p)

	a := runner.New(l, cfg)
	b := runner.New(l, cfg)

	dts := []float64{1.0 / 60, 1.0 / 120, 1.0 / 30, 1.0 / 60}
	script := make([]stepScript, 120)
	for i := range script {
		script[i] = stepScript{
			in: runner.Input{
				Left:  i%9 == 0,
				Right: i%11 == 0,
				Jump:  i%17 == 0,
				Slide: i%13 == 0,
			},
			dt: dts[i%len(dts)],
		}
	}

	for i, s := range script {
		ea := a.Step(s.in, s.dt)
		eb := b.Step(s.in, s.dt)
		if !reflect.DeepEqual(ea, eb) {
			t.Fatalf("step %d events differ: %v vs %v", i, ea, eb)
		}
		if a.Lane() != b.Lane() || a.Z() != b.Z() || a.Height() != b.Height() ||
			a.Speed() != b.Speed() || a.OnContainer() != b.OnContainer() {
			t.Fatalf("step %d states differ", i)
		}
	}
}

func TestScriptedRunFinishesGeneratedLevel(t *testing.T) {
	presets := []levelgen.Params{levelgen.Mangroves(), levelgen.Harbor(), levelgen.Neon()}
	for _, p := range presets {
		for _, seed := range []uint64{0, 1, 2, 3, 4, 5} {
			l := levelgen.New(seed, p)
			cfg := runner.DefaultConfig()
			cfg.BaseSpeed = p.BaseSpeed
			r := drivePlan(t, l, cfg)
			if r.Z() != l.FinishZ() {
				t.Fatalf("seed %d: z %v != finish %v", seed, r.Z(), l.FinishZ())
			}
		}
	}
}

func TestFailureEventsStopForwardMotion(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	cases := []struct {
		name  string
		match func(*levelgen.Level) bool
		want  runner.Event
	}{
		{
			name: "barrier",
			match: func(l *levelgen.Level) bool {
				e, ok := firstForced(l)
				return ok && (e.Kind == levelgen.LowBarrier || e.Kind == levelgen.HighBarrier) && e.Lane == 1
			},
			want: runner.EventHitBarrier,
		},
		{
			name: "enemy",
			match: func(l *levelgen.Level) bool {
				e, ok := firstForced(l)
				return ok && e.Kind == levelgen.Enemy && e.Lane == 1
			},
			want: runner.EventHitEnemy,
		},
		{
			name: "pit",
			match: func(l *levelgen.Level) bool {
				e, ok := firstForced(l)
				return ok && e.Kind == levelgen.Pit
			},
			want: runner.EventFellInPit,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := findLevel(t, p, tc.match)
			r := runner.New(l, cfg)
			dt := 1.0 / 60.0

			seen := false
			var hitZ float64
			for i := 0; i < 4000; i++ {
				for _, ev := range r.Step(runner.Input{}, dt) {
					if ev == tc.want {
						seen = true
						hitZ = r.Z()
					}
				}
				if seen {
					break
				}
			}
			if !seen {
				t.Fatalf("event %v never reported", tc.want)
			}

			for i := 0; i < 3; i++ {
				r.Step(runner.Input{}, dt)
			}
			if r.Z() != hitZ {
				t.Fatalf("z moved after %v: %v -> %v", tc.want, hitZ, r.Z())
			}
			if r.Speed() != 0 {
				t.Fatalf("Speed() = %v after %v, want 0", r.Speed(), tc.want)
			}
		})
	}
}

func TestRampTransitionsRaiseAndLower(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := findLevel(t, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 60, Density: 0}, func(l *levelgen.Level) bool {
		_, ok := firstRampOnLane(l, 1)
		return ok
	})
	ramp, _ := firstRampOnLane(l, 1)

	r := runner.New(l, cfg)
	dt := 1.0 / 60.0
	up, down := false, false
	sawOn := false
	for i := 0; i < 4000 && r.Z() < ramp.Z+ramp.Span+1; i++ {
		for _, ev := range r.Step(runner.Input{}, dt) {
			switch ev {
			case runner.EventRampUp:
				up = true
			case runner.EventRampDown:
				down = true
			}
		}
		if r.OnContainer() {
			sawOn = true
			if r.Height() != containerHeight {
				t.Fatalf("on container but Height() = %v, want %v", r.Height(), containerHeight)
			}
		}
	}
	if !up {
		t.Fatalf("no EventRampUp for ramp %+v", ramp)
	}
	if !down {
		t.Fatalf("no EventRampDown for ramp %+v", ramp)
	}
	if !sawOn {
		t.Fatalf("OnContainer() never turned true for ramp %+v", ramp)
	}
	if r.OnContainer() {
		t.Fatalf("OnContainer() still true after ramp %+v", ramp)
	}
	if r.Height() != 0 {
		t.Fatalf("Height() = %v after ramp exit, want 0", r.Height())
	}
	if r.Z() <= ramp.Z+ramp.Span {
		t.Fatalf("z %v did not pass ramp ending at %v", r.Z(), ramp.Z+ramp.Span)
	}
}

func TestRestartResetsLaneHeightSpeedStopAndZ(t *testing.T) {
	p := levelgen.Mangroves()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed

	l := findLevel(t, p, func(l *levelgen.Level) bool {
		e, ok := firstForced(l)
		if !ok || (e.Kind != levelgen.LowBarrier && e.Kind != levelgen.HighBarrier) || e.Lane != 1 {
			return false
		}
		if e.Z < 12 {
			return false
		}
		return clearForced(l, e.Z+barrierDepth, e.Z+barrierDepth+30)
	})
	bar, _ := firstForced(l)

	r := runner.New(l, cfg)
	dt := 1.0 / 60.0

	// Move to lane 2 so the lane-1 barrier is left behind for the jump.
	for i := 0; i < 4000 && r.Lane() < 2; i++ {
		r.Step(runner.Input{Right: true}, dt)
	}
	if r.Lane() != 2 {
		t.Fatalf("Lane() = %v after moving right, want 2", r.Lane())
	}

	jumped := false
	for _, ev := range r.Step(runner.Input{Jump: true}, dt) {
		if ev == runner.EventJump {
			jumped = true
		}
	}
	if !jumped {
		t.Fatalf("jump was not reported")
	}
	for i := 0; i < 4000 && r.Height() > 0; i++ {
		r.Step(runner.Input{}, dt)
	}
	if r.Height() != 0 {
		t.Fatalf("Height() = %v after landing, want 0", r.Height())
	}

	// Return to lane 1 and run into the barrier to stop forward motion.
	for i := 0; i < 4000 && r.Lane() > 1; i++ {
		r.Step(runner.Input{Left: true}, dt)
	}
	if r.Lane() != 1 {
		t.Fatalf("Lane() = %v after moving left, want 1", r.Lane())
	}

	hit := false
	var hitZ float64
	for i := 0; i < 4000; i++ {
		for _, ev := range r.Step(runner.Input{}, dt) {
			if ev == runner.EventHitBarrier {
				hit = true
				hitZ = r.Z()
			}
		}
		if hit {
			break
		}
	}
	if !hit {
		t.Fatalf("never hit lane-1 barrier %+v", bar)
	}
	if r.Speed() != 0 {
		t.Fatalf("Speed() = %v after hitting barrier, want 0", r.Speed())
	}
	for i := 0; i < 3; i++ {
		r.Step(runner.Input{}, dt)
	}
	if r.Z() != hitZ {
		t.Fatalf("Z moved after stop: %v -> %v", hitZ, r.Z())
	}

	startZ := bar.Z + barrierDepth + 0.1
	r.Restart(startZ)
	if r.Lane() != 1 {
		t.Fatalf("Lane() = %v after Restart, want 1", r.Lane())
	}
	if r.Height() != 0 {
		t.Fatalf("Height() = %v after Restart, want 0", r.Height())
	}
	if r.Speed() != cfg.BaseSpeed {
		t.Fatalf("Speed() = %v after Restart, want %v", r.Speed(), cfg.BaseSpeed)
	}
	if r.Z() != startZ {
		t.Fatalf("Z() = %v after Restart, want %v", r.Z(), startZ)
	}

	before := r.Z()
	events := r.Step(runner.Input{}, dt)
	if r.Z() <= before {
		t.Fatalf("Z did not advance after Restart: %v -> %v", before, r.Z())
	}
	if r.Speed() <= 0 {
		t.Fatalf("Speed() = %v after Restart step, want positive", r.Speed())
	}
	for _, ev := range events {
		if ev == runner.EventHitBarrier {
			t.Fatalf("hit barrier right after Restart at z %v", r.Z())
		}
	}

	for i := 0; i < 60; i++ {
		for _, ev := range r.Step(runner.Input{}, dt) {
			switch ev {
			case runner.EventHitBarrier, runner.EventHitEnemy, runner.EventFellInPit:
				t.Fatalf("hit event %v within 60 frames of Restart at z %v", ev, r.Z())
			}
		}
	}
}

type stepScript struct {
	in runner.Input
	dt float64
}

const (
	actJump = iota
	actLeft
	actRight
)

type decision struct {
	z    float64
	kind int
}

// buildPlan turns levelgen's solver strategy into timed runner inputs: jump
// every pit, and dodge every barrier or enemy on the current lane by moving
// to an adjacent free lane. The solver's plan proves passability, so this is
// the scripted run the runner must be able to execute.
func buildPlan(l *levelgen.Level, cfg runner.Config) []decision {
	lane := 1
	var plan []decision
	for _, e := range forcedOf(l) {
		switch e.Kind {
		case levelgen.Pit:
			plan = append(plan, decision{z: e.Z - cfg.BaseSpeed*0.15, kind: actJump})
		case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
			if e.Lane == lane {
				target := lane - 1
				if target < 0 {
					target = lane + 1
				}
				plan = append(plan, decision{
					z:    e.Z - (cfg.BaseSpeed*cfg.LaneSwitchDuration + 1.0),
					kind: actForDirection(target - lane),
				})
				lane = target
			}
		}
	}
	return plan
}

func actForDirection(dir int) int {
	if dir < 0 {
		return actLeft
	}
	return actRight
}

func drivePlan(t *testing.T, l *levelgen.Level, cfg runner.Config) *runner.Runner {
	t.Helper()
	r := runner.New(l, cfg)
	plan := buildPlan(l, cfg)
	dt := 1.0 / 60.0
	idx := 0
	holdingJump := false

	const maxSteps = 30000
	for i := 0; i < maxSteps; i++ {
		var in runner.Input
		if idx < len(plan) && r.Z() >= plan[idx].z {
			switch plan[idx].kind {
			case actJump:
				holdingJump = true
			case actLeft:
				in.Left = true
				idx++
			case actRight:
				in.Right = true
				idx++
			}
		}
		if holdingJump {
			in.Jump = true
		}

		finished := false
		for _, ev := range r.Step(in, dt) {
			switch ev {
			case runner.EventJump:
				if holdingJump {
					holdingJump = false
					idx++
				}
			case runner.EventFinished:
				finished = true
			case runner.EventHitBarrier, runner.EventHitEnemy, runner.EventFellInPit:
				t.Fatalf("seed %d: hit event %v at z %v", l.Seed(), ev, r.Z())
			}
		}
		if finished {
			return r
		}
	}
	t.Fatalf("seed %d: did not finish, z %v of %v", l.Seed(), r.Z(), l.FinishZ())
	return nil
}

func stepsToLand(t *testing.T, r *runner.Runner, dt float64) int {
	t.Helper()
	for i := 1; i <= 2000; i++ {
		before := r.Height()
		r.Step(runner.Input{}, dt)
		if before > 0 && r.Height() <= 0 {
			return i
		}
	}
	t.Fatalf("never landed")
	return -1
}

func findLevel(t *testing.T, p levelgen.Params, match func(*levelgen.Level) bool) *levelgen.Level {
	t.Helper()
	for seed := uint64(0); seed < 1000; seed++ {
		l := levelgen.New(seed, p)
		if match(l) {
			return l
		}
	}
	t.Fatalf("no matching level within 1000 seeds")
	return nil
}

func firstForced(l *levelgen.Level) (levelgen.Element, bool) {
	for _, e := range l.Elements(0, l.Length()) {
		if isForced(e.Kind) {
			return e, true
		}
	}
	return levelgen.Element{}, false
}

// forcedSpan returns the half-open [start, end) z interval a forced element
// occupies, matching the runner's own spanOf contract.
func forcedSpan(e levelgen.Element) (float64, float64) {
	switch e.Kind {
	case levelgen.Pit:
		return e.Z, e.Z + e.Span
	case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
		return e.Z, e.Z + barrierDepth
	default:
		return e.Z, e.Z
	}
}

// clearForced reports whether no forced element (pit, barrier or enemy)
// overlaps the half-open z interval [from, to).
func clearForced(l *levelgen.Level, from, to float64) bool {
	for _, e := range forcedOf(l) {
		start, end := forcedSpan(e)
		if start < to && end > from {
			return false
		}
	}
	return true
}

func firstRampOnLane(l *levelgen.Level, lane int) (levelgen.Element, bool) {
	for _, e := range l.Elements(0, l.Length()) {
		if e.Kind == levelgen.Ramp && e.Lane == lane {
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
	default:
		return false
	}
}
