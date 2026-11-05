package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// Player is the courier's state. Movement arrives later; the stub only renders
// a small capsule at its position.
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

func (p *Player) Update(dt float32, g *Game) {}

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
