package campaign

import (
	"testing"

	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

const flowStepDT = 1.0 / 60.0

func TestFreshFlowOpensLanguagePickAndStoresChoice(t *testing.T) {
	c := finishingCampaign()

	f := NewFlow(c, DefaultSettings(), NewProgress(), true)
	if f.Scene() != SceneLanguagePick {
		t.Fatalf("Scene = %v, want SceneLanguagePick", f.Scene())
	}

	// The default selection is English.
	f.Handle(ActionConfirm)
	if f.Settings().Language != LanguageEnglish {
		t.Fatalf("Language = %v after default confirm, want LanguageEnglish", f.Settings().Language)
	}
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after language confirm, want SceneMainMenu", f.Scene())
	}

	// Choosing down selects Ukrainian and stores it.
	f2 := NewFlow(c, DefaultSettings(), NewProgress(), true)
	f2.Handle(ActionDown)
	f2.Handle(ActionConfirm)
	if f2.Settings().Language != LanguageUkrainian {
		t.Fatalf("Language = %v after Ukrainian pick, want LanguageUkrainian", f2.Settings().Language)
	}
	if f2.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after Ukrainian pick, want SceneMainMenu", f2.Scene())
	}
}

func TestNewGameOpensStoryIntroThenStartsLevel1(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), Progress{Unlocked: 5, Completed: true, BestEndless: 42}, false)

	// Select New Game (the second menu item).
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)

	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after New Game, want SceneNewGameConfirm", f.Scene())
	}
	if got := f.Progress(); got != (Progress{Unlocked: 5, Completed: true, BestEndless: 42}) {
		t.Fatalf("Progress changed before confirmation: %+v", got)
	}

	// Confirm Yes to erase progress.
	f.Handle(ActionUp)
	f.Handle(ActionConfirm)

	if f.Scene() != SceneStory {
		t.Fatalf("Scene = %v after New Game confirm, want SceneStory", f.Scene())
	}
	if f.Story() != StoryIntro {
		t.Fatalf("Story = %v after New Game, want StoryIntro", f.Story())
	}
	if got := f.Progress(); got != NewProgress() {
		t.Fatalf("Progress after New Game = %+v, want %+v", got, NewProgress())
	}

	// Confirming the intro card starts level 1.
	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after StoryIntro confirm, want ScenePlay", f.Scene())
	}
	lv, ok := f.CurrentLevel()
	if !ok || lv.ID != 1 {
		t.Fatalf("CurrentLevel = %+v, %v, want level 1", lv, ok)
	}
}

func TestNewGameStartsLevel1AndFinishUnlocksLevel2(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), Progress{Unlocked: 5, Completed: true, BestEndless: 42}, false)

	// Select New Game (the second menu item).
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)

	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after New Game, want SceneNewGameConfirm", f.Scene())
	}
	if got := f.Progress(); got != (Progress{Unlocked: 5, Completed: true, BestEndless: 42}) {
		t.Fatalf("Progress changed before confirmation: %+v", got)
	}

	// Confirm Yes to erase progress.
	f.Handle(ActionUp)
	f.Handle(ActionConfirm)

	if f.Scene() != SceneStory {
		t.Fatalf("Scene = %v after New Game confirm, want SceneStory", f.Scene())
	}
	if f.Story() != StoryIntro {
		t.Fatalf("Story = %v after New Game, want StoryIntro", f.Story())
	}
	if got := f.Progress(); got != NewProgress() {
		t.Fatalf("Progress after New Game = %+v, want %+v", got, NewProgress())
	}

	// Confirming the intro card starts level 1.
	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after StoryIntro confirm, want ScenePlay", f.Scene())
	}
	lv, ok := f.CurrentLevel()
	if !ok || lv.ID != 1 {
		t.Fatalf("CurrentLevel = %+v, %v, want level 1", lv, ok)
	}

	finishRun(t, f)

	if f.Scene() != SceneLevelEnd {
		t.Fatalf("Scene = %v after finishing level 1, want SceneLevelEnd", f.Scene())
	}
	if f.Progress().Unlocked != 2 {
		t.Fatalf("Unlocked = %d after finishing level 1, want 2", f.Progress().Unlocked)
	}

	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after level end confirm, want ScenePlay", f.Scene())
	}
	lv, ok = f.CurrentLevel()
	if !ok || lv.ID != 2 {
		t.Fatalf("CurrentLevel = %+v, %v, want level 2", lv, ok)
	}
}

func TestStoryCardsBetweenWorlds(t *testing.T) {
	cases := []struct {
		name       string
		finishID   int
		wantStory  StoryKind
		wantNextID int
	}{
		{name: "world one to two", finishID: 3, wantStory: StoryWorld2, wantNextID: 4},
		{name: "world two to three", finishID: 6, wantStory: StoryWorld3, wantNextID: 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := finishingCampaign()
			f := NewFlow(c, DefaultSettings(), Progress{Unlocked: tc.finishID}, false)
			f.StartLevel(tc.finishID)

			finishRun(t, f)
			if f.Scene() != SceneLevelEnd {
				t.Fatalf("Scene = %v after level %d, want SceneLevelEnd", f.Scene(), tc.finishID)
			}

			f.Handle(ActionConfirm)
			if f.Scene() != SceneStory {
				t.Fatalf("Scene = %v after level end confirm, want SceneStory", f.Scene())
			}
			if f.Story() != tc.wantStory {
				t.Fatalf("Story = %v, want %v", f.Story(), tc.wantStory)
			}

			f.Handle(ActionConfirm)
			if f.Scene() != ScenePlay {
				t.Fatalf("Scene = %v after story confirm, want ScenePlay", f.Scene())
			}
			lv, ok := f.CurrentLevel()
			if !ok || lv.ID != tc.wantNextID {
				t.Fatalf("CurrentLevel = %+v, %v, want level %d", lv, ok, tc.wantNextID)
			}
		})
	}
}

func TestLevel9UnlocksBossAndBossWinReachesCredits(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), Progress{Unlocked: 9}, false)

	f.StartLevel(9)
	finishRun(t, f)

	if f.Scene() != SceneLevelEnd {
		t.Fatalf("Scene = %v after level 9, want SceneLevelEnd", f.Scene())
	}
	if f.Progress().Unlocked != 10 {
		t.Fatalf("Unlocked = %d after level 9, want 10 (boss unlocked)", f.Progress().Unlocked)
	}

	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after level 9 confirm, want ScenePlay (boss)", f.Scene())
	}
	if f.Boss() == nil {
		t.Fatalf("Boss = nil after level 9, want a boss fight")
	}
	lv, ok := f.CurrentLevel()
	if !ok || lv.ID != 10 || lv.Kind != LevelBoss {
		t.Fatalf("CurrentLevel = %+v, %v, want boss level 10", lv, ok)
	}

	winBoss(t, f)

	if f.Scene() != SceneEnding {
		t.Fatalf("Scene = %v after boss win, want SceneEnding", f.Scene())
	}
	if !f.Progress().Completed {
		t.Fatalf("Completed = false after boss win, want true")
	}
	if f.Story() != StoryEnding {
		t.Fatalf("Story = %v after boss win, want StoryEnding", f.Story())
	}

	f.Handle(ActionConfirm)
	if f.Scene() != SceneCredits {
		t.Fatalf("Scene = %v after ending confirm, want SceneCredits", f.Scene())
	}
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after credits confirm, want SceneMainMenu", f.Scene())
	}
}

func TestBossLossOpensGameOverAndRetryRestartsBoss(t *testing.T) {
	c := finishingCampaign()
	c.StartLives = 1
	f := NewFlow(c, DefaultSettings(), Progress{Unlocked: 10}, false)

	f.StartBoss()
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after StartBoss, want ScenePlay", f.Scene())
	}
	b := f.Boss()
	if b == nil {
		t.Fatalf("Boss = nil after StartBoss")
	}

	// Announce the first attack, then stand on its lane without dodging.
	f.StepRun(runner.Input{}, flowStepDT)
	target := b.AttackLane()
	for i := 0; i < 4 && b.HeroLane() != target; i++ {
		var in runner.Input
		if b.HeroLane() < target {
			in.Right = true
		} else {
			in.Left = true
		}
		f.StepRun(in, flowStepDT)
	}

	// Let the announced attack land undodged: the single life is lost.
	for i := 0; i < 400 && f.Scene() == ScenePlay; i++ {
		f.StepRun(runner.Input{}, flowStepDT)
	}
	if f.Scene() != SceneGameOver {
		t.Fatalf("Scene = %v after boss loss, want SceneGameOver", f.Scene())
	}
	if b.State() != gameplay.BossLost {
		t.Fatalf("Boss state = %v after boss loss, want BossLost", b.State())
	}

	// Retry restarts the boss fight.
	f.Handle(ActionConfirm) // Retry is the first game-over item.
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after boss retry, want ScenePlay", f.Scene())
	}
	if got := f.Boss(); got == nil || got.State() != gameplay.BossFighting {
		t.Fatalf("Boss after retry = %v, want a fresh BossFighting fight", got)
	}
	if got := f.Boss().Lives(); got != c.StartLives {
		t.Fatalf("Boss lives after retry = %d, want %d", got, c.StartLives)
	}
}

func TestGameOverRetriesSameLevelWithFreshLivesAndPauseBlocks(t *testing.T) {
	c := pitCampaign(t)
	f := NewFlow(c, DefaultSettings(), NewProgress(), false)

	// New Game.
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)
	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after New Game, want SceneNewGameConfirm", f.Scene())
	}
	f.Handle(ActionUp) // Yes
	f.Handle(ActionConfirm)
	if f.Scene() != SceneStory {
		t.Fatalf("Scene = %v after New Game confirm, want SceneStory", f.Scene())
	}
	f.Handle(ActionConfirm) // StoryIntro → level 1.
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after StoryIntro confirm, want ScenePlay", f.Scene())
	}

	// Pause stops the run and blocks StepRun.
	f.Handle(ActionPause)
	if f.Scene() != ScenePause {
		t.Fatalf("Scene = %v after pause, want ScenePause", f.Scene())
	}
	before := f.Run().Runner.Z()
	if outs := f.StepRun(runner.Input{Jump: true}, flowStepDT); len(outs) != 0 {
		t.Fatalf("StepRun while paused returned %v, want none", outs)
	}
	if got := f.Run().Runner.Z(); got != before {
		t.Fatalf("runner Z = %v while paused, want unchanged %v", got, before)
	}
	f.Handle(ActionPause)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after resume, want ScenePlay", f.Scene())
	}

	// Die in the level's first pit with one life remaining.
	pit, ok := firstPit(f.Run().Level)
	if !ok {
		t.Fatalf("level has no pit to die in")
	}
	f.Run().Runner.Restart(pit.Z - 4)
	f.Run().Lives = 1
	stepUntilScene(t, f, SceneGameOver)
	if f.Scene() != SceneGameOver {
		t.Fatalf("Scene = %v after last life, want SceneGameOver", f.Scene())
	}

	// Retry restarts the same level with fresh lives.
	f.Handle(ActionConfirm) // Retry is the first game-over item.
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after retry, want ScenePlay", f.Scene())
	}
	lv, ok := f.CurrentLevel()
	if !ok || lv.ID != 1 {
		t.Fatalf("CurrentLevel = %+v, %v after retry, want level 1", lv, ok)
	}
	if f.Run().Lives != c.StartLives {
		t.Fatalf("Lives = %d after retry, want %d", f.Run().Lives, c.StartLives)
	}
	if f.Run().Runner.Z() != 0 {
		t.Fatalf("runner Z = %v after retry, want 0", f.Run().Runner.Z())
	}
}

func TestMapRefusesLockedLevels(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), NewProgress(), false)

	// Open the map (third menu item).
	f.Handle(ActionDown)
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMap {
		t.Fatalf("Scene = %v after map selection, want SceneMap", f.Scene())
	}

	// Level 2 is locked.
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMap {
		t.Fatalf("Scene = %v after locked level confirm, want SceneMap", f.Scene())
	}
	if f.Run() != nil {
		t.Fatalf("Run started for a locked level, want none")
	}

	// The boss (level 10) is locked while Unlocked is below 10.
	for i := 0; i < 9; i++ {
		f.Handle(ActionDown)
	}
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMap {
		t.Fatalf("Scene = %v after locked boss confirm, want SceneMap", f.Scene())
	}
	if f.Boss() != nil {
		t.Fatalf("Boss started while locked, want none")
	}

	// Level 1 is unlocked.
	f.mapSel = 0
	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after unlocked level confirm, want ScenePlay", f.Scene())
	}
	lv, ok := f.CurrentLevel()
	if !ok || lv.ID != 1 {
		t.Fatalf("CurrentLevel = %+v, %v, want level 1", lv, ok)
	}
}

func TestMapAllowsUnlockedBoss(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), Progress{Unlocked: 10}, false)

	f.Handle(ActionDown)
	f.Handle(ActionDown)
	f.Handle(ActionConfirm) // Map.
	for i := 0; i < 9; i++ {
		f.Handle(ActionDown) // Level 10, the boss.
	}
	f.Handle(ActionConfirm)

	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after boss confirm, want ScenePlay", f.Scene())
	}
	if f.Boss() == nil {
		t.Fatalf("Boss = nil after unlocked boss confirm, want a boss fight")
	}
}

func TestSettingsRoundTripThroughFlow(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), NewProgress(), false)

	if f.Settings() != DefaultSettings() {
		t.Fatalf("Settings = %+v, want %+v", f.Settings(), DefaultSettings())
	}
	if f.Progress() != NewProgress() {
		t.Fatalf("Progress = %+v, want %+v", f.Progress(), NewProgress())
	}

	// Open Settings (fifth menu item).
	for i := 0; i < 4; i++ {
		f.Handle(ActionDown)
	}
	f.Handle(ActionConfirm)
	if f.Scene() != SceneSettings {
		t.Fatalf("Scene = %v after settings selection, want SceneSettings", f.Scene())
	}

	// Music volume down one step.
	f.Handle(ActionLeft)
	if got, want := f.Settings().MusicVolume, 1.0-volumeStep; !floatClose(got, want) {
		t.Fatalf("MusicVolume = %v after left, want %v", got, want)
	}

	// Sound volume right one step stays clamped at 1.
	f.Handle(ActionDown)
	f.Handle(ActionRight)
	if got := f.Settings().SoundVolume; !floatClose(got, 1) {
		t.Fatalf("SoundVolume = %v after right at max, want 1", got)
	}

	// Fullscreen toggles.
	f.Handle(ActionDown)
	f.Handle(ActionRight)
	if !f.Settings().Fullscreen {
		t.Fatalf("Fullscreen = false after toggle, want true")
	}

	// Language cycles to Ukrainian.
	f.Handle(ActionDown)
	f.Handle(ActionRight)
	if f.Settings().Language != LanguageUkrainian {
		t.Fatalf("Language = %v after toggle, want LanguageUkrainian", f.Settings().Language)
	}

	f.Handle(ActionBack)
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after settings back, want SceneMainMenu", f.Scene())
	}

	got := f.Settings()
	if !floatClose(got.MusicVolume, 1.0-volumeStep) || !got.Fullscreen || got.Language != LanguageUkrainian {
		t.Fatalf("settings did not round-trip: %+v", got)
	}
}

func TestContinueStartsHighestUnlockedAndBossRematch(t *testing.T) {
	c := finishingCampaign()

	f := NewFlow(c, DefaultSettings(), Progress{Unlocked: 4}, false)
	f.Handle(ActionConfirm) // Continue is the first menu item.
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after Continue, want ScenePlay", f.Scene())
	}
	lv, ok := f.CurrentLevel()
	if !ok || lv.ID != 4 {
		t.Fatalf("CurrentLevel = %+v, %v after Continue, want level 4", lv, ok)
	}

	f2 := NewFlow(c, DefaultSettings(), Progress{Unlocked: 10, Completed: true}, false)
	f2.Handle(ActionConfirm)
	if f2.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after completed Continue, want ScenePlay", f2.Scene())
	}
	if f2.Boss() == nil {
		t.Fatalf("Boss = nil after completed Continue, want a boss rematch")
	}
}

func TestEndlessDeterminism(t *testing.T) {
	c := finishingCampaign()

	f1 := NewFlow(c, DefaultSettings(), NewProgress(), false)
	f1.StartEndless(42)
	f2 := NewFlow(c, DefaultSettings(), NewProgress(), false)
	f2.StartEndless(42)

	if !f1.Endless() || !f2.Endless() {
		t.Fatalf("Endless = %v/%v, want true/true", f1.Endless(), f2.Endless())
	}
	if _, ok := f1.CurrentLevel(); ok {
		t.Fatalf("CurrentLevel reported a campaign level in endless mode")
	}

	for i := 0; i < 120; i++ {
		in := runner.Input{Jump: i%60 == 0}
		f1.StepRun(in, flowStepDT)
		f2.StepRun(in, flowStepDT)
		if got, want := f1.Run().Runner.Z(), f2.Run().Runner.Z(); got != want {
			t.Fatalf("step %d: endless Z differs: %v vs %v", i, got, want)
		}
	}
}

func TestEndlessDeathUpdatesBestEndless(t *testing.T) {
	c, seed, pit := endlessCampaign(t)
	f := NewFlow(c, DefaultSettings(), NewProgress(), false)

	f.StartEndless(seed)
	if !f.Endless() {
		t.Fatalf("Endless = false after StartEndless, want true")
	}

	f.Run().Runner.Restart(pit.Z - 4)
	f.Run().Lives = 1
	stepUntilScene(t, f, SceneLevelEnd)

	if f.Scene() != SceneLevelEnd {
		t.Fatalf("Scene = %v after endless death, want SceneLevelEnd", f.Scene())
	}
	if !f.Endless() {
		t.Fatalf("Endless = false after endless death, want true")
	}
	if f.Progress().BestEndless <= 0 {
		t.Fatalf("BestEndless = %v after endless death, want > 0", f.Progress().BestEndless)
	}

	// The endless level end's only exit is the menu.
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after endless level end confirm, want SceneMainMenu", f.Scene())
	}
	if f.Endless() {
		t.Fatalf("Endless = true after leaving endless level end, want false")
	}
}

func TestPauseOffersResumeRestartAndMenu(t *testing.T) {
	c := finishingCampaign()
	f := NewFlow(c, DefaultSettings(), NewProgress(), false)
	f.StartLevel(1)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v, want ScenePlay", f.Scene())
	}

	// Pause opens the menu with Resume highlighted.
	f.Handle(ActionPause)
	if f.Scene() != ScenePause {
		t.Fatalf("Scene = %v after pause, want ScenePause", f.Scene())
	}
	if f.MenuIndex() != pauseResume {
		t.Fatalf("MenuIndex = %d at pause, want %d (Resume)", f.MenuIndex(), pauseResume)
	}

	// Resume returns to play.
	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after Resume, want ScenePlay", f.Scene())
	}

	// Restart rebuilds the run with fresh lives at the start.
	f.Handle(ActionPause)
	f.Run().Lives = 1
	f.Run().Runner.Restart(12)
	f.Handle(ActionDown) // Restart
	f.Handle(ActionConfirm)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after Restart, want ScenePlay", f.Scene())
	}
	if f.Run() == nil {
		t.Fatal("Run = nil after Restart")
	}
	if got := f.Run().Lives; got != c.StartLives {
		t.Fatalf("Lives = %d after Restart, want %d", got, c.StartLives)
	}
	if got := f.Run().Runner.Z(); got != 0 {
		t.Fatalf("runner Z = %v after Restart, want 0", got)
	}

	// Back to menu returns to the main menu and clears the run.
	f.Handle(ActionPause)
	f.Handle(ActionDown)
	f.Handle(ActionDown) // Back to menu
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after Back to menu, want SceneMainMenu", f.Scene())
	}
	if f.Run() != nil || f.Boss() != nil {
		t.Fatalf("active run/boss remains after Back to menu")
	}

	// ActionPause stays an immediate resume shortcut.
	f.StartLevel(1)
	f.Handle(ActionPause)
	if f.Scene() != ScenePause {
		t.Fatalf("Scene = %v after second pause, want ScenePause", f.Scene())
	}
	f.Handle(ActionPause)
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after ActionPause resume, want ScenePlay", f.Scene())
	}
}

func TestNewGameConfirmsBeforeErasingProgress(t *testing.T) {
	c := finishingCampaign()
	before := Progress{Unlocked: 5, Completed: true, BestEndless: 42}
	f := NewFlow(c, DefaultSettings(), before, false)

	// Selecting New Game opens the confirmation without erasing progress.
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)
	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after New Game, want SceneNewGameConfirm", f.Scene())
	}
	if got := f.Progress(); got != before {
		t.Fatalf("Progress changed on opening confirmation: %+v", got)
	}

	// The default selection is No; confirming it cancels with progress intact.
	f.Handle(ActionConfirm)
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after cancel, want SceneMainMenu", f.Scene())
	}
	if got := f.Progress(); got != before {
		t.Fatalf("Progress erased on cancel: %+v", got)
	}

	// Back also cancels without erasing progress.
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)
	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after second New Game, want SceneNewGameConfirm", f.Scene())
	}
	f.Handle(ActionBack)
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after Back, want SceneMainMenu", f.Scene())
	}
	if got := f.Progress(); got != before {
		t.Fatalf("Progress erased on Back: %+v", got)
	}

	// Confirming Yes erases progress and opens the story intro.
	f.Handle(ActionDown)
	f.Handle(ActionConfirm)
	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after third New Game, want SceneNewGameConfirm", f.Scene())
	}
	f.Handle(ActionUp) // Yes
	f.Handle(ActionConfirm)
	if f.Scene() != SceneStory {
		t.Fatalf("Scene = %v after Yes, want SceneStory", f.Scene())
	}
	if f.Story() != StoryIntro {
		t.Fatalf("Story = %v after Yes, want StoryIntro", f.Story())
	}
	if got := f.Progress(); got != NewProgress() {
		t.Fatalf("Progress = %+v after Yes, want %+v", got, NewProgress())
	}
}

// finishingCampaign returns a full ten-level campaign whose run levels are
// short and empty, so a run finishes by simply auto-running forward.
func finishingCampaign() *Campaign {
	c := &Campaign{
		Worlds: []WorldDef{
			{Number: 1, Theme: ThemeMangroves},
			{Number: 2, Theme: ThemeHarbor},
			{Number: 3, Theme: ThemeNeon},
		},
		EndlessParams: flatParams(),
		StartLives:    3,
	}
	for id := 1; id <= 9; id++ {
		c.Levels = append(c.Levels, LevelDef{
			ID:     id,
			World:  (id-1)/3 + 1,
			Index:  (id-1)%3 + 1,
			Kind:   LevelRun,
			Seed:   uint64(id),
			Params: flatParams(),
		})
	}
	c.Levels = append(c.Levels, LevelDef{ID: 10, World: 3, Index: 4, Kind: LevelBoss})
	return c
}

// flatParams returns a short, obstacle-free level preset.
func flatParams() levelgen.Params {
	return levelgen.Params{
		BaseSpeed: 8,
		Length:    60,
		Density:   0,
	}
}

// pitCampaign returns the finishing campaign with level 1 replaced by a level
// that definitely contains a pit.
func pitCampaign(t *testing.T) *Campaign {
	t.Helper()
	p, seed := pitParams(t)
	c := finishingCampaign()
	c.Levels[0] = LevelDef{ID: 1, World: 1, Index: 1, Kind: LevelRun, Seed: seed, Params: p}
	return c
}

// endlessCampaign returns the finishing campaign with an endless preset that
// contains a pit, the seed it was found at, and that pit.
func endlessCampaign(t *testing.T) (*Campaign, uint64, levelgen.Element) {
	t.Helper()
	p, seed := pitParams(t)
	c := finishingCampaign()
	c.EndlessParams = p
	l := levelgen.New(seed, p)
	pit, ok := firstPit(l)
	if !ok {
		t.Fatalf("no pit in endless level for seed %d", seed)
	}
	return c, seed, pit
}

// pitParams finds a preset and seed whose generated level contains a pit.
func pitParams(t *testing.T) (levelgen.Params, uint64) {
	t.Helper()
	p := levelgen.Params{BaseSpeed: 8, Length: 120, Density: 0.4}
	for seed := uint64(0); seed < 200; seed++ {
		if _, ok := firstPit(levelgen.New(seed, p)); ok {
			return p, seed
		}
	}
	t.Fatalf("no pit within 200 seeds")
	return p, 0
}

// firstPit returns the first pit in level l, if any.
func firstPit(l *levelgen.Level) (levelgen.Element, bool) {
	for _, e := range l.Elements(0, l.Length()) {
		if e.Kind == levelgen.Pit {
			return e, true
		}
	}
	return levelgen.Element{}, false
}

// finishRun steps the flow with no input until it leaves ScenePlay, which on
// a flat level means the run finished.
func finishRun(t *testing.T, f *Flow) {
	t.Helper()
	for i := 0; i < 5000 && f.Scene() == ScenePlay; i++ {
		f.StepRun(runner.Input{}, flowStepDT)
	}
	if f.Scene() == ScenePlay {
		t.Fatalf("run never left ScenePlay")
	}
}

// stepUntilScene steps the flow with no input until it reaches want.
func stepUntilScene(t *testing.T, f *Flow, want Scene) {
	t.Helper()
	for i := 0; i < 5000 && f.Scene() != want; i++ {
		f.StepRun(runner.Input{}, flowStepDT)
	}
}

// winBoss drives the active boss fight with a dodge-and-stomp script until it
// wins.
func winBoss(t *testing.T, f *Flow) {
	t.Helper()
	b := f.Boss()
	if b == nil {
		t.Fatalf("no boss fight to win")
	}
	for i := 0; i < 4000 && b.State() == gameplay.BossFighting; i++ {
		f.StepRun(bossWinningInput(b), flowStepDT)
	}
	if b.State() != gameplay.BossWon {
		t.Fatalf("boss did not win: state=%v", b.State())
	}
}

// bossWinningInput reacts to the fight: dodge attacks, and stomp Copycat when
// he is vulnerable.
func bossWinningInput(b *gameplay.BossFight) runner.Input {
	if b.Vulnerable() {
		var in runner.Input
		target := b.AttackLane()
		switch {
		case b.HeroLane() < target:
			in.Right = true
		case b.HeroLane() > target:
			in.Left = true
		default:
			in.Jump = true
		}
		return in
	}
	return bossDodgeInput(b)
}

// bossDodgeInput moves toward the announced attack lane and applies the
// matching dodge.
func bossDodgeInput(b *gameplay.BossFight) runner.Input {
	var in runner.Input
	target := b.AttackLane()
	switch {
	case b.HeroLane() < target:
		in.Right = true
	case b.HeroLane() > target:
		in.Left = true
	default:
		if b.AttackKind() == gameplay.AttackLow {
			in.Jump = true
		} else if !b.Jumping() {
			in.Slide = true
		}
	}
	return in
}

// floatClose reports whether a and b are within a small tolerance.
func floatClose(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-9
}
