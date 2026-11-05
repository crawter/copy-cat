package collide

import (
	"testing"

	"copycat/leveldata"
	"copycat/move"
)

// farCrane returns a crane placed well off the take, so tests that are not
// about the crane never touch it.
func farCrane() Crane {
	return Crane{
		Pos:       leveldata.Point{X: 0, Y: 0, Z: -1000},
		HalfLen:   1.5,
		HalfWidth: 1.0,
		Height:    0.7,
	}
}

func hasEvent(events []Event, want Event) bool {
	for _, e := range events {
		if e == want {
			return true
		}
	}
	return false
}

func eventsEqual(got, want []Event) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestWallClampsHero(t *testing.T) {
	take := leveldata.Take9()
	tests := []struct {
		name string
		x    float64
		want float64
	}{
		{name: "far right", x: 100, want: 3.6},
		{name: "far left", x: -100, want: -3.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := Check(move.Hero{X: tt.x}, take, farCrane())
			if got.X != tt.want {
				t.Fatalf("Check X = %v, want %v", got.X, tt.want)
			}
		})
	}
}

func TestHeroInsideKeepsX(t *testing.T) {
	take := leveldata.Take9()
	tests := []struct {
		name string
		x    float64
	}{
		{name: "center", x: 0},
		{name: "right of center", x: 2},
		{name: "at the wall bound", x: 3.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := Check(move.Hero{X: tt.x}, take, farCrane())
			if got.X != tt.x {
				t.Fatalf("Check X = %v, want unchanged %v", got.X, tt.x)
			}
		})
	}
}

func TestClampHoldsInAir(t *testing.T) {
	take := leveldata.Take9()
	got, _ := Check(move.Hero{X: 100, Y: 0.5, Airborne: true}, take, farCrane())
	if want := 3.6; got.X != want {
		t.Fatalf("airborne Check X = %v, want %v", got.X, want)
	}
}

func TestRigHitOnFloor(t *testing.T) {
	take := leveldata.Take9()
	rig := take.Rigs[0]
	_, events := Check(move.Hero{X: 0, Y: 0, Z: rig.Position.Z}, take, farCrane())
	want := []Event{{Kind: EventRigHit}}
	if !eventsEqual(events, want) {
		t.Fatalf("events = %+v, want %+v", events, want)
	}
}

func TestJumpClearsRig(t *testing.T) {
	take := leveldata.Take9()
	rig := take.Rigs[0]
	const dt = 1.0 / 240.0

	h := move.NewHero()
	h.Z = rig.Position.Z - 5.5
	h = move.Step(h, move.Input{Jump: true}, dt)
	if !h.Airborne {
		t.Fatal("takeoff should start a jump")
	}

	var events []Event
	for {
		h, events = Check(h, take, farCrane())
		for _, e := range events {
			if e.Kind == EventRigHit {
				t.Fatalf("jump hit the rig at Z=%v Y=%v", h.Z, h.Y)
			}
		}
		if !h.Airborne && h.Z > rig.Position.Z+2 {
			break
		}
		h = move.Step(h, move.Input{}, dt)
	}
}

func TestCraneHit(t *testing.T) {
	take := leveldata.Take9()
	crane := Crane{
		Pos:       leveldata.Point{X: 2, Y: 0, Z: 10},
		HalfLen:   1.5,
		HalfWidth: 1.0,
		Height:    0.7,
	}

	_, events := Check(move.Hero{X: 2, Y: 0, Z: 10}, take, crane)
	if want := []Event{{Kind: EventCraneHit}}; !eventsEqual(events, want) {
		t.Fatalf("events = %+v, want %+v", events, want)
	}

	_, events = Check(move.Hero{X: 2, Y: crane.Height + 0.4, Z: 10}, take, crane)
	if len(events) != 0 {
		t.Fatalf("above events = %+v, want empty", events)
	}
}

func TestReelTouch(t *testing.T) {
	take := leveldata.Take9()
	tests := []struct {
		name  string
		index int
	}{
		{name: "floor reel", index: 0},
		{name: "nook reel", index: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reel := take.Reels[tt.index]
			h := move.Hero{X: reel.Position.X, Y: reel.Position.Y, Z: reel.Position.Z}
			_, events := Check(h, take, farCrane())
			if !hasEvent(events, Event{Kind: EventReel, Index: tt.index}) {
				t.Fatalf("events = %+v, want EventReel index %d", events, tt.index)
			}
		})
	}
}

func TestCheckpointTouch(t *testing.T) {
	take := leveldata.Take9()
	cp := take.Checkpoints[0]
	tests := []struct {
		name string
		z    float64
		want bool
	}{
		{name: "just before", z: cp.Position.Z - 0.6, want: false},
		{name: "at the near edge", z: cp.Position.Z - 0.5, want: true},
		{name: "on it", z: cp.Position.Z, want: true},
		{name: "at the far edge", z: cp.Position.Z + 0.5, want: true},
		{name: "just after", z: cp.Position.Z + 0.6, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, events := Check(move.Hero{X: 0, Y: 0, Z: tt.z}, take, farCrane())
			got := hasEvent(events, Event{Kind: EventCheckpoint, Index: 0})
			if got != tt.want {
				t.Fatalf("checkpoint event = %v at Z=%v, want %v (events %+v)", got, tt.z, tt.want, events)
			}
		})
	}
}

func TestExitTouch(t *testing.T) {
	take := leveldata.Take9()
	tests := []struct {
		name string
		z    float64
		want bool
	}{
		{name: "just before", z: take.Exit.Z - 0.01, want: false},
		{name: "on the line", z: take.Exit.Z, want: true},
		{name: "past the line", z: take.Exit.Z + 1, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, events := Check(move.Hero{X: 0, Y: 0, Z: tt.z}, take, farCrane())
			got := hasEvent(events, Event{Kind: EventExit})
			if got != tt.want {
				t.Fatalf("exit event = %v at Z=%v, want %v (events %+v)", got, tt.z, tt.want, events)
			}
		})
	}
}

func TestNothingTouches(t *testing.T) {
	take := leveldata.Take9()
	h := move.Hero{X: 2, Y: 0, Z: 0}
	got, events := Check(h, take, farCrane())
	if len(events) != 0 {
		t.Fatalf("events = %+v, want empty", events)
	}
	if got.X != h.X {
		t.Fatalf("Check X = %v, want unchanged %v", got.X, h.X)
	}
}

func TestSimultaneousTouches(t *testing.T) {
	take := leveldata.Take{
		Width:  8,
		Length: 100,
		Rigs: []leveldata.Rig{
			{Position: leveldata.Point{X: 0, Y: 0, Z: 50}},
		},
		Reels: []leveldata.Reel{
			{Position: leveldata.Point{X: 0, Y: 0, Z: 50}, Kind: leveldata.Floor},
		},
		Checkpoints: []leveldata.Checkpoint{
			{Position: leveldata.Point{X: 0, Y: 1, Z: 50}},
		},
		Exit: leveldata.Exit{Z: 100},
	}

	_, events := Check(move.Hero{X: 0, Y: 0, Z: 50}, take, farCrane())
	want := []Event{
		{Kind: EventRigHit},
		{Kind: EventReel, Index: 0},
		{Kind: EventCheckpoint, Index: 0},
	}
	if !eventsEqual(events, want) {
		t.Fatalf("events = %+v, want %+v", events, want)
	}
}
