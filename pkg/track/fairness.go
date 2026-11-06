package track

import "copy-cat/pkg/move"

// MaxGapWidth is the widest gap the hero can jump, in world units. It is
// 0.75 of move.MaxJumpDistance so every generated gap keeps a safety margin.
const MaxGapWidth = 0.75 * move.MaxJumpDistance

// GapWidth returns g's width in world units: its end minus its start.
func GapWidth(g Gap) float64 {
	return g.End - g.Start
}

// GapJumpable reports whether the hero can clear g: its width must not
// exceed MaxGapWidth.
func GapJumpable(g Gap) bool {
	return GapWidth(g) <= MaxGapWidth
}

// CountReachableJunk returns how many jar lids in t lie at or before limit
// along the corridor and on solid track, that is outside every gap.
func CountReachableJunk(t Track, limit float64) int {
	count := 0
	for _, j := range t.Junk {
		if j.Z > limit {
			continue
		}
		insideGap := false
		for _, g := range t.Gaps {
			if j.Z >= g.Start && j.Z <= g.End {
				insideGap = true
				break
			}
		}
		if !insideGap {
			count++
		}
	}
	return count
}
