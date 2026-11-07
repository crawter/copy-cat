// Package campaign defines the game's worlds, levels and endless mode.
//
// The campaign is fixed: three worlds of three run levels each, followed by
// the Copycat boss finale. Every run level carries a fixed seed, so each
// attempt replays the same learnable corridor; the endless mode draws a fresh
// seed per run. The package is pure Go: it imports only levelgen, never
// raylib or audio, and it holds no player-facing strings.
package campaign

import "copycat/internal/levelgen"

// Theme identifies a world's visual theme.
type Theme int

// The visual themes a campaign world can own.
const (
	// ThemeMangroves is the wet mangrove world theme.
	ThemeMangroves Theme = iota
	// ThemeHarbor is the rusty harbor world theme.
	ThemeHarbor
	// ThemeNeon is the neon Chic-Chic street world theme.
	ThemeNeon
)

// WorldDef is one campaign world. Each world owns one theme.
type WorldDef struct {
	// Number is the world's one-based position in the campaign.
	Number int
	// Theme is the world's visual theme.
	Theme Theme
}

// LevelKind identifies whether a level is a corridor run or the boss fight.
type LevelKind int

// The kinds a campaign level can be.
const (
	// LevelRun is a corridor level the hero runs through.
	LevelRun LevelKind = iota
	// LevelBoss is the Copycat boss fight, which has no corridor.
	LevelBoss
)

// LevelDef describes one campaign level.
type LevelDef struct {
	// ID is the level's one-based position in the campaign, from 1 to 10.
	ID int
	// World is the number of the world the level belongs to.
	World int
	// Index is the level's one-based position within its world. Run levels
	// use 1 through 3; the boss level follows them at 4.
	Index int
	// Kind reports whether the level is a run or the boss fight.
	Kind LevelKind
	// Seed is the fixed generation seed of a run level. The boss level has
	// no corridor and carries the zero seed.
	Seed uint64
	// Params tunes the generated corridor of a run level. The boss level
	// carries the zero Params.
	Params levelgen.Params
}

// Campaign is the fixed campaign: three worlds, nine run levels, the boss
// finale, and the parameters for the endless mode.
type Campaign struct {
	// Worlds holds the three campaign worlds in order.
	Worlds []WorldDef
	// Levels holds the ten campaign levels in order.
	Levels []LevelDef
	// EndlessParams tunes every endless-mode level.
	EndlessParams levelgen.Params
	// StartLives is how many lives every run and the boss fight start with.
	StartLives int
}

// startLives is the number of lives a new run and the boss fight start with.
const startLives = 3

// runSeeds are the fixed, distinct seeds for campaign levels 1 through 9, in
// level order. They never change, so every attempt replays the same learnable
// corridor.
var runSeeds = [9]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}

// Default returns the fixed campaign: three worlds, nine seeded run levels
// and the boss finale, with the endless mode set to the hard preset.
func Default() *Campaign {
	c := &Campaign{
		Worlds: []WorldDef{
			{Number: 1, Theme: ThemeMangroves},
			{Number: 2, Theme: ThemeHarbor},
			{Number: 3, Theme: ThemeNeon},
		},
		EndlessParams: levelgen.Neon(),
		StartLives:    startLives,
	}

	presets := []levelgen.Params{levelgen.Mangroves(), levelgen.Harbor(), levelgen.Neon()}
	id := 1
	for w := range c.Worlds {
		for i := 1; i <= 3; i++ {
			c.Levels = append(c.Levels, LevelDef{
				ID:     id,
				World:  c.Worlds[w].Number,
				Index:  i,
				Kind:   LevelRun,
				Seed:   runSeeds[id-1],
				Params: presets[w],
			})
			id++
		}
	}

	c.Levels = append(c.Levels, LevelDef{
		ID:    id,
		World: 3,
		Index: 4,
		Kind:  LevelBoss,
	})
	return c
}

// LevelByID returns the campaign level whose ID is id, if any.
func (c *Campaign) LevelByID(id int) (LevelDef, bool) {
	for _, lv := range c.Levels {
		if lv.ID == id {
			return lv, true
		}
	}
	return LevelDef{}, false
}

// Next returns the campaign level that follows the level with ID id. After
// the final level, or for an id that is not a level, it reports not found.
func (c *Campaign) Next(id int) (LevelDef, bool) {
	for i, lv := range c.Levels {
		if lv.ID != id {
			continue
		}
		if i+1 < len(c.Levels) {
			return c.Levels[i+1], true
		}
		return LevelDef{}, false
	}
	return LevelDef{}, false
}

// BuildEndless builds an endless-mode level from seed using EndlessParams.
func (c *Campaign) BuildEndless(seed uint64) *levelgen.Level {
	return levelgen.New(seed, c.EndlessParams)
}
