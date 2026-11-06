package track

import (
	"math/rand"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/move"
)

// TrackLength is the total length of one generated swamp corridor, in world
// units. It is longer than the distance the hero covers while surviving the
// set survival time at full speed, so the corridor never runs out during a
// survival run.
const TrackLength = 1200.0

// MinObstacleSpacing is the minimum distance between consecutive obstacles —
// gaps, crates, enemies and junk rows — along the corridor, in world units.
// It is one and a half full-speed jump distances, so the hero always has room
// to land and react before the next obstacle.
const MinObstacleSpacing = 1.5 * move.MaxJumpDistance

// junkRowSpacing is the sideways distance between the closely placed jar lids
// in one quota row, in world units. It keeps the lids dense enough that any X
// crossing the corridor runs through at least one lid in the row.
const junkRowSpacing = 0.5

// minGapWidth is the smallest gap the generator cuts, in world units. It keeps
// every gap a real hole rather than a sliver while still respecting
// MaxGapWidth.
const minGapWidth = 1.0

// enemyHoverHeight is the height above the floor at which the small machine
// line hovers, in world units.
const enemyHoverHeight = 2.0

// featureKind names the random features placed between the quota rows.
type featureKind int

const (
	featureGap featureKind = iota
	featureCrate
	featureEnemy
	featureJunk
	featureKindCount
)

// Generate builds one deterministic swamp Track for seed. The same seed always
// gives the same layout; different seeds give different layouts.
func Generate(seed uint64) Track {
	rng := rand.New(rand.NewSource(int64(seed)))
	t := Track{Length: TrackLength}

	reachWindow := move.RunSpeed * 2 * conditions.SurvivalTime
	spacing := MinObstacleSpacing

	// The quota jar-lid rows sit first, spread across the reach window, so the
	// collection target is always reachable before the run can end.
	anchors := make([]float64, 0, conditions.CollectionQuota)
	first := spacing
	last := reachWindow - spacing
	step := (last - first) / float64(conditions.CollectionQuota+1)
	for i := 0; i < conditions.CollectionQuota; i++ {
		z := first + step*float64(i+1)
		anchors = append(anchors, z)
		t.Junk = append(t.Junk, junkRow(z)...)
	}

	// The rows are anchors. The remaining features are generated only in the
	// space between neighbouring anchors, keeping MinObstacleSpacing clear on
	// both sides of every row.
	boundaries := make([]float64, 0, len(anchors)+2)
	boundaries = append(boundaries, 0)
	boundaries = append(boundaries, anchors...)
	boundaries = append(boundaries, TrackLength)

	for i := 0; i+1 < len(boundaries); i++ {
		start := boundaries[i] + spacing
		end := boundaries[i+1] - spacing
		if end <= start {
			continue
		}
		placeFeatures(rng, &t, start, end)
	}

	return t
}

// junkRow returns the closely placed jar lids of one quota row at z. They span
// the whole corridor width so any X crosses at least one lid.
func junkRow(z float64) []Junk {
	half := move.CorridorHalfWidth
	count := int(2*half/junkRowSpacing) + 1
	lids := make([]Junk, 0, count)
	for i := 0; i < count; i++ {
		lids = append(lids, Junk{X: -half + float64(i)*junkRowSpacing, Z: z})
	}
	return lids
}

// placeFeatures fills one clear stretch of corridor with random gaps, crates,
// enemies and extra junk. Every feature keeps MinObstacleSpacing from the next
// one, and every gap's end stays inside the stretch.
func placeFeatures(rng *rand.Rand, t *Track, start, end float64) {
	z := start
	for z <= end {
		kind := featureKind(rng.Intn(int(featureKindCount)))
		if kind == featureGap {
			width := randomGapWidth(rng)
			if z+width > end {
				// Not enough room left for a gap; put a lid there instead.
				z = placePointFeature(rng, t, z, featureJunk)
				continue
			}
			t.Gaps = append(t.Gaps, Gap{Start: z, End: z + width})
			z += width + MinObstacleSpacing
			continue
		}
		z = placePointFeature(rng, t, z, kind)
	}
}

// placePointFeature places one crate, enemy or extra lid at z and returns the
// next free z, MinObstacleSpacing further along.
func placePointFeature(rng *rand.Rand, t *Track, z float64, kind featureKind) float64 {
	x := randomX(rng)
	switch kind {
	case featureCrate:
		t.Crates = append(t.Crates, Crate{X: x, Z: z})
		// Each crate carries one more jar lid on top of the quota rows. It is
		// kept at the crate so the runner can release it when the crate breaks.
		t.Junk = append(t.Junk, Junk{X: x, Z: z})
	case featureEnemy:
		enemyKind := EnemyPest
		y := 0.0
		if rng.Intn(2) == 0 {
			enemyKind = EnemyMachine
			y = enemyHoverHeight
		}
		t.Enemies = append(t.Enemies, Enemy{Kind: enemyKind, X: x, Y: y, Z: z})
	case featureJunk:
		t.Junk = append(t.Junk, Junk{X: x, Z: z})
	}
	return z + MinObstacleSpacing
}

// randomX returns a sideways position inside the corridor bounds.
func randomX(rng *rand.Rand) float64 {
	half := move.CorridorHalfWidth
	return -half + rng.Float64()*2*half
}

// randomGapWidth returns a gap width in [minGapWidth, MaxGapWidth].
func randomGapWidth(rng *rand.Rand) float64 {
	return minGapWidth + rng.Float64()*(MaxGapWidth-minGapWidth)
}
