package enemy

import (
	"math"
	"testing"

	"copycat/collide"
	"copycat/leveldata"
	"copycat/move"
)

func TestNewStartsAtTrackStartAndDropsNothing(t *testing.T) {
	track := leveldata.Take9().Track
	c := New(track)

	if c.pos != track.Start {
		t.Fatalf("New pos = %+v, want track start %+v", c.pos, track.Start)
	}
	if c.defeated {
		t.Fatal("New crane is defeated, want living")
	}
	if c.Dropped() {
		t.Fatal("New crane dropped a reel before any defeat")
	}
}

func TestPatrolStaysInsideTrack(t *testing.T) {
	track := leveldata.Take9().Track
	c := New(track)
	lowZ := math.Min(track.Start.Z, track.End.Z)
	highZ := math.Max(track.Start.Z, track.End.Z)

	const dt = 1.0 / 240.0
	for i := 0; i < 240*300; i++ {
		c = Step(c, dt)
		if c.pos.Z < lowZ || c.pos.Z > highZ {
			t.Fatalf("step %d: Z = %v, want within [%v, %v]", i, c.pos.Z, lowZ, highZ)
		}
	}
}

func TestPatrolReversesAtEachEnd(t *testing.T) {
	track := leveldata.Take9().Track
	c := New(track)
	lowZ := math.Min(track.Start.Z, track.End.Z)
	highZ := math.Max(track.Start.Z, track.End.Z)

	const dt = 1.0 / 240.0
	reversals := 0
	prevDir := c.dir
	for i := 0; i < 240*300; i++ {
		c = Step(c, dt)
		if c.dir == prevDir {
			continue
		}
		reversals++
		prevDir = c.dir
		switch reversals {
		case 1:
			if c.pos.Z != highZ {
				t.Fatalf("first reversal at Z = %v, want high end %v", c.pos.Z, highZ)
			}
		case 2:
			if c.pos.Z != lowZ {
				t.Fatalf("second reversal at Z = %v, want low end %v", c.pos.Z, lowZ)
			}
			return
		}
	}
	t.Fatalf("reversals = %d, want at least 2", reversals)
}

func TestDefeatFromTop(t *testing.T) {
	c := New(leveldata.Take9().Track)
	h := move.Hero{X: c.pos.X, Y: bodyHeightMeters + counterweightHalfMeters, Z: c.pos.Z}

	got, ok := Defeat(h, c)
	if !ok {
		t.Fatal("hero on the counterweight should defeat the crane")
	}
	if !got.defeated {
		t.Fatal("defeated crane is not marked down")
	}
	if !got.Dropped() {
		t.Fatal("defeated crane should drop its reel")
	}
}

func TestNoDefeatFromFloor(t *testing.T) {
	c := New(leveldata.Take9().Track)
	h := move.Hero{X: c.pos.X, Y: 0, Z: c.pos.Z}

	got, ok := Defeat(h, c)
	if ok {
		t.Fatal("floor hero below the counterweight should not defeat the crane")
	}
	if got != c {
		t.Fatalf("Defeat changed the crane without a defeat: %+v != %+v", got, c)
	}
	if got.Dropped() {
		t.Fatal("undefeated crane dropped a reel")
	}
}

func TestJumpLandsTheDefeat(t *testing.T) {
	c := New(leveldata.Take9().Track)
	const dt = 1.0 / 240.0

	h := move.NewHero()
	h.X = c.pos.X
	h.Z = c.pos.Z - 5.5
	h = move.Step(h, move.Input{Jump: true}, dt)
	if !h.Airborne {
		t.Fatal("takeoff should start a jump")
	}

	var defeated Crane
	var ok bool
	for {
		defeated, ok = Defeat(h, c)
		if ok {
			break
		}
		if !h.Airborne && h.Z > c.pos.Z+2 {
			t.Fatalf("jump never defeated the crane; last hero %+v", h)
		}
		h = move.Step(h, move.Input{}, dt)
	}

	if !defeated.Dropped() {
		t.Fatal("defeated crane should drop its reel")
	}
	if h.Z < c.pos.Z-1 || h.Z > c.pos.Z+1 {
		t.Fatalf("defeat at hero Z = %v, want near the counterweight at %v", h.Z, c.pos.Z)
	}
}

func TestDefeatedCraneStaysDown(t *testing.T) {
	c := New(leveldata.Take9().Track)
	h := move.Hero{X: c.pos.X, Y: bodyHeightMeters + counterweightHalfMeters, Z: c.pos.Z}

	var ok bool
	c, ok = Defeat(h, c)
	if !ok {
		t.Fatal("hero on the counterweight should defeat the crane")
	}

	before := c
	got := Step(c, 0.5)
	if got.pos != before.pos {
		t.Fatalf("Step moved a defeated crane: %+v -> %+v", before.pos, got.pos)
	}
	if got.dir != before.dir {
		t.Fatalf("Step changed a defeated crane's direction: %v -> %v", before.dir, got.dir)
	}
	if !got.Dropped() {
		t.Fatal("defeated crane should keep its dropped reel")
	}
}

func TestDroppedFalseBeforeDefeatTrueAfter(t *testing.T) {
	c := New(leveldata.Take9().Track)
	if c.Dropped() {
		t.Fatal("Dropped = true before the defeat")
	}

	h := move.Hero{X: c.pos.X, Y: bodyHeightMeters + counterweightHalfMeters, Z: c.pos.Z}
	var ok bool
	c, ok = Defeat(h, c)
	if !ok {
		t.Fatal("hero on the counterweight should defeat the crane")
	}
	if !c.Dropped() {
		t.Fatal("Dropped = false after the defeat")
	}
}

func TestDefeatOnlyOnce(t *testing.T) {
	c := New(leveldata.Take9().Track)
	h := move.Hero{X: c.pos.X, Y: bodyHeightMeters + counterweightHalfMeters, Z: c.pos.Z}

	var ok bool
	c, ok = Defeat(h, c)
	if !ok {
		t.Fatal("first Defeat should defeat the crane")
	}
	if _, ok = Defeat(h, c); ok {
		t.Fatal("second Defeat = true, want false for a down crane")
	}
	if !c.Dropped() {
		t.Fatal("crane should still report its drop after the defeat")
	}
}

func TestBodyBoxMatchesAssumedSizes(t *testing.T) {
	c := New(leveldata.Take9().Track)
	got := c.Body()

	if got.Pos != c.pos {
		t.Fatalf("Body Pos = %+v, want crane position %+v", got.Pos, c.pos)
	}
	if got.HalfLen != bodyHalfLenMeters {
		t.Fatalf("Body HalfLen = %v, want %v", got.HalfLen, bodyHalfLenMeters)
	}
	if got.HalfWidth != bodyHalfWidthMeters {
		t.Fatalf("Body HalfWidth = %v, want %v", got.HalfWidth, bodyHalfWidthMeters)
	}
	if got.Height != bodyHeightMeters {
		t.Fatalf("Body Height = %v, want %v", got.Height, bodyHeightMeters)
	}
}

// TestJumpedApproachDefeatsWithoutCraneHit replays the main loop's frame order
// — step hero, step crane, Defeat, then collide.Check with an off-corridor
// body once defeated — over the normal jumped approach. It fails on any
// EventCraneHit before or after the defeat, and if the defeat never lands
// within 120 frames. With the old 0.7-meter body the approach clips the body
// in both patrol phases; with 0.4 meters it clears.
func TestJumpedApproachDefeatsWithoutCraneHit(t *testing.T) {
	const dt = 1.0 / 60.0
	take := leveldata.Take9()

	outbound := New(take.Track)
	inbound := New(take.Track)
	inbound.pos.Z = inbound.highZ
	inbound.dir = -1

	tests := []struct {
		name string
		c    Crane
	}{
		{name: "outbound", c: outbound},
		{name: "inbound", c: inbound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.c
			h := move.NewHero()
			h.Z = c.pos.Z - 5.5

			defeated := false
			for frame := 0; frame < 120; frame++ {
				in := move.Input{Jump: frame == 0}
				switch {
				case h.X < c.pos.X:
					in.Steer = 1
				case h.X > c.pos.X:
					in.Steer = -1
				}

				h = move.Step(h, in, dt)
				c = Step(c, dt)
				c, _ = Defeat(h, c)

				body := c.Body()
				if c.Dropped() {
					defeated = true
					body.Pos = leveldata.Point{X: 0, Y: 0, Z: -1000}
				}

				_, events := collide.Check(h, take, body)
				for _, ev := range events {
					if ev.Kind == collide.EventCraneHit {
						t.Fatalf("frame %d: jumped approach hit the crane body, want none", frame)
					}
				}
			}

			if !defeated {
				t.Fatal("crane never defeated within 120 frames on the jumped approach")
			}
		})
	}
}
