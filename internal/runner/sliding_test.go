package runner_test

import (
	"testing"

	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

const getterDT = 1.0 / 60.0

func TestAirborneGetterFollowsJump(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := levelgen.New(1, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 200, Density: 0})
	r := runner.New(l, cfg)

	if r.Airborne() {
		t.Fatalf("Airborne() = true at start, want false")
	}
	if outs := r.Step(runner.Input{Jump: true}, getterDT); len(outs) == 0 {
		t.Fatalf("jump step reported no events")
	}
	if !r.Airborne() {
		t.Fatalf("Airborne() = false right after jump, want true")
	}

	for i := 0; i < 2000 && r.Airborne(); i++ {
		r.Step(runner.Input{}, getterDT)
	}
	if r.Airborne() {
		t.Fatalf("Airborne() = true after landing, want false")
	}
}

func TestSlidingGetterFollowsSlide(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := levelgen.New(1, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 200, Density: 0})
	r := runner.New(l, cfg)

	if r.Sliding() {
		t.Fatalf("Sliding() = true at start, want false")
	}
	if outs := r.Step(runner.Input{Slide: true}, getterDT); len(outs) == 0 {
		t.Fatalf("slide step reported no events")
	}
	if !r.Sliding() {
		t.Fatalf("Sliding() = false right after slide, want true")
	}

	for i := 0; i < 2000 && r.Sliding(); i++ {
		r.Step(runner.Input{}, getterDT)
	}
	if r.Sliding() {
		t.Fatalf("Sliding() = true after the slide ended, want false")
	}
}

func TestSlammingGetterFollowsSlam(t *testing.T) {
	cfg := runner.DefaultConfig()
	l := levelgen.New(1, levelgen.Params{BaseSpeed: cfg.BaseSpeed, Length: 200, Density: 0})
	r := runner.New(l, cfg)

	if r.Slamming() {
		t.Fatalf("Slamming() = true at start, want false")
	}
	r.Step(runner.Input{Jump: true}, getterDT)
	for i := 0; i < 1000 && r.Height() < 0.8; i++ {
		r.Step(runner.Input{}, getterDT)
	}
	if r.Height() < 0.8 {
		t.Fatalf("jump never rose to slam height: Height()=%v", r.Height())
	}
	if outs := r.Step(runner.Input{Slide: true}, getterDT); len(outs) == 0 {
		t.Fatalf("slam step reported no events")
	}
	if !r.Slamming() {
		t.Fatalf("Slamming() = false right after slam, want true")
	}

	for i := 0; i < 2000 && r.Slamming(); i++ {
		r.Step(runner.Input{}, getterDT)
	}
	if r.Slamming() {
		t.Fatalf("Slamming() = true after the slam landed, want false")
	}
}
