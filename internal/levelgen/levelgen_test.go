package levelgen

import (
	"reflect"
	"testing"
)

var testPresets = []Params{Mangroves(), Harbor(), Neon()}

func testSeeds(n int) []uint64 {
	seeds := make([]uint64, n)
	for i := range seeds {
		seeds[i] = uint64(i)
	}
	return seeds
}

func TestIdenticalSeedsProduceIdenticalLevels(t *testing.T) {
	for _, p := range testPresets {
		a := New(123456789, p)
		b := New(123456789, p)
		if !reflect.DeepEqual(a.elems, b.elems) {
			t.Fatalf("preset %+v: same seed produced different elements", p)
		}
		if a.Seed() != b.Seed() || a.Length() != b.Length() || a.FinishZ() != b.FinishZ() {
			t.Fatalf("preset %+v: basic fields differ", p)
		}
		if a.Constraints() != b.Constraints() {
			t.Fatalf("preset %+v: constraints differ", p)
		}
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	p := Mangroves()
	pairs := [][2]uint64{{1, 2}, {2, 3}, {42, 43}, {100, 200}, {999999, 999998}}
	for _, pair := range pairs {
		a := New(pair[0], p)
		b := New(pair[1], p)
		if reflect.DeepEqual(a.elems, b.elems) {
			t.Fatalf("seeds %d and %d produced identical elements", pair[0], pair[1])
		}
	}
}

func TestElementFieldsValid(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			c := l.Constraints()
			for _, e := range l.elems {
				if e.Lane < 0 || e.Lane >= laneCount {
					t.Fatalf("seed %d: lane %d out of range in %+v", seed, e.Lane, e)
				}
				if e.Z < 0 || e.Z >= l.Length() {
					t.Fatalf("seed %d: z %v out of range in %+v", seed, e.Z, e)
				}
				if e.Span < 0 {
					t.Fatalf("seed %d: negative span %v in %+v", seed, e.Span, e)
				}
				if e.Kind == Pit {
					if e.Span > c.MaxPitWidth {
						t.Fatalf("seed %d: pit %+v wider than MaxPitWidth %v", seed, e, c.MaxPitWidth)
					}
					if e.Z+e.Span > l.Length() {
						t.Fatalf("seed %d: pit %+v runs past the finish", seed, e)
					}
				}
				if e.Kind == Ramp && e.Z+e.Span > l.Length() {
					t.Fatalf("seed %d: ramp %+v runs past the finish", seed, e)
				}
			}
		}
	}
}

func TestConstraintsMatchBounds(t *testing.T) {
	c := New(1, Mangroves()).Constraints()
	if c.MaxPitWidth != maxPitWidth || c.ReactionGap != reactionGap {
		t.Fatalf("Constraints() = %+v, want MaxPitWidth %v and ReactionGap %v", c, maxPitWidth, reactionGap)
	}
}

func TestElementsSortedHalfOpenAndCopied(t *testing.T) {
	l := New(7, Mangroves())
	all := l.Elements(0, l.Length())
	if len(all) != len(l.elems) {
		t.Fatalf("Elements(0, Length) returned %d elements, want %d", len(all), len(l.elems))
	}
	for i := 1; i < len(all); i++ {
		if all[i].Z < all[i-1].Z {
			t.Fatalf("elements not sorted by Z at index %d", i)
		}
	}
	mid := l.Length() / 2
	for _, e := range l.Elements(mid, l.Length()) {
		if e.Z < mid || e.Z >= l.Length() {
			t.Fatalf("element %+v outside half-open range [%v, %v)", e, mid, l.Length())
		}
	}
	if len(all) > 0 {
		before := l.elems[0]
		all[0].Z = -9999
		if l.elems[0] != before {
			t.Fatalf("Elements returned aliased elements")
		}
	}
}

func TestEveryPitHasFlatRunUp(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			for _, pit := range pitsOf(l) {
				if !flatRunUp(l, pit.Z, l.Constraints().ReactionGap) {
					t.Fatalf("seed %d: pit %+v lacks a flat run-up", seed, pit)
				}
			}
		}
	}
}

func TestForcedActionsSpacedByReactionGap(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			forced := forcedOf(l)
			for i := 1; i < len(forced); i++ {
				prev := forced[i-1]
				cur := forced[i]
				_, prevEnd := zSpan(prev)
				if gap := cur.Z - prevEnd; gap < l.Constraints().ReactionGap {
					t.Fatalf("seed %d: forced actions %+v and %+v too close (gap %v)", seed, prev, cur, gap)
				}
			}
		}
	}
}

// A pit spans every lane and is cleared by jumping, so the free-lane window
// invariant covers barriers and enemies; pits are covered by the run-up test.
func TestFreeLaneOfBarriersEnemiesInAnyReactionWindow(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			gap := l.Constraints().ReactionGap
			for z := 0.0; z < l.Length()-gap; z += 0.5 {
				if freeLaneClear(l, z, gap, -1) < 0 {
					t.Fatalf("seed %d: no free lane in window [%v, %v)", seed, z, z+gap)
				}
			}
		}
	}
}

func TestRampSpansHoldNoPitOrSameLaneBarrier(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			for _, ramp := range rampsOf(l) {
				rs, rt := zSpan(ramp)
				for _, e := range l.elems {
					if e.Kind == Ramp {
						continue
					}
					es, et := zSpan(e)
					if !(es < rt && et > rs) {
						continue
					}
					if e.Kind == Pit {
						t.Fatalf("seed %d: ramp %+v overlaps pit %+v", seed, ramp, e)
					}
					if (e.Kind == LowBarrier || e.Kind == HighBarrier) && e.Lane == ramp.Lane {
						t.Fatalf("seed %d: ramp %+v overlaps same-lane barrier %+v", seed, ramp, e)
					}
				}
			}
		}
	}
}

func TestEnemiesNotInsidePitsOrRamps(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			for _, enemy := range enemiesOf(l) {
				for _, e := range l.elems {
					if e.Kind == Pit && enemy.Z >= e.Z && enemy.Z < e.Z+e.Span {
						t.Fatalf("seed %d: enemy %+v inside pit %+v", seed, enemy, e)
					}
					if e.Kind == Ramp && e.Lane == enemy.Lane && enemy.Z >= e.Z && enemy.Z < e.Z+e.Span {
						t.Fatalf("seed %d: enemy %+v inside ramp %+v", seed, enemy, e)
					}
				}
			}
		}
	}
}

func TestNoIllegalOverlap(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(16) {
			l := New(seed, p)
			for i := 0; i < len(l.elems); i++ {
				for j := i + 1; j < len(l.elems); j++ {
					if illegalOverlap(l.elems[i], l.elems[j]) {
						t.Fatalf("seed %d: illegal overlap between %+v and %+v", seed, l.elems[i], l.elems[j])
					}
				}
			}
		}
	}
}

func TestSolverReachesFinishForManySeedsAndPresets(t *testing.T) {
	for _, p := range testPresets {
		for _, seed := range testSeeds(32) {
			l := New(seed, p)
			if !solverReaches(l) {
				t.Fatalf("seed %d preset %+v: solver could not reach the finish", seed, p)
			}
		}
	}
}

func TestPresetsRiseInDifficulty(t *testing.T) {
	m, h, n := Mangroves(), Harbor(), Neon()
	if !(m.BaseSpeed < h.BaseSpeed && h.BaseSpeed < n.BaseSpeed) {
		t.Fatalf("base speed not rising: %v %v %v", m.BaseSpeed, h.BaseSpeed, n.BaseSpeed)
	}
	if !(m.Density < h.Density && h.Density < n.Density) {
		t.Fatalf("density not rising: %v %v %v", m.Density, h.Density, n.Density)
	}
	if !(m.Length < h.Length && h.Length < n.Length) {
		t.Fatalf("length not rising: %v %v %v", m.Length, h.Length, n.Length)
	}
}

func TestZeroParamsProducesValidLevel(t *testing.T) {
	l := New(0, Params{})
	if l.Length() < minLength {
		t.Fatalf("Length() = %v, want at least %v", l.Length(), minLength)
	}
	if l.FinishZ() != l.Length() {
		t.Fatalf("FinishZ() = %v, want %v", l.FinishZ(), l.Length())
	}
	for _, e := range l.elems {
		if e.Z < 0 || e.Z >= l.Length() {
			t.Fatalf("element %+v out of level", e)
		}
	}
}

func TestPresetLevelsContainObstaclesPickupsAndRamps(t *testing.T) {
	for _, p := range testPresets {
		l := New(1234, p)
		if len(forcedOf(l)) == 0 {
			t.Fatalf("preset %+v: no forced actions", p)
		}
		if len(pickupsOf(l)) == 0 {
			t.Fatalf("preset %+v: no pickups", p)
		}
		if len(rampsOf(l)) == 0 {
			t.Fatalf("preset %+v: no ramps", p)
		}
	}
}

func forcedOf(l *Level) []Element {
	var out []Element
	for _, e := range l.elems {
		if isForced(e.Kind) {
			out = append(out, e)
		}
	}
	return out
}

func rampsOf(l *Level) []Element {
	var out []Element
	for _, e := range l.elems {
		if e.Kind == Ramp {
			out = append(out, e)
		}
	}
	return out
}

func enemiesOf(l *Level) []Element {
	var out []Element
	for _, e := range l.elems {
		if e.Kind == Enemy {
			out = append(out, e)
		}
	}
	return out
}

func pitsOf(l *Level) []Element {
	var out []Element
	for _, e := range l.elems {
		if e.Kind == Pit {
			out = append(out, e)
		}
	}
	return out
}

func pickupsOf(l *Level) []Element {
	var out []Element
	for _, e := range l.elems {
		if isPickup(e.Kind) {
			out = append(out, e)
		}
	}
	return out
}

// illegalOverlap reports whether two elements occupy the same space in a way
// the generator forbids.
func illegalOverlap(a, b Element) bool {
	if isPickup(a.Kind) && isPickup(b.Kind) {
		if a.Lane != b.Lane {
			return false
		}
		d := a.Z - b.Z
		if d < 0 {
			d = -d
		}
		return d < pickupSpace
	}
	if isPickup(a.Kind) {
		return pickupBlocked(a, b)
	}
	if isPickup(b.Kind) {
		return pickupBlocked(b, a)
	}
	as, at := zSpan(a)
	bs, bt := zSpan(b)
	if at <= bs || bt <= as {
		return false
	}
	if a.Kind == Pit || b.Kind == Pit {
		return true
	}
	return a.Lane == b.Lane
}

// pickupBlocked reports whether pickup p sits inside blocker b's footprint.
func pickupBlocked(p, b Element) bool {
	bs, bt := zSpan(b)
	if p.Z < bs || p.Z >= bt {
		return false
	}
	if b.Kind == Pit {
		return true
	}
	return p.Lane == b.Lane
}

// flatRunUp reports whether no forced action overlaps [pitZ-gap, pitZ).
func flatRunUp(l *Level, pitZ, gap float64) bool {
	start := pitZ - gap
	for _, e := range l.elems {
		if !isForced(e.Kind) {
			continue
		}
		s, t := zSpan(e)
		if s < pitZ && t > start {
			return false
		}
	}
	return true
}

// solverReaches proves a collision-free plan exists using only the exported
// Constraints bounds plus the element layout.
func solverReaches(l *Level) bool {
	c := l.Constraints()
	lane := 1
	z := 0.0
	for {
		next, ok := nextForced(l, z)
		if !ok {
			return true
		}
		if next.Kind == Pit {
			if next.Span > c.MaxPitWidth {
				return false
			}
			if !flatRunUp(l, next.Z, c.ReactionGap) {
				return false
			}
			z = next.Z + next.Span
			continue
		}
		if next.Lane == lane {
			free := freeLaneClear(l, next.Z, c.ReactionGap, lane)
			if free < 0 {
				return false
			}
			lane = free
		}
		z = next.Z + barrierDepth
	}
}

// nextForced returns the earliest forced action whose footprint is not yet
// fully passed.
func nextForced(l *Level, z float64) (Element, bool) {
	for _, e := range l.elems {
		if !isForced(e.Kind) {
			continue
		}
		_, trail := zSpan(e)
		if trail > z {
			return e, true
		}
	}
	return Element{}, false
}

// freeLaneClear returns a lane other than exclude that stays clear of barriers
// and enemies across the whole window [z, z+gap), or -1.
func freeLaneClear(l *Level, z, gap float64, exclude int) int {
	for lane := 0; lane < laneCount; lane++ {
		if lane == exclude {
			continue
		}
		if laneClearWindow(l, lane, z, z+gap) {
			return lane
		}
	}
	return -1
}

func laneClearWindow(l *Level, lane int, start, end float64) bool {
	for _, e := range l.elems {
		if e.Kind != LowBarrier && e.Kind != HighBarrier && e.Kind != Enemy {
			continue
		}
		if e.Lane != lane {
			continue
		}
		s, t := zSpan(e)
		if s < end && t > start {
			return false
		}
	}
	return true
}
