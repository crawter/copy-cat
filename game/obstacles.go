package game

// ObstacleKind names one obstacle the layout can place in a lane.
type ObstacleKind int

const (
	ObstacleNone ObstacleKind = iota
	ObstaclePit
	ObstacleSpikes
	ObstacleCrates
)

// Obstacles owns obstacle collision and rendering; inert until its feature.
type Obstacles struct{}

func NewObstacles(layout *Layout) *Obstacles {
	return &Obstacles{}
}

func (o *Obstacles) Update(dt float32, g *Game) {}

func (o *Obstacles) Draw(g *Game) {}
