package render

import (
	"fmt"
	"strings"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
)

// captionFlow returns a flow with a fresh endless run, so every caption test
// starts from a real gameplay.Run built by gameplay.NewRun.
func captionFlow(t *testing.T) *campaign.Flow {
	t.Helper()
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.StartEndless(7)
	if f.Run() == nil {
		t.Fatal("endless run did not start")
	}
	return f
}

// runSnap returns a caption snapshot matching r's current counters, as if the
// previous frame had already recorded them.
func runSnap(r *gameplay.Run) captionSnap {
	return captionSnap{
		run:       r,
		livesLost: r.LivesLost,
		stomps:    r.Stomps,
		bald:      r.Bald,
		chase:     r.Chase,
		shield:    r.ShieldHits,
	}
}

// loadingTip returns the deterministic tip introLines picks for idx.
func loadingTip(f *campaign.Flow, idx uint64) string {
	return textFor(f, campaign.LoadingTips[idx%uint64(len(campaign.LoadingTips))])
}

func TestCaptionLifeLossUsesCauseLine(t *testing.T) {
	cases := []struct {
		name  string
		cause gameplay.Cause
	}{
		{"pit", gameplay.CausePit},
		{"barrier", gameplay.CauseBarrier},
		{"enemy", gameplay.CauseEnemy},
		{"snip", gameplay.CauseSnip},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := captionFlow(t)
			r := f.Run()
			snap := runSnap(r)
			r.LivesLost++
			r.Cause = tc.cause

			line, ok := captionOnChange(f, &snap)
			if !ok {
				t.Fatal("captionOnChange reported no caption after a life loss")
			}
			if want := gameplay.CauseLines[tc.cause]; line != want {
				t.Fatalf("caption = %+v, want %+v", line, want)
			}
		})
	}
}

func TestCaptionStomp(t *testing.T) {
	f := captionFlow(t)
	r := f.Run()
	snap := runSnap(r)
	r.Stomps++

	line, ok := captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption after a stomp")
	}
	if line != gameplay.StompedMudskipper {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.StompedMudskipper)
	}
}

func TestCaptionSnip(t *testing.T) {
	f := captionFlow(t)
	r := f.Run()
	snap := runSnap(r)
	// Only Bald changes here; the life loss has already been recorded in the
	// snapshot, so the snip line is the transition left to report.
	r.Bald = true

	line, ok := captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption when Bald turned true")
	}
	if line != gameplay.SnipLine {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.SnipLine)
	}
}

func TestCaptionChase(t *testing.T) {
	f := captionFlow(t)
	r := f.Run()
	snap := runSnap(r)
	snap.chase = 0.4
	r.Chase = 1.0

	line, ok := captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption when the chase crossed the threshold")
	}
	if line != gameplay.ChaseLine {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.ChaseLine)
	}
}

func TestCaptionShield(t *testing.T) {
	f := captionFlow(t)
	r := f.Run()
	snap := runSnap(r)
	snap.shield = 1
	r.ShieldHits = 0

	line, ok := captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption when the shield dropped")
	}
	if line != gameplay.ShieldLine {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.ShieldLine)
	}
}

func TestCaptionBossPhases(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.StartBoss()
	b := f.Boss()
	if b == nil {
		t.Fatal("boss fight did not start")
	}

	var snap captionSnap
	line, ok := captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption at the boss fight start")
	}
	if line != gameplay.BossPhaseLines[0] {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.BossPhaseLines[0])
	}
	if snap.boss != b || snap.phase != 1 {
		t.Fatalf("snapshot not reset to phase 1: %+v", snap)
	}

	for i := 0; i < 4000 && b.Phase() == 1 && b.State() == gameplay.BossFighting; i++ {
		f.StepRun(bossWinningInput(b), testDT)
	}
	if b.Phase() != 2 {
		t.Fatalf("boss phase = %d, want 2", b.Phase())
	}
	line, ok = captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption at phase 2")
	}
	if line != gameplay.BossPhaseLines[1] {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.BossPhaseLines[1])
	}

	for i := 0; i < 4000 && b.Phase() == 2 && b.State() == gameplay.BossFighting; i++ {
		f.StepRun(bossWinningInput(b), testDT)
	}
	if b.Phase() != 3 {
		t.Fatalf("boss phase = %d, want 3", b.Phase())
	}
	line, ok = captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption at phase 3")
	}
	if line != gameplay.BossPhaseLines[2] {
		t.Fatalf("caption = %+v, want %+v", line, gameplay.BossPhaseLines[2])
	}
}

func TestCaptionPrefersLifeLossWhenSeveralFire(t *testing.T) {
	f := captionFlow(t)
	r := f.Run()
	snap := runSnap(r)
	snap.chase = 0.4
	// A snip changes the life counter, the bald flag and the chase at once.
	r.LivesLost++
	r.Cause = gameplay.CauseSnip
	r.Bald = true
	r.Chase = 0

	line, ok := captionOnChange(f, &snap)
	if !ok {
		t.Fatal("captionOnChange reported no caption")
	}
	if want := gameplay.CauseLines[gameplay.CauseSnip]; line != want {
		t.Fatalf("caption = %+v, want the snip cause line %+v", line, want)
	}
}

func TestCaptionResetsWhenRunChanges(t *testing.T) {
	f := captionFlow(t)
	r := f.Run()
	snap := runSnap(r)

	if line, ok := captionOnChange(f, &snap); ok {
		t.Fatalf("caption = %+v, want none for an unchanged run", line)
	}

	f.StartEndless(9)
	if line, ok := captionOnChange(f, &snap); ok {
		t.Fatalf("caption = %+v, want none after a new run", line)
	}
	if snap.run != f.Run() {
		t.Fatalf("snapshot run not reset: %v, want %v", snap.run, f.Run())
	}
	if snap.livesLost != f.Run().LivesLost || snap.stomps != f.Run().Stomps {
		t.Fatalf("snapshot counters not reset: %+v", snap)
	}
}

func TestPauseLinesIncludeKeyHints(t *testing.T) {
	f := newGameFlow(t)
	f.Handle(campaign.ActionPause)
	if f.Scene() != campaign.ScenePause {
		t.Fatalf("scene = %v, want ScenePause", f.Scene())
	}

	lines := MenuLines(f)
	if want := 2 + len(campaign.KeyHints); len(lines) != want {
		t.Fatalf("pause lines = %d, want %d", len(lines), want)
	}
	for i, hint := range campaign.KeyHints {
		if got, want := lines[2+i].Text, textFor(f, hint); got != want {
			t.Fatalf("pause line %d = %q, want %q", i, got, want)
		}
	}
}

func TestIntroLinesCampaignLevelOne(t *testing.T) {
	f := newGameFlow(t)
	lines := introLines(f)
	if want := 2 + len(campaign.KeyHints); len(lines) != want {
		t.Fatalf("intro lines = %d, want %d", len(lines), want)
	}
	if got, want := lines[0].Text, textFor(f, campaign.LevelNames[0]); got != want {
		t.Fatalf("intro name = %q, want %q", got, want)
	}
	if got, want := lines[1].Text, loadingTip(f, 1); got != want {
		t.Fatalf("intro tip = %q, want %q", got, want)
	}
	for i, hint := range campaign.KeyHints {
		if got, want := lines[2+i].Text, textFor(f, hint); got != want {
			t.Fatalf("hint %d = %q, want %q", i, got, want)
		}
	}
}

func TestIntroLinesCampaignLevelTwoOmitsKeyHints(t *testing.T) {
	f := campaign.NewFlow(flatCampaign(), campaign.DefaultSettings(), campaign.Progress{Unlocked: 2}, false)
	f.StartLevel(2)
	if f.Scene() != campaign.ScenePlay {
		t.Fatalf("scene = %v, want ScenePlay", f.Scene())
	}
	lines := introLines(f)
	if len(lines) != 2 {
		t.Fatalf("intro lines = %d, want 2", len(lines))
	}
	if got, want := lines[0].Text, textFor(f, campaign.LevelNames[1]); got != want {
		t.Fatalf("intro name = %q, want %q", got, want)
	}
	if got, want := lines[1].Text, loadingTip(f, 2); got != want {
		t.Fatalf("intro tip = %q, want %q", got, want)
	}
}

func TestIntroLinesEndlessUsesSeed(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.StartEndless(7)
	lines := introLines(f)
	if len(lines) != 2 {
		t.Fatalf("intro lines = %d, want 2", len(lines))
	}
	if got, want := lines[0].Text, levelText(f, levelgen.EndlessName); got != want {
		t.Fatalf("intro name = %q, want %q", got, want)
	}
	if got, want := lines[1].Text, loadingTip(f, 7); got != want {
		t.Fatalf("intro tip = %q, want %q", got, want)
	}
}

func TestLevelEndLinesShowRunCounters(t *testing.T) {
	t.Run("campaign", func(t *testing.T) {
		f := newGameFlow(t)
		r := f.Run()
		if r == nil {
			t.Fatal("no run to read counters from")
		}
		r.TotalCombs = 12
		r.LivesLost = 2
		r.Stomps = 3
		checkLevelEndCounters(t, f, joinLines(levelEndLines(f)), r)
	})
	t.Run("endless", func(t *testing.T) {
		f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
		f.StartEndless(7)
		r := f.Run()
		if r == nil {
			t.Fatal("no endless run to read counters from")
		}
		r.TotalCombs = 12
		r.LivesLost = 2
		r.Stomps = 3
		checkLevelEndCounters(t, f, joinLines(levelEndLines(f)), r)
	})
}

func checkLevelEndCounters(t *testing.T, f *campaign.Flow, text string, r *gameplay.Run) {
	t.Helper()
	for _, want := range []string{
		textFor(f, campaign.LevelEndCombs),
		textFor(f, campaign.LevelEndLivesLost),
		textFor(f, campaign.LevelEndStomps),
		fmt.Sprintf("%d", r.TotalCombs),
		fmt.Sprintf("%d", r.LivesLost),
		fmt.Sprintf("%d", r.Stomps),
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("level end missing %q in %q", want, text)
		}
	}
}
