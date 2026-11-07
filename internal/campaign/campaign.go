// Package campaign defines the game's worlds, levels and endless mode, and
// the scene flow that walks the player from the language pick and main menu
// through the campaign to the credits.
//
// The campaign is fixed: three worlds of three run levels each, followed by
// the Copycat boss finale. Every run level carries a fixed seed, so each
// attempt replays the same learnable corridor; the endless mode draws a fresh
// seed per run. The package is pure Go: it imports only the logic packages
// and the standard library, never raylib or audio, and it holds no
// player-facing strings.
package campaign

import (
	"time"

	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

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

// Scene identifies which screen the flow currently shows.
type Scene int

// The scenes the flow moves through.
const (
	// SceneLanguagePick is the first-run language choice screen.
	SceneLanguagePick Scene = iota
	// SceneMainMenu is the main menu.
	SceneMainMenu
	// SceneMap is the campaign level selection map.
	SceneMap
	// SceneSettings is the settings screen.
	SceneSettings
	// ScenePlay is an active run or boss fight.
	ScenePlay
	// ScenePause is a paused run; the run is not stepped while shown.
	ScenePause
	// SceneLevelEnd is the post-run screen. On the endless mode it is the
	// endless level end whose only exit is the menu.
	SceneLevelEnd
	// SceneGameOver is the death screen.
	SceneGameOver
	// SceneStory is a story card shown between worlds.
	SceneStory
	// SceneEnding is the ending story card shown after the boss is won.
	SceneEnding
	// SceneCredits is the credits screen after the ending card.
	SceneCredits
)

// Action is one menu or input action the player performed.
type Action int

// The actions the flow understands.
const (
	// ActionUp moves a menu selection up or backward.
	ActionUp Action = iota
	// ActionDown moves a menu selection down or forward.
	ActionDown
	// ActionLeft moves left or decreases a setting.
	ActionLeft
	// ActionRight moves right or increases a setting.
	ActionRight
	// ActionConfirm confirms the current selection.
	ActionConfirm
	// ActionBack returns to the previous screen.
	ActionBack
	// ActionPause toggles between playing and paused.
	ActionPause
)

// StoryKind identifies which story card the flow shows.
type StoryKind int

// The story cards the flow shows.
const (
	// StoryWorld2 is the card shown between worlds 1 and 2.
	StoryWorld2 StoryKind = iota
	// StoryWorld3 is the card shown between worlds 2 and 3.
	StoryWorld3
	// StoryEnding is the card shown after the boss is won.
	StoryEnding
)

// Language is a language the game's texts are shown in.
type Language int

// The languages the game ships with.
const (
	// LanguageEnglish is English.
	LanguageEnglish Language = iota
	// LanguageUkrainian is modern academic Ukrainian under the 2019
	// orthography.
	LanguageUkrainian
)

// Settings holds the player's preferences. It is plain data with exported
// fields so the save step can persist it as-is.
type Settings struct {
	// MusicVolume is the music loudness in [0, 1].
	MusicVolume float64
	// SoundVolume is the sound-effect loudness in [0, 1].
	SoundVolume float64
	// Fullscreen reports whether the window is fullscreen.
	Fullscreen bool
	// Language is the language every player-facing text is shown in.
	Language Language
}

// DefaultSettings returns the settings the game starts with: full volumes,
// windowed, English.
func DefaultSettings() Settings {
	return Settings{
		MusicVolume: 1,
		SoundVolume: 1,
		Fullscreen:  false,
		Language:    LanguageEnglish,
	}
}

// Progress holds the player's campaign progress. It is plain data with
// exported fields so the save step can persist it as-is.
type Progress struct {
	// Unlocked is the highest level ID the player may start. A fresh save
	// starts at 1; the boss, level 10, unlocks once level 9 is finished.
	Unlocked int
	// Completed reports that the boss has been defeated.
	Completed bool
	// BestEndless is the furthest distance, in meters, reached in the
	// endless mode.
	BestEndless float64
}

// NewProgress returns the progress a fresh save starts with: level 1
// unlocked, nothing completed, no endless best.
func NewProgress() Progress {
	return Progress{Unlocked: 1}
}

// Flow owns the scene state that drives the game from the language pick and
// main menu through the campaign to the ending. It is headless: Handle and
// StepRun take plain values and return plain values, and the main loop drives
// them without a window.
type Flow struct {
	c        *Campaign
	settings Settings
	progress Progress

	scene   Scene
	story   StoryKind
	quit    bool
	endless bool

	currentID int

	run  *gameplay.Run
	boss *gameplay.BossFight

	endlessSeed uint64
	maxZ        float64

	langSel int
	menuSel int
	mapSel  int
	setSel  int
	overSel int
}

// NewFlow returns a flow for campaign c carrying settings s and progress p.
// When firstRun is true it opens the language pick; otherwise it opens the
// main menu.
func NewFlow(c *Campaign, s Settings, p Progress, firstRun bool) *Flow {
	f := &Flow{
		c:        c,
		settings: s,
		progress: p,
		scene:    SceneMainMenu,
	}
	if firstRun {
		f.scene = SceneLanguagePick
	}
	return f
}

// Scene returns the scene the flow currently shows.
func (f *Flow) Scene() Scene { return f.scene }

// Settings returns the current settings.
func (f *Flow) Settings() Settings { return f.settings }

// Progress returns the current progress.
func (f *Flow) Progress() Progress { return f.progress }

// Story returns the story card the flow currently shows, or the ending card
// while SceneEnding is on screen. It is meaningful only on SceneStory and
// SceneEnding.
func (f *Flow) Story() StoryKind { return f.story }

// Endless reports whether the flow is in the endless mode.
func (f *Flow) Endless() bool { return f.endless }

// Quit reports whether the player asked to quit from the main menu.
func (f *Flow) Quit() bool { return f.quit }

// CurrentLevel returns the campaign level the flow is currently playing, or
// the last one it played, and false when it is in the endless mode or no
// level has been started.
func (f *Flow) CurrentLevel() (LevelDef, bool) {
	if f.currentID == 0 {
		return LevelDef{}, false
	}
	return f.c.LevelByID(f.currentID)
}

// Run returns the active run, or nil when the flow is not playing a run.
func (f *Flow) Run() *gameplay.Run { return f.run }

// Boss returns the active boss fight, or nil when the flow is not in one.
func (f *Flow) Boss() *gameplay.BossFight { return f.boss }

// MenuIndex returns the zero-based index of the highlighted selection in the
// current scene's list: the language pick, main menu, map, settings or game
// over. Scenes without a selection report 0.
func (f *Flow) MenuIndex() int {
	switch f.scene {
	case SceneLanguagePick:
		return f.langSel
	case SceneMainMenu:
		return f.menuSel
	case SceneMap:
		return f.mapSel
	case SceneSettings:
		return f.setSel
	case SceneGameOver:
		return f.overSel
	default:
		return 0
	}
}

// Handle applies one menu action to the current scene.
func (f *Flow) Handle(a Action) {
	switch f.scene {
	case SceneLanguagePick:
		f.handleLanguagePick(a)
	case SceneMainMenu:
		f.handleMainMenu(a)
	case SceneMap:
		f.handleMap(a)
	case SceneSettings:
		f.handleSettings(a)
	case ScenePlay:
		if a == ActionPause {
			f.scene = ScenePause
		}
	case ScenePause:
		if a == ActionPause {
			f.scene = ScenePlay
		}
	case SceneGameOver:
		f.handleGameOver(a)
	case SceneLevelEnd:
		f.handleLevelEnd(a)
	case SceneStory:
		f.handleStory(a)
	case SceneEnding:
		if a == ActionConfirm {
			f.scene = SceneCredits
		}
	case SceneCredits:
		if a == ActionConfirm {
			f.toMainMenu()
		}
	}
}

// StepRun advances the active run or boss fight by dt seconds with input in
// and returns the run's outcomes in order. When the flow is not playing, it
// changes nothing and returns no outcomes. During a boss fight it returns no
// run outcomes: the boss fight's own outcomes drive the win and loss
// transitions, and the render layer reads Boss() for its state.
func (f *Flow) StepRun(in runner.Input, dt float64) []gameplay.Outcome {
	if f.scene != ScenePlay {
		return nil
	}
	if f.run != nil {
		out := f.run.Step(in, dt)
		if z := f.run.Runner.Z(); z > f.maxZ {
			f.maxZ = z
		}
		switch f.run.State {
		case gameplay.StateDead:
			f.onRunDeath()
		case gameplay.StateFinished:
			f.onRunFinished()
		}
		return out
	}
	if f.boss != nil {
		f.boss.Step(in, dt)
		switch f.boss.State() {
		case gameplay.BossWon:
			f.onBossWon()
		case gameplay.BossLost:
			f.onBossLost()
		}
	}
	return nil
}

// StartLevel starts a run of the campaign level with ID id using the
// campaign's start lives. It does nothing when id is not unlocked. The boss
// level is started through StartBoss instead.
func (f *Flow) StartLevel(id int) {
	lv, ok := f.c.LevelByID(id)
	if !ok || id > f.progress.Unlocked {
		return
	}
	if lv.Kind == LevelBoss {
		f.StartBoss()
		return
	}
	f.startRun(lv)
}

// StartBoss starts the Copycat boss fight with the campaign's start lives.
func (f *Flow) StartBoss() {
	f.boss = gameplay.NewBossFight(f.c.StartLives)
	f.run = nil
	f.currentID = 0
	for _, lv := range f.c.Levels {
		if lv.Kind == LevelBoss {
			f.currentID = lv.ID
			break
		}
	}
	f.endless = false
	f.maxZ = 0
	f.scene = ScenePlay
}

// StartEndless starts an endless run generated from seed with the campaign's
// start lives.
func (f *Flow) StartEndless(seed uint64) {
	f.endlessSeed = seed
	l := f.c.BuildEndless(seed)
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = f.c.EndlessParams.BaseSpeed
	f.run = gameplay.NewRun(runner.New(l, cfg), l, f.c.StartLives)
	f.boss = nil
	f.currentID = 0
	f.endless = true
	f.maxZ = 0
	f.scene = ScenePlay
}

// Retry restarts the current run, boss fight or endless run with fresh lives
// and full pickups.
func (f *Flow) Retry() {
	switch {
	case f.endless:
		f.StartEndless(f.endlessSeed)
	case f.boss != nil:
		f.StartBoss()
	case f.run != nil:
		if lv, ok := f.c.LevelByID(f.currentID); ok {
			f.startRun(lv)
		}
	}
}

// startRun begins a run of the given campaign level.
func (f *Flow) startRun(lv LevelDef) {
	l := levelgen.New(lv.Seed, lv.Params)
	cfg := runner.DefaultConfig()
	cfg.BaseSpeed = lv.Params.BaseSpeed
	f.run = gameplay.NewRun(runner.New(l, cfg), l, f.c.StartLives)
	f.boss = nil
	f.currentID = lv.ID
	f.endless = false
	f.maxZ = 0
	f.scene = ScenePlay
}

// onRunDeath moves the flow past a dead run. An endless death records the
// distance and opens the endless level end; a campaign death opens the game
// over screen.
func (f *Flow) onRunDeath() {
	if f.endless {
		f.recordEndlessBest()
		f.scene = SceneLevelEnd
		return
	}
	f.scene = SceneGameOver
	f.overSel = 0
}

// onRunFinished moves the flow past a finished run. A campaign run unlocks
// the next level and opens the level end; an endless run that reaches the
// finish records its distance and opens the endless level end.
func (f *Flow) onRunFinished() {
	if f.endless {
		f.recordEndlessBest()
		f.scene = SceneLevelEnd
		return
	}
	if f.currentID > 0 && f.progress.Unlocked < f.currentID+1 {
		f.progress.Unlocked = f.currentID + 1
	}
	f.scene = SceneLevelEnd
}

// onBossWon completes the campaign and shows the ending card.
func (f *Flow) onBossWon() {
	f.progress.Completed = true
	if f.progress.Unlocked < 10 {
		f.progress.Unlocked = 10
	}
	f.scene = SceneEnding
	f.story = StoryEnding
}

// onBossLost opens the game over screen.
func (f *Flow) onBossLost() {
	f.scene = SceneGameOver
	f.overSel = 0
}

// recordEndlessBest raises BestEndless to the furthest distance the current
// endless run reached.
func (f *Flow) recordEndlessBest() {
	if f.maxZ > f.progress.BestEndless {
		f.progress.BestEndless = f.maxZ
	}
}

// toMainMenu returns to the main menu and clears any active run.
func (f *Flow) toMainMenu() {
	f.scene = SceneMainMenu
	f.menuSel = 0
	f.run = nil
	f.boss = nil
	f.endless = false
	f.currentID = 0
	f.maxZ = 0
}

// continueGame starts the highest unlocked level, or a boss rematch when the
// campaign is complete.
func (f *Flow) continueGame() {
	if f.progress.Completed {
		f.StartBoss()
		return
	}
	highest := f.progress.Unlocked
	if highest > 10 {
		highest = 10
	}
	if highest < 1 {
		highest = 1
	}
	if highest == 10 {
		f.StartBoss()
		return
	}
	f.StartLevel(highest)
}

// The main menu items, in display order.
const (
	menuContinue = iota
	menuNewGame
	menuMap
	menuEndless
	menuSettings
	menuQuit
	menuCount
)

// handleLanguagePick moves the first-run language choice.
func (f *Flow) handleLanguagePick(a Action) {
	switch a {
	case ActionUp, ActionDown, ActionLeft, ActionRight:
		f.langSel = 1 - f.langSel
	case ActionConfirm:
		if f.langSel == 1 {
			f.settings.Language = LanguageUkrainian
		} else {
			f.settings.Language = LanguageEnglish
		}
		f.scene = SceneMainMenu
		f.menuSel = 0
	}
}

// handleMainMenu moves the main menu selection and activates it.
func (f *Flow) handleMainMenu(a Action) {
	switch a {
	case ActionUp:
		f.menuSel = (f.menuSel - 1 + menuCount) % menuCount
	case ActionDown:
		f.menuSel = (f.menuSel + 1) % menuCount
	case ActionConfirm:
		switch f.menuSel {
		case menuContinue:
			f.continueGame()
		case menuNewGame:
			f.progress = NewProgress()
			f.StartLevel(1)
		case menuMap:
			f.scene = SceneMap
			f.mapSel = 0
		case menuEndless:
			f.StartEndless(freshSeed())
		case menuSettings:
			f.scene = SceneSettings
			f.setSel = 0
		case menuQuit:
			f.quit = true
		}
	}
}

// handleMap moves the level selection map. Locked levels refuse selection;
// the boss level, ID 10, unlocks after level 9 is finished.
func (f *Flow) handleMap(a Action) {
	switch a {
	case ActionUp, ActionLeft:
		if f.mapSel > 0 {
			f.mapSel--
		}
	case ActionDown, ActionRight:
		if f.mapSel < len(f.c.Levels)-1 {
			f.mapSel++
		}
	case ActionConfirm:
		id := f.mapSel + 1
		lv, ok := f.c.LevelByID(id)
		if !ok || id > f.progress.Unlocked {
			return
		}
		if lv.Kind == LevelBoss {
			f.StartBoss()
			return
		}
		f.StartLevel(id)
	case ActionBack:
		f.toMainMenu()
	}
}

// The settings items, in display order.
const (
	setMusic = iota
	setSound
	setFullscreen
	setLanguage
	setCount
)

// volumeStep is how much one left/right press changes a volume, and it keeps
// the two settings inside [0, 1].
const volumeStep = 0.1

// handleSettings moves and edits the settings screen.
func (f *Flow) handleSettings(a Action) {
	switch a {
	case ActionUp:
		f.setSel = (f.setSel - 1 + setCount) % setCount
	case ActionDown:
		f.setSel = (f.setSel + 1) % setCount
	case ActionLeft:
		f.adjustSetting(f.setSel, -1)
	case ActionRight:
		f.adjustSetting(f.setSel, 1)
	case ActionConfirm:
		f.toggleSetting(f.setSel)
	case ActionBack:
		f.toMainMenu()
	}
}

// adjustSetting changes the selected setting by dir in the direction of left
// or right.
func (f *Flow) adjustSetting(sel, dir int) {
	switch sel {
	case setMusic:
		f.settings.MusicVolume = clampVolume(f.settings.MusicVolume + float64(dir)*volumeStep)
	case setSound:
		f.settings.SoundVolume = clampVolume(f.settings.SoundVolume + float64(dir)*volumeStep)
	case setFullscreen:
		f.settings.Fullscreen = !f.settings.Fullscreen
	case setLanguage:
		f.settings.Language = otherLanguage(f.settings.Language)
	}
}

// toggleSetting flips the boolean settings on confirm.
func (f *Flow) toggleSetting(sel int) {
	switch sel {
	case setFullscreen:
		f.settings.Fullscreen = !f.settings.Fullscreen
	case setLanguage:
		f.settings.Language = otherLanguage(f.settings.Language)
	}
}

// The game over items, in display order.
const (
	overRetry = iota
	overMenu
	overCount
)

// handleGameOver moves the game over screen. Retry restarts the current run
// with fresh lives; Menu returns to the main menu.
func (f *Flow) handleGameOver(a Action) {
	switch a {
	case ActionUp:
		f.overSel = (f.overSel - 1 + overCount) % overCount
	case ActionDown:
		f.overSel = (f.overSel + 1) % overCount
	case ActionConfirm:
		switch f.overSel {
		case overRetry:
			f.Retry()
		case overMenu:
			f.toMainMenu()
		}
	}
}

// handleLevelEnd continues past the level end screen. After levels 3 and 6 a
// story card follows; after level 9 the boss follows; otherwise the next
// level follows. The endless level end's only exit is the menu.
func (f *Flow) handleLevelEnd(a Action) {
	if a != ActionConfirm {
		return
	}
	if f.endless {
		f.toMainMenu()
		return
	}
	switch f.currentID {
	case 3:
		f.scene = SceneStory
		f.story = StoryWorld2
	case 6:
		f.scene = SceneStory
		f.story = StoryWorld3
	case 9:
		f.StartBoss()
	default:
		f.StartLevel(f.currentID + 1)
	}
}

// handleStory continues past a world story card into the next world.
func (f *Flow) handleStory(a Action) {
	if a != ActionConfirm {
		return
	}
	switch f.story {
	case StoryWorld2:
		f.StartLevel(4)
	case StoryWorld3:
		f.StartLevel(7)
	}
}

// clampVolume clamps v to [0, 1].
func clampVolume(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// otherLanguage returns the language that is not l.
func otherLanguage(l Language) Language {
	if l == LanguageEnglish {
		return LanguageUkrainian
	}
	return LanguageEnglish
}

// freshSeed draws a new seed for an endless run started from the menu. The
// campaign levels stay fixed; only the endless mode is meant to differ every
// run.
func freshSeed() uint64 {
	return uint64(time.Now().UnixNano())
}
