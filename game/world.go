package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// Static corridor geometry, all measured in world meters.
const (
	floorThickness = float32(0.25)
	floorY         = -floorThickness / 2
	wallHeight     = float32(3)
	wallThickness  = float32(0.3)
	lineWidth      = float32(0.06)
	lineHeight     = float32(0.02)
)

var (
	colorFloor    = rl.Color{R: 88, G: 94, B: 104, A: 255}
	colorWall     = rl.Color{R: 62, G: 68, B: 78, A: 255}
	colorLine     = rl.Color{R: 150, G: 156, B: 164, A: 255}
	colorDeskTop  = rl.Color{R: 118, G: 84, B: 52, A: 255}
	colorDeskLeg  = rl.Color{R: 84, G: 60, B: 40, A: 255}
	colorLampStem = rl.Color{R: 52, G: 56, B: 62, A: 255}
	colorLampBulb = rl.Color{R: 255, G: 214, B: 120, A: 255}
)

type box struct {
	center rl.Vector3
	size   rl.Vector3
	color  rl.Color
}

type cylinderShape struct {
	base         rl.Vector3
	radiusTop    float32
	radiusBottom float32
	height       float32
	color        rl.Color
}

type sphereShape struct {
	center rl.Vector3
	radius float32
	color  rl.Color
}

// World owns the static corridor geometry. It is built once in NewWorld and
// drawn every frame from raylib primitives.
type World struct {
	boxes []box
	stem  cylinderShape
	bulb  sphereShape
}

func NewWorld(layout *Layout) *World {
	w := &World{}

	corridorWidth := layout.LaneWidth * float32(layout.LaneCount)
	floorHalf := corridorWidth / 2

	for _, seg := range layout.Segments {
		length := seg.Length
		centerZ := seg.Start + length/2

		// One wall slab per side per segment; the corridor stays open on top.
		wallSize := rl.Vector3{X: wallThickness, Y: wallHeight, Z: length}
		w.boxes = append(w.boxes,
			box{
				center: rl.Vector3{X: floorHalf + wallThickness/2, Y: wallHeight / 2, Z: centerZ},
				size:   wallSize,
				color:  colorWall,
			},
			box{
				center: rl.Vector3{X: -(floorHalf + wallThickness/2), Y: wallHeight / 2, Z: centerZ},
				size:   wallSize,
				color:  colorWall,
			},
		)

		if seg.Obstacle == ObstaclePit {
			// Cut the floor out of the pit lane only; the other lanes stay solid.
			pitCenter := layout.LaneX(seg.ObstacleLane)
			pitLeft := pitCenter - layout.LaneWidth/2
			pitRight := pitCenter + layout.LaneWidth/2
			if pitLeft > -floorHalf {
				w.addFloor(-floorHalf, pitLeft, centerZ, length)
			}
			if pitRight < floorHalf {
				w.addFloor(pitRight, floorHalf, centerZ, length)
			}
			continue
		}

		w.addFloor(-floorHalf, floorHalf, centerZ, length)
	}

	// Two lane divider lines run the whole corridor.
	dividerSize := rl.Vector3{X: lineWidth, Y: lineHeight, Z: layout.EndZ}
	dividerCenter := rl.Vector3{Y: lineHeight / 2, Z: layout.EndZ / 2}
	for lane := 1; lane < layout.LaneCount; lane++ {
		dividerCenter.X = layout.LaneX(lane) - layout.LaneWidth/2
		w.boxes = append(w.boxes, box{center: dividerCenter, size: dividerSize, color: colorLine})
	}

	w.buildDesk(layout)

	return w
}

// addFloor appends one floor slab covering x0..x1 in the given Z range.
func (w *World) addFloor(x0, x1, centerZ, length float32) {
	w.boxes = append(w.boxes, box{
		center: rl.Vector3{X: (x0 + x1) / 2, Y: floorY, Z: centerZ},
		size:   rl.Vector3{X: x1 - x0, Y: floorThickness, Z: length},
		color:  colorFloor,
	})
}

// buildDesk places the level-end desk on the final floor slab at EndZ.
func (w *World) buildDesk(layout *Layout) {
	topWidth := layout.LaneWidth*float32(layout.LaneCount) - 0.8
	topDepth := float32(1.0)
	topThickness := float32(0.12)
	topHeight := float32(0.94)
	deskZ := layout.EndZ - topDepth/2

	w.boxes = append(w.boxes, box{
		center: rl.Vector3{X: 0, Y: topHeight, Z: deskZ},
		size:   rl.Vector3{X: topWidth, Y: topThickness, Z: topDepth},
		color:  colorDeskTop,
	})

	legSize := rl.Vector3{X: 0.25, Y: topHeight, Z: 0.8}
	for _, x := range []float32{-(topWidth/2 - 0.2), topWidth/2 - 0.2} {
		w.boxes = append(w.boxes, box{
			center: rl.Vector3{X: x, Y: topHeight / 2, Z: deskZ},
			size:   legSize,
			color:  colorDeskLeg,
		})
	}

	stemX := -topWidth / 4
	w.stem = cylinderShape{
		base:         rl.Vector3{X: stemX, Y: topHeight + topThickness/2, Z: deskZ},
		radiusTop:    0.05,
		radiusBottom: 0.07,
		height:       0.5,
		color:        colorLampStem,
	}
	w.bulb = sphereShape{
		center: rl.Vector3{X: stemX, Y: topHeight + topThickness/2 + 0.5 + 0.12, Z: deskZ},
		radius: 0.14,
		color:  colorLampBulb,
	}
}

func (w *World) Update(dt float32, g *Game) {
	g.Cam.Follow(g.Player, dt)
}

func (w *World) Draw(g *Game) {
	rl.BeginMode3D(g.Cam.Camera3D)
	defer rl.EndMode3D()

	for _, b := range w.boxes {
		rl.DrawCubeV(b.center, b.size, b.color)
	}
	rl.DrawCylinder(w.stem.base, w.stem.radiusTop, w.stem.radiusBottom, w.stem.height, 16, w.stem.color)
	rl.DrawSphere(w.bulb.center, w.bulb.radius, w.bulb.color)
}
