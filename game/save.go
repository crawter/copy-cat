package game

// Save owns progress persistence; inert until its feature.
type Save struct{}

func NewSave() *Save {
	return &Save{}
}

func (s *Save) Update(dt float32, g *Game) {}

func (s *Save) Draw(g *Game) {}

// Write persists progress. It is a no-op in the scaffold.
func (s *Save) Write(g *Game) {}
