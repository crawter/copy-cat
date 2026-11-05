// Package collide checks the hero against the corridor's walls, the fallen
// light rigs, the camera crane, the film reels, the checkpoints and the exit,
// and reports each touch as a plain event. It never imports raylib, so the
// rules stay testable headless and reusable unchanged on every platform.
package collide

import (
	"copycat/leveldata"
	"copycat/move"
)

// The hero's box is a half-width of 0.4 m sideways, a half-depth of 0.3 m
// forward and a 0.6 m height around its position, so the half-height is 0.3 m.
const (
	heroHalfWidth  = 0.4
	heroHalfDepth  = 0.3
	heroHalfHeight = 0.3
)

// A fallen light rig stands 0.6 m high with a half-depth of 0.3 m.
const (
	rigHeight    = 0.6
	rigHalfDepth = 0.3
)

// A film reel is a box of half-extent 0.5 m around its position, and a
// checkpoint is passed within 0.5 m of its Z.
const (
	reelHalfExtent  = 0.5
	checkpointReach = 0.5
)

// EventKind is the kind of touch collide reports.
type EventKind int

// The kinds of touch collide reports.
const (
	// EventRigHit reports the hero hitting a fallen light rig.
	EventRigHit EventKind = iota
	// EventCraneHit reports the hero hitting the camera crane's rolling body.
	EventCraneHit
	// EventReel reports the hero touching a film reel; Index names the reel.
	EventReel
	// EventCheckpoint reports the hero passing a checkpoint; Index names it.
	EventCheckpoint
	// EventExit reports the hero crossing the exit line.
	EventExit
)

// Event is one touch of the frame. Kind says what touched. Index names the
// reel or checkpoint in the take's slices for the reel and checkpoint kinds;
// it is zero for the other kinds.
type Event struct {
	Kind  EventKind
	Index int
}

// Crane is the camera crane's rolling body, the part the hero must not touch.
// Pos is its center. HalfLen is its half length along Z, HalfWidth its half
// width along X, and Height how high it stands above the floor, all in meters.
type Crane struct {
	Pos                        leveldata.Point
	HalfLen, HalfWidth, Height float64
}

// Check clamps the hero inside the corridor's walls and reports every touch
// of this frame. The hero's X is clamped between minus and plus the take's
// half width minus the hero's half extent, so he can never steer off the set.
// The touches are reported in one fixed order: a rig hit, a crane hit, then
// each touched reel, then each passed checkpoint in the take's slice order,
// then the exit. The returned hero is the clamped hero, and events is empty
// when nothing touches.
func Check(h move.Hero, t leveldata.Take, crane Crane) (move.Hero, []Event) {
	halfWidth := t.Width / 2
	maxX := halfWidth - heroHalfWidth
	if h.X > maxX {
		h.X = maxX
	} else if h.X < -maxX {
		h.X = -maxX
	}

	hMinX, hMaxX, hMinY, hMaxY, hMinZ, hMaxZ := heroBox(h)

	events := []Event{}

	for _, rig := range t.Rigs {
		if overlaps(hMinX, hMaxX, -halfWidth, halfWidth) &&
			overlaps(hMinY, hMaxY, 0, rigHeight) &&
			overlaps(hMinZ, hMaxZ, rig.Position.Z-rigHalfDepth, rig.Position.Z+rigHalfDepth) {
			events = append(events, Event{Kind: EventRigHit})
			break
		}
	}

	if overlaps(hMinX, hMaxX, crane.Pos.X-crane.HalfWidth, crane.Pos.X+crane.HalfWidth) &&
		overlaps(hMinY, hMaxY, 0, crane.Height) &&
		overlaps(hMinZ, hMaxZ, crane.Pos.Z-crane.HalfLen, crane.Pos.Z+crane.HalfLen) {
		events = append(events, Event{Kind: EventCraneHit})
	}

	for i, reel := range t.Reels {
		if overlaps(hMinX, hMaxX, reel.Position.X-reelHalfExtent, reel.Position.X+reelHalfExtent) &&
			overlaps(hMinY, hMaxY, reel.Position.Y-reelHalfExtent, reel.Position.Y+reelHalfExtent) &&
			overlaps(hMinZ, hMaxZ, reel.Position.Z-reelHalfExtent, reel.Position.Z+reelHalfExtent) {
			events = append(events, Event{Kind: EventReel, Index: i})
		}
	}

	for i, cp := range t.Checkpoints {
		if h.Z >= cp.Position.Z-checkpointReach && h.Z <= cp.Position.Z+checkpointReach {
			events = append(events, Event{Kind: EventCheckpoint, Index: i})
		}
	}

	if h.Z >= t.Exit.Z {
		events = append(events, Event{Kind: EventExit})
	}

	return h, events
}

// heroBox returns the hero's box as its minima and maxima on each axis.
func heroBox(h move.Hero) (minX, maxX, minY, maxY, minZ, maxZ float64) {
	return h.X - heroHalfWidth, h.X + heroHalfWidth,
		h.Y - heroHalfHeight, h.Y + heroHalfHeight,
		h.Z - heroHalfDepth, h.Z + heroHalfDepth
}

// overlaps reports whether two ranges on one axis overlap, treating mere
// touching at an endpoint as no overlap.
func overlaps(aMin, aMax, bMin, bMax float64) bool {
	return aMin < bMax && bMin < aMax
}
