package runner_test

import (
	"math"
	"testing"

	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

const slideTestDT = 1.0 / 60.0

func TestSlidingGetter(t *testing.T) {
	r := newGettersRunner(t, levelgen.Params{BaseSpeed: 8, Length: 60, Density: 0})
	if r.Sliding() || r.Slamming() || r.Airborne() {
		t.Fatal("fresh runner reports a move")
	}

	r.Step(runner.Input{Slide: true}, slideTestDT)
	if !r.Sliding() {
		t.Fatal("Sliding() = false right after a slide")
	}
	if r.Slamming() || r.Airborne() {
		t.Fatal("slide reports slam or airborne")
	}

	for i := 0; i < 2000 && r.Sliding(); i++ {
		r.Step(runner.Input{}, slideTestDT)
	}
	if r.Sliding() {
		t.Fatal("Sliding() still true after the slide duration")
	}
	if r.Slamming() || r.Airborne() {
		t.Fatal("grounded after slide reports slam or airborne")
	}
}

func TestSlamGetter(t *testing.T) {
	r := newGettersRunner(t, levelgen.Params{BaseSpeed: 8, Length: 60, Density: 0})
	r.Step(runner.Input{Jump: true}, slideTestDT)

	// Rise to near the jump apex before slamming: a slam from ground height
	// lands within a single step, so it would be over before the getter is
	// read.
	cfg := runner.DefaultConfig()
	for i := 0; i < int(math.Ceil(cfg.JumpVelocity/cfg.Gravity/slideTestDT)); i++ {
		r.Step(runner.Input{}, slideTestDT)
	}

	r.Step(runner.Input{Slide: true}, slideTestDT)
	if !r.Slamming() {
		t.Fatal("Slamming() = false right after an airborne slide")
	}
	if !r.Airborne() {
		t.Fatal("slam is not airborne")
	}
	if r.Sliding() {
		t.Fatal("slam reports sliding")
	}

	for i := 0; i < 2000 && r.Airborne(); i++ {
		r.Step(runner.Input{}, slideTestDT)
	}
	if r.Airborne() || r.Slamming() {
		t.Fatal("after slam lands, airborne or slamming still true")
	}
}

func TestAirborneGetter(t *testing.T) {
	r := newGettersRunner(t, levelgen.Params{BaseSpeed: 8, Length: 60, Density: 0})
	r.Step(runner.Input{Jump: true}, slideTestDT)
	if !r.Airborne() {
		t.Fatal("Airborne() = false right after a jump")
	}
	if r.Sliding() || r.Slamming() {
		t.Fatal("jump reports slide or slam")
	}

	for i := 0; i < 2000 && r.Airborne(); i++ {
		r.Step(runner.Input{}, slideTestDT)
	}
	if r.Airborne() {
		t.Fatal("Airborne() still true after landing")
	}
}

func TestContainerRideGetters(t *testing.T) {
	l := findGetterRamp(t)
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = 8
	r := runner.New(l, cfg)

	for i := 0; i < 4000 && !r.OnContainer(); i++ {
		r.Step(runner.Input{}, slideTestDT)
	}
	if !r.OnContainer() {
		t.Fatal("never reached the container ramp")
	}
	if r.Airborne() || r.Sliding() || r.Slamming() {
		t.Fatalf("container ride reports a move: airborne=%v sliding=%v slamming=%v",
			r.Airborne(), r.Sliding(), r.Slamming())
	}
	if r.Height() != containerHeight {
		t.Fatalf("Height() = %v on container, want %v", r.Height(), containerHeight)
	}
}

func newGettersRunner(t *testing.T, p levelgen.Params) *runner.Runner {
	t.Helper()
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = p.BaseSpeed
	return runner.New(levelgen.New(1, p), cfg)
}

func findGetterRamp(t *testing.T) *levelgen.Level {
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
