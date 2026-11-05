package game

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/v5/raylib"
)

// HUD owns the on-screen overlays.
type HUD struct{}

func NewHUD() *HUD {
	return &HUD{}
}

func (h *HUD) Update(dt float32, g *Game) {}

// Draw writes the stamp counts, distance, and best distance top-left in 2D.
func (h *HUD) Draw(g *Game) {
	const (
		x          int32 = 16
		y          int32 = 16
		fontSize   int32 = 20
		lineHeight int32 = 26
	)

	lines := []string{
		fmt.Sprintf("%s: %d/%d", HUDStamps, g.Run.Stamps, g.Run.TotalStamps),
		fmt.Sprintf("%s: %.0f м", HUDDistance, g.Run.Distance),
		fmt.Sprintf("%s: %.0f м", HUDBest, g.Run.BestDistance),
	}
	for i, line := range lines {
		rl.DrawText(line, x, y+lineHeight*int32(i), fontSize, rl.RayWhite)
	}
}
