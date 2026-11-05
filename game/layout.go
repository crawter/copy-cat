package game

// Layout describes the whole corridor level as data. Feature subsystems read
// this shared contract instead of placing their own content.
type Layout struct {
	LaneWidth float32
	LaneCount int
	EndZ      float32
	Segments  []Segment
}

// LaneX returns the world X coordinate of a lane index.
func (l *Layout) LaneX(lane int) float32 {
	return (float32(lane) - float32(l.LaneCount-1)/2) * l.LaneWidth
}

// StampSpot places one collectible stamp in a lane at a depth Z.
type StampSpot struct {
	Lane int
	Z    float32
}

// EnemySpawn places one enemy of a kind in a lane at a depth Z.
type EnemySpawn struct {
	Kind EnemyKind
	Lane int
	Z    float32
}

// Segment is one slice of the corridor.
type Segment struct {
	Start  float32
	Length float32

	Obstacle     ObstacleKind
	ObstacleLane int

	Stamps []StampSpot

	Enemy *EnemySpawn
}

// NewLayout builds the concrete POC level: twenty segments totaling 240
// meters, ending with a clear three-segment stretch reserved for the desk.
func NewLayout() *Layout {
	return &Layout{
		LaneWidth: 2.5,
		LaneCount: 3,
		EndZ:      240,
		Segments: []Segment{
			{Start: 0, Length: 15, Obstacle: ObstacleNone,
				Stamps: []StampSpot{{Lane: 0, Z: 10}, {Lane: 1, Z: 10}, {Lane: 2, Z: 10}}},
			{Start: 15, Length: 12, Obstacle: ObstacleCrates, ObstacleLane: 0},
			{Start: 27, Length: 15, Obstacle: ObstacleNone,
				Stamps: []StampSpot{{Lane: 1, Z: 40}},
				Enemy:  &EnemySpawn{Kind: EnemyCopier, Lane: 0, Z: 35}},
			{Start: 42, Length: 10, Obstacle: ObstacleSpikes, ObstacleLane: 2},
			{Start: 52, Length: 15, Obstacle: ObstacleNone},
			{Start: 67, Length: 3.5, Obstacle: ObstaclePit, ObstacleLane: 1},
			{Start: 70.5, Length: 15, Obstacle: ObstacleNone},
			{Start: 85.5, Length: 12, Obstacle: ObstacleCrates, ObstacleLane: 2},
			{Start: 97.5, Length: 15, Obstacle: ObstacleNone,
				Stamps: []StampSpot{{Lane: 0, Z: 100}, {Lane: 2, Z: 100}},
				Enemy:  &EnemySpawn{Kind: EnemyCabinet, Lane: 2, Z: 105}},
			{Start: 112.5, Length: 10, Obstacle: ObstacleSpikes, ObstacleLane: 0},
			{Start: 122.5, Length: 15, Obstacle: ObstacleNone},
			{Start: 137.5, Length: 4, Obstacle: ObstaclePit, ObstacleLane: 0},
			{Start: 141.5, Length: 15, Obstacle: ObstacleNone,
				Stamps: []StampSpot{{Lane: 1, Z: 145}}},
			{Start: 156.5, Length: 12, Obstacle: ObstacleCrates, ObstacleLane: 1},
			{Start: 168.5, Length: 15, Obstacle: ObstacleNone,
				Stamps: []StampSpot{{Lane: 0, Z: 180}, {Lane: 1, Z: 180}, {Lane: 2, Z: 180}},
				Enemy:  &EnemySpawn{Kind: EnemyCoffee, Lane: 1, Z: 175}},
			{Start: 183.5, Length: 10, Obstacle: ObstacleSpikes, ObstacleLane: 1},
			{Start: 193.5, Length: 3.5, Obstacle: ObstaclePit, ObstacleLane: 2},
			{Start: 197, Length: 15, Obstacle: ObstacleNone},
			{Start: 212, Length: 15, Obstacle: ObstacleNone},
			{Start: 227, Length: 13, Obstacle: ObstacleNone},
		},
	}
}
