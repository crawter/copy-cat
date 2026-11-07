package main

import (
	"testing"

	"copycat/internal/audio"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

func TestSlideSoundSelection(t *testing.T) {
	tests := []struct {
		name     string
		airborne bool
		want     audio.Sound
	}{
		{name: "grounded slide", want: audio.SoundSlide},
		{name: "airborne slam", airborne: true, want: audio.SoundSlam},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slideSound(tt.airborne); got != tt.want {
				t.Fatalf("slideSound(%v) = %v, want %v", tt.airborne, got, tt.want)
			}
		})
	}
}

func TestSlideOnContainerMapsToSlide(t *testing.T) {
	r := runnerOnContainer(t)
	if airborne(r, nil) {
		t.Fatal("hero riding a container ramp must be grounded, not airborne")
	}
	if got := slideSound(airborne(r, nil)); got != audio.SoundSlide {
		t.Fatalf("slide on container = %v, want SoundSlide", got)
	}
}

// runnerOnContainer steps a runner onto a center-lane container ramp and
// returns it, using the first seed whose mangroves level has a reachable one.
func runnerOnContainer(t *testing.T) *runner.Runner {
	t.Helper()
	for seed := uint64(1); seed <= 500; seed++ {
		l := levelgen.New(seed, levelgen.Mangroves())
		ramp, ok := reachableCenterRamp(l)
		if !ok {
			continue
		}
		r := runner.New(l, runner.DefaultConfig())
		for step := 0; step < 6000; step++ {
			r.Step(runner.Input{}, 1.0/60.0)
			if r.OnContainer() {
				return r
			}
			if r.Z() > ramp.Z+ramp.Span+1 {
				break
			}
		}
	}
	t.Fatal("no seed produced a reachable center-lane ramp")
	return nil
}

// reachableCenterRamp returns the first center-lane ramp reachable by running
// straight from the start: no pit or center-lane forced obstacle before it.
func reachableCenterRamp(l *levelgen.Level) (levelgen.Element, bool) {
	var ramp levelgen.Element
	found := false
	for _, e := range l.Elements(0, l.Length()) {
		if e.Kind == levelgen.Ramp && e.Lane == 1 {
			ramp = e
			found = true
			break
		}
	}
	if !found {
		return levelgen.Element{}, false
	}
	for _, e := range l.Elements(0, ramp.Z) {
		if e.Kind == levelgen.Pit {
			return levelgen.Element{}, false
		}
		switch e.Kind {
		case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
			if e.Lane == 1 {
				return levelgen.Element{}, false
			}
		}
	}
	return ramp, true
}
