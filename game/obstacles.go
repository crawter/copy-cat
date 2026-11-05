package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// ObstacleKind names one obstacle the layout can place in a lane.
type ObstacleKind int

const (
	ObstacleNone ObstacleKind = iota
	ObstaclePit
	ObstacleSpikes
	ObstacleCrates
)

// obstacle is one obstacle built from a layout segment.
type obstacle struct {
	kind  ObstacleKind
	lane  int
	start float32
	end   float32
}

// Obstacles owns obstacle collision and rendering.
type Obstacles struct {
	items []obstacle
}

// NewObstacles builds one entry per obstacle segment in the layout.
func NewObstacles(layout *Layout) *Obstacles {
	o := &Obstacles{}
	for _, seg := range layout.Segments {
		if seg.Obstacle == ObstacleNone {
			continue
		}
		o.items = append(o.items, obstacle{
			kind:  seg.Obstacle,
			lane:  seg.ObstacleLane,
			start: seg.Start,
			end:   seg.Start + seg.Length,
		})
	}
	return o
}

// Update checks the player against each obstacle once per frame. A fatal
// collision only switches the scene; the death scene and respawn own the rest.
func (o *Obstacles) Update(dt float32, g *Game) {
	if g.Scene != ScenePlay {
		return
	}

	for _, ob := range o.items {
		switch ob.kind {
		case ObstaclePit:
			if g.Player.Z >= ob.start && g.Player.Z <= ob.end && g.Player.Y <= 0 {
				g.Scene = SceneDeath
				return
			}
		case ObstacleSpikes:
			if g.Player.Lane == ob.lane && g.Player.Grounded &&
				g.Player.Z >= ob.start && g.Player.Z <= ob.end {
				g.Scene = SceneDeath
				return
			}
		case ObstacleCrates:
			if g.Player.Lane == ob.lane && g.Player.Grounded &&
				g.Player.Z >= ob.start && g.Player.Z <= ob.end {
				g.Player.Z = ob.start
			}
		}
	}
}

// Draw renders every obstacle in one 3D pass.
func (o *Obstacles) Draw(g *Game) {
	rl.BeginMode3D(g.Cam.Camera3D)
	defer rl.EndMode3D()

	for _, ob := range o.items {
		switch ob.kind {
		case ObstaclePit:
			o.drawPit(g, ob)
		case ObstacleSpikes:
			o.drawSpikes(g, ob)
		case ObstacleCrates:
			o.drawCrates(g, ob)
		}
	}
}

// drawPit draws a dark recessed box across all lanes, edged with hazard stripes.
func (o *Obstacles) drawPit(g *Game, ob obstacle) {
	width := float32(g.Layout.LaneCount) * g.Layout.LaneWidth
	length := ob.end - ob.start
	const depth = float32(1.5)

	rl.DrawCube(
		rl.Vector3{X: 0, Y: -depth / 2, Z: (ob.start + ob.end) / 2},
		width, depth, length,
		rl.Color{R: 18, G: 20, B: 26, A: 255},
	)

	o.drawHazardEdge(g, width, ob.start)
	o.drawHazardEdge(g, width, ob.end)
}

// drawHazardEdge draws alternating yellow and black stripes along one pit edge.
func (o *Obstacles) drawHazardEdge(g *Game, width, z float32) {
	const stripeWidth = float32(0.3)
	const stripeHeight = float32(0.06)
	const stripeLength = float32(0.35)

	count := int(width/stripeWidth) + 1
	for i := 0; i < count; i++ {
		color := rl.Color{R: 220, G: 190, B: 40, A: 255}
		if i%2 == 1 {
			color = rl.Color{R: 20, G: 20, B: 20, A: 255}
		}
		x := -width/2 + float32(i)*stripeWidth + stripeWidth/2
		rl.DrawCube(
			rl.Vector3{X: x, Y: stripeHeight / 2, Z: z},
			stripeWidth, stripeHeight, stripeLength,
			color,
		)
	}
}

// drawSpikes draws a row of gray cones across the segment in the obstacle lane.
func (o *Obstacles) drawSpikes(g *Game, ob obstacle) {
	x := g.Layout.LaneX(ob.lane)
	const spacing = float32(0.6)

	for z := ob.start; z <= ob.end; z += spacing {
		rl.DrawCylinder(
			rl.Vector3{X: x, Y: 0.25, Z: z},
			0.02, 0.25, 0.5, 12,
			rl.Color{R: 150, G: 155, B: 165, A: 255},
		)
	}
}

// drawCrates draws a meter-tall brown crate stack in the obstacle lane.
func (o *Obstacles) drawCrates(g *Game, ob obstacle) {
	x := g.Layout.LaneX(ob.lane)
	z := (ob.start + ob.end) / 2

	rl.DrawCube(
		rl.Vector3{X: x, Y: 0.5, Z: z},
		1, 1, 1,
		rl.Color{R: 130, G: 90, B: 45, A: 255},
	)
	rl.DrawCube(
		rl.Vector3{X: x, Y: 1.3, Z: z},
		0.7, 0.6, 0.7,
		rl.Color{R: 155, G: 112, B: 60, A: 255},
	)
}
