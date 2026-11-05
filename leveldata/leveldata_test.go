package leveldata

import (
	"fmt"
	"testing"
)

func checkX(t *testing.T, name string, x, width float64) {
	t.Helper()
	if x < 0 || x > width {
		t.Errorf("%s X = %v, want within [0, %v]", name, x, width)
	}
}

func checkInside(t *testing.T, name string, take Take, p Point) {
	t.Helper()
	if p.X < 0 || p.X > take.Width {
		t.Errorf("%s X = %v, want within [0, %v]", name, p.X, take.Width)
	}
	if p.Z < 0 || p.Z > take.Length {
		t.Errorf("%s Z = %v, want within [0, %v]", name, p.Z, take.Length)
	}
}

func TestCorridorWidth(t *testing.T) {
	take := Take9()
	if take.Width != 8.0 {
		t.Fatalf("Width = %v, want exactly 8", take.Width)
	}

	for i, rig := range take.Rigs {
		checkX(t, fmt.Sprintf("rig %d", i), rig.Position.X, take.Width)
	}
	for i, reel := range take.Reels {
		checkX(t, fmt.Sprintf("reel %d", i), reel.Position.X, take.Width)
	}
	for i, cp := range take.Checkpoints {
		checkX(t, fmt.Sprintf("checkpoint %d", i), cp.Position.X, take.Width)
	}
	checkX(t, "track start", take.Track.Start.X, take.Width)
	checkX(t, "track end", take.Track.End.X, take.Width)
}

func TestPiecesSitInsideCorridor(t *testing.T) {
	take := Take9()

	for i, rig := range take.Rigs {
		checkInside(t, fmt.Sprintf("rig %d", i), take, rig.Position)
	}
	for i, reel := range take.Reels {
		checkInside(t, fmt.Sprintf("reel %d", i), take, reel.Position)
	}
	for i, cp := range take.Checkpoints {
		checkInside(t, fmt.Sprintf("checkpoint %d", i), take, cp.Position)
	}
	checkInside(t, "track start", take, take.Track.Start)
	checkInside(t, "track end", take, take.Track.End)
	if take.Exit.Z < 0 || take.Exit.Z > take.Length {
		t.Errorf("exit Z = %v, want within [0, %v]", take.Exit.Z, take.Length)
	}
}

func TestRunLengthInsideTwoToThreeMinutes(t *testing.T) {
	take := Take9()

	seconds := take.Length / 12.0
	if seconds < 120 || seconds > 180 {
		t.Fatalf("run length = %v s, want within [120, 180]", seconds)
	}
}

func TestSectionsCoverCorridor(t *testing.T) {
	take := Take9()

	if len(take.Sections) != 3 {
		t.Fatalf("sections = %d, want 3", len(take.Sections))
	}
	if take.Sections[0].StartZ != 0 {
		t.Errorf("first section StartZ = %v, want 0", take.Sections[0].StartZ)
	}
	for i, s := range take.Sections {
		if s.StartZ >= s.EndZ {
			t.Errorf("section %d is empty or reversed: %v..%v", i, s.StartZ, s.EndZ)
		}
	}
	for i := 1; i < len(take.Sections); i++ {
		prev := take.Sections[i-1]
		cur := take.Sections[i]
		if prev.EndZ != cur.StartZ {
			t.Errorf("sections %d and %d do not meet without gap or overlap: %v != %v", i-1, i, prev.EndZ, cur.StartZ)
		}
	}
	if last := take.Sections[len(take.Sections)-1].EndZ; last != take.Length {
		t.Errorf("last section EndZ = %v, want corridor Length %v", last, take.Length)
	}
}

func TestNineteenReelsWithEachPlacement(t *testing.T) {
	take := Take9()

	if len(take.Reels) != 19 {
		t.Fatalf("reels = %d, want 19", len(take.Reels))
	}

	seen := map[Placement]bool{}
	for i, reel := range take.Reels {
		switch reel.Kind {
		case Floor, RigTop, Nook:
			seen[reel.Kind] = true
		default:
			t.Errorf("reel %d kind = %q, want floor, rig top or nook", i, reel.Kind)
		}
	}
	for _, kind := range []Placement{Floor, RigTop, Nook} {
		if !seen[kind] {
			t.Errorf("no reel placed %q", kind)
		}
	}
}

func TestRigTopReelsSitAtRigPosition(t *testing.T) {
	take := Take9()

	for i, reel := range take.Reels {
		if reel.Kind != RigTop {
			continue
		}
		found := false
		for j, rig := range take.Rigs {
			if reel.Position.X != rig.Position.X || reel.Position.Z != rig.Position.Z {
				continue
			}
			found = true
			if reel.Position.Y <= rig.Position.Y {
				t.Errorf("rig-top reel %d Y = %v, want above rig %d Y %v", i, reel.Position.Y, j, rig.Position.Y)
			}
			break
		}
		if !found {
			t.Errorf("rig-top reel %d at %+v sits on no rig", i, reel.Position)
		}
	}
}

func TestCheckpointsInOrderBeforeExit(t *testing.T) {
	take := Take9()

	if len(take.Checkpoints) != 3 {
		t.Fatalf("checkpoints = %d, want 3", len(take.Checkpoints))
	}

	prev := -1.0
	for i, cp := range take.Checkpoints {
		if cp.Position.Z <= prev {
			t.Errorf("checkpoint %d Z = %v, want after previous %v", i, cp.Position.Z, prev)
		}
		prev = cp.Position.Z
		if cp.Position.Z >= take.Exit.Z {
			t.Errorf("checkpoint %d Z = %v, want before exit %v", i, cp.Position.Z, take.Exit.Z)
		}
	}
}

func TestTrackInsideOneSection(t *testing.T) {
	take := Take9()

	checkInside(t, "track start", take, take.Track.Start)
	checkInside(t, "track end", take, take.Track.End)

	startSection := -1
	endSection := -1
	for i, s := range take.Sections {
		if take.Track.Start.Z >= s.StartZ && take.Track.Start.Z <= s.EndZ {
			startSection = i
		}
		if take.Track.End.Z >= s.StartZ && take.Track.End.Z <= s.EndZ {
			endSection = i
		}
	}
	if startSection == -1 || endSection == -1 {
		t.Fatalf("track ends are not inside any section: %+v, %+v", take.Track.Start, take.Track.End)
	}
	if startSection != endSection {
		t.Errorf("track starts in section %d and ends in section %d, want one section", startSection, endSection)
	}
}
