// Package enemy holds Take 9's one enemy, the derailed camera crane. It owns
// the crane's back-and-forth patrol along its broken track, the defeat when
// the hero jumps onto its counterweight, and the dropped film reel, as plain
// values that never import raylib. That keeps the enemy rules testable
// headless and reusable unchanged on every platform.
package enemy

import (
	"copycat/collide"
	"copycat/leveldata"
	"copycat/move"
)

// The crane rolls along Z at four meters a second between the track's two
// ends. The rolling body is a box of half-length 1.5 m, half-width 1.0 m and
// height 0.4 m. The counterweight is a box of half-extent 0.6 m centered on
// top of the body, its bottom at the body's top. The hero's box is the same
// size collide uses: half-width 0.4 m, half-height 0.3 m and half-depth 0.3 m.
const (
	patrolSpeedMetersPerSecond = 4.0

	bodyHalfLenMeters   = 1.5
	bodyHalfWidthMeters = 1.0
	bodyHeightMeters    = 0.4

	counterweightHalfMeters = 0.6

	heroHalfWidthMeters  = 0.4
	heroHalfDepthMeters  = 0.3
	heroHalfHeightMeters = 0.3
)

// Box is an axis-aligned box centered at Pos with the given half-extents along
// X, Y and Z.
type Box struct {
	Pos                 leveldata.Point
	HalfX, HalfY, HalfZ float64
}

// Crane is the derailed camera crane. pos is its patrol position, dir is its
// patrol direction along Z (+1 forward, -1 back), defeated records whether
// the hero has jumped onto its counterweight, and lowZ and highZ are the
// track's two ends ordered along Z so Step knows where to reverse.
type Crane struct {
	pos      leveldata.Point
	dir      float64
	defeated bool
	lowZ     float64
	highZ    float64
}

// New returns a crane at track's start, ready to patrol between the track's
// two ends along Z.
func New(track leveldata.Track) Crane {
	lowZ, highZ := track.Start.Z, track.End.Z
	if highZ < lowZ {
		lowZ, highZ = highZ, lowZ
	}
	dir := 1.0
	if track.End.Z < track.Start.Z {
		dir = -1.0
	}
	return Crane{
		pos:   track.Start,
		dir:   dir,
		lowZ:  lowZ,
		highZ: highZ,
	}
}

// Step advances c one frame of dt seconds and returns the result. A living
// crane rolls along Z at patrolSpeedMetersPerSecond and reverses direction at
// each end of the track, clamping to the end it reaches. A defeated crane
// does not move: its position and direction come back unchanged.
func Step(c Crane, dt float64) Crane {
	if c.defeated || dt <= 0 {
		return c
	}

	c.pos.Z += c.dir * patrolSpeedMetersPerSecond * dt
	if c.dir > 0 && c.pos.Z >= c.highZ {
		c.pos.Z = c.highZ
		c.dir = -1
	} else if c.dir < 0 && c.pos.Z <= c.lowZ {
		c.pos.Z = c.lowZ
		c.dir = 1
	}
	return c
}

// Body returns the crane's rolling body box for collide.Check. The body is
// bodyHalfLenMeters half-length along Z, bodyHalfWidthMeters half-width along
// X and bodyHeightMeters high, centered at the crane's position.
func (c Crane) Body() collide.Crane {
	return collide.Crane{
		Pos:       c.pos,
		HalfLen:   bodyHalfLenMeters,
		HalfWidth: bodyHalfWidthMeters,
		Height:    bodyHeightMeters,
	}
}

// Defeat reports whether h lands on c's counterweight and defeats it. When
// the hero's box overlaps the counterweight box, the crane is set down and
// stays down for the rest of the take; the returned crane then reports
// Dropped. A crane that is already down returns unchanged with false, so the
// drop is only ever signalled once.
func Defeat(h move.Hero, c Crane) (Crane, bool) {
	if c.defeated {
		return c, false
	}
	if boxesOverlap(heroBox(h), c.counterweight()) {
		c.defeated = true
		return c, true
	}
	return c, false
}

// Dropped reports whether the crane has been defeated and dropped its one
// film reel. It is false before the defeat and true from the defeat on.
func (c Crane) Dropped() bool {
	return c.defeated
}

// counterweight returns the box the hero must land on to defeat c: a box of
// half-extent counterweightHalfMeters centered on top of the body, with its
// bottom at the body's top.
func (c Crane) counterweight() Box {
	return Box{
		Pos: leveldata.Point{
			X: c.pos.X,
			Y: c.pos.Y + bodyHeightMeters + counterweightHalfMeters,
			Z: c.pos.Z,
		},
		HalfX: counterweightHalfMeters,
		HalfY: counterweightHalfMeters,
		HalfZ: counterweightHalfMeters,
	}
}

// heroBox returns the hero's box as collide measures it.
func heroBox(h move.Hero) Box {
	return Box{
		Pos:   leveldata.Point{X: h.X, Y: h.Y, Z: h.Z},
		HalfX: heroHalfWidthMeters,
		HalfY: heroHalfHeightMeters,
		HalfZ: heroHalfDepthMeters,
	}
}

// boxesOverlap reports whether two axis-aligned boxes overlap, treating mere
// touching at a face as no overlap.
func boxesOverlap(a, b Box) bool {
	return a.Pos.X-a.HalfX < b.Pos.X+b.HalfX && b.Pos.X-b.HalfX < a.Pos.X+a.HalfX &&
		a.Pos.Y-a.HalfY < b.Pos.Y+b.HalfY && b.Pos.Y-b.HalfY < a.Pos.Y+a.HalfY &&
		a.Pos.Z-a.HalfZ < b.Pos.Z+b.HalfZ && b.Pos.Z-b.HalfZ < a.Pos.Z+a.HalfZ
}
