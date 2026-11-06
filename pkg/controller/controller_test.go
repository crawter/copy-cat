package controller

import (
	"testing"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/runner"
)

// testDT is the fixed small step the simulations use so the results are
// deterministic and headless.
const testDT = 1.0 / 120.0

// fixedSeeds is the seed set shared with the track generator: 0, 1, 7, 13,
// 42 and 99.
var fixedSeeds = []uint64{0, 1, 7, 13, 42, 99}

func TestSurvivalRunHasNoForcedStumble(t *testing.T) {
	maxSteps := int(conditions.SurvivalTime/testDT) + 2

	for _, seed := range fixedSeeds {
		r := runner.NewRun(seed, runner.ModeSurvival)
		steps := 0
		for r.Elapsed < conditions.SurvivalTime && !r.Caught && steps < maxSteps {
			r.Step(IdealInput(r), testDT)
			steps++
		}

		if r.Caught {
			t.Fatalf("seed %d: caught after %d steps (%.2fs)", seed, steps, r.Elapsed)
		}
		if r.StumbleCount != 0 {
			t.Fatalf("seed %d: StumbleCount = %d, want 0", seed, r.StumbleCount)
		}
		if !r.Complete {
			t.Fatalf("seed %d: Complete = false after %d steps (%.2fs), want true", seed, steps, r.Elapsed)
		}
	}
}

func TestIdealInputDeterministic(t *testing.T) {
	a := runner.NewRun(42, runner.ModeSurvival)
	b := runner.NewRun(42, runner.ModeSurvival)

	for i := 0; i < 500; i++ {
		ia := IdealInput(a)
		ib := IdealInput(b)
		if ia != ib {
			t.Fatalf("step %d: IdealInput differs between identical runs: %+v vs %+v", i, ia, ib)
		}
		a.Step(ia, testDT)
		b.Step(ib, testDT)
	}
}
