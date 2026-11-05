package game

// Pickups owns collectible stamps; inert until its feature.
type Pickups struct{}

func NewPickups(layout *Layout) *Pickups {
	return &Pickups{}
}

func (p *Pickups) Update(dt float32, g *Game) {}

func (p *Pickups) Draw(g *Game) {}
