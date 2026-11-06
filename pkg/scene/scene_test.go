package scene

import (
	"testing"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/controller"
	"copy-cat/pkg/move"
	"copy-cat/pkg/runner"
)

// testDT is the fixed small step the simulations use so the results are
// deterministic and headless.
const testDT = 1.0 / 120.0

func TestStartFromMenuBeginsPlaying(t *testing.T) {
	g := NewGame()

	if g.State != StateMenu {
		t.Fatalf("NewGame State = %v, want StateMenu", g.State)
	}
	if g.Run != nil {
		t.Fatalf("NewGame Run = %#v, want nil", g.Run)
	}

	g.Start(42, runner.ModeCollection)

	if g.State != StatePlaying {
		t.Fatalf("State after Start = %v, want StatePlaying", g.State)
	}
	if g.Run == nil {
		t.Fatal("Run after Start = nil, want a fresh run")
	}
	if g.Run.Elapsed != 0 {
		t.Fatalf("Elapsed = %v, want 0", g.Run.Elapsed)
	}
	if g.Run.Collected != 0 {
		t.Fatalf("Collected = %d, want 0", g.Run.Collected)
	}
	if g.Run.Caught {
		t.Fatalf("Caught = true, want false")
	}
	if g.Run.Complete {
		t.Fatalf("Complete = true, want false")
	}
	if g.Seed != 42 {
		t.Fatalf("Seed = %d, want 42", g.Seed)
	}
	if g.Mode != runner.ModeCollection {
		t.Fatalf("Mode = %v, want ModeCollection", g.Mode)
	}
}

func TestCompleteTransition(t *testing.T) {
	g := NewGame()
	g.Start(42, runner.ModeCollection)

	stepUntilComplete(t, g)

	if g.Run == nil {
		t.Fatal("Run = nil after completing")
	}
	if !g.Run.Complete {
		t.Fatalf("Run.Complete = false, want true")
	}
	if g.Run.Caught {
		t.Fatalf("Run.Caught = true, want false on the way to completion")
	}
}

func TestFailTransition(t *testing.T) {
	g := NewGame()
	g.Start(7, runner.ModeSurvival)

	stepUntilFailed(t, g)

	if g.Run == nil {
		t.Fatal("Run = nil after failing")
	}
	if !g.Run.Caught {
		t.Fatalf("Run.Caught = false, want true")
	}
	if g.Run.Complete {
		t.Fatalf("Run.Complete = true, want false on the way to failure")
	}
}

func TestRestartRestoresFreshRun(t *testing.T) {
	t.Run("from complete", func(t *testing.T) {
		g := NewGame()
		g.Start(42, runner.ModeCollection)
		stepUntilComplete(t, g)

		g.Restart()

		if g.State != StatePlaying {
			t.Fatalf("State after Restart = %v, want StatePlaying", g.State)
		}
		assertFreshRun(t, g, 42, runner.ModeCollection)
	})

	t.Run("from failed", func(t *testing.T) {
		g := NewGame()
		g.Start(7, runner.ModeSurvival)
		stepUntilFailed(t, g)

		g.Restart()

		if g.State != StatePlaying {
			t.Fatalf("State after Restart = %v, want StatePlaying", g.State)
		}
		assertFreshRun(t, g, 7, runner.ModeSurvival)
	})
}

func TestToMenuDropsRun(t *testing.T) {
	g := NewGame()
	g.Start(42, runner.ModeSurvival)

	if g.State != StatePlaying {
		t.Fatalf("State after Start = %v, want StatePlaying", g.State)
	}

	g.ToMenu()

	if g.State != StateMenu {
		t.Fatalf("State after ToMenu = %v, want StateMenu", g.State)
	}
	if g.Run != nil {
		t.Fatalf("Run after ToMenu = %#v, want nil", g.Run)
	}
}

func stepUntilComplete(t *testing.T, g *Game) {
	t.Helper()

	maxSteps := int(2*conditions.SurvivalTime/testDT) + 2
	steps := 0
	for g.State == StatePlaying && steps < maxSteps {
		g.Step(controller.IdealInput(g.Run), testDT)
		steps++
	}
	if g.State != StateComplete {
		t.Fatalf("State = %v after %d steps, want StateComplete", g.State, steps)
	}
}

func stepUntilFailed(t *testing.T, g *Game) {
	t.Helper()

	maxSteps := int(4*conditions.SurvivalTime/testDT) + 2
	steps := 0
	for g.State == StatePlaying && steps < maxSteps {
		g.Step(move.Input{}, testDT)
		steps++
	}
	if g.State != StateFailed {
		t.Fatalf("State = %v after %d steps, want StateFailed", g.State, steps)
	}
}

func assertFreshRun(t *testing.T, g *Game, wantSeed uint64, wantMode runner.Mode) {
	t.Helper()

	if g.Seed != wantSeed {
		t.Fatalf("Seed = %d, want %d", g.Seed, wantSeed)
	}
	if g.Mode != wantMode {
		t.Fatalf("Mode = %v, want %v", g.Mode, wantMode)
	}
	if g.Run == nil {
		t.Fatal("Run = nil after Restart, want a fresh run")
	}
	if g.Run.Elapsed != 0 {
		t.Fatalf("Elapsed = %v, want 0", g.Run.Elapsed)
	}
	if g.Run.Collected != 0 {
		t.Fatalf("Collected = %d, want 0", g.Run.Collected)
	}
	if g.Run.Caught {
		t.Fatalf("Caught = true, want false")
	}
	if g.Run.Complete {
		t.Fatalf("Complete = true, want false")
	}
	if g.Run.Mode != wantMode {
		t.Fatalf("Run.Mode = %v, want %v", g.Run.Mode, wantMode)
	}
}
