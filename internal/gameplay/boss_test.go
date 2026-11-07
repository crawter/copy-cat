package gameplay_test

import (
	"reflect"
	"testing"

	"copycat/internal/gameplay"
	"copycat/internal/runner"
)

const (
	bossDT        = 1.0 / 60.0
	bossMaxSteps  = 4000
	bossWaitSteps = 400
)

func TestBossScriptedDodgeSurvivesEveryAttack(t *testing.T) {
	b := gameplay.NewBossFight(3)
	outs, won := runWinningScript(b, bossMaxSteps)
	if !won {
		t.Fatalf("scripted fight did not win: state=%v outcomes=%v", b.State(), outs)
	}
	if got := countOutcome(outs, gameplay.BossHit); got != 0 {
		t.Fatalf("scripted dodge took %d hits, want 0; outcomes=%v", got, outs)
	}
	if b.Lives() != 3 {
		t.Fatalf("Lives = %d after scripted win, want 3", b.Lives())
	}
	if got := countOutcome(outs, gameplay.BossStomped); got != 3 {
		t.Fatalf("BossStomped fired %d times, want 3", got)
	}
	if got := countOutcome(outs, gameplay.BossPhaseUp); got != 2 {
		t.Fatalf("BossPhaseUp fired %d times, want 2", got)
	}
	if got := countOutcome(outs, gameplay.BossWonOutcome); got != 1 {
		t.Fatalf("BossWonOutcome fired %d times, want 1", got)
	}
	if b.State() != gameplay.BossWon {
		t.Fatalf("State = %v after scripted win, want BossWon", b.State())
	}
}

func TestBossStandingOnAttackLaneWithoutDodgeLosesOneLife(t *testing.T) {
	b := gameplay.NewBossFight(3)

	// The first step announces the opening attack, so the getters carry the
	// real attack lane and kind before we commit the hero to the lane.
	if outs := b.Step(runner.Input{}, bossDT); !hasOutcome(outs, gameplay.BossTelegraph) {
		t.Fatalf("first step did not announce an attack: %v", outs)
	}

	target := b.AttackLane()
	for i := 0; i < 4 && b.HeroLane() != target; i++ {
		var in runner.Input
		if b.HeroLane() < target {
			in.Right = true
		} else {
			in.Left = true
		}
		b.Step(in, bossDT)
	}
	if b.HeroLane() != target {
		t.Fatalf("hero never reached attack lane %d; HeroLane=%d", target, b.HeroLane())
	}

	hits := 0
	for i := 0; i < bossWaitSteps && b.State() == gameplay.BossFighting; i++ {
		for _, o := range b.Step(runner.Input{}, bossDT) {
			if o == gameplay.BossHit {
				hits++
			}
		}
		if hits > 0 {
			break
		}
	}
	if hits != 1 {
		t.Fatalf("BossHit fired %d times, want exactly 1", hits)
	}
	if b.Lives() != 2 {
		t.Fatalf("Lives = %d after one undodged attack, want 2", b.Lives())
	}
	if b.State() != gameplay.BossFighting {
		t.Fatalf("State = %v after one hit with lives left, want BossFighting", b.State())
	}
}

func TestBossStompAdvancesPhase(t *testing.T) {
	b := gameplay.NewBossFight(3)
	if !stepUntilVulnerable(t, b) {
		t.Fatalf("fight never opened a vulnerable window; state=%v phase=%d", b.State(), b.Phase())
	}
	if b.Phase() != 1 || !b.Vulnerable() || b.Lives() != 3 {
		t.Fatalf("pre-stomp state: phase=%d vulnerable=%v lives=%d, want phase 1, vulnerable, 3 lives",
			b.Phase(), b.Vulnerable(), b.Lives())
	}

	stomped, phaseUp := false, false
	for i := 0; i < bossWaitSteps && b.State() == gameplay.BossFighting; i++ {
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
		for _, o := range b.Step(in, bossDT) {
			switch o {
			case gameplay.BossStomped:
				stomped = true
			case gameplay.BossPhaseUp:
				phaseUp = true
			}
		}
		if b.Phase() > 1 {
			break
		}
	}
	if !stomped || !phaseUp {
		t.Fatalf("stomp outcomes: stomped=%v phaseUp=%v, want both true", stomped, phaseUp)
	}
	if b.Phase() != 2 {
		t.Fatalf("Phase = %d after first stomp, want 2", b.Phase())
	}
	if b.Lives() != 3 {
		t.Fatalf("Lives = %d after stomp, want 3", b.Lives())
	}
	if b.Vulnerable() {
		t.Fatalf("Vulnerable still true after stomp advanced the phase")
	}
}

func TestBossThreeStompsWinAndScriptWinsEveryTime(t *testing.T) {
	for run := 0; run < 3; run++ {
		b := gameplay.NewBossFight(3)
		outs, won := runWinningScript(b, bossMaxSteps)
		if !won {
			t.Fatalf("run %d: script did not win; state=%v outcomes=%v", run, b.State(), outs)
		}
		if b.State() != gameplay.BossWon {
			t.Fatalf("run %d: State = %v, want BossWon", run, b.State())
		}
		if b.Lives() != 3 {
			t.Fatalf("run %d: Lives = %d, want 3", run, b.Lives())
		}
		if got := countOutcome(outs, gameplay.BossHit); got != 0 {
			t.Fatalf("run %d: scripted dodge took %d hits", run, got)
		}
		if got := countOutcome(outs, gameplay.BossStomped); got != 3 {
			t.Fatalf("run %d: BossStomped fired %d times, want 3", run, got)
		}
	}
}

func TestBossZeroLivesLoses(t *testing.T) {
	b := gameplay.NewBossFight(0)
	outs := b.Step(runner.Input{}, bossDT)
	if !hasOutcome(outs, gameplay.BossLostOutcome) {
		t.Fatalf("first step with zero lives did not fire BossLostOutcome: %v", outs)
	}
	if b.State() != gameplay.BossLost {
		t.Fatalf("State = %v with zero lives, want BossLost", b.State())
	}
	if b.Lives() != 0 {
		t.Fatalf("Lives = %d with zero lives, want 0", b.Lives())
	}
	if again := b.Step(runner.Input{}, bossDT); len(again) != 0 {
		t.Fatalf("Step after loss returned outcomes: %v", again)
	}
}

func TestBossLastHitLoses(t *testing.T) {
	b := gameplay.NewBossFight(1)

	// Announce the first attack, then stand on its lane without dodging.
	if outs := b.Step(runner.Input{}, bossDT); !hasOutcome(outs, gameplay.BossTelegraph) {
		t.Fatalf("first step did not announce an attack: %v", outs)
	}
	target := b.AttackLane()
	for i := 0; i < 4 && b.HeroLane() != target; i++ {
		var in runner.Input
		if b.HeroLane() < target {
			in.Right = true
		} else {
			in.Left = true
		}
		b.Step(in, bossDT)
	}

	hit, lost := false, false
	for i := 0; i < bossWaitSteps && b.State() == gameplay.BossFighting; i++ {
		for _, o := range b.Step(runner.Input{}, bossDT) {
			switch o {
			case gameplay.BossHit:
				hit = true
			case gameplay.BossLostOutcome:
				lost = true
			}
		}
		if lost {
			break
		}
	}
	if !hit || !lost {
		t.Fatalf("last-life hit outcomes: hit=%v lost=%v, want both true", hit, lost)
	}
	if b.State() != gameplay.BossLost {
		t.Fatalf("State = %v after losing last life, want BossLost", b.State())
	}
	if b.Lives() != 0 {
		t.Fatalf("Lives = %d after losing last life, want 0", b.Lives())
	}
}

func TestBossMissedWindowDealsNoDamage(t *testing.T) {
	b := gameplay.NewBossFight(3)
	if !stepUntilVulnerable(t, b) {
		t.Fatalf("fight never opened a vulnerable window")
	}
	lives := b.Lives()
	phase := b.Phase()

	var outs []gameplay.BossOutcome
	resumed := false
	for i := 0; i < bossWaitSteps && b.State() == gameplay.BossFighting && !resumed; i++ {
		for _, o := range b.Step(runner.Input{}, bossDT) {
			outs = append(outs, o)
			if o == gameplay.BossTelegraph {
				resumed = true
			}
		}
	}
	if !resumed {
		t.Fatalf("fight never resumed after missed window")
	}
	if got := countOutcome(outs, gameplay.BossHit); got != 0 {
		t.Fatalf("missed window dealt damage: %v", outs)
	}
	if b.Lives() != lives {
		t.Fatalf("Lives = %d after missed window, want %d", b.Lives(), lives)
	}
	if b.Phase() != phase {
		t.Fatalf("Phase = %d after missed window, want %d", b.Phase(), phase)
	}
	if b.Vulnerable() {
		t.Fatalf("Vulnerable still true after missed window resumed attacks")
	}
}

func TestBossSameInputScriptGivesIdenticalStates(t *testing.T) {
	b1 := gameplay.NewBossFight(3)
	b2 := gameplay.NewBossFight(3)

	for i := 0; i < bossMaxSteps; i++ {
		in := winningInput(b1)
		o1 := b1.Step(in, bossDT)
		o2 := b2.Step(in, bossDT)
		if !reflect.DeepEqual(o1, o2) {
			t.Fatalf("step %d outcomes differ: %v vs %v", i, o1, o2)
		}
		if s1, s2 := snapshot(b1), snapshot(b2); s1 != s2 {
			t.Fatalf("step %d states differ: %+v vs %+v", i, s1, s2)
		}
		if b1.State() != gameplay.BossFighting {
			break
		}
	}
	if b1.State() != gameplay.BossWon || b2.State() != gameplay.BossWon {
		t.Fatalf("identical script states: %v and %v, want BossWon", b1.State(), b2.State())
	}
}

type bossSnapshot struct {
	State      gameplay.BossState
	Phase      int
	Lives      int
	HeroLane   int
	Jumping    bool
	AttackLane int
	AttackKind gameplay.AttackKind
	Vulnerable bool
}

func snapshot(b *gameplay.BossFight) bossSnapshot {
	return bossSnapshot{
		State:      b.State(),
		Phase:      b.Phase(),
		Lives:      b.Lives(),
		HeroLane:   b.HeroLane(),
		Jumping:    b.Jumping(),
		AttackLane: b.AttackLane(),
		AttackKind: b.AttackKind(),
		Vulnerable: b.Vulnerable(),
	}
}

// runWinningScript drives b with a dodge-and-stomp script until it wins or
// runs out of steps, returning every reported outcome.
func runWinningScript(b *gameplay.BossFight, maxSteps int) ([]gameplay.BossOutcome, bool) {
	var all []gameplay.BossOutcome
	for i := 0; i < maxSteps && b.State() == gameplay.BossFighting; i++ {
		all = append(all, b.Step(winningInput(b), bossDT)...)
		if b.State() == gameplay.BossWon {
			return all, true
		}
	}
	return all, false
}

// stepUntilVulnerable dodges attacks without stomping until Copycat crouches.
func stepUntilVulnerable(t *testing.T, b *gameplay.BossFight) bool {
	t.Helper()
	for i := 0; i < bossMaxSteps && b.State() == gameplay.BossFighting; i++ {
		b.Step(dodgeInput(b), bossDT)
		if b.Vulnerable() {
			return true
		}
	}
	return false
}

// winningInput reacts to the current fight state: dodge attacks and, when
// Copycat is vulnerable, move to his lane and jump to stomp him.
func winningInput(b *gameplay.BossFight) runner.Input {
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
	return dodgeInput(b)
}

// dodgeInput moves toward the announced attack lane and applies the matching
// dodge: jump for a low attack, slide for a high attack once grounded.
func dodgeInput(b *gameplay.BossFight) runner.Input {
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

func hasOutcome(outs []gameplay.BossOutcome, want gameplay.BossOutcome) bool {
	for _, o := range outs {
		if o == want {
			return true
		}
	}
	return false
}

func countOutcome(outs []gameplay.BossOutcome, want gameplay.BossOutcome) int {
	n := 0
	for _, o := range outs {
		if o == want {
			n++
		}
	}
	return n
}
