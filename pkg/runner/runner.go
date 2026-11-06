// Package runner simulates one corridor run: it advances the hero with
// move.Step, resolves gaps, crates, enemies and jar lids on the generated
// track, and drives the closing crew chase. It is pure Go and imports no
// raylib, so it can be tested headless and reused unchanged on Android.
package runner

import (
	"math"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/move"
	"copy-cat/pkg/track"
)

// Mode names the win condition of one level.
type Mode int

const (
	// ModeSurvival completes the run when the hero survives the set time.
	ModeSurvival Mode = iota
	// ModeCollection completes the run when the hero collects the set quota
	// of jar lids.
	ModeCollection
)

// ChaseStartGap is how far behind the hero the crew starts, in world units.
// A clean run keeps pulling away, so the chase only closes after stumbles.
const ChaseStartGap = 12.0

// CrewSpeed is how fast the crew advances along the corridor, in world units
// per second. It is slower than a clean run and faster than a stumble, so the
// crew closes the gap only while the hero stumbles.
const CrewSpeed = 0.9 * move.RunSpeed

// contactHalfWidth is how close the hero's X must be to an obstacle's X for
// the two to touch, in world units. It leaves the hero room to steer around
// a crate or an enemy.
const contactHalfWidth = 0.5

// Run is one playthrough of a generated track: the hero's live state, the
// progress of the two level conditions, and the chase.
type Run struct {
	Track        track.Track
	Player       move.State
	Mode         Mode
	Elapsed      float64
	Collected    int
	Broken       int
	Smashed      int
	EnemyHits    int
	StumbleCount int
	CrewZ        float64
	Caught       bool
	Complete     bool
}

// NewRun builds a fresh run for seed and mode. It generates the track with
// track.Generate(seed) and places the crew ChaseStartGap behind the hero.
func NewRun(seed uint64, mode Mode) *Run {
	return &Run{
		Track:  track.Generate(seed),
		Player: move.State{OnGround: true},
		Mode:   mode,
		CrewZ:  -ChaseStartGap,
	}
}

// Step advances the run by dt seconds using in as the hero's intent. It
// resolves the track, advances the crew and updates the level condition. Once
// the hero is caught or the run is complete, further steps change nothing.
func (r *Run) Step(in move.Input, dt float64) {
	if r.Caught || r.Complete {
		return
	}

	prevZ := r.Player.Z
	move.Step(&r.Player, in, dt)
	r.Elapsed += dt

	r.resolveObstacles(prevZ)

	r.CrewZ += CrewSpeed * dt
	r.wrapLoop()
	if r.CrewZ >= r.Player.Z {
		r.Caught = true
	}

	if !r.Caught {
		r.updateComplete()
	}
}

// wrapLoop keeps the hero and the crew inside one loop of the corridor. When
// the hero passes the end of the track, both move back by the track length,
// so the corridor wraps seamlessly.
func (r *Run) wrapLoop() {
	for r.Track.Length > 0 && r.Player.Z > r.Track.Length {
		r.Player.Z -= r.Track.Length
		r.CrewZ -= r.Track.Length
	}
}

// updateComplete sets Complete from the run's mode: survival reads the elapsed
// time, collection reads the collected jar lids.
func (r *Run) updateComplete() {
	switch r.Mode {
	case ModeSurvival:
		r.Complete = conditions.SurvivalMet(r.Elapsed)
	case ModeCollection:
		r.Complete = conditions.CollectionMet(r.Collected)
	}
}

// resolveObstacles handles every track feature the hero reached this step.
// prevZ is the hero's Z before move.Step, used to detect one crossing per
// feature.
func (r *Run) resolveObstacles(prevZ float64) {
	z := r.Player.Z

	// A grounded hero over a hole falls in, stumbles and is put back at the
	// gap start. Airborne heroes jump the gap untouched.
	for _, g := range r.Track.Gaps {
		if r.Player.OnGround && z > g.Start && z < g.End {
			r.Player.Z = g.Start
			z = g.Start
			move.Stumble(&r.Player)
			r.StumbleCount++
			break
		}
	}

	// A grounded hero hits a crate sideways and stumbles. An airborne hero
	// lands on it, breaks it and collects the jar lid inside.
	for i := 0; i < len(r.Track.Crates); {
		c := r.Track.Crates[i]
		remove := false
		if crossed(prevZ, z, c.Z) && nearX(r.Player.X, c.X) {
			if r.Player.OnGround {
				move.Stumble(&r.Player)
				r.StumbleCount++
			} else {
				r.Broken++
				r.Collected++
				r.removeJunkAt(c.X, c.Z)
				remove = true
			}
		}
		if remove {
			r.Track.Crates = append(r.Track.Crates[:i], r.Track.Crates[i+1:]...)
			continue
		}
		i++
	}

	// A grounded hero running into a pest stumbles and takes one enemy hit. An
	// airborne hero landing on a pest smashes it. Machines only touch an
	// airborne hero; a grounded hero passes under them.
	for i := 0; i < len(r.Track.Enemies); {
		e := r.Track.Enemies[i]
		remove := false
		if crossed(prevZ, z, e.Z) && nearX(r.Player.X, e.X) {
			switch e.Kind {
			case track.EnemyPest:
				if r.Player.OnGround {
					move.Stumble(&r.Player)
					r.EnemyHits++
					r.StumbleCount++
				} else {
					r.Smashed++
					remove = true
				}
			case track.EnemyMachine:
				if !r.Player.OnGround {
					r.Smashed++
					remove = true
				}
			}
		}
		if remove {
			r.Track.Enemies = append(r.Track.Enemies[:i], r.Track.Enemies[i+1:]...)
			continue
		}
		i++
	}

	// Loose jar lids are collected one each as the hero runs through them. A
	// lid still sitting on an unbroken crate is collected with the crate, not
	// here.
	for i := 0; i < len(r.Track.Junk); {
		j := r.Track.Junk[i]
		if crossed(prevZ, z, j.Z) && nearX(r.Player.X, j.X) && !r.junkInsideCrate(j) {
			r.Collected++
			r.Track.Junk = append(r.Track.Junk[:i], r.Track.Junk[i+1:]...)
			continue
		}
		i++
	}
}

// removeJunkAt removes one jar lid at exactly x, z: the lid carried by a
// broken crate.
func (r *Run) removeJunkAt(x, z float64) {
	for i, j := range r.Track.Junk {
		if j.X == x && j.Z == z {
			r.Track.Junk = append(r.Track.Junk[:i], r.Track.Junk[i+1:]...)
			return
		}
	}
}

// junkInsideCrate reports whether j is the jar lid carried by an unbroken
// crate, which is collected only when the crate breaks.
func (r *Run) junkInsideCrate(j track.Junk) bool {
	for _, c := range r.Track.Crates {
		if c.X == j.X && c.Z == j.Z {
			return true
		}
	}
	return false
}

// crossed reports whether the hero's forward position moved across obstacleZ
// during this step, from prevZ to z.
func crossed(prevZ, z, obstacleZ float64) bool {
	return prevZ <= obstacleZ && obstacleZ <= z
}

// nearX reports whether two sideways positions are close enough to touch.
func nearX(a, b float64) bool {
	return math.Abs(a-b) <= contactHalfWidth
}
