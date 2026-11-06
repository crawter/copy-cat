package move

import (
	"math"
	"testing"
)

// testDT is the fixed small step the simulations use so the results are
// deterministic and headless.
const testDT = 1.0 / 120.0

func TestRunningMovesForward(t *testing.T) {
	s := State{OnGround: true}
	const steps = 30
	for i := 0; i < steps; i++ {
		Step(&s, Input{}, testDT)
	}
	want := RunSpeed * float64(steps) * testDT
	if math.Abs(s.Z-want) > 1e-9 {
		t.Fatalf("Z = %v, want %v after %d steps", s.Z, want, steps)
	}
}

func TestSteeringBounds(t *testing.T) {
	s := State{OnGround: true}

	Step(&s, Input{Right: true}, testDT)
	if math.Abs(s.X-SteerSpeed*testDT) > 1e-9 {
		t.Fatalf("one right step moved X to %v, want %v", s.X, SteerSpeed*testDT)
	}

	for i := 0; i < 2000; i++ {
		Step(&s, Input{Right: true}, testDT)
		if s.X > CorridorHalfWidth || s.X < -CorridorHalfWidth {
			t.Fatalf("X escaped the corridor: %v", s.X)
		}
	}
	if s.X != CorridorHalfWidth {
		t.Fatalf("X = %v, want right wall %v", s.X, CorridorHalfWidth)
	}

	for i := 0; i < 2000; i++ {
		Step(&s, Input{Left: true}, testDT)
		if s.X > CorridorHalfWidth || s.X < -CorridorHalfWidth {
			t.Fatalf("X escaped the corridor: %v", s.X)
		}
	}
	if s.X != -CorridorHalfWidth {
		t.Fatalf("X = %v, want left wall %v", s.X, -CorridorHalfWidth)
	}
}

func TestJumpDistanceAndLanding(t *testing.T) {
	s := State{OnGround: true}

	Step(&s, Input{Jump: true}, testDT)
	if s.OnGround {
		t.Fatal("a jump from the ground did not leave the ground")
	}
	elapsed := testDT

	for i := 0; !s.OnGround && i < 10000; i++ {
		Step(&s, Input{}, testDT)
		elapsed += testDT
		if s.Y < 0 {
			t.Fatalf("Y went negative: %v", s.Y)
		}
	}
	if !s.OnGround {
		t.Fatal("the hero never landed")
	}
	if s.Y != 0 || s.VY != 0 {
		t.Fatalf("landing not at rest on the ground: Y=%v VY=%v", s.Y, s.VY)
	}

	wantTime := 2 * JumpSpeed / Gravity
	if math.Abs(elapsed-wantTime) > 2*testDT {
		t.Fatalf("landed after %v, want %v", elapsed, wantTime)
	}
	if math.Abs(s.Z-MaxJumpDistance) > 2*testDT*RunSpeed {
		t.Fatalf("Z after landing = %v, want %v", s.Z, MaxJumpDistance)
	}
}

func TestJumpOnlyFromGround(t *testing.T) {
	s := State{OnGround: true}

	Step(&s, Input{Jump: true}, testDT)
	if s.OnGround {
		t.Fatal("a jump from the ground did not leave the ground")
	}
	afterTakeoff := s.VY

	Step(&s, Input{Jump: true}, testDT)
	if s.OnGround {
		t.Fatal("a second jump input while airborne landed the hero")
	}
	if s.VY >= afterTakeoff {
		t.Fatalf("a second jump input reset the arc: VY = %v, want below %v", s.VY, afterTakeoff)
	}
}

func TestStumbleSlowsAndRecovers(t *testing.T) {
	s := State{OnGround: true}
	Stumble(&s)
	if s.StumbleLeft != StumbleTime {
		t.Fatalf("StumbleLeft = %v, want %v", s.StumbleLeft, StumbleTime)
	}

	for i := 0; s.StumbleLeft > 0 && i < 10000; i++ {
		Step(&s, Input{}, testDT)
	}
	if s.StumbleLeft != 0 {
		t.Fatalf("StumbleLeft = %v, want 0 after stumble", s.StumbleLeft)
	}

	want := RunSpeed * StumbleSpeedFraction * StumbleTime
	if math.Abs(s.Z-want) > 2*RunSpeed*StumbleSpeedFraction*testDT {
		t.Fatalf("advance over stumble = %v, want %v", s.Z, want)
	}

	before := s.Z
	Step(&s, Input{}, testDT)
	if math.Abs((s.Z-before)-RunSpeed*testDT) > 1e-9 {
		t.Fatalf("post-stumble step advanced %v, want %v", s.Z-before, RunSpeed*testDT)
	}
}
