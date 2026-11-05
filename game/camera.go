package game

import (
	"math"
)

// Follow places the camera behind the player and aims it down the corridor.
// Z is pinned to the player so the view never outruns them; X and Y ease toward
// the player's lane and jump arc so lane changes and jumps stay smooth.
func (c *Camera) Follow(player *Player, dt float32) {
	const (
		backwardOffset = float32(-8)
		followHeight   = float32(3.2)
		eyeHeight      = float32(1.5)
		lookAhead      = float32(6)
		easeX          = float32(14)
		easeY          = float32(10)
	)

	c.Position.Z = player.Z + backwardOffset
	c.Position.X = damp(c.Position.X, player.X, easeX, dt)
	c.Position.Y = damp(c.Position.Y, player.Y+followHeight, easeY, dt)

	c.Target.X = damp(c.Target.X, player.X, easeX, dt)
	c.Target.Y = player.Y + eyeHeight
	c.Target.Z = player.Z + lookAhead
}

// damp moves current toward target with an exponential ease for the frame dt.
func damp(current, target, rate, dt float32) float32 {
	if dt <= 0 {
		return current
	}
	k := float32(1 - math.Exp(-float64(rate*dt)))
	return current + (target-current)*k
}
