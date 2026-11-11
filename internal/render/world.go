package render

import (
	"math"

	"copycat/internal/campaign"
	"copycat/internal/levelgen"
)

// Corridor geometry, in meters. It states the render layer's half of the
// shared world contract the runner and levelgen use: three lanes at x = -2,
// 0 and 2, the floor at y = 0, and a container ramp raising the hero to
// y = 1.5.
const (
	laneWidth      = 2.0
	laneGap        = 0.2
	stripWidth     = laneWidth - laneGap
	corridorHalf   = 3.5
	wallThickness  = 0.4
	wallHeight     = 4.0
	floorThickness = 0.2

	barrierDepth   = 1.0
	lowBarrierH    = 0.7
	highBarrierGap = 1.2
	highBarrierTop = 2.6
	containerH     = 1.5
	pitDepth       = 1.0
	rampStepDepth  = 0.5
	rampSteps      = 3

	drawAhead      = 40.0
	drawBehind     = 12.0
	maxElementSpan = 8.0
	floorTileLen   = 2.0
	propSpacing    = 6.0

	// fogNear and fogFar are the Z distances, in meters, where corridor fog
	// starts and where it fully reaches the backdrop color.
	fogNear = 6.0
	fogFar  = 38.0
)

// laneX returns the x offset, in meters, of a lane index in [0, 2].
func laneX(lane float64) float64 {
	return (lane - 1) * laneWidth
}

// spanOf returns the half-open [start, end) z interval an element occupies.
func spanOf(e levelgen.Element) (float64, float64) {
	switch e.Kind {
	case levelgen.Pit, levelgen.Ramp:
		return e.Z, e.Z + e.Span
	case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
		return e.Z, e.Z + barrierDepth
	default:
		return e.Z, e.Z
	}
}

// DrawWorld builds the corridor draw list for l with theme and draws it.
// Callers own the 3D camera; cam only drives culling.
func DrawWorld(l *levelgen.Level, theme campaign.Theme, cam Cam) {
	DrawCmds(BuildWorld(l, theme, cam))
}

// BuildWorld returns the draw list for a corridor slice around cam. It draws
// three lane strips, side walls, a themed backdrop and, for each element
// inside the draw distance, its themed primitive. Everything farther than
// drawAhead ahead of or drawBehind behind the camera is culled.
func BuildWorld(l *levelgen.Level, theme campaign.Theme, cam Cam) []Cmd {
	b := worldBuilder{
		level: l,
		theme: theme,
		pal:   Palette(theme),
		cam:   cam,
	}
	return b.build()
}

// ApplyFog returns cs with each command's color blended toward the theme
// backdrop by its Z distance from cam.Z. Commands near the camera keep their
// colors; commands at or beyond fogFar are fully backdrop-colored. It copies
// cs, so BuildWorld's exact palette colors stay available to callers.
func ApplyFog(cs []Cmd, cam Cam, pal ThemePalette) []Cmd {
	out := make([]Cmd, len(cs))
	copy(out, cs)
	for i := range out {
		c := &out[i]
		dist := math.Abs(float64(c.Pos.Z) - cam.Z)
		c.Color = blendColor(c.Color, pal.Backdrop, fogAmount(dist))
	}
	return out
}

// fogAmount returns how much to blend toward the backdrop at z distance
// dist, from 0 inside fogNear to 1 at or beyond fogFar.
func fogAmount(dist float64) float64 {
	if dist <= fogNear {
		return 0
	}
	if dist >= fogFar {
		return 1
	}
	return (dist - fogNear) / (fogFar - fogNear)
}

type worldBuilder struct {
	level *levelgen.Level
	theme campaign.Theme
	pal   ThemePalette
	cam   Cam
	cmds  []Cmd
}

func (b *worldBuilder) build() []Cmd {
	b.cmds = make([]Cmd, 0, 256)
	b.addBackdrop()
	b.addFloor()
	b.addWalls()
	for _, e := range b.elementsInWindow() {
		b.addElement(e)
	}
	return b.cmds
}

// window returns the corridor slice visible around the camera, clamped to the
// level bounds.
func (b *worldBuilder) window() (float64, float64) {
	start := b.cam.Z - drawBehind
	if start < 0 {
		start = 0
	}
	end := b.cam.Z + drawAhead
	if end > b.level.Length() {
		end = b.level.Length()
	}
	if end < start {
		end = start
	}
	return start, end
}

// camWindow returns the raw cull window, unclamped to the level bounds.
func camWindow(z float64) (float64, float64) {
	return z - drawBehind, z + drawAhead
}

// elementsInWindow returns the elements whose footprint overlaps the cull
// window around the camera, in level order.
func (b *worldBuilder) elementsInWindow() []levelgen.Element {
	start, end := camWindow(b.cam.Z)
	from := start - maxElementSpan
	if from < 0 {
		from = 0
	}
	all := b.level.Elements(from, end)
	out := make([]levelgen.Element, 0, len(all))
	for _, e := range all {
		s, t := spanOf(e)
		if t <= start || s >= end {
			continue
		}
		out = append(out, e)
	}
	return out
}

func (b *worldBuilder) addBackdrop() {
	start, end := b.window()
	if end <= start {
		return
	}
	mid := (start + end) / 2
	length := end - start
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(0, wallHeight+3, mid),
		Size:  Size{X: corridorHalf*2 + 20, Y: 6, Z: length},
		Color: b.pal.Backdrop,
	})
	b.addThemeProps(start, end)
}

func (b *worldBuilder) addThemeProps(start, end float64) {
	for z := math.Ceil(start/propSpacing) * propSpacing; z < end; z += propSpacing {
		b.addThemePropPair(z)
	}
}

func (b *worldBuilder) addThemePropPair(z float64) {
	inner := corridorHalf - 0.3
	switch b.theme {
	case campaign.ThemeHarbor:
		for _, x := range []float64{-inner, inner} {
			b.cmds = append(b.cmds,
				Cmd{Shape: ShapeBox, Pos: v3(x, 0.5, z), Size: Size{X: 1, Y: 1, Z: 1}, Color: b.pal.Ramp},
				Cmd{Shape: ShapeBox, Pos: v3(x, 1.5, z), Size: Size{X: 1, Y: 1, Z: 1}, Color: b.pal.RampTop},
			)
		}
	case campaign.ThemeNeon:
		for _, x := range []float64{-inner, inner} {
			b.cmds = append(b.cmds,
				Cmd{Shape: ShapeBox, Pos: v3(x, 2, z), Size: Size{X: 0.2, Y: 2.5, Z: 0.4}, Color: b.pal.Accent},
			)
		}
	default: // ThemeMangroves and any unknown theme.
		for _, x := range []float64{-inner, inner} {
			b.cmds = append(b.cmds,
				Cmd{Shape: ShapeCylinder, Pos: v3(x, 2.5, z), Size: Size{X: 0.25, Y: 5}, Color: b.pal.LowBarrier},
			)
		}
	}
}

func (b *worldBuilder) addFloor() {
	start, end := b.window()
	pits := b.pitsOverlapping(start, end)
	for z := start; z < end; {
		t1 := z + floorTileLen
		if t1 > end {
			t1 = end
		}
		mid := (z + t1) / 2
		if !overlapsAnyPit(pits, z, t1) {
			for lane := 0; lane < 3; lane++ {
				col := b.pal.Floor
				if lane == 1 {
					col = b.pal.FloorAlt
				}
				b.cmds = append(b.cmds, Cmd{
					Shape: ShapeBox,
					Pos:   v3(laneX(float64(lane)), -floorThickness/2, mid),
					Size:  Size{X: stripWidth, Y: floorThickness, Z: t1 - z},
					Color: col,
				})
			}
		}
		z = t1
	}
}

func (b *worldBuilder) pitsOverlapping(start, end float64) []levelgen.Element {
	from := start - 4
	if from < 0 {
		from = 0
	}
	all := b.level.Elements(from, end)
	var pits []levelgen.Element
	for _, e := range all {
		if e.Kind != levelgen.Pit {
			continue
		}
		if e.Z+e.Span <= start || e.Z >= end {
			continue
		}
		pits = append(pits, e)
	}
	return pits
}

func overlapsAnyPit(pits []levelgen.Element, z0, z1 float64) bool {
	for _, p := range pits {
		if z0 < p.Z+p.Span && p.Z < z1 {
			return true
		}
	}
	return false
}

func (b *worldBuilder) addWalls() {
	start, end := b.window()
	if end <= start {
		return
	}
	mid := (start + end) / 2
	length := end - start
	for _, x := range []float64{-corridorHalf, corridorHalf} {
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x, wallHeight/2, mid),
			Size:  Size{X: wallThickness, Y: wallHeight, Z: length},
			Color: b.pal.Wall,
		})
	}
}

func (b *worldBuilder) addElement(e levelgen.Element) {
	switch e.Kind {
	case levelgen.Pit:
		b.addPit(e)
	case levelgen.LowBarrier:
		b.addLowBarrier(e)
	case levelgen.HighBarrier:
		b.addHighBarrier(e)
	case levelgen.Ramp:
		b.addRamp(e)
	case levelgen.Enemy:
		b.addEnemy(e)
	case levelgen.Comb:
		b.addComb(e)
	case levelgen.Hairspray:
		b.addHairspray(e)
	case levelgen.Mirror:
		b.addMirror(e)
	}
}

func (b *worldBuilder) addPit(e levelgen.Element) {
	z := e.Z + e.Span/2
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(0, -pitDepth/2, z),
		Size:  Size{X: corridorHalf * 2, Y: pitDepth, Z: e.Span},
		Color: b.pal.Hazard,
	})
}

func (b *worldBuilder) addLowBarrier(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z := e.Z + barrierDepth/2
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, lowBarrierH/2, z),
		Size:  Size{X: stripWidth, Y: lowBarrierH, Z: barrierDepth},
		Color: b.pal.LowBarrier,
	})
	switch b.theme {
	case campaign.ThemeHarbor:
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x, lowBarrierH+0.075, z),
			Size:  Size{X: stripWidth, Y: 0.15, Z: barrierDepth},
			Color: b.pal.RampTop,
		})
	case campaign.ThemeNeon:
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x, lowBarrierH+0.45, z+barrierDepth/2-0.1),
			Size:  Size{X: stripWidth, Y: 0.9, Z: 0.2},
			Color: b.pal.LowBarrier,
		})
	default:
		for _, dx := range []float64{-0.6, 0.6} {
			b.cmds = append(b.cmds, Cmd{
				Shape: ShapeCone,
				Pos:   v3(x+dx, lowBarrierH+0.25, z),
				Size:  Size{X: 0.15, Y: 0.5},
				Color: b.pal.Accent,
			})
		}
	}
}

func (b *worldBuilder) addHighBarrier(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z := e.Z + barrierDepth/2
	beamY := (highBarrierGap + highBarrierTop) / 2
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, beamY, z),
		Size:  Size{X: stripWidth, Y: highBarrierTop - highBarrierGap, Z: barrierDepth},
		Color: b.pal.HighBarrier,
	})
	for _, dx := range []float64{-0.8, 0.8} {
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x+dx, highBarrierGap/2, z),
			Size:  Size{X: 0.2, Y: highBarrierGap, Z: barrierDepth},
			Color: b.pal.Wall,
		})
	}
}

func (b *worldBuilder) addRamp(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z0 := e.Z
	z1 := e.Z + e.Span
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, containerH/2, (z0+z1)/2),
		Size:  Size{X: stripWidth, Y: containerH, Z: e.Span},
		Color: b.pal.Ramp,
	})
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, containerH+0.05, (z0+z1)/2),
		Size:  Size{X: stripWidth - 0.2, Y: 0.1, Z: e.Span},
		Color: b.pal.RampTop,
	})
	for i := 1; i <= rampSteps; i++ {
		stepH := containerH * float64(i) / float64(rampSteps+1)
		stepZ := z0 - rampStepDepth*float64(i) + rampStepDepth/2
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x, stepH/2, stepZ),
			Size:  Size{X: stripWidth, Y: stepH, Z: rampStepDepth},
			Color: b.pal.RampTop,
		})
	}
}

func (b *worldBuilder) addEnemy(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z := e.Z + barrierDepth/2
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeSphere,
		Pos:   v3(x, 0.45, z),
		Size:  Size{X: 0.45},
		Color: b.pal.Enemy,
	})
	switch b.theme {
	case campaign.ThemeHarbor:
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeSphere,
			Pos:   v3(x, 0.85, z),
			Size:  Size{X: 0.2},
			Color: b.pal.Accent,
		})
	case campaign.ThemeNeon:
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x, 0.2, z),
			Size:  Size{X: 0.7, Y: 0.4, Z: 0.7},
			Color: b.pal.Accent,
		})
	default:
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeCone,
			Pos:   v3(x, 0.3, z-0.4),
			Size:  Size{X: 0.2, Y: 0.5},
			Color: b.pal.Accent,
		})
	}
}

func (b *worldBuilder) addComb(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z := e.Z
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, 0.15, z),
		Size:  Size{X: 0.7, Y: 0.3, Z: 0.25},
		Color: b.pal.Comb,
	})
	for i := 0; i < 5; i++ {
		dx := -0.3 + 0.15*float64(i)
		b.cmds = append(b.cmds, Cmd{
			Shape: ShapeBox,
			Pos:   v3(x+dx, 0.475, z),
			Size:  Size{X: 0.06, Y: 0.35, Z: 0.25},
			Color: b.pal.Comb,
		})
	}
}

func (b *worldBuilder) addHairspray(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z := e.Z
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeCylinder,
		Pos:   v3(x, 0.3, z),
		Size:  Size{X: 0.2, Y: 0.6},
		Color: b.pal.Hairspray,
	})
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeSphere,
		Pos:   v3(x, 0.65, z),
		Size:  Size{X: 0.07},
		Color: b.pal.Hairspray,
	})
}

func (b *worldBuilder) addMirror(e levelgen.Element) {
	x := laneX(float64(e.Lane))
	z := e.Z
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, 0.75, z),
		Size:  Size{X: 0.9, Y: 1.5, Z: 0.1},
		Color: b.pal.Mirror,
	})
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeBox,
		Pos:   v3(x, 0.75, z-0.07),
		Size:  Size{X: 0.7, Y: 1.3, Z: 0.04},
		Color: b.pal.Mirror,
	})
	b.cmds = append(b.cmds, Cmd{
		Shape: ShapeCylinder,
		Pos:   v3(x, 0.25, z),
		Size:  Size{X: 0.08, Y: 0.5},
		Color: b.pal.Mirror,
	})
}
