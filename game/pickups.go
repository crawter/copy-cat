package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// Stamp is one live collectible in the corridor.
type Stamp struct {
	Lane     int
	X        float32
	Z        float32
	Rotation float32
}

// Pickups owns the collectible stamps: their placement, collision and drawing.
type Pickups struct {
	stamps []Stamp
}

// Stamp dimensions and pickup reach, in meters.
const (
	stampSize       = 0.8
	stampThickness  = 0.06
	stampBorderSize = 0.95
	stampDepth      = 0.9
	stampPickupY    = 1.5
)

// NewPickups builds a live stamp for every stamp spot in the layout.
func NewPickups(layout *Layout) *Pickups {
	pickups := &Pickups{}
	i := 0
	for _, segment := range layout.Segments {
		for _, spot := range segment.Stamps {
			pickups.stamps = append(pickups.stamps, Stamp{
				Lane:     spot.Lane,
				X:        layout.LaneX(spot.Lane),
				Z:        spot.Z,
				Rotation: float32(15 + (i%3)*15),
			})
			i++
		}
	}
	return pickups
}

// Update collects stamps the player touches.
func (p *Pickups) Update(dt float32, g *Game) {
	player := g.Player
	if player == nil {
		return
	}

	halfLane := g.Layout.LaneWidth / 2
	for i := len(p.stamps) - 1; i >= 0; i-- {
		stamp := p.stamps[i]
		if absf(player.X-stamp.X) <= halfLane &&
			absf(player.Z-stamp.Z) <= stampDepth/2 &&
			player.Y < stampPickupY {
			g.Run.Stamps++
			g.Run.TotalStamps++
			p.stamps = append(p.stamps[:i], p.stamps[i+1:]...)
		}
	}
}

// Draw renders the live stamps as flat squares with a lighter border.
func (p *Pickups) Draw(g *Game) {
	rl.BeginMode3D(g.Cam.Camera3D)
	defer rl.EndMode3D()

	for _, stamp := range p.stamps {
		p.drawStamp(stamp)
	}
}

func (p *Pickups) drawStamp(stamp Stamp) {
	rl.PushMatrix()
	defer rl.PopMatrix()

	rl.Translatef(stamp.X, stampThickness/2, stamp.Z)
	rl.Rotatef(stamp.Rotation, 0, 1, 0)

	rl.DrawCube(
		rl.Vector3{},
		stampSize, stampThickness, stampSize,
		rl.Color{R: 214, G: 150, B: 42, A: 255},
	)
	rl.DrawCubeWires(
		rl.Vector3{},
		stampBorderSize, stampThickness+0.01, stampBorderSize,
		rl.Color{R: 255, G: 224, B: 150, A: 255},
	)
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
