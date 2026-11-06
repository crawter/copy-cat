package track

import (
	"reflect"
	"testing"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/move"
)

// fixedSeeds is the seed set shared by the generator and the later ideal
// controller: 0, 1, 7, 13, 42 and 99.
var fixedSeeds = []uint64{0, 1, 7, 13, 42, 99}

func TestDifferentSeedsGiveDifferentLayouts(t *testing.T) {
	a := Generate(0)
	b := Generate(1)
	if reflect.DeepEqual(a, b) {
		t.Fatalf("Generate(0) and Generate(1) produced the same layout: %#v", a)
	}
}

func TestEveryGapIsJumpable(t *testing.T) {
	for _, seed := range fixedSeeds {
		tr := Generate(seed)
		for i, g := range tr.Gaps {
			if !GapJumpable(g) {
				t.Fatalf("seed %d gap %d %+v is not jumpable (MaxGapWidth %v)", seed, i, g, MaxGapWidth)
			}
		}
	}
}

func TestCollectionTargetIsReachable(t *testing.T) {
	window := move.RunSpeed * 2 * conditions.SurvivalTime
	for _, seed := range fixedSeeds {
		tr := Generate(seed)
		if got := CountReachableJunk(tr, window); got < conditions.CollectionQuota {
			t.Fatalf("seed %d: CountReachableJunk within %v = %d, want at least %d", seed, window, got, conditions.CollectionQuota)
		}
	}
}

func TestSameSeedSameTrack(t *testing.T) {
	const seed = uint64(42)
	a := Generate(seed)
	b := Generate(seed)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Generate(%d) is not deterministic: %#v vs %#v", seed, a, b)
	}
}
