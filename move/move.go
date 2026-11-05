// Package move holds the hero's running, jumping and gravity as plain values,
// independent of raylib, so the movement rules can be tested headless and
// reused unchanged on every platform.
package move

// RunSpeed is the hero's forward speed in meters per second.
const RunSpeed = 12.0

// SteerSpeed is the hero's sideways speed in meters per second at full steer.
const SteerSpeed = 8.0

// Gravity is the downward acceleration in meters per second squared applied
// while the hero is airborne.
const Gravity = 9.5

// JumpSpeed is the upward vertical speed in meters per second given at the
// start of a jump.
const JumpSpeed = 4.4

// Hero is the hero's position and vertical state.
//
// X, Y and Z are the sideways, height and forward positions in meters. Vy is
// the vertical speed in meters per second. Airborne reports whether the hero
// is off the floor.
type Hero struct {
	X, Y, Z, Vy float64
	Airborne    bool
}

// Input is one frame's steering and jump input.
//
// Steer ranges from -1 (full left) to 1 (full right). Jump is true on the
// press frame.
type Input struct {
	Steer float64
	Jump  bool
}

// NewHero returns a hero on the floor, not airborne, at the origin.
func NewHero() Hero {
	return Hero{}
}

// Step advances h one frame of dt seconds.
//
// The hero always runs forward at RunSpeed. Steering is clamped to -1..1 and
// moves the hero sideways on the ground and in the air. A jump press while
// grounded starts a jump; a press while airborne changes nothing. Airborne,
// gravity pulls Vy down and Y follows Vy. The hero lands when Y reaches zero
// and keeps running forward.
func Step(h Hero, in Input, dt float64) Hero {
	h.Z += RunSpeed * dt

	steer := in.Steer
	if steer < -1 {
		steer = -1
	} else if steer > 1 {
		steer = 1
	}
	h.X += steer * SteerSpeed * dt

	if !h.Airborne && in.Jump {
		h.Vy = JumpSpeed
		h.Airborne = true
	}

	if h.Airborne {
		h.Vy -= Gravity * dt
		h.Y += h.Vy * dt
		if h.Y <= 0 {
			h.Y = 0
			h.Vy = 0
			h.Airborne = false
		}
	}

	return h
}
