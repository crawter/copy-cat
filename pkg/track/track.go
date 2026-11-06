// Package track holds the corridor's data model and fairness helpers. The
// seeded generator builds a Track, and the runner, controller and render
// layers read it. It is pure Go and imports no raylib, so it can be tested
// headless and reused unchanged on Android.
package track

// EnemyKind names one line of the enemy cast. The first version ships two
// lines: a small pest on the floor and a small machine in the air.
type EnemyKind int

const (
	// EnemyPest is the small swamp pest that moves on the track floor.
	EnemyPest EnemyKind = iota
	// EnemyMachine is the small machine that flies through the air.
	EnemyMachine
)

// Track is one generated corridor run: its length, the gaps that cut the
// floor, and the crates, jar lids and enemies placed along it.
type Track struct {
	Length  float64
	Gaps    []Gap
	Crates  []Crate
	Junk    []Junk
	Enemies []Enemy
}

// Gap is one hole in the track floor. Start and End are distances along the
// corridor, in world units; the hero must jump the gap between them.
type Gap struct {
	Start float64
	End   float64
}

// Crate is one crate on the track. X is the sideways position and Z the
// distance run along the corridor.
type Crate struct {
	X float64
	Z float64
}

// Junk is one jar lid on the track. X is the sideways position and Z the
// distance run along the corridor. There is one junk type, so no kind field.
type Junk struct {
	X float64
	Z float64
}

// Enemy is one enemy on the track. Kind picks the cast line, X is the
// sideways position, Y the height above the floor and Z the distance run.
type Enemy struct {
	Kind EnemyKind
	X    float64
	Y    float64
	Z    float64
}
