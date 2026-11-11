package campaign

import (
	"math"
	"testing"

	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
	"copycat/internal/runner"
)

// playStepDT is the fixed frame the headless playthrough and endless smoke
// tests drive the flow at.
const playStepDT = 1.0 / 60.0

// Controller tuning. The values are chosen so the reactive controller stays
// inside the passability window levelgen guarantees for every campaign preset.
const (
	// pitJumpLead is how many meters before a pit's leading edge the
	// controller starts the clearing jump. It is less than the shortest jump
	// distance minus the widest pit, so the landing always clears the pit.
	pitJumpLead = 1.0
	// switchLeadScale turns the current speed into how many meters before a
	// barrier or enemy the controller starts a lane change. The switch takes
	// a quarter of a second, so 0.4 seconds of headroom finishes it early.
	switchLeadScale = 0.4
	// forcedDepth is the footprint depth the runner and levelgen use for a
	// barrier or enemy.
	forcedDepth = 1.0
)

// Budgets keep every headless run finite; they are far larger than a clean
// run needs.
const (
	playthroughStepBudget = 20000
	endlessStepBudget     = 20000
	bossStepBudget        = 10000
	endlessSmokeSeed      = 42
)

func TestFullCampaignPlaythrough(t *testing.T) {
	c := Default()
	f := NewFlow(c, DefaultSettings(), NewProgress(), true)

	// First run: keep the default English, then start a new game.
	if f.Scene() != SceneLanguagePick {
		t.Fatalf("Scene = %v, want SceneLanguagePick", f.Scene())
	}
	f.Handle(ActionConfirm)
	if f.Settings().Language != LanguageEnglish {
		t.Fatalf("Language = %v, want LanguageEnglish", f.Settings().Language)
	}
	if f.Scene() != SceneMainMenu {
		t.Fatalf("Scene = %v after language pick, want SceneMainMenu", f.Scene())
	}

	f.Handle(ActionDown) // New Game
	f.Handle(ActionConfirm)
	if f.Scene() != SceneNewGameConfirm {
		t.Fatalf("Scene = %v after New Game, want SceneNewGameConfirm", f.Scene())
	}
	f.Handle(ActionUp) // Yes
	f.Handle(ActionConfirm)
	if f.Scene() != SceneStory {
		t.Fatalf("Scene = %v after New Game confirm, want SceneStory", f.Scene())
	}
	if f.Story() != StoryIntro {
		t.Fatalf("Story = %v after New Game, want StoryIntro", f.Story())
	}
	f.Handle(ActionConfirm) // StoryIntro → level 1
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after StoryIntro confirm, want ScenePlay", f.Scene())
	}

	for levelID := 1; levelID <= 9; levelID++ {
		if lv, ok := f.CurrentLevel(); !ok || lv.ID != levelID {
			t.Fatalf("before level %d: CurrentLevel = %+v, %v", levelID, lv, ok)
		}

		playRunLevel(t, f, levelID)

		if f.Scene() != SceneLevelEnd {
			t.Fatalf("after level %d: Scene = %v, want SceneLevelEnd", levelID, f.Scene())
		}
		if got := f.Progress().Unlocked; got != levelID+1 {
			t.Fatalf("after level %d: Unlocked = %d, want %d", levelID, got, levelID+1)
		}

		f.Handle(ActionConfirm)
		switch levelID {
		case 3, 6:
			if f.Scene() != SceneStory {
				t.Fatalf("after level %d confirm: Scene = %v, want SceneStory", levelID, f.Scene())
			}
			f.Handle(ActionConfirm)
			if f.Scene() != ScenePlay {
				t.Fatalf("after story confirm for level %d: Scene = %v, want ScenePlay", levelID, f.Scene())
			}
		case 9:
			if f.Scene() != ScenePlay || f.Boss() == nil {
				t.Fatalf("after level 9 confirm: Scene = %v, Boss = %v, want the boss fight", f.Scene(), f.Boss())
			}
		default:
			if f.Scene() != ScenePlay {
				t.Fatalf("after level %d confirm: Scene = %v, want ScenePlay", levelID, f.Scene())
			}
		}
	}

	winBossScripted(t, f)

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
}

func TestEndlessSmoke(t *testing.T) {
	c := Default()
	f := NewFlow(c, DefaultSettings(), NewProgress(), false)

	f.StartEndless(endlessSmokeSeed)
	if !f.Endless() {
		t.Fatalf("Endless = false after StartEndless, want true")
	}
	if f.Scene() != ScenePlay {
		t.Fatalf("Scene = %v after StartEndless, want ScenePlay", f.Scene())
	}
	run := f.Run()
	if run == nil {
		t.Fatalf("Run = nil after StartEndless")
	}

	ctl := newPlayController(run)
	steps := 0
	for steps < endlessStepBudget && f.Scene() == ScenePlay {
		f.StepRun(ctl.input(), playStepDT)
		steps++
	}

	// The endless level end opens on death or on finishing the corridor.
	if f.Scene() != SceneLevelEnd {
		t.Fatalf("endless smoke: Scene = %v after %d steps, want SceneLevelEnd", f.Scene(), steps)
	}
	if f.Progress().BestEndless <= 0 {
		t.Fatalf("endless smoke: BestEndless = %v, want > 0", f.Progress().BestEndless)
	}
	if !f.Endless() {
		t.Fatalf("endless smoke: Endless = false at the level end, want true")
	}
}

// playRunLevel drives one campaign run level with the reactive controller
// until it finishes. It fails on any hit and on a game over.
func playRunLevel(t *testing.T, f *Flow, levelID int) {
	t.Helper()
	run := f.Run()
	if run == nil {
		t.Fatalf("level %d: Run = nil", levelID)
	}
	ctl := newPlayController(run)
	for step := 0; step < playthroughStepBudget && f.Scene() == ScenePlay; step++ {
		out := f.StepRun(ctl.input(), playStepDT)
		for _, o := range out {
			switch o {
			case gameplay.OutcomeStumble, gameplay.OutcomeSnip, gameplay.OutcomePit, gameplay.OutcomeDied:
				t.Fatalf("level %d: controller took a hit at z=%.2f (outcome %v)", levelID, run.Runner.Z(), o)
			}
		}
	}
	if f.Scene() == SceneGameOver {
		t.Fatalf("level %d: game over; the controller failed", levelID)
	}
	if run.State != gameplay.StateFinished {
		t.Fatalf("level %d: did not finish (state=%v scene=%v)", levelID, run.State, f.Scene())
	}
}

// winBossScripted drives the deterministic boss fight with a dodge-and-stomp
// script until it wins. Copycat's attacks follow the fixed lane and kind
// patterns, so the same reaction wins every fight.
func winBossScripted(t *testing.T, f *Flow) {
	t.Helper()
	b := f.Boss()
	if b == nil {
		t.Fatalf("no boss fight to win")
	}
	for i := 0; i < bossStepBudget && b.State() == gameplay.BossFighting; i++ {
		f.StepRun(bossScriptInput(b), playStepDT)
	}
	if b.State() != gameplay.BossWon {
		t.Fatalf("boss did not win: state=%v", b.State())
	}
}

// bossScriptInput reacts to the telegraphed, deterministic attack: move onto
// the announced lane, dodge with the matching move, and stomp every
// vulnerable window.
func bossScriptInput(b *gameplay.BossFight) runner.Input {
	var in runner.Input
	if b.Vulnerable() {
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

// playController is the reactive run controller. It reads the runner and the
// level elements ahead once per step and returns the input that keeps the
// hero alive by construction: jump pits, and change lanes away from barriers
// and enemies.
type playController struct {
	run    *gameplay.Run
	elems  []levelgen.Element
	cons   levelgen.Constraints
	cursor int
}

func newPlayController(run *gameplay.Run) *playController {
	return &playController{
		run:   run,
		elems: run.Level.Elements(0, run.Level.Length()),
		cons:  run.Level.Constraints(),
	}
}

// input returns the input for the next step.
func (c *playController) input() runner.Input {
	r := c.run.Runner
	z := r.Z()

	for c.cursor < len(c.elems) && footprintEnd(c.elems[c.cursor]) <= z {
		c.cursor++
	}

	next, ok := c.nextForced()
	if !ok {
		return runner.Input{}
	}

	var in runner.Input
	if next.Kind == levelgen.Pit {
		if z >= next.Z-pitJumpLead {
			in.Jump = true
		}
		return in
	}

	lane := currentLane(r.Lane())
	if lane != next.Lane {
		return in
	}
	if dir := c.dodgeDirection(lane, next.Z); dir != 0 {
		if z >= next.Z-switchLead(r.Speed()) {
			if dir < 0 {
				in.Left = true
			} else {
				in.Right = true
			}
		}
	}
	return in
}

// nextForced returns the next pit, barrier or enemy whose footprint the hero
// has not yet passed.
func (c *playController) nextForced() (levelgen.Element, bool) {
	for i := c.cursor; i < len(c.elems); i++ {
		switch c.elems[i].Kind {
		case levelgen.Pit, levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
			return c.elems[i], true
		}
	}
	return levelgen.Element{}, false
}

// dodgeDirection picks the neighbouring lane that stays free of barriers and
// enemies for one reaction gap, and returns the switch direction toward it.
func (c *playController) dodgeDirection(lane int, z float64) int {
	for _, dir := range []int{-1, 1} {
		target := lane + dir
		if target < 0 || target > 2 {
			continue
		}
		if c.laneFree(target, z, z+c.cons.ReactionGap) {
			return dir
		}
	}
	return 0
}

// laneFree reports whether lane holds no barrier or enemy overlapping the
// half-open window [from, to).
func (c *playController) laneFree(lane int, from, to float64) bool {
	for _, e := range c.elems {
		switch e.Kind {
		case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
		default:
			continue
		}
		if e.Lane != lane {
			continue
		}
		if e.Z < to && e.Z+forcedDepth > from {
			return false
		}
	}
	return true
}

// footprintEnd returns the z just past an element's footprint.
func footprintEnd(e levelgen.Element) float64 {
	switch e.Kind {
	case levelgen.Pit, levelgen.Ramp:
		return e.Z + e.Span
	case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
		return e.Z + forcedDepth
	default:
		return e.Z
	}
}

// switchLead returns how many meters before an obstacle the controller starts
// its lane change.
func switchLead(speed float64) float64 {
	return speed * switchLeadScale
}

// currentLane returns the lane the runner currently centers on.
func currentLane(l float64) int {
	return int(math.Round(l))
}
