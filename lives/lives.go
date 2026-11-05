// Package lives holds the hero's nine lives, the spending of one life per
// death and the respawn at the last checkpoint, as plain values that never
// import raylib. The session's lives live in memory only and end with the
// window; calling New again is the retry. Keeping the rules here makes them
// testable headless and reusable unchanged on every platform.
package lives

import "copycat/leveldata"

// startLives is the number of lives the hero begins each take with.
const startLives = 9

// Lives is the hero's remaining lives and the last checkpoint passed. Its
// values are plain: the caller copies the whole struct and reads it through
// its methods.
type Lives struct {
	remaining  int
	checkpoint leveldata.Point
}

// New returns a fresh session: nine lives and the respawn at the take's
// start, the origin. Calling New again is the retry after a game over.
func New() Lives {
	return Lives{
		remaining:  startLives,
		checkpoint: leveldata.Point{},
	}
}

// Lives reports the remaining lives.
func (l Lives) Lives() int {
	return l.remaining
}

// GameOver reports whether zero lives remain.
func (l Lives) GameOver() bool {
	return l.remaining == 0
}

// Respawn returns the last checkpoint passed, or the take's start before any
// checkpoint has been passed.
func (l Lives) Respawn() leveldata.Point {
	return l.checkpoint
}

// AtCheckpoint returns l with p remembered as the last checkpoint passed,
// replacing the older one.
func (l Lives) AtCheckpoint(p leveldata.Point) Lives {
	l.checkpoint = p
	return l
}

// Die spends one life and reports whether the hero respawns. It returns true
// with the hero respawning at Respawn while lives remain. The ninth death
// returns false: zero lives mean game over instead of a respawn. A death once
// the game is already over spends nothing.
func (l Lives) Die() (Lives, bool) {
	if l.remaining == 0 {
		return l, false
	}
	l.remaining--
	return l, l.remaining > 0
}
