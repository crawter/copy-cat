// Package controller holds the deterministic ideal player for the runner.
// IdealInput steers the hero back to the corridor centre and jumps with a
// fixed lead before every gap, crate and pest, so a headless simulation can
// prove no generated track forces a stumble. It is pure Go and imports no
// raylib.
package controller

import (
	"math"

	"copy-cat/pkg/move"
	"copy-cat/pkg/runner"
	"copy-cat/pkg/track"
)

// jumpLead is how far before an obstacle the ideal controller starts a jump,
// in world units. It is the part of one full jump that remains after the
// widest generated gap, so the hero lands just past the far edge of even the
// widest gap.
const jumpLead = move.MaxJumpDistance - track.MaxGapWidth

// IdealInput returns the deterministic ideal intent for r: steer the hero
// back to X 0 and jump at jumpLead before the next gap, crate or pest.
// Machines are passed under on the ground, so they are never jumped. Equal
// run states always give equal inputs.
func IdealInput(r *runner.Run) move.Input {
	in := move.Input{}

	switch {
	case r.Player.X > 0:
		in.Left = true
	case r.Player.X < 0:
		in.Right = true
	}

	if r.Player.OnGround {
		next := nextJumpObstacleZ(r)
		if next > r.Player.Z && next-r.Player.Z <= jumpLead {
			in.Jump = true
		}
	}

	return in
}

// nextJumpObstacleZ returns the Z of the nearest obstacle ahead of the hero
// that must be jumped: the start of a gap, or the position of a crate or a
// pest. Machines are passed under and are ignored. It returns +Inf when there
// is nothing left to jump.
func nextJumpObstacleZ(r *runner.Run) float64 {
	z := r.Player.Z
	best := math.Inf(1)

	for _, g := range r.Track.Gaps {
		if g.Start > z && g.Start < best {
			best = g.Start
		}
	}
	for _, c := range r.Track.Crates {
		if c.Z > z && c.Z < best {
			best = c.Z
		}
	}
	for _, e := range r.Track.Enemies {
		if e.Kind == track.EnemyPest && e.Z > z && e.Z < best {
			best = e.Z
		}
	}

	return best
}
