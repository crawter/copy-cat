// Package gameplay applies the run's rules on top of the runner: hits,
// stomps, the Copycat chase with its snip, pickups, lives and death causes.
//
// The runner simulates movement only and reports events; gameplay reads those
// events and turns them into outcomes. It is pure Go and imports only the
// runner and levelgen packages, so its tests run headless.
package gameplay

import (
	"math"

	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

// Cause names the last damage or death cause of a run.
type Cause int

// The causes a run can record.
const (
	CauseNone Cause = iota
	CausePit
	CauseBarrier
	CauseEnemy
	CauseSnip
)

// State is the current state of a run.
type State int

// The states a run can be in. StateDead is the game-over condition the
// campaign reads.
const (
	StateRunning State = iota
	StateDead
	StateFinished
)

// Outcome names one thing that happened during a run step.
type Outcome int

// The outcomes a run reports.
const (
	OutcomeStumble Outcome = iota
	OutcomeSnip
	OutcomePit
	OutcomeStomp
	OutcomeFinished
	OutcomeDied
	OutcomeShieldSaved
	OutcomeComb
	OutcomeHairspray
	OutcomeCheckpoint
	OutcomeLife
)

// Run owns the rules the runner reports into: hits become stumbles, a close
// chase turns a second stumble into a snip, pits and snips cost lives, stomps
// defeat enemies, and pickups grant combs, a shield and checkpoints. The boss
// fight and the campaign build on it unchanged.
type Run struct {
	// Runner is the movement simulation the run steps.
	Runner *runner.Runner
	// Level is the level the runner runs on.
	Level *levelgen.Level
	// Lives is how many lives remain; 0 means the run is dead.
	Lives int
	// Combs is how many combs the run has collected toward the next extra
	// life. Every 100 combs grant one life, keeping the remainder.
	Combs int
	// ShieldHits is how many snips a hairspray shield still absorbs.
	ShieldHits int
	// Chase is how close Copycat is, from 0 (far) to 1 (close). A stumble
	// while it is above 0.5 becomes a snip.
	Chase float64
	// Bald reports whether a snip has shaved the hero. Only a mirror clears
	// it later.
	Bald bool
	// CheckpointZ is where a lost life respawns the hero. It starts at the
	// level start and a mirror updates it later.
	CheckpointZ float64
	// Cause records the last damage or death cause.
	Cause Cause
	// State is the run's current state.
	State State

	// pickups is the level's pickup elements in level order, kept so Step
	// never allocates to scan them.
	pickups []levelgen.Element
	// collected records which pickups the run already collected, keyed by
	// kind, lane and Z, so a respawn never collects one twice.
	collected map[pickupKey]bool
}

// NewRun returns a run driving r over level with lives starting lives. The
// hero starts intact: no shield, no chase, no combs, hair on, at the level
// start.
func NewRun(r *runner.Runner, level *levelgen.Level, lives int) *Run {
	if lives < 0 {
		lives = 0
	}
	return &Run{
		Runner:      r,
		Level:       level,
		Lives:       lives,
		CheckpointZ: 0,
		State:       StateRunning,
		pickups:     pickupsOf(level),
		collected:   make(map[pickupKey]bool),
	}
}

// Step advances the run by dt seconds with input in and returns every outcome
// that happened this step, in order. When State is StateDead or StateFinished
// it changes nothing and returns no outcomes.
func (r *Run) Step(in runner.Input, dt float64) []Outcome {
	if r.State != StateRunning {
		return nil
	}

	events := r.Runner.Step(in, dt)
	var out []Outcome
	out = append(out, r.collectPickups()...)

	stumbled := false
	finished := false

	for _, ev := range events {
		if finished {
			break
		}
		switch ev {
		case runner.EventFinished:
			r.State = StateFinished
			out = append(out, OutcomeFinished)
			finished = true
		case runner.EventHitBarrier:
			stumbled = true
			out = append(out, r.stumble(CauseBarrier)...)
		case runner.EventHitEnemy:
			stumbled = true
			out = append(out, r.stumble(CauseEnemy)...)
		case runner.EventFellInPit:
			out = append(out, r.fellInPit()...)
		case runner.EventStomp:
			out = append(out, OutcomeStomp)
		}
	}

	if !finished && !stumbled && r.State == StateRunning {
		r.decayChase(dt)
	}
	return out
}

// collectPickups collects the pickups the hero currently reaches. It runs
// right after the runner moves and before the events become outcomes, so a
// pickup the hero crosses on the way into a pit still counts.
func (r *Run) collectPickups() []Outcome {
	z := r.Runner.Z()
	lane := laneOf(r.Runner.Lane())
	var out []Outcome
	for _, e := range r.pickups {
		if e.Lane != lane {
			continue
		}
		if math.Abs(e.Z-z) > pickupReach {
			continue
		}
		key := pickupKey{kind: e.Kind, lane: e.Lane, z: e.Z}
		if r.collected[key] {
			continue
		}
		r.collected[key] = true
		switch e.Kind {
		case levelgen.Comb:
			r.Combs++
			out = append(out, OutcomeComb)
			if r.Combs >= combsPerLife {
				r.Combs -= combsPerLife
				r.Lives++
				out = append(out, OutcomeLife)
			}
		case levelgen.Hairspray:
			r.ShieldHits = shieldHitsCap
			out = append(out, OutcomeHairspray)
		case levelgen.Mirror:
			r.CheckpointZ = e.Z
			r.Bald = false
			out = append(out, OutcomeCheckpoint)
		}
	}
	return out
}

// stumble handles a barrier or enemy hit. A stumble while the chase is low
// starts the chase without costing a life; a stumble while it is high becomes
// a snip.
func (r *Run) stumble(cause Cause) []Outcome {
	if r.Chase > chaseThreshold {
		return r.snip(cause)
	}
	r.Cause = cause
	r.Chase = 1
	r.resumeAfterStumble()
	return []Outcome{OutcomeStumble}
}

// snip applies the Copycat snip: a life is lost and the hero turns bald, or
// the shield absorbs the snip instead.
func (r *Run) snip(cause Cause) []Outcome {
	if r.ShieldHits > 0 {
		r.ShieldHits--
		r.Chase = 0
		r.Cause = cause
		r.resumeAfterStumble()
		return []Outcome{OutcomeShieldSaved}
	}
	r.loseLife()
	r.Bald = true
	r.Chase = 0
	r.Cause = CauseSnip
	r.respawn()
	out := []Outcome{OutcomeSnip}
	if r.Lives <= 0 {
		r.State = StateDead
		out = append(out, OutcomeDied)
	}
	return out
}

// fellInPit applies a pit fall: a life is lost and the hero respawns at the
// checkpoint. The shield never stops a pit.
func (r *Run) fellInPit() []Outcome {
	r.loseLife()
	r.Chase = 0
	r.Cause = CausePit
	r.respawn()
	out := []Outcome{OutcomePit}
	if r.Lives <= 0 {
		r.State = StateDead
		out = append(out, OutcomeDied)
	}
	return out
}

// loseLife decrements Lives without letting it fall below zero.
func (r *Run) loseLife() {
	if r.Lives > 0 {
		r.Lives--
	}
}

// respawn restarts the runner at CheckpointZ.
func (r *Run) respawn() {
	r.Runner.Restart(r.CheckpointZ)
}

// resumeAfterStumble restarts the runner just past the obstacle the hero
// stumbled into, so the run keeps going forward without re-colliding. A
// stumble does not respawn at the checkpoint.
func (r *Run) resumeAfterStumble() {
	r.Runner.Restart(r.obstacleEnd())
}

// obstacleEnd returns the z just past the barrier or enemy the hero's front
// edge is currently inside, falling back to a short step forward when none is
// found.
func (r *Run) obstacleEnd() float64 {
	z := r.Runner.Z()
	lane := laneOf(r.Runner.Lane())
	for _, e := range r.Level.Elements(z-obstacleDepth, z+obstacleDepth) {
		if e.Lane != lane {
			continue
		}
		switch e.Kind {
		case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
			return e.Z + obstacleDepth
		}
	}
	return z + obstacleDepth
}

// decayChase lowers Chase over chaseDuration seconds of clean running. It
// stops at 0.
func (r *Run) decayChase(dt float64) {
	if r.Chase <= 0 || dt <= 0 {
		return
	}
	r.Chase -= dt / chaseDuration
	if r.Chase < 0 {
		r.Chase = 0
	}
}

// laneOf returns the lane the hero currently centers on.
func laneOf(lane float64) int {
	return int(math.Round(lane))
}

// pickupKey identifies one placed pickup so a collected pickup is never
// collected twice after a respawn.
type pickupKey struct {
	kind levelgen.Kind
	lane int
	z    float64
}

// pickupsOf returns the level's pickup elements in level order. It copies
// once at construction so Step never allocates while scanning.
func pickupsOf(level *levelgen.Level) []levelgen.Element {
	all := level.Elements(0, level.Length())
	pickups := make([]levelgen.Element, 0, len(all))
	for _, e := range all {
		switch e.Kind {
		case levelgen.Comb, levelgen.Hairspray, levelgen.Mirror:
			pickups = append(pickups, e)
		}
	}
	return pickups
}

const (
	// chaseThreshold is the chase level at or above which a stumble becomes a
	// snip instead of starting the chase.
	chaseThreshold = 0.5
	// chaseDuration is how many seconds the chase takes to decay from 1 to 0.
	chaseDuration = 4.0
	// obstacleDepth matches the barrier and enemy footprint the runner and
	// levelgen use, in meters.
	obstacleDepth = 1.0
)

const (
	// pickupReach is how close the hero's front edge must be to a pickup's Z,
	// in meters, to collect it.
	pickupReach = 0.5
	// combsPerLife is how many combs grant one extra life.
	combsPerLife = 100
	// shieldHitsCap is the most ShieldHits a hairspray can grant; the shield
	// never stacks beyond it.
	shieldHitsCap = 1
)
