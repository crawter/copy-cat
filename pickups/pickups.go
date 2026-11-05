// Package pickups tracks the film reels Шкряб has collected in Take 9 and the
// found-of-twenty count. It works on plain values and never imports raylib, so
// the rules stay testable headless and reusable unchanged on every platform.
package pickups

// PlacedReels is the number of reels leveldata places in Take 9: nineteen,
// laid out in the take's Reels slice order as ids 0 through 18.
const PlacedReels = 19

// DroppedReelID is the id of the twentieth reel, the one the defeated camera
// crane drops. It follows the nineteen placed reels.
const DroppedReelID = 19

// TotalReels is how many reels Take 9 holds: nineteen placed plus the crane's
// drop. The HUD and the recap read found of this total, never twenty-one.
const TotalReels = 20

// Reels tracks which of the take's reels have been collected.
type Reels struct {
	collected [TotalReels]bool
}

// New returns a Reels with nothing collected.
func New() Reels {
	return Reels{}
}

// Collect marks id collected and reports whether this collection was new. It
// returns true only when id was not yet collected; a repeated id returns false
// and leaves the count unchanged. An id outside 0 through DroppedReelID
// returns false and changes nothing.
func (r *Reels) Collect(id int) bool {
	if id < 0 || id > DroppedReelID {
		return false
	}
	if r.collected[id] {
		return false
	}
	r.collected[id] = true
	return true
}

// CollectDropped collects the defeated crane's dropped reel once. It returns
// true on the first call and false on every later call.
func (r *Reels) CollectDropped() bool {
	return r.Collect(DroppedReelID)
}

// Count returns the number of distinct reels collected so far.
func (r Reels) Count() int {
	n := 0
	for _, got := range r.collected {
		if got {
			n++
		}
	}
	return n
}

// Total returns the number of reels the take holds: always TotalReels, twenty.
func (r Reels) Total() int {
	return TotalReels
}

// Collected reports whether id has been collected.
func (r Reels) Collected(id int) bool {
	if id < 0 || id > DroppedReelID {
		return false
	}
	return r.collected[id]
}
