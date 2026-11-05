package game

// Scenes owns the death and end screens; inert until its feature.
type Scenes struct{}

func NewScenes() *Scenes {
	return &Scenes{}
}

func (s *Scenes) Update(dt float32, g *Game) {}

func (s *Scenes) Draw(g *Game) {}
