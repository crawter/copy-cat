package render

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"copycat/internal/campaign"
	"copycat/internal/font"
	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

const testDT = 1.0 / 60.0

func TestMenuLinesEverySceneYieldsLines(t *testing.T) {
	cases := []struct {
		name string
		flow func(t *testing.T) *campaign.Flow
	}{
		{
			name: "language pick",
			flow: func(t *testing.T) *campaign.Flow {
				return campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), true)
			},
		},
		{
			name: "main menu",
			flow: func(t *testing.T) *campaign.Flow {
				return campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
			},
		},
		{
			name: "new game confirmation",
			flow: func(t *testing.T) *campaign.Flow {
				f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
				f.Handle(campaign.ActionDown) // New Game
				f.Handle(campaign.ActionConfirm)
				return f
			},
		},
		{
			name: "map",
			flow: func(t *testing.T) *campaign.Flow {
				return mapFlow(t)
			},
		},
		{
			name: "settings",
			flow: func(t *testing.T) *campaign.Flow {
				return settingsFlow(t)
			},
		},
		{
			name: "pause",
			flow: func(t *testing.T) *campaign.Flow {
				f := newGameFlow(t)
				f.Handle(campaign.ActionPause)
				return f
			},
		},
		{
			name: "level end",
			flow: func(t *testing.T) *campaign.Flow {
				f := newGameFlow(t)
				stepToScene(t, f, campaign.SceneLevelEnd)
				return f
			},
		},
		{
			name: "game over",
			flow: func(t *testing.T) *campaign.Flow {
				return gameOverFlow(t)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := MenuLines(tc.flow(t))
			if len(lines) == 0 {
				t.Fatal("MenuLines returned no lines")
			}
			for _, l := range lines {
				if l.Text == "" {
					t.Fatal("MenuLines returned an empty line")
				}
			}
		})
	}
}

func TestMenuLinesLanguageDiffers(t *testing.T) {
	c := campaign.Default()
	en := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), false)
	ua := campaign.NewFlow(c, ukrainianSettings(), campaign.NewProgress(), false)

	if got, want := joinLines(MenuLines(en)), joinLines(MenuLines(ua)); got == want {
		t.Fatalf("EN and UA main menu lines are identical: %q", got)
	}
}

func TestHighlightMatchesMenuIndex(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), true)
	if got, want := highlighted(t, MenuLines(f)).Text, textFor(f, campaign.LanguageEnglishName); got != want {
		t.Fatalf("language pick highlight = %q, want %q", got, want)
	}
	f.Handle(campaign.ActionDown)
	if got, want := highlighted(t, MenuLines(f)).Text, textFor(f, campaign.LanguageUkrainianName); got != want {
		t.Fatalf("language pick highlight after down = %q, want %q", got, want)
	}

	m := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	if got, want := highlighted(t, MenuLines(m)).Text, textFor(m, campaign.MainMenuItems[0]); got != want {
		t.Fatalf("main menu highlight = %q, want %q", got, want)
	}
	m.Handle(campaign.ActionDown)
	if got, want := highlighted(t, MenuLines(m)).Text, textFor(m, campaign.MainMenuItems[1]); got != want {
		t.Fatalf("main menu highlight after down = %q, want %q", got, want)
	}

	s := settingsFlow(t)
	if got := highlighted(t, MenuLines(s)).Text; !strings.Contains(got, textFor(s, campaign.SettingsItems[0])) {
		t.Fatalf("settings highlight = %q, missing %q", got, textFor(s, campaign.SettingsItems[0]))
	}
	s.Handle(campaign.ActionDown)
	if got := highlighted(t, MenuLines(s)).Text; !strings.Contains(got, textFor(s, campaign.SettingsItems[1])) {
		t.Fatalf("settings highlight after down = %q, missing %q", got, textFor(s, campaign.SettingsItems[1]))
	}

	mp := mapFlow(t)
	if got := highlighted(t, MenuLines(mp)).Text; !strings.Contains(got, textFor(mp, campaign.LevelNames[0])) {
		t.Fatalf("map highlight = %q, missing %q", got, textFor(mp, campaign.LevelNames[0]))
	}
	mp.Handle(campaign.ActionDown)
	if got := highlighted(t, MenuLines(mp)).Text; !strings.Contains(got, textFor(mp, campaign.LevelNames[1])) {
		t.Fatalf("map highlight after down = %q, missing %q", got, textFor(mp, campaign.LevelNames[1]))
	}

	g := gameOverFlow(t)
	if got, want := highlighted(t, MenuLines(g)).Text, textFor(g, campaign.GameOverItems[0]); got != want {
		t.Fatalf("game over highlight = %q, want %q", got, want)
	}
	g.Handle(campaign.ActionDown)
	if got, want := highlighted(t, MenuLines(g)).Text, textFor(g, campaign.GameOverItems[1]); got != want {
		t.Fatalf("game over highlight after down = %q, want %q", got, want)
	}
}

func TestGameOverCarriesRightEpitaph(t *testing.T) {
	f := gameOverFlow(t)
	if got := f.Run().Cause; got != gameplay.CausePit {
		t.Fatalf("run cause = %v, want CausePit", got)
	}
	want := textFor(f, campaign.Epitaphs[gameplay.CausePit])
	if !containsLine(MenuLines(f), want) {
		t.Fatalf("game over lines missing epitaph %q", want)
	}
}

func TestMapListsTenLevelsBossLast(t *testing.T) {
	f := mapFlow(t)
	lines := MenuLines(f)
	if got, want := len(lines), 1+len(campaign.LevelNames); got != want {
		t.Fatalf("map lines = %d, want %d", got, want)
	}
	for i, name := range campaign.LevelNames {
		if got := lines[1+i].Text; !strings.Contains(got, textFor(f, name)) {
			t.Fatalf("map line %d = %q, missing %q", i, got, textFor(f, name))
		}
	}
	bossName := textFor(f, campaign.LevelNames[len(campaign.LevelNames)-1])
	if got := lines[len(lines)-1].Text; !strings.Contains(got, bossName) {
		t.Fatalf("boss is not the last map entry: %q", got)
	}
}

func TestHUDCoversLivesCombsShield(t *testing.T) {
	f := newGameFlow(t)
	text := joinLines(HUDLines(f))
	for _, want := range []string{
		hudText(f, gameplay.HUDLives),
		hudText(f, gameplay.HUDCombs),
		hudText(f, gameplay.HUDShield),
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("HUD missing %q in %q", want, text)
		}
	}
}

func TestHUDBossCoversLivesAndPhase(t *testing.T) {
	f := campaign.NewFlow(flatCampaign(), campaign.DefaultSettings(), campaign.Progress{Unlocked: 10}, false)
	f.StartBoss()
	if f.Scene() != campaign.ScenePlay || f.Boss() == nil {
		t.Fatalf("boss not started: scene=%v boss=%v", f.Scene(), f.Boss())
	}

	text := joinLines(HUDLines(f))
	for _, want := range []string{
		hudText(f, gameplay.HUDLives),
		textFor(f, uiPhase),
		"1/3",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("boss HUD missing %q in %q", want, text)
		}
	}
}

func TestHUDCampaignShowsProgressEndlessShowsDistance(t *testing.T) {
	f := newGameFlow(t)
	text := joinLines(HUDLines(f))
	if !strings.Contains(text, textFor(f, uiLevel)) {
		t.Fatalf("campaign HUD missing level progress: %q", text)
	}
	if strings.Contains(text, hudText(f, gameplay.HUDDistance)) {
		t.Fatalf("campaign HUD shows distance: %q", text)
	}

	e := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	e.StartEndless(7)
	etext := joinLines(HUDLines(e))
	if !strings.Contains(etext, hudText(e, gameplay.HUDDistance)) {
		t.Fatalf("endless HUD missing distance: %q", etext)
	}
	if strings.Contains(etext, textFor(e, uiLevel)) {
		t.Fatalf("endless HUD shows level progress: %q", etext)
	}
}

func TestStoryEndingCreditsNonEmpty(t *testing.T) {
	s := storyFlow(t)
	if lines := StoryLines(s); len(lines) == 0 || lines[0].Text == "" {
		t.Fatal("story lines are empty")
	}

	e := endingFlow(t)
	if lines := StoryLines(e); len(lines) == 0 || lines[0].Text == "" {
		t.Fatal("ending lines are empty")
	}

	c := creditsFlow(t)
	lines := StoryLines(c)
	if len(lines) < len(campaign.Credits)+1 {
		t.Fatalf("credits lines = %d, want at least %d", len(lines), len(campaign.Credits)+1)
	}
	for _, l := range lines {
		if l.Text == "" {
			t.Fatal("credits contains an empty line")
		}
	}
}

func TestStoryLinesWrapLongParagraph(t *testing.T) {
	f := storyFlow(t)
	lines := StoryLines(f)
	if len(lines) <= 1 {
		t.Fatalf("story paragraph was not wrapped: %d lines", len(lines))
	}
	for i, l := range lines {
		if n := utf8.RuneCountInString(l.Text); n > storyWrapChars {
			t.Fatalf("line %d has %d runes, exceeds %d: %q", i, n, storyWrapChars, l.Text)
		}
	}
}

func TestMainMenuLogoAndTagline(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	lines := MenuLines(f)
	if len(lines) < 2+len(campaign.MainMenuItems) {
		t.Fatalf("main menu lines = %d, want at least %d", len(lines), 2+len(campaign.MainMenuItems))
	}
	if got, want := lines[0].Text, textFor(f, campaign.GameTitle); got != want {
		t.Fatalf("logo text = %q, want %q", got, want)
	}
	if got, want := lines[1].Text, textFor(f, campaign.TitleTagline); got != want {
		t.Fatalf("tagline text = %q, want %q", got, want)
	}
	if lines[0].Scale <= 1 {
		t.Fatalf("logo scale = %d, want > 1", lines[0].Scale)
	}
	if lines[1].Scale <= 1 || lines[1].Scale >= lines[0].Scale {
		t.Fatalf("tagline scale = %d, want between 1 and logo scale %d", lines[1].Scale, lines[0].Scale)
	}
	if lines[1].Y <= lines[0].Y {
		t.Fatalf("tagline y = %d, must be below logo y = %d", lines[1].Y, lines[0].Y)
	}
	if got, want := lines[2].Text, textFor(f, campaign.MainMenuItems[0]); got != want {
		t.Fatalf("first menu item = %q, want %q", got, want)
	}
	if lines[2].Y <= lines[1].Y {
		t.Fatalf("first menu item y = %d, must be below tagline y = %d", lines[2].Y, lines[1].Y)
	}
}

func TestMainMenuScaledLayoutKeepsMenuBelowLogo(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	lines := MenuLines(f)
	tagline := lines[1]
	taglineBottom := tagline.Y + uiLineGap*tagline.Scale
	for _, l := range lines[2:] {
		if l.Y < taglineBottom {
			t.Fatalf("menu item %q y = %d, above tagline bottom %d", l.Text, l.Y, taglineBottom)
		}
	}
}

func TestLevelIntroLines(t *testing.T) {
	f := newGameFlow(t)
	lines := LevelIntroLines(f)
	if len(lines) != 2 {
		t.Fatalf("level intro lines = %d, want 2", len(lines))
	}
	lv, ok := f.CurrentLevel()
	if !ok {
		t.Fatal("no current level")
	}
	if got, want := lines[0].Text, fmt.Sprintf("%s %d", textFor(f, uiLevel), lv.ID); got != want {
		t.Fatalf("level number line = %q, want %q", got, want)
	}
	if got, want := lines[1].Text, textFor(f, campaign.LevelNames[lv.ID-1]); got != want {
		t.Fatalf("level name line = %q, want %q", got, want)
	}
}

func TestLevelIntroEndless(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.StartEndless(7)
	lines := LevelIntroLines(f)
	if len(lines) != 1 {
		t.Fatalf("endless intro lines = %d, want 1", len(lines))
	}
	if got, want := lines[0].Text, textFor(f, uiEndless); got != want {
		t.Fatalf("endless intro = %q, want %q", got, want)
	}
}

func TestLevelEndShowsStats(t *testing.T) {
	f := newGameFlow(t)
	stepToScene(t, f, campaign.SceneLevelEnd)
	r := f.Run()
	if r == nil {
		t.Fatal("no run at level end")
	}
	r.TotalCombs = 7
	r.LivesLost = 2
	r.Stomps = 3

	text := joinLines(MenuLines(f))
	for _, want := range []string{
		textFor(f, uiTotalCombs),
		textFor(f, uiLivesLost),
		textFor(f, uiStomps),
		hudText(f, gameplay.HUDDistance),
		"7", "2", "3",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("level end missing %q in %q", want, text)
		}
	}
}

func TestLandingBurstFiresOnLandingNotOnStomp(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	pos := Vec3{}

	// A hero landing: an empty level lets a jump arc start and end cleanly.
	flat := levelgen.New(1, levelgen.Params{BaseSpeed: 8, Length: 30})
	run := gameplay.NewRun(runner.New(flat, runner.DefaultConfig()), flat, 3)
	u := NewUI(font.New())

	u.updateRunOverlays(f, run, pos) // grounded first frame
	if hasBurst(u, BuildLandingBurst()) {
		t.Fatal("landing burst spawned on the first grounded frame")
	}

	run.Step(runner.Input{Jump: true}, testDT)
	u.updateRunOverlays(f, run, pos)
	if !run.Runner.Airborne() {
		t.Fatal("hero did not become airborne")
	}
	if hasBurst(u, BuildLandingBurst()) {
		t.Fatal("landing burst spawned on takeoff")
	}

	for i := 0; i < 200 && run.Runner.Airborne(); i++ {
		run.Step(runner.Input{}, testDT)
		u.updateRunOverlays(f, run, pos)
	}
	if run.Runner.Airborne() {
		t.Fatal("hero never landed")
	}
	if !hasBurst(u, BuildLandingBurst()) {
		t.Fatal("landing burst not spawned on hero landing")
	}

	// An enemy stomp keeps the hero airborne, so it must not fire the landing
	// burst. Search for a deterministic stomp that stays airborne.
	assertStompDoesNotFireLanding(t, f)
}

// assertStompDoesNotFireLanding finds a level where a jump lands on an enemy
// while the hero is still airborne, and pins that updateRunOverlays does not
// treat the stomp as a landing.
func assertStompDoesNotFireLanding(t *testing.T, f *campaign.Flow) {
	t.Helper()
	for seed := uint64(1); seed <= 100; seed++ {
		l := levelgen.New(seed, levelgen.Params{BaseSpeed: 8, Length: 40, Density: 0.6})
		for _, e := range l.Elements(8, 28) {
			if e.Kind != levelgen.Enemy || e.Lane != 1 {
				continue
			}
			for jumpAt := e.Z - 4.0; jumpAt <= e.Z-1.0; jumpAt += 0.25 {
				u := NewUI(font.New())
				run := gameplay.NewRun(runner.New(l, runner.DefaultConfig()), l, 3)
				u.updateRunOverlays(f, run, Vec3{})

				for i := 0; i < 400 && run.State == gameplay.StateRunning && run.Runner.Z() < jumpAt; i++ {
					run.Step(runner.Input{}, testDT)
					u.updateRunOverlays(f, run, Vec3{})
				}
				if run.State != gameplay.StateRunning {
					continue
				}

				run.Step(runner.Input{Jump: true}, testDT)
				u.updateRunOverlays(f, run, Vec3{})
				if !run.Runner.Airborne() {
					continue
				}
				if hasBurst(u, BuildLandingBurst()) {
					t.Fatal("landing burst spawned on takeoff")
				}

				for i := 0; i < 200 && run.State == gameplay.StateRunning; i++ {
					out := run.Step(runner.Input{}, testDT)
					u.updateRunOverlays(f, run, Vec3{})
					if !hasOutcome(out, gameplay.OutcomeStomp) {
						if !run.Runner.Airborne() {
							break
						}
						continue
					}
					if !run.Runner.Airborne() {
						break // stomp and landing coincided; keep searching
					}
					if hasBurst(u, BuildLandingBurst()) {
						t.Fatal("landing burst fired at an enemy stomp")
					}
					return
				}
			}
		}
	}
	t.Fatal("no airborne enemy stomp found within 100 seeds")
}

func TestPauseLinesEntriesAndHighlight(t *testing.T) {
	f := newGameFlow(t)
	f.Handle(campaign.ActionPause)
	if f.Scene() != campaign.ScenePause {
		t.Fatalf("scene = %v, want ScenePause", f.Scene())
	}

	lines := MenuLines(f)
	if len(lines) < 2+len(campaign.PauseItems) {
		t.Fatalf("pause lines = %d, want at least %d", len(lines), 2+len(campaign.PauseItems))
	}
	if got, want := highlighted(t, lines).Text, textFor(f, campaign.PauseItems[0]); got != want {
		t.Fatalf("pause highlight = %q, want %q", got, want)
	}
	for i, item := range campaign.PauseItems {
		if got, want := lines[2+i].Text, textFor(f, item); got != want {
			t.Fatalf("pause item %d = %q, want %q", i, got, want)
		}
	}

	f.Handle(campaign.ActionDown)
	if got, want := highlighted(t, MenuLines(f)).Text, textFor(f, campaign.PauseItems[1]); got != want {
		t.Fatalf("pause highlight after down = %q, want %q", got, want)
	}
	f.Handle(campaign.ActionDown)
	if got, want := highlighted(t, MenuLines(f)).Text, textFor(f, campaign.PauseItems[2]); got != want {
		t.Fatalf("pause highlight after two downs = %q, want %q", got, want)
	}
}

func TestNewGameConfirmLines(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.Handle(campaign.ActionDown) // New Game
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneNewGameConfirm {
		t.Fatalf("scene = %v, want SceneNewGameConfirm", f.Scene())
	}

	lines := MenuLines(f)
	if got, want := len(lines), 1+len(campaign.NewGameConfirmItems); got != want {
		t.Fatalf("confirm lines = %d, want %d", got, want)
	}
	if got, want := lines[0].Text, textFor(f, campaign.NewGameConfirmTitle); got != want {
		t.Fatalf("confirm title = %q, want %q", got, want)
	}
	for i, item := range campaign.NewGameConfirmItems {
		if got, want := lines[1+i].Text, textFor(f, item); got != want {
			t.Fatalf("confirm item %d = %q, want %q", i, got, want)
		}
	}
	if got, want := highlighted(t, lines).Text, textFor(f, campaign.NewGameConfirmItems[f.MenuIndex()]); got != want {
		t.Fatalf("confirm highlight = %q, want %q", got, want)
	}
}

func TestSettingsShowVolumePercentages(t *testing.T) {
	f := settingsFlow(t)
	text := joinLines(MenuLines(f))
	if !strings.Contains(text, "100%") {
		t.Fatalf("settings lines missing 100%%: %q", text)
	}

	f.Handle(campaign.ActionLeft) // music 100% -> 90%
	if got := joinLines(MenuLines(f)); !strings.Contains(got, "90%") {
		t.Fatalf("settings lines missing 90%% after music down: %q", got)
	}
}

func TestVolumeTextPercentage(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{
		{1, "100%"},
		{0.8, "80%"},
		{0.05, "5%"},
		{0, "0%"},
	} {
		if got := volumeText(tc.in); got != tc.want {
			t.Errorf("volumeText(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestVolumeBarGeometry(t *testing.T) {
	bar := volumeBar(18, 0.8)
	if bar.X != uiMarginX+uiIndent {
		t.Errorf("bar X = %d, want %d", bar.X, uiMarginX+uiIndent)
	}
	if bar.Y != 18+uiLineGap {
		t.Errorf("bar Y = %d, want %d", bar.Y, 18+uiLineGap)
	}
	if bar.Width != volumeBarWidth || bar.Height != volumeBarHeight {
		t.Errorf("bar size = %dx%d, want %dx%d", bar.Width, bar.Height, volumeBarWidth, volumeBarHeight)
	}
	if bar.Fill != 0.8 {
		t.Errorf("bar Fill = %v, want 0.8", bar.Fill)
	}
	if got := volumeBar(0, -1).Fill; got != 0 {
		t.Errorf("negative fraction clamped to %v, want 0", got)
	}
	if got := volumeBar(0, 2).Fill; got != 1 {
		t.Errorf("oversized fraction clamped to %v, want 1", got)
	}
}

func TestLogoOffsetChangesOverTime(t *testing.T) {
	first := logoOffset(0)
	changed := false
	for step := 1; step <= 200; step++ {
		if logoOffset(float64(step)*0.05) != first {
			changed = true
			break
		}
	}
	if !changed {
		t.Fatalf("logoOffset did not change over time (first = %d)", first)
	}

	amplitude := int(logoBobAmplitude)
	for _, clock := range []float64{0, 0.1, 0.5, 1, 3, 10} {
		off := logoOffset(clock)
		if off < -amplitude || off > amplitude {
			t.Fatalf("logoOffset(%v) = %d outside ±%d", clock, off, amplitude)
		}
	}
}

// --- helpers ---

func ukrainianSettings() campaign.Settings {
	s := campaign.DefaultSettings()
	s.Language = campaign.LanguageUkrainian
	return s
}

func flatCampaign() *campaign.Campaign {
	c := campaign.Default()
	p := levelgen.Params{BaseSpeed: 8, Length: 60}
	for i := range c.Levels {
		if c.Levels[i].Kind == campaign.LevelRun {
			c.Levels[i].Params = p
		}
	}
	return c
}

func newGameFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := campaign.NewFlow(flatCampaign(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.Handle(campaign.ActionDown) // New Game
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneNewGameConfirm {
		t.Fatalf("scene = %v, want SceneNewGameConfirm", f.Scene())
	}
	f.Handle(campaign.ActionUp) // Yes
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneStory {
		t.Fatalf("scene = %v, want SceneStory", f.Scene())
	}
	f.Handle(campaign.ActionConfirm) // StoryIntro → level 1
	if f.Scene() != campaign.ScenePlay {
		t.Fatalf("scene = %v, want ScenePlay", f.Scene())
	}
	return f
}

func settingsFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	for i := 0; i < 4; i++ {
		f.Handle(campaign.ActionDown)
	}
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneSettings {
		t.Fatalf("scene = %v, want SceneSettings", f.Scene())
	}
	return f
}

func mapFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.Handle(campaign.ActionDown)
	f.Handle(campaign.ActionDown)
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneMap {
		t.Fatalf("scene = %v, want SceneMap", f.Scene())
	}
	return f
}

func gameOverFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	c, pit := pitCampaign(t)
	f := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.Handle(campaign.ActionDown) // New Game
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneNewGameConfirm {
		t.Fatalf("scene = %v, want SceneNewGameConfirm", f.Scene())
	}
	f.Handle(campaign.ActionUp) // Yes
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneStory {
		t.Fatalf("scene = %v, want SceneStory", f.Scene())
	}
	f.Handle(campaign.ActionConfirm) // StoryIntro → level 1
	if f.Scene() != campaign.ScenePlay {
		t.Fatalf("scene = %v, want ScenePlay", f.Scene())
	}
	r := f.Run()
	r.Runner.Restart(pit.Z - 4)
	r.Lives = 1
	stepToScene(t, f, campaign.SceneGameOver)
	return f
}

func storyFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := campaign.NewFlow(flatCampaign(), campaign.DefaultSettings(), campaign.Progress{Unlocked: 3}, false)
	f.StartLevel(3)
	stepToScene(t, f, campaign.SceneLevelEnd)
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneStory {
		t.Fatalf("scene = %v, want SceneStory", f.Scene())
	}
	return f
}

func endingFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := campaign.NewFlow(flatCampaign(), campaign.DefaultSettings(), campaign.Progress{Unlocked: 10}, false)
	f.StartBoss()
	winBoss(t, f)
	if f.Scene() != campaign.SceneEnding {
		t.Fatalf("scene = %v, want SceneEnding", f.Scene())
	}
	return f
}

func creditsFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := endingFlow(t)
	f.Handle(campaign.ActionConfirm)
	if f.Scene() != campaign.SceneCredits {
		t.Fatalf("scene = %v, want SceneCredits", f.Scene())
	}
	return f
}

func pitCampaign(t *testing.T) (*campaign.Campaign, levelgen.Element) {
	t.Helper()
	p := levelgen.Params{BaseSpeed: 8, Length: 120, Density: 0.4}
	for seed := uint64(0); seed < 200; seed++ {
		l := levelgen.New(seed, p)
		if pit, ok := firstPit(l); ok {
			c := flatCampaign()
			c.Levels[0] = campaign.LevelDef{ID: 1, World: 1, Index: 1, Kind: campaign.LevelRun, Seed: seed, Params: p}
			return c, pit
		}
	}
	t.Fatal("no pit found within 200 seeds")
	return nil, levelgen.Element{}
}

func firstPit(l *levelgen.Level) (levelgen.Element, bool) {
	for _, e := range l.Elements(0, l.Length()) {
		if e.Kind == levelgen.Pit {
			return e, true
		}
	}
	return levelgen.Element{}, false
}

func stepToScene(t *testing.T, f *campaign.Flow, want campaign.Scene) {
	t.Helper()
	for i := 0; i < 5000 && f.Scene() != want; i++ {
		f.StepRun(runner.Input{}, testDT)
	}
	if f.Scene() != want {
		t.Fatalf("scene = %v, want %v", f.Scene(), want)
	}
}

func winBoss(t *testing.T, f *campaign.Flow) {
	t.Helper()
	b := f.Boss()
	if b == nil {
		t.Fatal("no boss fight to win")
	}
	for i := 0; i < 4000 && b.State() == gameplay.BossFighting; i++ {
		f.StepRun(bossWinningInput(b), testDT)
	}
	if b.State() != gameplay.BossWon {
		t.Fatalf("boss state = %v, want BossWon", b.State())
	}
}

func bossWinningInput(b *gameplay.BossFight) runner.Input {
	if b.Vulnerable() {
		var in runner.Input
		switch {
		case b.HeroLane() < b.AttackLane():
			in.Right = true
		case b.HeroLane() > b.AttackLane():
			in.Left = true
		default:
			in.Jump = true
		}
		return in
	}
	return bossDodgeInput(b)
}

func bossDodgeInput(b *gameplay.BossFight) runner.Input {
	var in runner.Input
	switch {
	case b.HeroLane() < b.AttackLane():
		in.Right = true
	case b.HeroLane() > b.AttackLane():
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

func highlighted(t *testing.T, lines []ScreenLine) ScreenLine {
	t.Helper()
	var out []ScreenLine
	for _, l := range lines {
		if l.Highlighted {
			out = append(out, l)
		}
	}
	if len(out) != 1 {
		t.Fatalf("want exactly one highlighted line, got %d", len(out))
	}
	return out[0]
}

func joinLines(lines []ScreenLine) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

func containsLine(lines []ScreenLine, text string) bool {
	for _, l := range lines {
		if l.Text == text {
			return true
		}
	}
	return false
}

func hasBurst(u *UI, want []Cmd) bool {
	for _, b := range u.bursts {
		if reflect.DeepEqual(b.cmds, want) {
			return true
		}
	}
	return false
}

func hasOutcome(out []gameplay.Outcome, want gameplay.Outcome) bool {
	for _, o := range out {
		if o == want {
			return true
		}
	}
	return false
}
