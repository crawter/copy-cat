package game

// Audio owns synthesized sound; inert until its feature.
type Audio struct{}

func NewAudio() *Audio {
	return &Audio{}
}

func (a *Audio) Update(dt float32, g *Game) {}

func (a *Audio) Draw(g *Game) {}
