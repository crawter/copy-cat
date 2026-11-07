package campaign

import (
	"testing"

	"copycat/internal/levelgen"
)

func TestDefaultWorlds(t *testing.T) {
	c := Default()

	if c.StartLives != 3 {
		t.Errorf("StartLives = %d, want 3", c.StartLives)
	}
	if len(c.Worlds) != 3 {
		t.Fatalf("len(Worlds) = %d, want 3", len(c.Worlds))
	}

	wantThemes := []Theme{ThemeMangroves, ThemeHarbor, ThemeNeon}
	for i, want := range wantThemes {
		w := c.Worlds[i]
		if w.Number != i+1 {
			t.Errorf("Worlds[%d].Number = %d, want %d", i, w.Number, i+1)
		}
		if w.Theme != want {
			t.Errorf("Worlds[%d].Theme = %v, want %v", i, w.Theme, want)
		}
	}
}

func TestDefaultLevels(t *testing.T) {
	c := Default()

	if len(c.Levels) != 10 {
		t.Fatalf("len(Levels) = %d, want 10", len(c.Levels))
	}

	seen := make(map[int]bool, len(c.Levels))
	for _, lv := range c.Levels {
		if seen[lv.ID] {
			t.Errorf("duplicate level ID %d", lv.ID)
		}
		seen[lv.ID] = true
	}
}

func TestRunLevelsPresetAndSeeds(t *testing.T) {
	c := Default()

	presets := []levelgen.Params{levelgen.Mangroves(), levelgen.Harbor(), levelgen.Neon()}
	seenSeeds := make(map[uint64]bool)

	for _, lv := range c.Levels {
		if lv.Kind != LevelRun {
			continue
		}
		if lv.Seed == 0 {
			t.Errorf("run level %d carries no seed", lv.ID)
		}
		if seenSeeds[lv.Seed] {
			t.Errorf("run seed %d is not distinct", lv.Seed)
		}
		seenSeeds[lv.Seed] = true

		if lv.World < 1 || lv.World > 3 {
			t.Errorf("run level %d has World %d, want 1..3", lv.ID, lv.World)
			continue
		}
		if want := presets[lv.World-1]; lv.Params != want {
			t.Errorf("run level %d has Params %+v, want %+v", lv.ID, lv.Params, want)
		}
		if lv.Index < 1 || lv.Index > 3 {
			t.Errorf("run level %d has Index %d, want 1..3", lv.ID, lv.Index)
		}
	}

	if len(seenSeeds) != 9 {
		t.Fatalf("distinct run seeds = %d, want 9", len(seenSeeds))
	}
}

func TestBuildRunLevels(t *testing.T) {
	c := Default()
	for _, lv := range c.Levels {
		if lv.Kind != LevelRun {
			continue
		}
		l := levelgen.New(lv.Seed, lv.Params)
		if l == nil {
			t.Fatalf("levelgen.New for level %d returned nil", lv.ID)
		}
		if got, want := l.Length(), lv.Params.Length; got != want {
			t.Errorf("level %d Length = %v, want %v", lv.ID, got, want)
		}
	}
}

func TestBossLevelLast(t *testing.T) {
	c := Default()

	for i, lv := range c.Levels[:9] {
		if lv.ID != i+1 {
			t.Errorf("Levels[%d].ID = %d, want %d", i, lv.ID, i+1)
		}
		if lv.Kind != LevelRun {
			t.Errorf("Levels[%d].Kind = %v, want LevelRun", i, lv.Kind)
		}
	}

	last := c.Levels[len(c.Levels)-1]
	if last.ID != 10 {
		t.Errorf("last level ID = %d, want 10", last.ID)
	}
	if last.Kind != LevelBoss {
		t.Errorf("last level Kind = %v, want LevelBoss", last.Kind)
	}
	if last.Seed != 0 {
		t.Errorf("boss level Seed = %d, want 0 (no seed)", last.Seed)
	}
	if last.Params != (levelgen.Params{}) {
		t.Errorf("boss level Params = %+v, want zero (no corridor)", last.Params)
	}
}

func TestLevelByID(t *testing.T) {
	c := Default()
	for _, want := range c.Levels {
		got, ok := c.LevelByID(want.ID)
		if !ok {
			t.Fatalf("LevelByID(%d) not found", want.ID)
		}
		if got != want {
			t.Errorf("LevelByID(%d) = %+v, want %+v", want.ID, got, want)
		}
	}
	if _, ok := c.LevelByID(11); ok {
		t.Errorf("LevelByID(11) found, want not found")
	}
}

func TestNext(t *testing.T) {
	c := Default()
	for id := 1; id <= 9; id++ {
		next, ok := c.Next(id)
		if !ok {
			t.Fatalf("Next(%d) not found", id)
		}
		if next.ID != id+1 {
			t.Errorf("Next(%d).ID = %d, want %d", id, next.ID, id+1)
		}
	}

	for _, id := range []int{0, 10, 11} {
		if _, ok := c.Next(id); ok {
			t.Errorf("Next(%d) found, want not found", id)
		}
	}
}

func TestBuildEndlessDeterminism(t *testing.T) {
	c := Default()
	if c.EndlessParams != levelgen.Neon() {
		t.Errorf("EndlessParams = %+v, want the hard neon preset %+v", c.EndlessParams, levelgen.Neon())
	}

	seeds := make([]uint64, 0, 103)
	for s := uint64(0); s < 100; s++ {
		seeds = append(seeds, s)
	}
	seeds = append(seeds, 1<<63, ^uint64(0), 0x9E3779B97F4A7C15)

	first := make([]*levelgen.Level, 0, len(seeds))
	for _, seed := range seeds {
		l := c.BuildEndless(seed)
		if l == nil {
			t.Fatalf("BuildEndless(%d) returned nil", seed)
		}
		if l.Seed() != seed {
			t.Errorf("BuildEndless(%d).Seed() = %d", seed, l.Seed())
		}
		if got, want := l.Length(), c.EndlessParams.Length; got != want {
			t.Errorf("BuildEndless(%d).Length() = %v, want %v", seed, got, want)
		}
		first = append(first, l)
	}

	for i, seed := range seeds {
		again := c.BuildEndless(seed)
		got := first[i].Elements(0, first[i].Length())
		want := again.Elements(0, again.Length())
		if len(got) != len(want) {
			t.Fatalf("BuildEndless(%d) element counts differ: %d vs %d", seed, len(got), len(want))
		}
		for j := range got {
			if got[j] != want[j] {
				t.Fatalf("BuildEndless(%d) element %d differs: %+v vs %+v", seed, j, got[j], want[j])
			}
		}
	}
}
