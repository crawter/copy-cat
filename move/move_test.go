package move

import (
	"math"
	"testing"
)

func TestStepRunsForward(t *testing.T) {
	const dt = 0.5

	got := Step(NewHero(), Input{}, dt)
	if want := RunSpeed * dt; got.Z != want {
		t.Fatalf("Step Z = %v, want %v", got.Z, want)
	}
}

func TestStepJumpsFromGround(t *testing.T) {
	const dt = 1.0 / 240.0

	got := Step(NewHero(), Input{Jump: true}, dt)
	if !got.Airborne {
		t.Fatal("hero should be airborne after a jump press on the ground")
	}
	if got.Vy <= 0 {
		t.Fatalf("Step Vy = %v, want a positive upward speed", got.Vy)
	}
}

func TestStepClampsSteer(t *testing.T) {
	tests := []struct {
		name  string
		steer float64
		want  float64
	}{
		{name: "above one", steer: 2.0, want: SteerSpeed},
		{name: "below minus one", steer: -3.0, want: -SteerSpeed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Step(NewHero(), Input{Steer: tt.steer}, 1.0)
			if got.X != tt.want {
				t.Fatalf("Step X = %v, want %v", got.X, tt.want)
			}
		})
	}
}

func TestStepSteersOnGroundAndInAir(t *testing.T) {
	tests := []struct {
		name     string
		airborne bool
		steer    float64
		dt       float64
		want     float64
	}{
		{name: "ground right", steer: 1.0, dt: 0.5, want: SteerSpeed * 0.5},
		{name: "ground left", steer: -1.0, dt: 0.5, want: -SteerSpeed * 0.5},
		{name: "air right", airborne: true, steer: 1.0, dt: 0.5, want: SteerSpeed * 0.5},
		{name: "air left", airborne: true, steer: -0.5, dt: 0.5, want: -0.5 * SteerSpeed * 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHero()
			if tt.airborne {
				h = Step(h, Input{Jump: true}, 1.0/240.0)
				if !h.Airborne {
					t.Fatal("hero should be airborne after a jump press")
				}
			}
			got := Step(h, Input{Steer: tt.steer}, tt.dt)
			if got.X != tt.want {
				t.Fatalf("Step X = %v, want %v", got.X, tt.want)
			}
		})
	}
}

func TestJumpHeight(t *testing.T) {
	const dt = 1.0 / 240.0

	h := NewHero()
	h = Step(h, Input{Jump: true}, dt)
	peak := h.Y
	for h.Airborne {
		h = Step(h, Input{}, dt)
		if h.Y > peak {
			peak = h.Y
		}
	}

	if math.Abs(peak-1.0) > 0.1 {
		t.Fatalf("peak Y = %v m, want within 0.1 m of 1 m", peak)
	}
}

func TestJumpLandsAboutElevenMetersPastTakeoff(t *testing.T) {
	const dt = 1.0 / 240.0

	h := NewHero()
	takeoffZ := h.Z
	h = Step(h, Input{Jump: true}, dt)
	for h.Airborne {
		h = Step(h, Input{}, dt)
	}

	forward := h.Z - takeoffZ
	if math.Abs(forward-11.0) > 1.0 {
		t.Fatalf("jump landed %v m past takeoff, want within 1 m of 11 m", forward)
	}
}

func TestNoDoubleJump(t *testing.T) {
	const dt = 1.0 / 240.0

	start := NewHero()
	start = Step(start, Input{Jump: true}, dt)
	if !start.Airborne {
		t.Fatal("hero should be airborne after a jump press")
	}

	plain := start
	pressed := start
	for step := 0; plain.Airborne; step++ {
		in := Input{}
		if step == 20 {
			in.Jump = true
		}
		plain = Step(plain, Input{}, dt)
		pressed = Step(pressed, in, dt)
		if pressed != plain {
			t.Fatalf("step %d: mid-air jump press changed the arc: got %+v, want %+v", step, pressed, plain)
		}
	}
}

func TestLanding(t *testing.T) {
	const dt = 1.0 / 240.0

	h := NewHero()
	h = Step(h, Input{Jump: true}, dt)
	for h.Airborne {
		h = Step(h, Input{}, dt)
	}

	if h.Y != 0 {
		t.Fatalf("Y after landing = %v, want 0", h.Y)
	}
	if h.Airborne {
		t.Fatal("hero should not be airborne after landing")
	}
	if h.Vy != 0 {
		t.Fatalf("Vy after landing = %v, want 0", h.Vy)
	}

	before := h.Z
	h = Step(h, Input{}, dt)
	if want := before + RunSpeed*dt; h.Z != want {
		t.Fatalf("Z after landing step = %v, want %v (hero should keep running)", h.Z, want)
	}
}
