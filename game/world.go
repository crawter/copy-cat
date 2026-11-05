package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// World owns the corridor geometry. For the scaffold it draws only a ground
// plane so the launch check has a visible 3D scene.
type World struct{}

func NewWorld(layout *Layout) *World {
	return &World{}
}

func (w *World) Update(dt float32, g *Game) {}

func (w *World) Draw(g *Game) {
	rl.BeginMode3D(g.Cam.Camera3D)
	defer rl.EndMode3D()

	rl.DrawPlane(
		rl.Vector3{X: 0, Y: 0, Z: g.Layout.EndZ / 2},
		rl.Vector2{X: g.Layout.LaneWidth * float32(g.Layout.LaneCount+2), Y: g.Layout.EndZ + 16},
		rl.Color{R: 96, G: 104, B: 118, A: 255},
	)
}
