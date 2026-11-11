package main

import (
	"testing"

	"copycat/internal/audio"
	"copycat/internal/campaign"
	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

// flowDT is the fixed per-frame step used by the flow tests.
const flowDT = 1.0 / 60.0

func TestSlideSoundSelection(t *testing.T) {
	tests := []struct {
		name     string
		airborne bool
		want     audio.Sound
	}{
		{name: "grounded slide", want: audio.SoundSlide},
		{name: "airborne slam", airborne: true, want: audio.SoundSlam},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slideSound(tt.airborne); got != tt.want {
				t.Fatalf("slideSound(%v) = %v, want %v", tt.airborne, got, tt.want)
			}
		})
	}
}

func TestIsMenuScene(t *testing.T) {
	tests := []struct {
		name  string
		scene campaign.Scene
		want  bool
	}{
		{name: "language pick", scene: campaign.SceneLanguagePick, want: true},
		{name: "main menu", scene: campaign.SceneMainMenu, want: true},
		{name: "map", scene: campaign.SceneMap, want: true},
		{name: "settings", scene: campaign.SceneSettings, want: true},
		{name: "game over", scene: campaign.SceneGameOver, want: true},
		{name: "play", scene: campaign.ScenePlay, want: false},
		{name: "pause", scene: campaign.ScenePause, want: false},
		{name: "level end", scene: campaign.SceneLevelEnd, want: false},
		{name: "story", scene: campaign.SceneStory, want: false},
		{name: "ending", scene: campaign.SceneEnding, want: false},
		{name: "credits", scene: campaign.SceneCredits, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMenuScene(tt.scene); got != tt.want {
				t.Fatalf("isMenuScene(%v) = %v, want %v", tt.scene, got, tt.want)
			}
		})
	}
}

func TestSlideOnContainerMapsToSlide(t *testing.T) {
	r := runnerOnContainer(t)
	if airborne(r, nil) {
		t.Fatal("hero riding a container ramp must be grounded, not airborne")
	}
	if got := slideSound(airborne(r, nil)); got != audio.SoundSlide {
		t.Fatalf("slide on container = %v, want SoundSlide", got)
	}
}

func TestOutcomeSounds(t *testing.T) {
	tests := []struct {
		name string
		in   gameplay.Outcome
		want []audio.Sound
	}{
		{name: "stumble", in: gameplay.OutcomeStumble, want: []audio.Sound{audio.SoundHit}},
		{name: "snip", in: gameplay.OutcomeSnip, want: []audio.Sound{audio.SoundSnip, audio.SoundMeow}},
		{name: "pit", in: gameplay.OutcomePit, want: []audio.Sound{audio.SoundFall, audio.SoundHonk}},
		{name: "stomp", in: gameplay.OutcomeStomp, want: []audio.Sound{audio.SoundStomp}},
		{name: "finished", in: gameplay.OutcomeFinished, want: []audio.Sound{audio.SoundFanfare}},
		{name: "died", in: gameplay.OutcomeDied, want: []audio.Sound{audio.SoundGameOver}},
		{name: "shield saved", in: gameplay.OutcomeShieldSaved, want: []audio.Sound{audio.SoundHairspray}},
		{name: "comb", in: gameplay.OutcomeComb, want: []audio.Sound{audio.SoundComb}},
		{name: "hairspray", in: gameplay.OutcomeHairspray, want: []audio.Sound{audio.SoundHairspray}},
		{name: "checkpoint", in: gameplay.OutcomeCheckpoint, want: []audio.Sound{audio.SoundCheckpoint}},
		{name: "life", in: gameplay.OutcomeLife, want: []audio.Sound{audio.SoundExtraLife}},
		{name: "unknown", in: gameplay.Outcome(99), want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outcomeSounds(tt.in)
			assertSounds(t, got, tt.want)
		})
	}
}

func TestBossTransitionSounds(t *testing.T) {
	t.Run("hit", func(t *testing.T) {
		hit := gameplay.NewBossFight(3)
		// Three half-second steps drive the first two phase-1 attacks; the
		// second lands on the center lane where the hero stands, so the fight
		// takes one hit: three lives drop to two, still in phase 1.
		for i := 0; i < 3; i++ {
			hit.Step(runner.Input{}, 0.5)
		}
		if hit.Lives() != 2 || hit.Phase() != 1 {
			t.Fatalf("scripted hit = lives %d phase %d, want lives 2 phase 1", hit.Lives(), hit.Phase())
		}
		got := bossTransitionSounds(gameplay.NewBossFight(3), hit)
		assertSounds(t, got, []audio.Sound{audio.SoundBossHit})
	})

	t.Run("phase up", func(t *testing.T) {
		phase := gameplay.NewBossFight(3)
		// Drive the fight through its two phase-1 attacks: it takes one hit
		// and opens a vulnerable window. The window stays open into the next
		// step, where moving to the crouch lane and jumping stomps Copycat
		// from phase 1 to phase 2.
		for i := 0; i < 3; i++ {
			phase.Step(runner.Input{}, 0.5)
		}
		phase.Step(runner.Input{Right: true, Jump: true}, 0.5)
		if phase.Phase() != 2 {
			t.Fatalf("Phase = %d after scripted stomp, want 2", phase.Phase())
		}
		if phase.Lives() != 2 {
			t.Fatalf("Lives = %d after scripted stomp, want 2", phase.Lives())
		}
		got := bossTransitionSounds(gameplay.NewBossFight(2), phase)
		assertSounds(t, got, []audio.Sound{audio.SoundMeow})
	})
}

func TestTrackForScene(t *testing.T) {
	c := campaign.Default()
	tests := []struct {
		name  string
		setup func(*campaign.Flow)
		want  audio.Track
	}{
		{
			name:  "main menu",
			setup: func(f *campaign.Flow) {},
			want:  audio.TrackMenu,
		},
		{
			name:  "mangroves run",
			setup: func(f *campaign.Flow) { f.StartLevel(1) },
			want:  audio.TrackMangroves,
		},
		{
			name:  "harbor run",
			setup: func(f *campaign.Flow) { f.StartLevel(4) },
			want:  audio.TrackHarbor,
		},
		{
			name:  "neon run",
			setup: func(f *campaign.Flow) { f.StartLevel(7) },
			want:  audio.TrackNeon,
		},
		{
			name:  "endless run",
			setup: func(f *campaign.Flow) { f.StartEndless(42) },
			want:  audio.TrackNeon,
		},
		{
			name:  "boss fight",
			setup: func(f *campaign.Flow) { f.StartBoss() },
			want:  audio.TrackBoss,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.Progress{Unlocked: 10}, false)
			tt.setup(f)
			if got := trackForScene(f); got != tt.want {
				t.Fatalf("trackForScene(%v) = %v, want %v", f.Scene(), got, tt.want)
			}
		})
	}

	t.Run("ending and credits", func(t *testing.T) {
		f := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.Progress{Unlocked: 10}, false)
		f.StartBoss()
		winBossFlow(t, f)
		if f.Scene() != campaign.SceneEnding {
			t.Fatalf("Scene = %v after boss win, want SceneEnding", f.Scene())
		}
		if got := trackForScene(f); got != audio.TrackEnding {
			t.Fatalf("trackForScene(SceneEnding) = %v, want TrackEnding", got)
		}

		f.Handle(campaign.ActionConfirm)
		if f.Scene() != campaign.SceneCredits {
			t.Fatalf("Scene = %v after ending confirm, want SceneCredits", f.Scene())
		}
		if got := trackForScene(f); got != audio.TrackEnding {
			t.Fatalf("trackForScene(SceneCredits) = %v, want TrackEnding", got)
		}
	})
}

// winBossFlow drives the active boss fight with a dodge-and-stomp script until
// it wins and the flow reaches the ending card.
func winBossFlow(t *testing.T, f *campaign.Flow) {
	t.Helper()
	b := f.Boss()
	if b == nil {
		t.Fatalf("no boss fight to win")
	}
	for i := 0; i < 4000 && b.State() == gameplay.BossFighting; i++ {
		f.StepRun(bossWinningInput(b), flowDT)
	}
	if b.State() != gameplay.BossWon {
		t.Fatalf("boss state = %v, want BossWon", b.State())
	}
}

// bossWinningInput reacts to the fight: dodge attacks, and stomp Copycat when
// he is vulnerable.
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

// assertSounds fails when got differs from want in order or length.
func assertSounds(t *testing.T, got, want []audio.Sound) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("sounds = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("sounds = %v, want %v", got, want)
		}
	}
}

// runnerOnContainer steps a runner onto a center-lane container ramp and
// returns it, using the first seed whose mangroves level has a reachable one.
func runnerOnContainer(t *testing.T) *runner.Runner {
	t.Helper()
	for seed := uint64(1); seed <= 500; seed++ {
		l := levelgen.New(seed, levelgen.Mangroves())
		ramp, ok := reachableCenterRamp(l)
		if !ok {
			continue
		}
		r := runner.New(l, runner.DefaultConfig())
		for step := 0; step < 6000; step++ {
			r.Step(runner.Input{}, 1.0/60.0)
			if r.OnContainer() {
				return r
			}
			if r.Z() > ramp.Z+ramp.Span+1 {
				break
			}
		}
	}
	t.Fatal("no seed produced a reachable center-lane ramp")
	return nil
}

// reachableCenterRamp returns the first center-lane ramp reachable by running
// straight from the start: no pit or center-lane forced obstacle before it.
func reachableCenterRamp(l *levelgen.Level) (levelgen.Element, bool) {
	var ramp levelgen.Element
	found := false
	for _, e := range l.Elements(0, l.Length()) {
		if e.Kind == levelgen.Ramp && e.Lane == 1 {
			ramp = e
			found = true
			break
		}
	}
	if !found {
		return levelgen.Element{}, false
	}
	for _, e := range l.Elements(0, ramp.Z) {
		if e.Kind == levelgen.Pit {
			return levelgen.Element{}, false
		}
		switch e.Kind {
		case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
			if e.Lane == 1 {
				return levelgen.Element{}, false
			}
		}
	}
	return ramp, true
}

func TestIsMenuScene(t *testing.T) {
	c := campaign.Default()

	languagePick := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), true)
	if !isMenuScene(languagePick) {
		t.Fatal("language pick should be a menu scene")
	}

	menu := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), false)
	if !isMenuScene(menu) {
		t.Fatal("main menu should be a menu scene")
	}

	menu.Handle(campaign.ActionDown)
	menu.Handle(campaign.ActionDown)
	menu.Handle(campaign.ActionConfirm) // map
	if !isMenuScene(menu) {
		t.Fatal("map should be a menu scene")
	}

	menu.Handle(campaign.ActionBack) // back to main menu
	menu.Handle(campaign.ActionDown)
	menu.Handle(campaign.ActionDown)
	menu.Handle(campaign.ActionDown)
	menu.Handle(campaign.ActionDown)
	menu.Handle(campaign.ActionConfirm) // settings
	if !isMenuScene(menu) {
		t.Fatal("settings should be a menu scene")
	}

	play := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), false)
	play.StartLevel(1)
	if isMenuScene(play) {
		t.Fatal("play should not be a menu scene")
	}
	play.Handle(campaign.ActionPause)
	if isMenuScene(play) {
		t.Fatal("pause should not be a menu scene")
	}
}

func TestOutcomeSounds(t *testing.T) {
	tests := []struct {
		name    string
		outcome gameplay.Outcome
		want    []audio.Sound
	}{
		{name: "stumble", outcome: gameplay.OutcomeStumble, want: []audio.Sound{audio.SoundHit}},
		{name: "snip", outcome: gameplay.OutcomeSnip, want: []audio.Sound{audio.SoundSnip, audio.SoundMeow}},
		{name: "pit", outcome: gameplay.OutcomePit, want: []audio.Sound{audio.SoundFall, audio.SoundHonk}},
		{name: "stomp", outcome: gameplay.OutcomeStomp, want: []audio.Sound{audio.SoundStomp}},
		{name: "finished", outcome: gameplay.OutcomeFinished, want: []audio.Sound{audio.SoundFanfare}},
		{name: "died", outcome: gameplay.OutcomeDied, want: []audio.Sound{audio.SoundGameOver}},
		{name: "shield saved", outcome: gameplay.OutcomeShieldSaved, want: []audio.Sound{audio.SoundHairspray}},
		{name: "comb", outcome: gameplay.OutcomeComb, want: []audio.Sound{audio.SoundComb}},
		{name: "hairspray", outcome: gameplay.OutcomeHairspray, want: []audio.Sound{audio.SoundHairspray}},
		{name: "checkpoint", outcome: gameplay.OutcomeCheckpoint, want: []audio.Sound{audio.SoundCheckpoint}},
		{name: "life", outcome: gameplay.OutcomeLife, want: []audio.Sound{audio.SoundExtraLife}},
		{name: "unknown", outcome: gameplay.Outcome(999), want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSounds(t, tt.want, outcomeSounds(tt.outcome))
		})
	}
}

func TestBossTransitionSounds(t *testing.T) {
	tests := []struct {
		name                                         string
		beforeLives, beforePhase, afterLives, afterPhase int
		want                                         []audio.Sound
	}{
		{name: "no change", beforeLives: 3, beforePhase: 1, afterLives: 3, afterPhase: 1, want: nil},
		{name: "lives drop", beforeLives: 3, beforePhase: 1, afterLives: 2, afterPhase: 1, want: []audio.Sound{audio.SoundBossHit}},
		{name: "phase rises", beforeLives: 3, beforePhase: 1, afterLives: 3, afterPhase: 2, want: []audio.Sound{audio.SoundMeow}},
		{name: "both", beforeLives: 3, beforePhase: 1, afterLives: 2, afterPhase: 2, want: []audio.Sound{audio.SoundBossHit, audio.SoundMeow}},
		{name: "no boss", beforeLives: -1, beforePhase: -1, afterLives: -1, afterPhase: -1, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSounds(t, tt.want, bossTransitionSounds(tt.beforeLives, tt.beforePhase, tt.afterLives, tt.afterPhase))
		})
	}
}

func TestVolumePreviewSound(t *testing.T) {
	tests := []struct {
		name      string
		prev, cur campaign.Settings
		want      audio.Sound
		play      bool
	}{
		{
			name: "music change",
			prev: campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.5},
			cur:  campaign.Settings{MusicVolume: 0.75, SoundVolume: 0.5},
			want: audio.SoundMenuSelect,
			play: true,
		},
		{
			name: "sound change",
			prev: campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.5},
			cur:  campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.75},
			want: audio.SoundMenuSelect,
			play: true,
		},
		{
			name: "no change",
			prev: campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.5},
			cur:  campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.5},
		},
		{
			name: "non-volume change",
			prev: campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.5, Fullscreen: false},
			cur:  campaign.Settings{MusicVolume: 0.5, SoundVolume: 0.5, Fullscreen: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := volumePreviewSound(tt.prev, tt.cur)
			if ok != tt.play {
				t.Fatalf("volumePreviewSound(%+v, %+v) ok = %v, want %v", tt.prev, tt.cur, ok, tt.play)
			}
			if ok && got != tt.want {
				t.Fatalf("volumePreviewSound(%+v, %+v) sound = %v, want %v", tt.prev, tt.cur, got, tt.want)
			}
		})
	}
}

func TestTrackForScene(t *testing.T) {
	c := campaign.Default()

	menu := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), false)
	if got := trackForScene(menu); got != audio.TrackMenu {
		t.Fatalf("main menu track = %v, want TrackMenu", got)
	}

	flow := campaign.NewFlow(c, campaign.DefaultSettings(), campaign.NewProgress(), false)
	flow.StartLevel(1)
	if got := trackForScene(flow); got != audio.TrackMangroves {
		t.Fatalf("mangroves run track = %v, want TrackMangroves", got)
	}

	if !winBossFlow(t, flow) {
		t.Fatal("boss fight was not won")
	}
	if got := trackForScene(flow); got != audio.TrackEnding {
		t.Fatalf("ending track = %v, want TrackEnding", got)
	}

	flow.Handle(campaign.ActionConfirm) // ending -> credits
	if got := trackForScene(flow); got != audio.TrackEnding {
		t.Fatalf("credits track = %v, want TrackEnding", got)
	}
}

// assertSounds checks that got matches want exactly.
func assertSounds(t *testing.T, want, got []audio.Sound) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("sounds = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sound %d = %v, want %v", i, got[i], want[i])
		}
	}
}

// winBossFlow starts the boss fight and drives it until it leaves ScenePlay,
// then reports whether the hero won. It fails the test when the fight does
// not resolve within a generous step budget.
func winBossFlow(t *testing.T, f *campaign.Flow) bool {
	t.Helper()
	f.StartBoss()
	for step := 0; step < 100000 && f.Scene() == campaign.ScenePlay; step++ {
		f.StepRun(bossWinningInput(f, step%2 == 0), flowDT)
	}
	if f.Scene() != campaign.SceneEnding && f.Scene() != campaign.SceneGameOver {
		t.Fatalf("boss fight did not resolve: scene = %v", f.Scene())
	}
	return f.Scene() == campaign.SceneEnding
}

// bossWinningInput returns the input for one boss frame that drives the hero
// toward a win: it dodges attacks on the hero's lane and moves to Copycat's
// vulnerable lane to stomp. movePulse alternates so a held lane move releases
// and the hero can cross more than one lane during a window.
func bossWinningInput(f *campaign.Flow, movePulse bool) runner.Input {
	b := f.Boss()
	if b == nil {
		return runner.Input{}
	}
	in := runner.Input{}
	switch {
	case b.Vulnerable():
		switch {
		case b.AttackLane() > b.HeroLane():
			in.Right = movePulse
		case b.AttackLane() < b.HeroLane():
			in.Left = movePulse
		default:
			in.Jump = true
		}
	case b.AttackLane() == b.HeroLane():
		if b.AttackKind() == gameplay.AttackHigh {
			in.Slide = true
		} else {
			in.Jump = true
		}
	}
	return in
}
