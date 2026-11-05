package pickups

import "testing"

func collectAllPlaced(r *Reels) {
	for id := 0; id < PlacedReels; id++ {
		r.Collect(id)
	}
}

func TestNewStartsAtZero(t *testing.T) {
	r := New()
	if got := r.Count(); got != 0 {
		t.Fatalf("New Count = %d, want 0", got)
	}
}

func TestOneCountPerPlacedReel(t *testing.T) {
	r := New()
	for id := 0; id < PlacedReels; id++ {
		if got := r.Collect(id); !got {
			t.Fatalf("Collect(%d) = false, want true", id)
		}
		if want := id + 1; r.Count() != want {
			t.Fatalf("after Collect(%d) Count = %d, want %d", id, r.Count(), want)
		}
	}
	if want := PlacedReels; r.Count() != want {
		t.Fatalf("Count after all placed reels = %d, want %d", r.Count(), want)
	}
}

func TestDroppedReelCompletesTwenty(t *testing.T) {
	r := New()
	collectAllPlaced(&r)

	if got := r.CollectDropped(); !got {
		t.Fatal("first CollectDropped = false, want true")
	}
	if want := TotalReels; r.Count() != want {
		t.Fatalf("Count after the drop = %d, want %d", r.Count(), want)
	}
}

func TestCollectDroppedTwiceCountsOnce(t *testing.T) {
	r := New()
	if got := r.CollectDropped(); !got {
		t.Fatal("first CollectDropped = false, want true")
	}
	if got := r.CollectDropped(); got {
		t.Fatal("second CollectDropped = true, want false")
	}
	if want := 1; r.Count() != want {
		t.Fatalf("Count after two drops = %d, want %d", r.Count(), want)
	}
}

func TestNoDoubleCount(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{name: "first placed", id: 0},
		{name: "middle placed", id: 9},
		{name: "last placed", id: PlacedReels - 1},
		{name: "dropped", id: DroppedReelID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			if got := r.Collect(tt.id); !got {
				t.Fatalf("first Collect(%d) = false, want true", tt.id)
			}
			before := r.Count()
			if got := r.Collect(tt.id); got {
				t.Fatalf("second Collect(%d) = true, want false", tt.id)
			}
			if r.Count() != before {
				t.Fatalf("Count changed on double collect: before %d, after %d", before, r.Count())
			}
		})
	}
}

func TestUnknownIDsReturnFalseAndChangeNothing(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{name: "below zero", id: -1},
		{name: "far below zero", id: -100},
		{name: "one past the drop", id: DroppedReelID + 1},
		{name: "far above", id: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			collectAllPlaced(&r)
			before := r.Count()

			if got := r.Collect(tt.id); got {
				t.Fatalf("Collect(%d) = true, want false", tt.id)
			}
			if got := r.Collected(tt.id); got {
				t.Fatalf("Collected(%d) = true, want false", tt.id)
			}
			if r.Count() != before {
				t.Fatalf("Count changed on unknown id: before %d, after %d", before, r.Count())
			}
		})
	}
}

func TestCollectedReportsOnlyCollectedIDs(t *testing.T) {
	r := New()
	r.Collect(0)
	r.Collect(PlacedReels - 1)
	r.CollectDropped()

	collected := []int{0, PlacedReels - 1, DroppedReelID}
	notCollected := []int{1, PlacedReels / 2}

	for _, id := range collected {
		if got := r.Collected(id); !got {
			t.Fatalf("Collected(%d) = false, want true", id)
		}
	}
	for _, id := range notCollected {
		if got := r.Collected(id); got {
			t.Fatalf("Collected(%d) = true, want false", id)
		}
	}
}

func TestTotalIsTwenty(t *testing.T) {
	tests := []struct {
		name string
		r    Reels
	}{
		{name: "nothing collected", r: New()},
		{name: "every reel collected", r: func() Reels {
			r := New()
			collectAllPlaced(&r)
			r.CollectDropped()
			return r
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Total(); got != TotalReels {
				t.Fatalf("Total = %d, want %d", got, TotalReels)
			}
		})
	}
}
