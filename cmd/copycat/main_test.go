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

func TestSlideOnContainerMapsToSlide(t *testing.T) {
	r := runnerOnContainer(t)
	if airborne(r, nil) {
		t.Fatal("hero riding a container ramp must be grounded, not airborne")
	}
	if got := slideSound(airborne(r, nil)); got != audio.SoundSlide {
		t.Fatalf("slide on container = %v, want SoundSlide", got)
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
