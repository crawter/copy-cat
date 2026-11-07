// Package levelgen generates deterministic, provably passable three-lane
// corridor levels from a seed.
//
// A level is a corridor running along +z in meters, with three lanes at fixed
// x offsets, from z=0 to a finish at Length. A level is built only from its
// seed and parameters, so the same seed always yields an identical level. The
// package is pure Go: it never imports raylib, audio or assets, so its tests
// run headless.
//
// Generation guarantees passability by construction:
//
//   - No pit is wider than Constraints.MaxPitWidth.
//   - Every pit follows a flat run-up of at least Constraints.ReactionGap.
//   - In any window of Constraints.ReactionGap, at least one lane stays free
//     of barriers and enemies. A pit spans every lane and is cleared by
//     jumping, so pits are covered by the flat run-up rule instead.
//   - Two forced actions (pit, barrier or enemy) are never closer than
//     Constraints.ReactionGap.
//   - A ramp's span holds no pit and no same-lane barrier, and it ends on
//     flat ground.
//   - Enemies sit on lanes only, never inside pits or ramp spans.
//   - Mirrors appear at regular intervals, combs throughout, and the finish
//     is always reachable.
package levelgen

import "sort"

// Kind identifies what an Element is.
type Kind int

// The element kinds a level can hold.
const (
	// Pit is a gap spanning every lane; the hero must jump it. A pit ignores
	// Lane and its Span is its width.
	Pit Kind = iota
	// LowBarrier is a hurdle the hero must jump.
	LowBarrier
	// HighBarrier is a hurdle the hero must slide under.
	HighBarrier
	// Ramp is a container ramp. Running onto it on its lane raises the hero
	// for the ramp's run length, stored in Span.
	Ramp
	// Enemy sits on one lane; the hero may stomp it or dodge onto another lane.
	Enemy
	// Comb is a pickup that counts toward an extra life.
	Comb
	// Hairspray is a pickup that grants a one-hit shield.
	Hairspray
	// Mirror is a checkpoint pickup.
	Mirror
)

// Element is one thing placed in the corridor.
type Element struct {
	Kind Kind
	// Lane is the lane the element occupies, in the range [0, 2]. For a Pit,
	// which spans every lane, Lane is ignored.
	Lane int
	// Z is the element's position along the corridor, in meters.
	Z float64
	// Span is a Pit's width or a Ramp's run length, in meters. Other kinds
	// carry 0.
	Span float64
}

// Params tunes one generated level.
type Params struct {
	// BaseSpeed is the runner's starting forward speed in meters per second.
	BaseSpeed float64
	// Length is the corridor length in meters, from z=0 to the finish.
	Length float64
	// Density is the obstacle density in [0, 1]: higher means more barriers,
	// enemies and pits per meter.
	Density float64
	// Combs, Hairspray and Mirrors are the pickup counts to place.
	Combs     int
	Hairspray int
	Mirrors   int
}

// Constraints states the passability invariants as numbers.
type Constraints struct {
	// MaxPitWidth is the widest pit the generator places and the widest pit
	// the runner's jump must clear.
	MaxPitWidth float64
	// ReactionGap is the minimum flat run-up before a pit, the minimum gap
	// between consecutive forced actions, and the horizon within which at
	// least one lane stays free of barriers and enemies. A pit is cleared by
	// jumping, so it is covered by the run-up rule instead of the free-lane
	// rule.
	ReactionGap float64
}

// Level is a fully generated corridor. All elements are computed eagerly, so
// querying a level never allocates per frame.
type Level struct {
	seed   uint64
	length float64
	cons   Constraints
	elems  []Element // sorted by Z, then Kind, then Lane
}

// New generates the level for seed with parameters p and returns it. The same
// seed and parameters always produce the identical level.
func New(seed uint64, p Params) *Level {
	p = sanitize(p)
	l := &Level{
		seed:   seed,
		length: p.Length,
		cons: Constraints{
			MaxPitWidth: maxPitWidth,
			ReactionGap: reactionGap,
		},
	}
	rng := newRNG(seed)

	forced := l.placeForced(p, rng)
	ramps := l.placeRamps(forced, rng)

	placed := make([]Element, 0, len(forced)+len(ramps))
	placed = append(placed, forced...)
	placed = append(placed, ramps...)
	placed = append(placed, l.placePickups(placed, p, rng)...)

	sort.Slice(placed, func(i, j int) bool {
		a, b := placed[i], placed[j]
		if a.Z != b.Z {
			return a.Z < b.Z
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Lane < b.Lane
	})
	l.elems = placed
	return l
}

// Seed returns the seed the level was generated from.
func (l *Level) Seed() uint64 { return l.seed }

// Length returns the corridor length in meters.
func (l *Level) Length() float64 { return l.length }

// FinishZ returns the z position of the finish, equal to Length.
func (l *Level) FinishZ() float64 { return l.length }

// Constraints returns the passability invariants the level was built under.
func (l *Level) Constraints() Constraints { return l.cons }

// Elements returns a copy of the elements whose Z lies in the half-open range
// [from, to), sorted by Z. The returned slice is safe to mutate.
func (l *Level) Elements(from, to float64) []Element {
	out := make([]Element, 0)
	for _, e := range l.elems {
		if e.Z >= from && e.Z < to {
			out = append(out, e)
		}
	}
	return out
}

// Mangroves returns the easy world preset for the wet mangroves.
func Mangroves() Params {
	return Params{
		BaseSpeed: 7,
		Length:    280,
		Density:   0.35,
		Combs:     90,
		Hairspray: 3,
		Mirrors:   3,
	}
}

// Harbor returns the medium world preset for the rusty harbor.
func Harbor() Params {
	return Params{
		BaseSpeed: 9,
		Length:    340,
		Density:   0.5,
		Combs:     80,
		Hairspray: 2,
		Mirrors:   3,
	}
}

// Neon returns the hard world preset for the neon Chic-Chic street.
func Neon() Params {
	return Params{
		BaseSpeed: 11,
		Length:    420,
		Density:   0.65,
		Combs:     70,
		Hairspray: 2,
		Mirrors:   4,
	}
}

const (
	laneCount = 3
	// pitLane is the ignored Lane value stored on pits; it keeps every
	// element's Lane inside [0, 2].
	pitLane = 0

	maxPitWidth = 3.0
	reactionGap = 8.0

	barrierDepth = 1.0
	minPitWidth  = 1.5

	rampSpan   = 7.0
	exitClear  = 1.0
	rampChance = 0.6

	startSafe = reactionGap
	endSafe   = reactionGap

	// densityGapScale turns Density into extra flat gap between forced
	// actions: higher density shortens the gaps.
	densityGapScale = 40.0

	pickupSpace = 1.0

	defaultBaseSpeed = 8.0
	maxPickupCount   = 2000
	minLength        = startSafe + endSafe + rampSpan + exitClear
)

const (
	pitShare  = 0.20
	lowShare  = 0.30
	highShare = 0.30
)

func sanitize(p Params) Params {
	if p.BaseSpeed <= 0 {
		p.BaseSpeed = defaultBaseSpeed
	}
	if p.Length < minLength {
		p.Length = minLength
	}
	p.Density = clampUnit(p.Density)
	p.Combs = clampCount(p.Combs)
	p.Hairspray = clampCount(p.Hairspray)
	p.Mirrors = clampCount(p.Mirrors)
	return p
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampCount(n int) int {
	if n < 0 {
		return 0
	}
	if n > maxPickupCount {
		return maxPickupCount
	}
	return n
}

func isForced(k Kind) bool {
	return k == Pit || k == LowBarrier || k == HighBarrier || k == Enemy
}

func isPickup(k Kind) bool {
	return k == Comb || k == Hairspray || k == Mirror
}

// zSpan returns the half-open [start, end) z interval an element occupies.
func zSpan(e Element) (float64, float64) {
	switch e.Kind {
	case Pit, Ramp:
		return e.Z, e.Z + e.Span
	case LowBarrier, HighBarrier, Enemy:
		return e.Z, e.Z + barrierDepth
	default:
		return e.Z, e.Z
	}
}

// placeForced lays out pits, barriers and enemies with the spacing guarantees.
func (l *Level) placeForced(p Params, rng *rng) []Element {
	if p.Density <= 0 {
		return nil
	}
	var forced []Element
	nextAllowed := startSafe
	for nextAllowed < l.length-endSafe {
		z := nextAllowed + rng.unit()*(1-p.Density)*densityGapScale
		kind := chooseForced(rng)
		if kind == Pit {
			if z+maxPitWidth > l.length-endSafe {
				break
			}
			width := minPitWidth + rng.unit()*(maxPitWidth-minPitWidth)
			forced = append(forced, Element{Kind: Pit, Lane: pitLane, Z: z, Span: width})
			nextAllowed = z + width + reactionGap
			continue
		}
		if z+barrierDepth > l.length-endSafe {
			break
		}
		forced = append(forced, Element{Kind: kind, Lane: rng.intn(laneCount), Z: z})
		nextAllowed = z + barrierDepth + reactionGap
	}
	return forced
}

func chooseForced(rng *rng) Kind {
	roll := rng.unit()
	switch {
	case roll < pitShare:
		return Pit
	case roll < pitShare+lowShare:
		return LowBarrier
	case roll < pitShare+lowShare+highShare:
		return HighBarrier
	default:
		return Enemy
	}
}

type interval struct {
	start, end float64
}

// rampBlocked returns the z intervals that must stay ramp-free: forced action
// footprints and the flat run-up before each pit.
func rampBlocked(forced []Element) []interval {
	ivs := make([]interval, 0, len(forced)*2)
	for _, e := range forced {
		if e.Kind == Pit {
			runUp := interval{start: e.Z - reactionGap, end: e.Z}
			if runUp.start < 0 {
				runUp.start = 0
			}
			ivs = append(ivs, runUp)
		}
		s, t := zSpan(e)
		ivs = append(ivs, interval{start: s, end: t})
	}
	return ivs
}

func mergeIntervals(ivs []interval) []interval {
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].start < ivs[j].start })
	var out []interval
	for _, iv := range ivs {
		if len(out) == 0 || iv.start > out[len(out)-1].end {
			out = append(out, iv)
			continue
		}
		if iv.end > out[len(out)-1].end {
			out[len(out)-1].end = iv.end
		}
	}
	return out
}

// placeRamps fills the clear stretches between forced actions with container
// ramps. At least one ramp is placed whenever the corridor has room for one.
func (l *Level) placeRamps(forced []Element, rng *rng) []Element {
	blocked := mergeIntervals(rampBlocked(forced))
	type stretch struct{ start, end float64 }
	var stretches []stretch

	pos := startSafe
	end := l.length - endSafe
	for _, b := range blocked {
		if b.end <= pos {
			continue
		}
		if b.start > pos {
			segEnd := b.start
			if segEnd > end {
				segEnd = end
			}
			if pos < segEnd {
				stretches = append(stretches, stretch{pos, segEnd})
			}
		}
		if b.end > pos {
			pos = b.end
		}
	}
	if pos < end {
		stretches = append(stretches, stretch{pos, end})
	}

	ramps := make([]Element, 0, len(stretches))
	placedAny := false
	for _, s := range stretches {
		if s.end-s.start < rampSpan+exitClear {
			continue
		}
		if rng.unit() > rampChance {
			continue
		}
		l.placeRampIn(s.start, s.end, rng, &ramps)
		placedAny = true
	}
	if !placedAny {
		for _, s := range stretches {
			if s.end-s.start >= rampSpan+exitClear {
				l.placeRampIn(s.start, s.end, rng, &ramps)
				break
			}
		}
	}
	return ramps
}

func (l *Level) placeRampIn(start, end float64, rng *rng, ramps *[]Element) {
	if end-start < rampSpan+exitClear {
		return
	}
	z := start + rng.unit()*(end-start-rampSpan-exitClear)
	*ramps = append(*ramps, Element{Kind: Ramp, Lane: rng.intn(laneCount), Z: z, Span: rampSpan})
}

// placePickups scatters combs and hairspray and lays mirrors at regular
// intervals, always on clear ground.
func (l *Level) placePickups(placed []Element, p Params, rng *rng) []Element {
	var out []Element
	for i := 0; i < p.Mirrors; i++ {
		target := (float64(i) + 1) * l.length / float64(p.Mirrors+1)
		if e, ok := findClearSpot(target, l.length, Mirror, placed); ok {
			out = append(out, e)
			placed = append(placed, e)
		}
	}
	for i := 0; i < p.Combs; i++ {
		if e, ok := randomClearSpot(rng, l.length, Comb, placed); ok {
			out = append(out, e)
			placed = append(placed, e)
		}
	}
	for i := 0; i < p.Hairspray; i++ {
		if e, ok := randomClearSpot(rng, l.length, Hairspray, placed); ok {
			out = append(out, e)
			placed = append(placed, e)
		}
	}
	return out
}

// findClearSpot finds a clear pickup position at or just after target.
func findClearSpot(target, length float64, kind Kind, placed []Element) (Element, bool) {
	for dz := 0.0; dz <= reactionGap; dz += 0.5 {
		z := target + dz
		if z >= length {
			break
		}
		for lane := 0; lane < laneCount; lane++ {
			if clearSpot(lane, z, placed) {
				return Element{Kind: kind, Lane: lane, Z: z}, true
			}
		}
	}
	return Element{}, false
}

// randomClearSpot picks a random clear pickup position, then falls back to a
// linear scan so a pickup count is honored whenever room exists.
func randomClearSpot(rng *rng, length float64, kind Kind, placed []Element) (Element, bool) {
	for attempt := 0; attempt < 100; attempt++ {
		lane := rng.intn(laneCount)
		z := startSafe + rng.unit()*(length-startSafe-endSafe)
		if clearSpot(lane, z, placed) {
			return Element{Kind: kind, Lane: lane, Z: z}, true
		}
	}
	for z := startSafe; z < length-endSafe; z += 0.5 {
		for lane := 0; lane < laneCount; lane++ {
			if clearSpot(lane, z, placed) {
				return Element{Kind: kind, Lane: lane, Z: z}, true
			}
		}
	}
	return Element{}, false
}

// clearSpot reports whether a pickup may sit at (lane, z) without colliding
// with anything already placed.
func clearSpot(lane int, z float64, placed []Element) bool {
	for _, e := range placed {
		switch e.Kind {
		case Pit:
			if z >= e.Z && z < e.Z+e.Span {
				return false
			}
		case Ramp:
			if e.Lane == lane && z >= e.Z && z < e.Z+e.Span {
				return false
			}
		case LowBarrier, HighBarrier, Enemy:
			if e.Lane == lane && z >= e.Z && z < e.Z+barrierDepth {
				return false
			}
		case Comb, Hairspray, Mirror:
			if e.Lane == lane {
				d := z - e.Z
				if d < 0 {
					d = -d
				}
				if d < pickupSpace {
					return false
				}
			}
		}
	}
	return true
}

// rng is a self-contained splitmix64 generator. The package never reads the
// global rand or the wall clock.
type rng struct {
	state uint64
}

func newRNG(seed uint64) *rng {
	return &rng{state: seed}
}

func (r *rng) next() uint64 {
	r.state += 0x9E3779B97F4A7C15
	z := r.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// unit returns a value in [0, 1).
func (r *rng) unit() float64 {
	return float64(r.next()>>11) / (1 << 53)
}

// intn returns an integer in [0, n).
func (r *rng) intn(n int) int {
	return int(r.next() % uint64(n))
}
