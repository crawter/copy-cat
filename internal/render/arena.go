package render

import (
	"copycat/internal/gameplay"
)

// Salon arena geometry, in meters. The boss fight owns a three-lane arena on
// the same x offsets as the corridor, so the hero and Copycat land on the
// lanes the simulation already uses.
const (
	arenaHalfX    = 6.0
	arenaNearZ    = -6.0
	arenaFarZ     = 12.0
	arenaTile     = 2.0
	arenaTileSize = 1.9
)

// The Last Haircut salon palette, kept separate from the corridor themes.
var (
	bossDark   = Color{R: 24, G: 20, B: 34, A: 255}
	bossLight  = Color{R: 232, G: 224, B: 236, A: 255}
	bossChair  = Color{R: 200, G: 40, B: 90, A: 255}
	bossMirror = Color{R: 150, G: 180, B: 200, A: 255}

	// The boss telegraph paints the attack lane so the player can read which
	// dodge the next attack needs: orange to jump a low swipe, red to slide a
	// high swipe.
	bossTelegraphLow  = Color{R: 250, G: 170, B: 60, A: 220}
	bossTelegraphHigh = Color{R: 235, G: 60, B: 60, A: 220}
)

// BuildBossArena returns the draw list for the Last Haircut salon: a
// checkered floor, barber chairs and standing mirrors. It draws only the
// arena, never the corridor, so the boss scene has a stage of its own.
func BuildBossArena(cam Cam) []Cmd {
	var cs []Cmd
	xTiles := int(arenaHalfX * 2 / arenaTile)
	zTiles := int((arenaFarZ - arenaNearZ) / arenaTile)
	for ix := 0; ix < xTiles; ix++ {
		x0 := -arenaHalfX + float64(ix)*arenaTile
		for iz := 0; iz < zTiles; iz++ {
			z0 := arenaNearZ + float64(iz)*arenaTile
			col := bossDark
			if (ix+iz)%2 == 0 {
				col = bossLight
			}
			cs = append(cs, Cmd{
				Shape: ShapePlane,
				Pos:   v3(x0+arenaTile/2, 0, z0+arenaTile/2),
				Size:  Size{X: arenaTileSize, Z: arenaTileSize},
				Color: col,
			})
		}
	}
	for _, x := range []float64{-4, 4} {
		for _, z := range []float64{-2, 2, 6} {
			cs = append(cs, barberChairCmds(x, z)...)
		}
	}
	for _, x := range []float64{-3, 0, 3} {
		cs = append(cs, standingMirrorCmds(x, 9)...)
	}
	return cs
}

// BuildArenaTelegraph returns the draw list highlighting the attack lane
// during Copycat's telegraphed attack. lane is the lane index in [0, 2];
// kind picks the dodge-specific highlight color.
func BuildArenaTelegraph(lane int, kind gameplay.AttackKind) []Cmd {
	if lane < 0 {
		lane = 0
	}
	if lane > 2 {
		lane = 2
	}
	col := bossTelegraphLow
	if kind == gameplay.AttackHigh {
		col = bossTelegraphHigh
	}
	return []Cmd{{
		Shape: ShapePlane,
		Pos:   v3(laneX(float64(lane)), 0.02, (arenaNearZ+arenaFarZ)/2),
		Size:  Size{X: arenaTileSize, Z: arenaFarZ - arenaNearZ},
		Color: col,
	}}
}

func barberChairCmds(x, z float64) []Cmd {
	return []Cmd{
		{Shape: ShapeCylinder, Pos: v3(x, 0.25, z), Size: Size{X: 0.3, Y: 0.5}, Color: bossChair},
		{Shape: ShapeBox, Pos: v3(x, 0.75, z), Size: Size{X: 0.9, Y: 0.3, Z: 0.9}, Color: bossChair},
		{Shape: ShapeBox, Pos: v3(x, 1.25, z+0.5), Size: Size{X: 0.9, Y: 0.9, Z: 0.2}, Color: bossChair},
	}
}

func standingMirrorCmds(x, z float64) []Cmd {
	return []Cmd{
		{Shape: ShapeBox, Pos: v3(x, 0.95, z), Size: Size{X: 1.1, Y: 1.9, Z: 0.1}, Color: bossMirror},
		{Shape: ShapeBox, Pos: v3(x, 0.95, z-0.06), Size: Size{X: 0.9, Y: 1.7, Z: 0.04}, Color: bossLight},
		{Shape: ShapeBox, Pos: v3(x, 0.05, z), Size: Size{X: 1.3, Y: 0.1, Z: 0.4}, Color: bossMirror},
	}
}
