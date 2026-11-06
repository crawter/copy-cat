// Package scene holds the game's screen state machine: the menu, the playing
// state, and the complete and failed outcomes with a clean restart. It is
// pure Go and imports no raylib, so it can be tested headless and reused
// unchanged on Android.
package scene

import (
	"copy-cat/pkg/move"
	"copy-cat/pkg/runner"
)

// State names the screen the game is currently showing.
type State int

const (
	// StateMenu is the opening menu, before a run exists.
	StateMenu State = iota
	// StatePlaying is an active run, where input advances the runner.
	StatePlaying
	// StateComplete is the completion outcome of a finished run.
	StateComplete
	// StateFailed is the failure outcome of a finished run.
	StateFailed
)

// Game is the live state machine. It owns the current screen and, while
// playing, the active run.
type Game struct {
	State State
	Mode  runner.Mode
	Seed  uint64
	Run   *runner.Run
}

// NewGame returns a game sitting at the menu with no run.
func NewGame() *Game {
	return &Game{State: StateMenu}
}

// Start builds a fresh run for seed and mode and enters the playing state.
func (g *Game) Start(seed uint64, mode runner.Mode) {
	g.Seed = seed
	g.Mode = mode
	g.Run = runner.NewRun(seed, mode)
	g.State = StatePlaying
}

// Step forwards in to the active run and moves the game to its outcome once
// the run completes or the crew catches the hero. It changes nothing outside
// the playing state.
func (g *Game) Step(in move.Input, dt float64) {
	if g.State != StatePlaying || g.Run == nil {
		return
	}

	g.Run.Step(in, dt)
	if g.Run.Complete {
		g.State = StateComplete
	} else if g.Run.Caught {
		g.State = StateFailed
	}
}

// Restart builds a fresh run with the same seed and mode and returns to the
// playing state. It works from the complete and failed outcomes only.
func (g *Game) Restart() {
	if g.State != StateComplete && g.State != StateFailed {
		return
	}

	g.Run = runner.NewRun(g.Seed, g.Mode)
	g.State = StatePlaying
}

// ToMenu returns to the menu and drops the active run.
func (g *Game) ToMenu() {
	g.State = StateMenu
	g.Run = nil
}
