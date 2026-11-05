package game

// HUD owns on-screen overlays; inert until its feature.
type HUD struct{}

func NewHUD() *HUD {
	return &HUD{}
}

func (h *HUD) Update(dt float32, g *Game) {}

func (h *HUD) Draw(g *Game) {}
