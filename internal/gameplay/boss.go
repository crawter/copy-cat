package gameplay

import "copycat/internal/runner"

// BossState is the current state of the Copycat boss fight.
type BossState int

// The states a boss fight can be in.
const (
	// BossFighting means the fight is in progress.
	BossFighting BossState = iota
	// BossWon means Copycat has been stomped for the third time.
	BossWon
	// BossLost means the hero's lives reached zero.
	BossLost
)

// BossOutcome names one thing that happened during a boss fight step.
type BossOutcome int

// The outcomes a boss fight reports, in the order they happen.
const (
	// BossTelegraph reports that Copycat announced an attack on AttackLane.
	BossTelegraph BossOutcome = iota
	// BossAttack reports that the announced attack landed.
	BossAttack
	// BossHit reports that the landed attack cost the hero one life.
	BossHit
	// BossVulnerable reports that Copycat crouched, opening a stomp window.
	BossVulnerable
	// BossStomped reports that the hero landed on the crouching Copycat.
	BossStomped
	// BossPhaseUp reports that a stomp advanced the fight to the next phase.
	BossPhaseUp
	// BossWonOutcome reports that the third stomp defeated Copycat.
	BossWonOutcome
	// BossLostOutcome reports that the hero's lives reached zero.
	BossLostOutcome
)

// AttackKind is the kind of attack Copycat has announced.
type AttackKind int

// The kinds of attacks Copycat can announce.
const (
	// AttackLow is a low swipe; a jump dodges it.
	AttackLow AttackKind = iota
	// AttackHigh is a high swipe; a slide dodges it.
	AttackHigh
)

// BossFight is the deterministic Copycat boss simulation. It owns the
// three-lane arena, Copycat's phased attacks and vulnerable windows, and the
// hero's dodge and stomp moves. It never creates a runner.
type BossFight struct {
	lives int
	state BossState

	phase int

	heroLane  int
	jumping   bool
	jumpLeft  float64
	sliding   bool
	slideLeft float64

	vulnerable bool
	vulnLeft   float64

	announced     bool
	attackLane    int
	attackKind    AttackKind
	telegraphLeft float64
	attacksDone   int

	attackSeq      int
	patternCounter int

	prevLeft  bool
	prevRight bool
	prevJump  bool
	prevSlide bool
}

// NewBossFight returns a Copycat boss fight with lives lives. It starts at
// phase 1 with the hero on the center lane. A fight with zero lives loses on
// its first step.
func NewBossFight(lives int) *BossFight {
	if lives < 0 {
		lives = 0
	}
	return &BossFight{
		lives:    lives,
		state:    BossFighting,
		phase:    1,
		heroLane: 1,
	}
}

// Step advances the boss fight by dt seconds with input in and returns every
// outcome that happened this step, in order. When the fight is over it
// changes nothing and returns no outcomes.
func (b *BossFight) Step(in runner.Input, dt float64) []BossOutcome {
	if b.state != BossFighting {
		return nil
	}
	if b.lives <= 0 {
		b.state = BossLost
		return []BossOutcome{BossLostOutcome}
	}
	if dt <= 0 {
		return nil
	}

	var out []BossOutcome
	out = append(out, b.handleInput(in)...)

	b.prevLeft = in.Left
	b.prevRight = in.Right
	b.prevJump = in.Jump
	b.prevSlide = in.Slide

	if b.state != BossFighting {
		return out
	}

	out = append(out, b.advanceAttack(dt)...)
	if b.state != BossFighting {
		return out
	}

	out = append(out, b.advanceJump(dt)...)
	if b.state != BossFighting {
		return out
	}

	b.advanceSlide(dt)
	out = append(out, b.advanceVulnerable(dt)...)
	return out
}

// State returns the fight's current state.
func (b *BossFight) State() BossState {
	return b.state
}

// Phase returns the current phase, from 1 to 3.
func (b *BossFight) Phase() int {
	return b.phase
}

// Lives returns how many lives the hero still has.
func (b *BossFight) Lives() int {
	return b.lives
}

// HeroLane returns the lane the hero currently centers on: 0, 1 or 2.
func (b *BossFight) HeroLane() int {
	return b.heroLane
}

// Jumping reports whether the hero is airborne in a jump.
func (b *BossFight) Jumping() bool {
	return b.jumping
}

// Sliding reports whether the hero is in a grounded slide.
func (b *BossFight) Sliding() bool {
	return b.sliding
}

// AttackLane returns the lane Copycat occupies. During a telegraphed attack it
// is the attack lane; during a vulnerable window it is the crouch lane to
// stomp.
func (b *BossFight) AttackLane() int {
	return b.attackLane
}

// AttackKind returns the kind of the currently announced attack.
func (b *BossFight) AttackKind() AttackKind {
	return b.attackKind
}

// Vulnerable reports whether Copycat is crouched and can be stomped.
func (b *BossFight) Vulnerable() bool {
	return b.vulnerable
}

// handleInput applies edge-triggered hero moves: one lane per press, a jump,
// a grounded slide, or a slam that lands an airborne slide immediately.
func (b *BossFight) handleInput(in runner.Input) []BossOutcome {
	var out []BossOutcome
	if in.Left && !b.prevLeft && b.heroLane > 0 {
		b.heroLane--
	}
	if in.Right && !b.prevRight && b.heroLane < 2 {
		b.heroLane++
	}

	if in.Slide && !b.prevSlide {
		if b.jumping {
			b.jumping = false
			b.jumpLeft = 0
			out = append(out, b.landing()...)
		} else {
			b.sliding = true
			b.slideLeft = bossSlideDuration
		}
	}

	if in.Jump && !b.prevJump && !b.jumping {
		b.jumping = true
		b.jumpLeft = bossJumpDuration
		b.sliding = false
		b.slideLeft = 0
	}
	return out
}

// advanceAttack announces the next attack, counts down its telegraph and
// resolves the hit when it lands.
func (b *BossFight) advanceAttack(dt float64) []BossOutcome {
	if b.vulnerable {
		return nil
	}
	if !b.announced {
		b.startAttack()
		return []BossOutcome{BossTelegraph}
	}

	b.telegraphLeft -= dt
	if b.telegraphLeft > 0 {
		return nil
	}
	b.telegraphLeft = 0

	out := []BossOutcome{BossAttack}
	if b.heroLane == b.attackLane && !b.dodges() {
		b.lives--
		out = append(out, BossHit)
		if b.lives <= 0 {
			b.state = BossLost
			out = append(out, BossLostOutcome)
			return out
		}
	}

	b.attacksDone++
	if b.attacksDone >= b.phaseAttacks() {
		b.attacksDone = 0
		b.attackLane = b.nextLane()
		b.vulnerable = true
		b.vulnLeft = b.phaseWindow()
		b.announced = false
		out = append(out, BossVulnerable)
	} else {
		b.startAttack()
		out = append(out, BossTelegraph)
	}
	return out
}

// advanceJump counts down a jump and resolves its landing.
func (b *BossFight) advanceJump(dt float64) []BossOutcome {
	if !b.jumping {
		return nil
	}
	b.jumpLeft -= dt
	if b.jumpLeft > 0 {
		return nil
	}
	b.jumpLeft = 0
	b.jumping = false
	return b.landing()
}

// advanceSlide counts down a grounded slide.
func (b *BossFight) advanceSlide(dt float64) {
	if !b.sliding {
		return
	}
	b.slideLeft -= dt
	if b.slideLeft <= 0 {
		b.sliding = false
		b.slideLeft = 0
	}
}

// advanceVulnerable counts down a stomp window and resumes attacks when it
// expires without a stomp.
func (b *BossFight) advanceVulnerable(dt float64) []BossOutcome {
	if !b.vulnerable {
		return nil
	}
	b.vulnLeft -= dt
	if b.vulnLeft > 0 {
		return nil
	}
	b.vulnLeft = 0
	b.vulnerable = false
	b.announced = false
	b.startAttack()
	return []BossOutcome{BossTelegraph}
}

// landing resolves a jump or slam landing. A landing on Copycat's crouch lane
// during the vulnerable window stomps him.
func (b *BossFight) landing() []BossOutcome {
	if !b.vulnerable || b.heroLane != b.attackLane {
		return nil
	}

	b.vulnerable = false
	b.vulnLeft = 0
	out := []BossOutcome{BossStomped}

	if b.phase >= 3 {
		b.state = BossWon
		return append(out, BossWonOutcome)
	}

	b.phase++
	out = append(out, BossPhaseUp)
	b.attacksDone = 0
	b.announced = false
	b.startAttack()
	out = append(out, BossTelegraph)
	return out
}

// dodges reports whether the hero's current move dodges the announced attack.
func (b *BossFight) dodges() bool {
	switch b.attackKind {
	case AttackLow:
		return b.jumping
	case AttackHigh:
		return b.sliding
	default:
		return false
	}
}

// startAttack chooses the next attack and starts its telegraph.
func (b *BossFight) startAttack() {
	b.attackLane = b.nextLane()
	b.attackKind = attackKindPattern[b.attackSeq%len(attackKindPattern)]
	b.attackSeq++
	b.telegraphLeft = b.phaseTelegraph()
	b.announced = true
}

// nextLane returns the next lane from the fixed cycle, so no lane choice is
// ever random.
func (b *BossFight) nextLane() int {
	lane := lanePattern[b.patternCounter%len(lanePattern)]
	b.patternCounter++
	return lane
}

// phaseAttacks returns how many telegraphed attacks the current phase throws.
func (b *BossFight) phaseAttacks() int {
	return b.phase + 1
}

// phaseTelegraph returns the attack telegraph delay of the current phase.
func (b *BossFight) phaseTelegraph() float64 {
	switch b.phase {
	case 1:
		return bossPhase1Telegraph
	case 2:
		return bossPhase2Telegraph
	default:
		return bossPhase3Telegraph
	}
}

// phaseWindow returns the vulnerable window length of the current phase.
func (b *BossFight) phaseWindow() float64 {
	switch b.phase {
	case 1:
		return bossPhase1Window
	case 2:
		return bossPhase2Window
	default:
		return bossPhase3Window
	}
}

const (
	// bossJumpDuration is how long a jump keeps the hero airborne, in
	// seconds. It is short enough that a jump started at the start of any
	// vulnerable window lands inside it.
	bossJumpDuration = 0.5
	// bossSlideDuration is how long a grounded slide keeps the hero low, in
	// seconds. It covers every attack telegraph.
	bossSlideDuration = 0.6

	// bossPhase1Telegraph, bossPhase2Telegraph and bossPhase3Telegraph are
	// the attack telegraph delays in seconds; they shorten each phase.
	bossPhase1Telegraph = 0.4
	bossPhase2Telegraph = 0.35
	bossPhase3Telegraph = 0.3

	// bossPhase1Window, bossPhase2Window and bossPhase3Window are the
	// vulnerable window lengths in seconds; they shrink each phase.
	bossPhase1Window = 1.0
	bossPhase2Window = 0.9
	bossPhase3Window = 0.8
)

// lanePattern is the fixed cycle Copycat's lane choices follow.
var lanePattern = [...]int{0, 1, 2, 1}

// attackKindPattern is the fixed cycle Copycat's attack kinds follow.
var attackKindPattern = [...]AttackKind{AttackLow, AttackHigh}
