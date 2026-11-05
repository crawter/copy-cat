package render

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"copycat/enemy"
	"copycat/leveldata"
	"copycat/move"
)

// Camera placement in meters, relative to the hero.
const (
	cameraHeightMeters       = 2.5
	cameraBackMeters         = 6.0
	cameraTargetHeightMeters = 0.5
	cameraLookAheadMeters    = 6.0
)

// The counterweight's half extent in meters. It mirrors the size enemy's
// defeat reads, so the drawn counterweight matches what the hero can land on.
const counterweightHalfMeters = 0.6

// The cheap film set's palette.
var (
	colorSky       = rl.Color{R: 16, G: 14, B: 22, A: 255}
	colorFloor     = rl.Color{R: 54, G: 42, B: 32, A: 255}
	colorWall      = rl.Color{R: 86, G: 60, B: 40, A: 255}
	colorWallTrim  = rl.Color{R: 64, G: 46, B: 32, A: 255}
	colorDoor      = rl.Color{R: 46, G: 34, B: 26, A: 255}
	colorMoon      = rl.Color{R: 222, G: 214, B: 176, A: 255}
	colorCat       = rl.Color{R: 22, G: 20, B: 26, A: 255}
	colorCatEye    = rl.Color{R: 238, G: 214, B: 88, A: 255}
	colorBandage   = rl.Color{R: 224, G: 224, B: 214, A: 255}
	colorFluff     = rl.Color{R: 236, G: 230, B: 220, A: 255}
	colorRig       = rl.Color{R: 120, G: 116, B: 124, A: 255}
	colorRigLight  = rl.Color{R: 240, G: 226, B: 140, A: 255}
	colorReel      = rl.Color{R: 180, G: 176, B: 168, A: 255}
	colorReelHub   = rl.Color{R: 60, G: 56, B: 54, A: 255}
	colorCrane     = rl.Color{R: 208, G: 168, B: 40, A: 255}
	colorCraneDark = rl.Color{R: 40, G: 36, B: 34, A: 255}
	colorCounter   = rl.Color{R: 70, G: 66, B: 74, A: 255}
	colorPipe      = rl.Color{R: 110, G: 108, B: 116, A: 255}
	colorBoard     = rl.Color{R: 30, G: 28, B: 32, A: 255}
	colorBoardTop  = rl.Color{R: 238, G: 236, B: 230, A: 255}
	colorSign      = rl.Color{R: 24, G: 22, B: 28, A: 255}
	colorSignPanel = rl.Color{R: 238, G: 236, B: 230, A: 255}
	colorText      = rl.Color{R: 240, G: 236, B: 228, A: 255}
	colorTextDim   = rl.Color{R: 180, G: 172, B: 160, A: 255}
	colorHUD       = rl.Color{R: 250, G: 246, B: 236, A: 255}
)

func vec3(p leveldata.Point) rl.Vector3 {
	return rl.Vector3{X: float32(p.X), Y: float32(p.Y), Z: float32(p.Z)}
}

func drawSet(w World, dying bool) {
	cam := buildCamera(w.Hero)
	rl.BeginMode3D(cam)
	drawCorridor(w.Take)
	drawSetDressing(w.Take)
	drawCrane(w.Crane)
	if dying {
		drawFluff(w.Hero)
	} else {
		drawHero(w.Hero)
	}
	rl.EndMode3D()
}

func buildCamera(h move.Hero) rl.Camera3D {
	px, py, pz := cameraPos(h)
	tx, ty, tz := cameraTarget(h)
	return rl.Camera3D{
		Position:   rl.Vector3{X: float32(px), Y: float32(py), Z: float32(pz)},
		Target:     rl.Vector3{X: float32(tx), Y: float32(ty), Z: float32(tz)},
		Up:         rl.Vector3{X: 0, Y: 1, Z: 0},
		Fovy:       60,
		Projection: rl.CameraPerspective,
	}
}

func drawCorridor(t leveldata.Take) {
	half := float32(t.Width / 2)
	length := float32(t.Length)
	center := float32(t.Length / 2)
	wallThick := float32(0.3)
	wallHeight := float32(4.0)

	// Floor.
	rl.DrawPlane(rl.Vector3{X: 0, Y: 0, Z: center}, rl.Vector2{X: half * 2, Y: length}, colorFloor)
	// Side walls.
	rl.DrawCubeV(rl.Vector3{X: -half, Y: wallHeight / 2, Z: center}, rl.Vector3{X: wallThick, Y: wallHeight, Z: length}, colorWall)
	rl.DrawCubeV(rl.Vector3{X: half, Y: wallHeight / 2, Z: center}, rl.Vector3{X: wallThick, Y: wallHeight, Z: length}, colorWall)
	// Ceiling.
	rl.DrawCubeV(rl.Vector3{X: 0, Y: wallHeight + wallThick/2, Z: center}, rl.Vector3{X: half * 2, Y: wallThick, Z: length}, colorWallTrim)

	// Fake doors painted on both inner wall faces every 80 m.
	inner := half - wallThick/2
	doorW := float32(0.06)
	doorH := float32(2.0)
	doorD := float32(0.9)
	for z := float32(40); z < length; z += 80 {
		rl.DrawCubeV(rl.Vector3{X: -inner + doorW/2, Y: doorH/2 + 0.2, Z: z}, rl.Vector3{X: doorW, Y: doorH, Z: doorD}, colorDoor)
		rl.DrawCubeV(rl.Vector3{X: inner - doorW/2, Y: doorH/2 + 0.2, Z: z}, rl.Vector3{X: doorW, Y: doorH, Z: doorD}, colorDoor)
	}
}

func drawSetDressing(t leveldata.Take) {
	drawMoon()
	drawClapboard(leveldata.Point{X: 3.2, Y: 1.0, Z: 12})
	for _, rig := range t.Rigs {
		drawRig(rig, t)
	}
	for _, reel := range t.Reels {
		drawReel(reel)
	}
	for _, cp := range t.Checkpoints {
		drawClapboard(cp.Position)
	}
	drawSign(t)
}

func drawMoon() {
	// A cardboard moon on a string, hanging from the tunnel ceiling.
	rl.DrawLine3D(rl.Vector3{X: 0, Y: 4, Z: 40}, rl.Vector3{X: 0, Y: 6.2, Z: 40}, colorRig)
	rl.DrawCircle3D(rl.Vector3{X: 0, Y: 6.6, Z: 40}, 1.1, rl.Vector3{X: 0, Y: 0, Z: 1}, 0, colorMoon)
}

func drawClapboard(p leveldata.Point) {
	x := float32(p.X)
	y := float32(p.Y)
	z := float32(p.Z)

	// Stand reaching up to the board's bottom edge.
	rl.DrawCubeV(rl.Vector3{X: x, Y: (y - 0.35) / 2, Z: z}, rl.Vector3{X: 0.08, Y: y - 0.35, Z: 0.08}, colorBoardTop)
	// Board.
	rl.DrawCubeV(rl.Vector3{X: x, Y: y, Z: z}, rl.Vector3{X: 1.1, Y: 0.7, Z: 0.08}, colorBoard)
	// Hinged top stripe.
	rl.DrawCubeV(rl.Vector3{X: x, Y: y + 0.37, Z: z}, rl.Vector3{X: 1.1, Y: 0.12, Z: 0.1}, colorBoardTop)
	// Two diagonal stripes on the face toward the camera.
	rl.DrawCubeV(rl.Vector3{X: x - 0.28, Y: y - 0.03, Z: z - 0.04}, rl.Vector3{X: 0.42, Y: 0.1, Z: 0.02}, colorBoardTop)
	rl.DrawCubeV(rl.Vector3{X: x + 0.28, Y: y + 0.07, Z: z - 0.04}, rl.Vector3{X: 0.42, Y: 0.1, Z: 0.02}, colorBoardTop)
}

func drawRig(r leveldata.Rig, t leveldata.Take) {
	half := float32(t.Width / 2)
	z := float32(r.Position.Z)
	// A fallen rig lies across the whole corridor, so the beam is centered on
	// the corridor's middle axis; r.Position.Z carries its forward spot.
	rl.DrawCubeV(rl.Vector3{X: 0, Y: 0.3, Z: z}, rl.Vector3{X: half * 2, Y: 0.6, Z: 0.6}, colorRig)
	rl.DrawSphere(rl.Vector3{X: -half + 0.5, Y: 0.5, Z: z}, 0.18, colorRigLight)
	rl.DrawSphere(rl.Vector3{X: 0, Y: 0.5, Z: z}, 0.18, colorRigLight)
	rl.DrawSphere(rl.Vector3{X: half - 0.5, Y: 0.5, Z: z}, 0.18, colorRigLight)
}

func drawReel(r leveldata.Reel) {
	x := float32(r.Position.X)
	y := float32(r.Position.Y + 0.05)
	z := float32(r.Position.Z)
	// A film reel standing on edge, its hub darker than its metal rim.
	rl.DrawCylinder(rl.Vector3{X: x, Y: y, Z: z}, 0.35, 0.35, 0.1, 16, colorReel)
	rl.DrawCylinder(rl.Vector3{X: x, Y: y + 0.05, Z: z}, 0.12, 0.12, 0.02, 10, colorReelHub)
}

func drawSign(t leveldata.Take) {
	half := float32(t.Width / 2)
	z := float32(t.Exit.Z)
	// Beam across the corridor carrying the sign.
	rl.DrawCubeV(rl.Vector3{X: 0, Y: 3.6, Z: z}, rl.Vector3{X: half * 2, Y: 0.15, Z: 0.15}, colorRig)
	// Board hanging from the beam.
	rl.DrawCubeV(rl.Vector3{X: 0, Y: 2.9, Z: z}, rl.Vector3{X: 1.6, Y: 1.1, Z: 0.1}, colorSign)
	// Painted panel where the word CUT sits.
	rl.DrawCubeV(rl.Vector3{X: 0, Y: 2.9, Z: z - 0.05}, rl.Vector3{X: 1.3, Y: 0.8, Z: 0.02}, colorSignPanel)
	// Two diagonal cuts hinting at the word CUT.
	rl.DrawCubeV(rl.Vector3{X: -0.35, Y: 2.95, Z: z - 0.05}, rl.Vector3{X: 0.4, Y: 0.12, Z: 0.02}, colorSign)
	rl.DrawCubeV(rl.Vector3{X: 0.35, Y: 2.85, Z: z - 0.05}, rl.Vector3{X: 0.4, Y: 0.12, Z: 0.02}, colorSign)
}

func drawCrane(c enemy.Crane) {
	body := c.Body()
	if c.Dropped() {
		drawHeap(body.Pos)
		return
	}
	pos := vec3(body.Pos)
	size := rl.Vector3{
		X: float32(body.HalfWidth * 2),
		Y: float32(body.Height),
		Z: float32(body.HalfLen * 2),
	}
	rl.DrawCubeV(pos, size, colorCrane)
	// Hazard stripe along the top of the rolling body.
	rl.DrawCubeV(
		rl.Vector3{X: pos.X, Y: pos.Y + float32(body.Height), Z: pos.Z},
		rl.Vector3{X: size.X, Y: 0.08, Z: size.Z},
		colorCraneDark,
	)
	// Counterweight sitting on top of the body, the size enemy's defeat uses.
	cw := float32(counterweightHalfMeters)
	rl.DrawCubeV(
		rl.Vector3{X: pos.X, Y: pos.Y + float32(body.Height) + cw, Z: pos.Z},
		rl.Vector3{X: cw * 2, Y: cw * 2, Z: cw * 2},
		colorCounter,
	)
}

func drawHeap(p leveldata.Point) {
	x := float32(p.X)
	z := float32(p.Z)
	// A collapsed heap of pipes stays where the crane fell for the rest of the
	// take.
	rl.DrawCylinderEx(rl.Vector3{X: x - 0.6, Y: 0.15, Z: z - 0.4}, rl.Vector3{X: x + 0.6, Y: 0.15, Z: z + 0.4}, 0.12, 0.12, 10, colorPipe)
	rl.DrawCylinderEx(rl.Vector3{X: x - 0.5, Y: 0.3, Z: z + 0.5}, rl.Vector3{X: x + 0.5, Y: 0.3, Z: z - 0.5}, 0.12, 0.12, 10, colorPipe)
	rl.DrawCylinderEx(rl.Vector3{X: x - 0.7, Y: 0.45, Z: z}, rl.Vector3{X: x + 0.7, Y: 0.45, Z: z}, 0.1, 0.1, 10, colorPipe)
	rl.DrawSphere(rl.Vector3{X: x, Y: 0.5, Z: z}, 0.25, colorCounter)
}

func drawHero(h move.Hero) {
	x := float32(h.X)
	y := float32(h.Y)
	z := float32(h.Z)

	// Body.
	rl.DrawCubeV(rl.Vector3{X: x, Y: y + 0.35, Z: z}, rl.Vector3{X: 0.5, Y: 0.4, Z: 0.9}, colorCat)
	// Head.
	rl.DrawSphere(rl.Vector3{X: x, Y: y + 0.62, Z: z + 0.32}, 0.2, colorCat)
	// Ears.
	rl.DrawCubeV(rl.Vector3{X: x - 0.11, Y: y + 0.83, Z: z + 0.28}, rl.Vector3{X: 0.07, Y: 0.13, Z: 0.07}, colorCat)
	rl.DrawCubeV(rl.Vector3{X: x + 0.11, Y: y + 0.83, Z: z + 0.28}, rl.Vector3{X: 0.07, Y: 0.13, Z: 0.07}, colorCat)
	// Bandage around the right ear.
	rl.DrawCubeV(rl.Vector3{X: x + 0.11, Y: y + 0.83, Z: z + 0.28}, rl.Vector3{X: 0.1, Y: 0.09, Z: 0.1}, colorBandage)
	// Eyes.
	rl.DrawSphere(rl.Vector3{X: x - 0.07, Y: y + 0.66, Z: z + 0.44}, 0.035, colorCatEye)
	rl.DrawSphere(rl.Vector3{X: x + 0.07, Y: y + 0.66, Z: z + 0.44}, 0.035, colorCatEye)
	// Tail.
	rl.DrawCylinderEx(rl.Vector3{X: x, Y: y + 0.4, Z: z - 0.45}, rl.Vector3{X: x, Y: y + 0.75, Z: z - 0.75}, 0.05, 0.05, 8, colorCat)
}

func drawFluff(h move.Hero) {
	x := float32(h.X)
	y := float32(h.Y + 0.5)
	z := float32(h.Z)
	// The farce death: the cat puffs into a cloud of fluff, nothing like real
	// pain.
	rl.DrawSphere(rl.Vector3{X: x, Y: y, Z: z}, 0.35, colorFluff)
	rl.DrawSphere(rl.Vector3{X: x - 0.25, Y: y - 0.1, Z: z - 0.1}, 0.25, colorFluff)
	rl.DrawSphere(rl.Vector3{X: x + 0.25, Y: y - 0.1, Z: z + 0.1}, 0.25, colorFluff)
	rl.DrawSphere(rl.Vector3{X: x + 0.15, Y: y + 0.2, Z: z - 0.15}, 0.2, colorFluff)
	rl.DrawSphere(rl.Vector3{X: x - 0.15, Y: y + 0.22, Z: z + 0.15}, 0.2, colorFluff)
}

func drawTitle() {
	centerText(gameName, screenHeight/2-60, 72, colorText)
	centerText(startPrompt, screenHeight/2+24, 24, colorTextDim)
}

func drawRecap(w World) {
	if w.Line != "" {
		centerText(w.Line, 240, 28, colorText)
	}
	centerText(RecapText(w.ReelsFound, w.ReelsTotal), 300, 32, colorHUD)
	centerText(startPrompt, 440, 22, colorTextDim)
}

func drawGameOver(w World) {
	if w.Line != "" {
		centerText(w.Line, 220, 28, colorText)
	}
	centerText(retryPrompt, 320, 26, colorText)
	centerText(leavePrompt, 372, 22, colorTextDim)
}

func drawHUD(w World) {
	rl.DrawText(TakeText(), 24, 24, 24, colorHUD)
	reels := ReelsText(w.ReelsFound, w.ReelsTotal)
	width := rl.MeasureText(reels, 24)
	rl.DrawText(reels, screenWidth-width-24, 24, 24, colorHUD)
	drawLives(w.Lives)
	if w.Line != "" {
		centerText(w.Line, screenHeight-56, 22, colorTextDim)
	}
}

func drawLives(n int) {
	for i := 0; i < n; i++ {
		drawCatHead(int32(24+i*30), 58)
	}
}

func drawCatHead(x, y int32) {
	// Ears.
	rl.DrawTriangle(rl.Vector2{X: float32(x + 2), Y: float32(y + 8)}, rl.Vector2{X: float32(x + 5), Y: float32(y + 2)}, rl.Vector2{X: float32(x + 9), Y: float32(y + 9)}, colorCat)
	rl.DrawTriangle(rl.Vector2{X: float32(x + 11), Y: float32(y + 9)}, rl.Vector2{X: float32(x + 15), Y: float32(y + 2)}, rl.Vector2{X: float32(x + 18), Y: float32(y + 8)}, colorCat)
	// Head.
	rl.DrawCircle(x+10, y+12, 8, colorCat)
	// Eyes.
	rl.DrawCircle(x+7, y+11, 1, colorCatEye)
	rl.DrawCircle(x+13, y+11, 1, colorCatEye)
}

func centerText(text string, y, fontSize int32, color rl.Color) {
	width := rl.MeasureText(text, fontSize)
	rl.DrawText(text, (screenWidth-width)/2, y, fontSize, color)
}
