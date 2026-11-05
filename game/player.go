package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/v5/raylib"
)

// Movement tuning. The auto-run speed plus the jump numbers below give each
// jump about 6.5 meters of forward travel, comfortably clearing the layout's
// four-meter pits at the fixed run speed.
const (
	runSpeed     float32 = 8
	gravity      float32 = 22
	jumpVelocity float32 = 9
	laneEaseRate float32 = 12
)

// Player is the courier's state: lane, position, velocity, and groundedness.
type Player struct {
	Lane     int
	X, Y, Z  float32
	VX, VY   float32
	Grounded bool
}

func NewPlayer() *Player {
	return &Player{
		Lane:     1,
		X:        0,
		Y:        0,
		Z:        0,
		Grounded: true,
	}
}

// Update advances the player one frame: lane selection, horizontal easing,
// forward auto-run, jumping with gravity, and run-distance tracking.
func (p *Player) Update(dt float32, g *Game) {
	// One lane per key press, clamped to the layout's lanes.
	if LaneLeft() {
		p.Lane--
	}
	if LaneRight() {
		p.Lane++
	}
	if p.Lane < 0 {
		p.Lane = 0
	}
	if p.Lane >= g.Layout.LaneCount {
		p.Lane = g.Layout.LaneCount - 1
	}

	// Ease the horizontal position toward the chosen lane's center.
	targetX := g.Layout.LaneX(p.Lane)
	p.X = easeTowards(p.X, targetX, laneEaseRate, dt)

	// Auto-run forward at a constant speed.
	p.Z += runSpeed * dt

	// A press while grounded gives the upward impulse; gravity then pulls the
	// player back down until landing resets Y.
	if Jump() && p.Grounded {
		p.VY = jumpVelocity
		p.Grounded = false
	}
	p.VY -= gravity * dt
	p.Y += p.VY * dt
	if p.Y <= 0 {
		p.Y = 0
		p.VY = 0
		p.Grounded = true
	}

	// Track the furthest point reached on this run.
	if p.Z > g.Run.Distance {
		g.Run.Distance = p.Z
	}
}

// easeTowards moves current toward target with a frame-rate-independent
// exponential ease, returning the new value.
func easeTowards(current, target, rate, dt float32) float32 {
	if dt <= 0 {
		return current
	}
	t := 1 - float32(math.Exp(float64(-rate*dt)))
	return current + (target-current)*t
}

// Draw renders the hero as a capsule at its current position inside its own
// 3D pass through the shared camera.
func (p *Player) Draw(g *Game) {
	rl.BeginMode3D(g.Cam.Camera3D)
	defer rl.EndMode3D()

	rl.DrawCapsule(
		rl.Vector3{X: p.X, Y: p.Y + 0.5, Z: p.Z},
		rl.Vector3{X: p.X, Y: p.Y + 1.6, Z: p.Z},
		0.4, 12, 12,
		rl.Color{R: 190, G: 188, B: 210, A: 255},
	)
}
