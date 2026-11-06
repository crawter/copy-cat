package track

import (
	"testing"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/move"
)

func TestMaxGapWidthFollowsMovementConstants(t *testing.T) {
	if MaxGapWidth != 0.75*move.MaxJumpDistance {
		t.Fatalf("MaxGapWidth = %v, want %v", MaxGapWidth, 0.75*move.MaxJumpDistance)
	}
}

func TestGapJumpableBoundsTheGap(t *testing.T) {
	if !GapJumpable(Gap{Start: 0, End: MaxGapWidth}) {
		t.Fatalf("GapJumpable(Gap{Start: 0, End: %v}) = false, want true", MaxGapWidth)
	}
	if GapJumpable(Gap{Start: 0, End: MaxGapWidth + MaxGapWidth}) {
		t.Fatalf("GapJumpable(Gap{Start: 0, End: %v}) = true, want false", MaxGapWidth+MaxGapWidth)
	}
}

func TestCountReachableJunk(t *testing.T) {
	quota := conditions.CollectionQuota
	window := move.RunSpeed * 2 * conditions.SurvivalTime

	gapStart := window / 4
	gap := Gap{Start: gapStart, End: gapStart + MaxGapWidth}
	tr := Track{
		Length: 2 * window,
		Gaps:   []Gap{gap},
	}

	// Floor lids within the reach window and on solid track.
	for i := 0; i < quota; i++ {
		tr.Junk = append(tr.Junk, Junk{X: 0, Z: float64(i)})
	}

	// The same count sitting inside the gap, so they are unreachable.
	for i := 0; i < quota; i++ {
		z := gapStart + GapWidth(gap)*float64(i+1)/float64(quota+1)
		tr.Junk = append(tr.Junk, Junk{X: 0, Z: z})
	}

	// The same count beyond the reach window.
	for i := 0; i < quota; i++ {
		tr.Junk = append(tr.Junk, Junk{X: 0, Z: window + MaxGapWidth + float64(i)})
	}

	if got := CountReachableJunk(tr, window); got != quota {
		t.Fatalf("CountReachableJunk = %d, want %d", got, quota)
	}
}
