// Package leveldata holds Take 9 of Шкряб's POC as hand-written plain Go
// data. It lays out the corridor, its three sections, the fallen light rigs,
// the film reels, the checkpoints, the camera crane's track and the exit, in
// meters. It imports nothing beyond the standard library, so the set is
// readable and testable headless and reusable unchanged on every platform.
package leveldata

// takeWidth is the corridor's width in meters.
const takeWidth = 8.0

// takeLength is the corridor's length in meters. At the hero's run speed of
// twelve meters a second it is covered in about 142 seconds, inside the
// two-to-three-minute run the design fixes.
const takeLength = 1700.0

// Point is a position in meters. X is sideways across the corridor, Y is the
// height above the floor, and Z is forward along the corridor.
type Point struct {
	X, Y, Z float64
}

// Section is one stretch of the corridor. StartZ and EndZ are its forward
// bounds in meters.
type Section struct {
	StartZ, EndZ float64
}

// Rig is a fallen light rig lying across the corridor. Position is the rig's
// center.
type Rig struct {
	Position Point
}

// Placement is the kind of spot where a film reel lies.
type Placement string

// The three kinds of spot where a film reel can lie.
const (
	// Floor places a reel on the corridor floor.
	Floor Placement = "floor"
	// RigTop places a reel on top of a fallen light rig.
	RigTop Placement = "rig top"
	// Nook places a reel in a side nook near a corridor edge.
	Nook Placement = "nook"
)

// Reel is one lost film reel. Position is where it lies and Kind is the kind
// of spot it lies in.
type Reel struct {
	Position Point
	Kind     Placement
}

// Checkpoint is a clapperboard checkpoint. Position is where it stands.
type Checkpoint struct {
	Position Point
}

// Track is one stretch of the camera crane's broken track. Start and End are
// its two ends, and the crane patrols back and forth between them.
type Track struct {
	Start, End Point
}

// Exit is the take's exit line. It runs across the corridor's width at the
// corridor's end, where the 'CUT!' sign hangs. Z is its forward position in
// meters.
type Exit struct {
	Z float64
}

// Take is the whole layout of Take 9 in meters.
type Take struct {
	Width       float64
	Length      float64
	Sections    []Section
	Rigs        []Rig
	Reels       []Reel
	Checkpoints []Checkpoint
	Track       Track
	Exit        Exit
}

// Take9 returns the whole layout of Take 9.
func Take9() Take {
	third := takeLength / 3.0

	rigs := []Rig{
		{Position: Point{X: takeWidth / 2, Y: 0, Z: third / 2}},
		{Position: Point{X: takeWidth / 2, Y: 0, Z: 1.5 * third}},
		{Position: Point{X: takeWidth / 2, Y: 0, Z: 2.5 * third}},
	}

	reels := []Reel{
		// Section one.
		{Position: Point{X: 2.0, Y: 0, Z: 60}, Kind: Floor},
		{Position: Point{X: 6.0, Y: 0, Z: 140}, Kind: Floor},
		{Position: Point{X: 0.6, Y: 0.7, Z: 220}, Kind: Nook},
		{Position: Point{X: rigs[0].Position.X, Y: 0.4, Z: rigs[0].Position.Z}, Kind: RigTop},
		{Position: Point{X: 3.0, Y: 0, Z: 360}, Kind: Floor},
		{Position: Point{X: 7.4, Y: 0.7, Z: 450}, Kind: Nook},
		// Section two.
		{Position: Point{X: 5.0, Y: 0, Z: 620}, Kind: Floor},
		{Position: Point{X: 1.5, Y: 0, Z: 700}, Kind: Floor},
		{Position: Point{X: 0.6, Y: 0.7, Z: 780}, Kind: Nook},
		{Position: Point{X: rigs[1].Position.X, Y: 0.4, Z: rigs[1].Position.Z}, Kind: RigTop},
		{Position: Point{X: 6.5, Y: 0, Z: 930}, Kind: Floor},
		{Position: Point{X: 2.5, Y: 0, Z: 1020}, Kind: Floor},
		{Position: Point{X: 7.4, Y: 0.7, Z: 1100}, Kind: Nook},
		// Section three.
		{Position: Point{X: 3.5, Y: 0, Z: 1180}, Kind: Floor},
		{Position: Point{X: 0.6, Y: 0.7, Z: 1260}, Kind: Nook},
		{Position: Point{X: 6.0, Y: 0, Z: 1330}, Kind: Floor},
		{Position: Point{X: rigs[2].Position.X, Y: 0.4, Z: rigs[2].Position.Z}, Kind: RigTop},
		{Position: Point{X: 2.0, Y: 0, Z: 1500}, Kind: Floor},
		{Position: Point{X: 7.4, Y: 0.7, Z: 1580}, Kind: Nook},
	}

	return Take{
		Width:  takeWidth,
		Length: takeLength,
		Sections: []Section{
			{StartZ: 0, EndZ: third},
			{StartZ: third, EndZ: 2 * third},
			{StartZ: 2 * third, EndZ: takeLength},
		},
		Rigs:  rigs,
		Reels: reels,
		Checkpoints: []Checkpoint{
			{Position: Point{X: takeWidth / 2, Y: 1.0, Z: third}},
			{Position: Point{X: takeWidth / 2, Y: 1.0, Z: 2 * third}},
			{Position: Point{X: takeWidth / 2, Y: 1.0, Z: 1680}},
		},
		Track: Track{
			Start: Point{X: 1.0, Y: 0, Z: 650},
			End:   Point{X: 7.0, Y: 0, Z: 1050},
		},
		Exit: Exit{Z: takeLength},
	}
}
