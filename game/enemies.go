package game

// EnemyKind names one enemy the layout can spawn.
type EnemyKind int

const (
	EnemyCopier EnemyKind = iota
	EnemyCabinet
	EnemyCoffee
)

// Enemies owns enemy behavior and rendering; inert until its feature.
type Enemies struct{}

func NewEnemies(layout *Layout) *Enemies {
	return &Enemies{}
}

func (e *Enemies) Update(dt float32, g *Game) {}

func (e *Enemies) Draw(g *Game) {}
