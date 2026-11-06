// Package move holds the hero's corridor movement physics: running,
// steering, jumping and stumbling. It is pure Go and imports no raylib, so
// it can be tested headless and reused unchanged on Android.
package move

// RunSpeed is how far the hero advances along the corridor each second,
// in world units per second, while running at full speed.
const RunSpeed = 8.0

// SteerSpeed is how fast the hero shifts sideways, in world units per second.
const SteerSpeed = 6.0

// JumpSpeed is the upward speed given when the hero jumps from the ground,
// in world units per second.
const JumpSpeed = 14.0

// Gravity is the constant downward acceleration applied while the hero is
// airborne, in world units per second squared.
const Gravity = 32.0

// CorridorHalfWidth is the distance from the corridor centre to each side
// wall, in world units. Steering clamps the hero inside this bound.
const CorridorHalfWidth = 3.0

// StumbleTime is how long one stumble slows the hero, in seconds.
const StumbleTime = 1.0

// StumbleSpeedFraction is the fraction of RunSpeed the hero keeps while
// stumbling.
const StumbleSpeedFraction = 0.5

// MaxJumpDistance is the forward distance covered by one full-speed jump,
// from takeoff to landing. It is derived from the movement constants so the
// track generator can bound every gap without a second hardcoded distance.
const MaxJumpDistance = RunSpeed * 2 * JumpSpeed / Gravity

// State holds the hero's position and movement state in the corridor.
// X is the sideways position, Y the height above the ground and Z the
// distance run along the corridor.
type State struct {
	X           float64
	Y           float64
	Z           float64
	VY          float64
	OnGround    bool
	StumbleLeft float64
}

// Input holds the player's movement intent for one step.
type Input struct {
	Left  bool
	Right bool
	Jump  bool
}

// Step advances s by dt seconds using in as the player's intent.
func Step(s *State, in Input, dt float64) {
	if s.StumbleLeft > 0 {
		s.Z += RunSpeed * StumbleSpeedFraction * dt
		s.StumbleLeft -= dt
		if s.StumbleLeft < 0 {
			s.StumbleLeft = 0
		}
	} else {
		s.Z += RunSpeed * dt
	}

	if in.Left {
		s.X -= SteerSpeed * dt
	}
	if in.Right {
		s.X += SteerSpeed * dt
	}
	if s.X > CorridorHalfWidth {
		s.X = CorridorHalfWidth
	}
	if s.X < -CorridorHalfWidth {
		s.X = -CorridorHalfWidth
	}

	if in.Jump && s.OnGround {
		s.VY = JumpSpeed
		s.OnGround = false
	}

	if !s.OnGround {
		s.VY -= Gravity * dt
		s.Y += s.VY * dt
		if s.Y <= 0 {
			s.Y = 0
			s.VY = 0
			s.OnGround = true
		}
	}
}

// Stumble begins a stumble, slowing the hero for StumbleTime seconds.
func Stumble(s *State) {
	s.StumbleLeft = StumbleTime
}
