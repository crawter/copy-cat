package runner

import (
	"math"
	"reflect"
	"testing"

	"copy-cat/pkg/conditions"
	"copy-cat/pkg/move"
	"copy-cat/pkg/track"
)

// testDT is the fixed small step the simulations use so the results are
// deterministic and headless.
const testDT = 1.0 / 120.0

func TestCrateBreakCollectsJunk(t *testing.T) {
	r := &Run{
		Track: track.Track{
			Length: move.RunSpeed * 2 * conditions.SurvivalTime,
			Crates: []track.Crate{{X: 0, Z: move.RunSpeed}},
			Junk:   []track.Junk{{X: 0, Z: move.RunSpeed}},
		},
		Player: move.State{X: 0, Y: 1, Z: move.RunSpeed, OnGround: false},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
	}

	r.Step(move.Input{}, testDT)

	if r.Broken != 1 {
		t.Fatalf("Broken = %d, want 1", r.Broken)
	}
	if r.Collected != 1 {
		t.Fatalf("Collected = %d, want 1", r.Collected)
	}
	if len(r.Track.Crates) != 0 {
		t.Fatalf("crate still on track after breaking: %#v", r.Track.Crates)
	}
	if len(r.Track.Junk) != 0 {
		t.Fatalf("crate's jar lid still on track after breaking: %#v", r.Track.Junk)
	}
}

func TestCrateSidewaysHitStumbles(t *testing.T) {
	r := &Run{
		Track: track.Track{
			Length: move.RunSpeed * 2 * conditions.SurvivalTime,
			Crates: []track.Crate{{X: 0, Z: move.RunSpeed}},
			Junk:   []track.Junk{{X: 0, Z: move.RunSpeed}},
		},
		Player: move.State{X: 0, Z: move.RunSpeed, OnGround: true},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
	}

	r.Step(move.Input{}, testDT)

	if r.StumbleCount != 1 {
		t.Fatalf("StumbleCount = %d, want 1", r.StumbleCount)
	}
	if r.Player.StumbleLeft != move.StumbleTime {
		t.Fatalf("StumbleLeft = %v, want %v", r.Player.StumbleLeft, move.StumbleTime)
	}
	if r.Collected != 0 {
		t.Fatalf("Collected = %d, want 0 after a sideways hit", r.Collected)
	}
	if r.Broken != 0 {
		t.Fatalf("Broken = %d, want 0 after a sideways hit", r.Broken)
	}
	if len(r.Track.Crates) != 1 {
		t.Fatalf("crate left the track after a sideways hit: %#v", r.Track.Crates)
	}
	if len(r.Track.Junk) != 1 {
		t.Fatalf("crate's jar lid left the track after a sideways hit: %#v", r.Track.Junk)
	}
}

func TestEnemySmashAndStumble(t *testing.T) {
	pestZ := move.RunSpeed * 2

	t.Run("smash from above", func(t *testing.T) {
		r := &Run{
			Track: track.Track{
				Length:  move.RunSpeed * 2 * conditions.SurvivalTime,
				Enemies: []track.Enemy{{Kind: track.EnemyPest, X: 0, Z: pestZ}},
			},
			Player: move.State{X: 0, Y: 1, Z: pestZ, OnGround: false},
			Mode:   ModeSurvival,
			CrewZ:  -ChaseStartGap,
		}

		r.Step(move.Input{}, testDT)

		if r.Smashed != 1 {
			t.Fatalf("Smashed = %d, want 1", r.Smashed)
		}
		if r.EnemyHits != 0 {
			t.Fatalf("EnemyHits = %d, want 0 after a smash", r.EnemyHits)
		}
		if len(r.Track.Enemies) != 0 {
			t.Fatalf("smashed pest still on track: %#v", r.Track.Enemies)
		}
	})

	t.Run("run into on the ground", func(t *testing.T) {
		r := &Run{
			Track: track.Track{
				Length:  move.RunSpeed * 2 * conditions.SurvivalTime,
				Enemies: []track.Enemy{{Kind: track.EnemyPest, X: 0, Z: pestZ}},
			},
			Player: move.State{X: 0, Z: pestZ, OnGround: true},
			Mode:   ModeSurvival,
			CrewZ:  -ChaseStartGap,
		}

		r.Step(move.Input{}, testDT)

		if r.EnemyHits != 1 {
			t.Fatalf("EnemyHits = %d, want 1", r.EnemyHits)
		}
		if r.Smashed != 0 {
			t.Fatalf("Smashed = %d, want 0 after running into a pest", r.Smashed)
		}
		if r.StumbleCount != 1 {
			t.Fatalf("StumbleCount = %d, want 1", r.StumbleCount)
		}
		if r.Player.StumbleLeft != move.StumbleTime {
			t.Fatalf("StumbleLeft = %v, want %v", r.Player.StumbleLeft, move.StumbleTime)
		}
		if len(r.Track.Enemies) != 1 {
			t.Fatalf("run-into pest left the track: %#v", r.Track.Enemies)
		}
	})
}

func TestEnemyHitCounterIgnoresGapFalls(t *testing.T) {
	gap := track.Gap{
		Start: move.RunSpeed * 4,
		End:   move.RunSpeed*4 + track.MaxGapWidth,
	}
	mid := gap.Start + track.GapWidth(gap)/2

	r := &Run{
		Track: track.Track{
			Length: move.RunSpeed * 2 * conditions.SurvivalTime,
			Gaps:   []track.Gap{gap},
		},
		Player: move.State{X: 0, Z: mid, OnGround: true},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
	}

	r.Step(move.Input{}, testDT)

	if r.EnemyHits != 0 {
		t.Fatalf("EnemyHits = %d, want 0 after a gap fall", r.EnemyHits)
	}
	if r.StumbleCount != 1 {
		t.Fatalf("StumbleCount = %d, want 1", r.StumbleCount)
	}
	if r.Player.Z != gap.Start {
		t.Fatalf("Player.Z = %v, want gap start %v", r.Player.Z, gap.Start)
	}
	if r.Player.StumbleLeft != move.StumbleTime {
		t.Fatalf("StumbleLeft = %v, want %v", r.Player.StumbleLeft, move.StumbleTime)
	}
}

func TestMachineSmashAirborne(t *testing.T) {
	machineZ := move.RunSpeed * 2
	machine := track.Enemy{Kind: track.EnemyMachine, X: 0, Y: 2, Z: machineZ}

	t.Run("airborne smash", func(t *testing.T) {
		r := &Run{
			Track: track.Track{
				Length:  move.RunSpeed * 2 * conditions.SurvivalTime,
				Enemies: []track.Enemy{machine},
			},
			Player: move.State{X: 0, Y: 2, Z: machineZ, OnGround: false},
			Mode:   ModeSurvival,
			CrewZ:  -ChaseStartGap,
		}

		r.Step(move.Input{}, testDT)

		if r.Smashed != 1 {
			t.Fatalf("Smashed = %d, want 1", r.Smashed)
		}
		if len(r.Track.Enemies) != 0 {
			t.Fatalf("smashed machine still on track: %#v", r.Track.Enemies)
		}
	})

	t.Run("grounded passes under", func(t *testing.T) {
		r := &Run{
			Track: track.Track{
				Length:  move.RunSpeed * 2 * conditions.SurvivalTime,
				Enemies: []track.Enemy{machine},
			},
			Player: move.State{X: 0, Z: machineZ, OnGround: true},
			Mode:   ModeSurvival,
			CrewZ:  -ChaseStartGap,
		}

		r.Step(move.Input{}, testDT)

		if r.Smashed != 0 {
			t.Fatalf("Smashed = %d, want 0", r.Smashed)
		}
		if r.EnemyHits != 0 {
			t.Fatalf("EnemyHits = %d, want 0", r.EnemyHits)
		}
		if len(r.Track.Enemies) != 1 {
			t.Fatalf("machine left the track after the hero passed under: %#v", r.Track.Enemies)
		}
	})
}

func TestChaseClosesOnStumble(t *testing.T) {
	r := &Run{
		Track:  track.Track{Length: move.RunSpeed * 2 * conditions.SurvivalTime},
		Player: move.State{OnGround: true},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
	}

	move.Stumble(&r.Player)
	gapBefore := r.Player.Z - r.CrewZ
	r.Step(move.Input{}, testDT)
	gapAfter := r.Player.Z - r.CrewZ

	wantShrink := (CrewSpeed - move.RunSpeed*move.StumbleSpeedFraction) * testDT
	if math.Abs((gapBefore-gapAfter)-wantShrink) > 1e-9 {
		t.Fatalf("gap shrank by %v, want %v", gapBefore-gapAfter, wantShrink)
	}

	caught := &Run{
		Track:  track.Track{Length: move.RunSpeed * 2 * conditions.SurvivalTime},
		Player: move.State{OnGround: true},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
	}
	for i := 0; i < 8 && !caught.Caught; i++ {
		move.Stumble(&caught.Player)
		for caught.Player.StumbleLeft > 0 && !caught.Caught {
			caught.Step(move.Input{}, testDT)
		}
	}
	if !caught.Caught {
		t.Fatalf("Caught = false after 8 stumbles, want true")
	}
}

func TestCaughtFreezesRun(t *testing.T) {
	r := &Run{
		Track:  track.Track{Length: move.RunSpeed * 2 * conditions.SurvivalTime},
		Player: move.State{OnGround: true},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
		Caught: true,
	}
	snapshot := *r

	r.Step(move.Input{Jump: true, Right: true}, testDT)

	if !reflect.DeepEqual(r, &snapshot) {
		t.Fatalf("run changed after Caught:\n got %#v\nwant %#v", r, &snapshot)
	}
}

func TestCompleteFreezesRun(t *testing.T) {
	r := &Run{
		Track:    track.Track{Length: move.RunSpeed * 2 * conditions.SurvivalTime},
		Player:   move.State{OnGround: true},
		Mode:     ModeSurvival,
		CrewZ:    -ChaseStartGap,
		Complete: true,
	}
	snapshot := *r

	r.Step(move.Input{Jump: true, Right: true}, testDT)

	if !reflect.DeepEqual(r, &snapshot) {
		t.Fatalf("run changed after Complete:\n got %#v\nwant %#v", r, &snapshot)
	}
}

func TestTrackWrapLoopsTheRun(t *testing.T) {
	length := move.RunSpeed * 4
	r := &Run{
		Track:  track.Track{Length: length},
		Player: move.State{OnGround: true, Z: length - move.RunSpeed*testDT/2},
		Mode:   ModeSurvival,
		CrewZ:  length - move.RunSpeed*testDT/2 - ChaseStartGap,
	}

	wantPlayerZ := r.Player.Z + move.RunSpeed*testDT - length
	wantCrewZ := r.CrewZ + CrewSpeed*testDT - length

	r.Step(move.Input{}, testDT)

	if math.Abs(r.Player.Z-wantPlayerZ) > 1e-9 {
		t.Fatalf("Player.Z = %v, want %v", r.Player.Z, wantPlayerZ)
	}
	if math.Abs(r.CrewZ-wantCrewZ) > 1e-9 {
		t.Fatalf("CrewZ = %v, want %v", r.CrewZ, wantCrewZ)
	}
	if r.Caught || r.Complete {
		t.Fatalf("run ended after wrapping: Caught=%v Complete=%v", r.Caught, r.Complete)
	}

	before := r.Player.Z
	r.Step(move.Input{}, testDT)
	if r.Player.Z <= before {
		t.Fatalf("run did not continue after wrapping: Z = %v before %v", r.Player.Z, before)
	}
}

func TestSurvivalRunCompletes(t *testing.T) {
	r := &Run{
		Track:  track.Track{Length: move.RunSpeed * 2 * conditions.SurvivalTime},
		Player: move.State{OnGround: true},
		Mode:   ModeSurvival,
		CrewZ:  -ChaseStartGap,
	}

	steps := int(conditions.SurvivalTime/testDT) + 1
	for i := 0; i < steps && !r.Complete; i++ {
		r.Step(move.Input{}, testDT)
	}

	if !r.Complete {
		t.Fatalf("Complete = false after %d steps (%v seconds), want true", steps, float64(steps)*testDT)
	}
	if r.Elapsed < conditions.SurvivalTime {
		t.Fatalf("Elapsed = %v, want at least %v", r.Elapsed, conditions.SurvivalTime)
	}
}

func TestCollectionRunCompletes(t *testing.T) {
	junk := make([]track.Junk, 0, conditions.CollectionQuota)
	for i := 0; i < conditions.CollectionQuota; i++ {
		junk = append(junk, track.Junk{X: 0, Z: float64(i+1)})
	}

	r := &Run{
		Track:  track.Track{Length: move.RunSpeed * 2 * conditions.SurvivalTime, Junk: junk},
		Player: move.State{X: 0, OnGround: true},
		Mode:   ModeCollection,
		CrewZ:  -ChaseStartGap,
	}

	maxSteps := int(2*conditions.SurvivalTime/testDT) + 1
	for i := 0; i < maxSteps && !r.Complete; i++ {
		r.Step(move.Input{}, testDT)
	}

	if !r.Complete {
		t.Fatalf("Complete = false after %d steps, want true", maxSteps)
	}
	if r.Collected < conditions.CollectionQuota {
		t.Fatalf("Collected = %d, want at least %d", r.Collected, conditions.CollectionQuota)
	}
}
