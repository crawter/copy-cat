// Package runner simulates the hero's three-lane auto-run against a
// generated level.
//
// The runner is pure Go and deterministic: the same input script and step
// durations always produce the identical state. It owns movement only — jump,
// slide, slam, lane switches and container ramps — and reports what happened
// each step as events. Damage, lives and death rules belong to the gameplay
// layer, which reads those events. The runner imports levelgen only, never
// raylib or audio, so its tests run headless.
package runner

import (
	"math"

	"copycat/internal/levelgen"
)

// Event names one thing that happened during a step.
type Event int

// The events the runner reports.
const (
	// EventJump reports that the hero started a jump.
	EventJump Event = iota
	// EventSlide reports that the hero started a slide.
	EventSlide
	// EventSlam reports that the hero started a slam while airborne.
	EventSlam
	// EventStomp reports that the hero landed on an enemy from above.
	EventStomp
	// EventHitEnemy reports that the hero ran into an enemy at ground height.
	EventHitEnemy
	// EventHitBarrier reports that the hero ran into a barrier without the
	// matching action.
	EventHitBarrier
	// EventFellInPit reports that the hero fell into a pit.
	EventFellInPit
	// EventRampUp reports that the hero climbed onto a container ramp.
	EventRampUp
	// EventRampDown reports that the hero left a container ramp.
	EventRampDown
	// EventFinished reports that the hero crossed the finish.
	EventFinished
)

// Input is the player's raw intent for one step.
type Input struct {
	Left, Right, Jump, Slide bool
}

// Config tunes the runner's movement. All fields are in SI units unless the
// comment says otherwise. DefaultConfig returns the values the game ships
// with.
type Config struct {
	// BaseSpeed is the starting forward speed in meters per second; it should
	// match the level's base speed.
	BaseSpeed float64
	// MaxSpeed is the forward speed cap in meters per second.
	MaxSpeed float64
	// SpeedRampDistance is the distance, in meters, over which speed eases
	// from BaseSpeed toward MaxSpeed.
	SpeedRampDistance float64

	// LaneSwitchDuration is the time, in seconds, to glide one lane.
	LaneSwitchDuration float64

	// JumpVelocity is the initial upward velocity of a jump, in meters per
	// second.
	JumpVelocity float64
	// Gravity is the downward acceleration, in meters per second squared.
	Gravity float64

	// SlideDuration is how long a slide lasts, in seconds.
	SlideDuration float64

	// SlamVelocity is the downward velocity applied by a slam, in meters per
	// second.
	SlamVelocity float64
	// SlamLockout is the recovery lockout after a slam lands, in seconds.
	SlamLockout float64
}

// DefaultConfig returns the movement tuning the finished game uses.
func DefaultConfig() Config {
	return Config{
		BaseSpeed:         8,
		MaxSpeed:          16,
		SpeedRampDistance: 240,
		LaneSwitchDuration: 0.25,
		JumpVelocity:      7,
		Gravity:           22,
		SlideDuration:     0.5,
		SlamVelocity:      30,
		SlamLockout:       0.35,
	}
}

// Runner simulates the hero's movement along one generated level.
type Runner struct {
	level *levelgen.Level
	cfg   Config
	elems []levelgen.Element // sorted copy of the level's elements

	z         float64 // front edge along the corridor, in meters
	lane      float64 // continuous lane index in [0, 2]
	height    float64 // feet above the current ground, in meters
	vy        float64 // vertical velocity, in meters per second
	air       bool    // airborne: jumping, slamming or falling
	sliding   bool
	slideTimer float64
	slamming  bool
	slamLock  float64

	onContainer bool

	switching     bool
	switchFrom    float64
	switchTarget  float64
	switchProgress float64

	speed  float64
	cursor int // first element whose footprint is not yet fully behind

	stopped  bool // forward motion halted by a hit
	finished bool // the finish has been crossed
}

// New returns a runner for level with movement tuning cfg. The hero starts on
// the center lane at z=0 on flat ground.
func New(level *levelgen.Level, cfg Config) *Runner {
	cfg = sanitizeConfig(cfg)
	return &Runner{
		level: level,
		cfg:   cfg,
		elems: level.Elements(0, level.Length()),
		lane:  1,
		speed: cfg.BaseSpeed,
	}
}

// Step advances the simulation by dt seconds with input in and returns every
// event that happened this step, in the order they happened. Passing the same
// input script and step durations always yields the same state and events.
func (r *Runner) Step(in Input, dt float64) []Event {
	if r.stopped || r.finished {
		return nil
	}
	if dt <= 0 {
		return nil
	}
	var events []Event

	r.handleInput(in, &events)

	speed := r.speedFor(r.z)
	nextZ := r.z + speed*dt
	if nextZ >= r.level.FinishZ() {
		r.z = r.level.FinishZ()
		r.speed = 0
		r.finished = true
		events = append(events, EventFinished)
		return events
	}
	r.z = nextZ
	r.speed = r.speedFor(r.z)

	r.stepLane(dt)
	r.resolve(&events)
	r.stepVertical(dt)
	r.tickTimers(dt)

	return events
}

// Lane returns the hero's lane as a continuous index in [0, 2]: 0 is left, 1
// center, 2 right. It moves between integers only during a lane switch.
func (r *Runner) Lane() float64 { return r.lane }

// Z returns the hero's front edge along the corridor, in meters.
func (r *Runner) Z() float64 { return r.z }

// Height returns the hero's feet height above the current ground, in meters.
func (r *Runner) Height() float64 { return r.height }

// OnContainer reports whether the hero currently rides a container ramp.
func (r *Runner) OnContainer() bool { return r.onContainer }

// Speed returns the current forward speed, in meters per second.
func (r *Runner) Speed() float64 { return r.speed }

// Sliding reports whether the hero is in a grounded slide.
func (r *Runner) Sliding() bool { return r.sliding }

// Slamming reports whether the hero is slamming downward from the air.
func (r *Runner) Slamming() bool { return r.slamming }

// Airborne reports whether the hero is off the ground: jumping, slamming or
// falling.
func (r *Runner) Airborne() bool { return r.air }

// Restart resets the runner so a run resumes cleanly from startZ: lane back
// to center, height and vertical motion to zero, speed to the base speed, all
// timers cleared, forward motion un-stopped and Z moved to startZ. It is the
// primitive the gameplay layer uses to respawn the hero at a checkpoint.
func (r *Runner) Restart(startZ float64) {
	r.z = startZ
	r.lane = 1
	r.height = 0
	r.vy = 0
	r.air = false
	r.sliding = false
	r.slideTimer = 0
	r.slamming = false
	r.slamLock = 0
	r.onContainer = false
	r.switching = false
	r.switchFrom = 0
	r.switchTarget = 0
	r.switchProgress = 0
	r.speed = r.cfg.BaseSpeed
	r.stopped = false
	r.finished = false

	r.cursor = 0
	for r.cursor < len(r.elems) {
		_, end := spanOf(r.elems[r.cursor])
		if end <= startZ {
			r.cursor++
			continue
		}
		break
	}
}

func sanitizeConfig(c Config) Config {
	d := DefaultConfig()
	if c.BaseSpeed <= 0 {
		c.BaseSpeed = d.BaseSpeed
	}
	if c.MaxSpeed <= 0 {
		c.MaxSpeed = d.MaxSpeed
	}
	if c.SpeedRampDistance <= 0 {
		c.SpeedRampDistance = d.SpeedRampDistance
	}
	if c.LaneSwitchDuration <= 0 {
		c.LaneSwitchDuration = d.LaneSwitchDuration
	}
	if c.JumpVelocity <= 0 {
		c.JumpVelocity = d.JumpVelocity
	}
	if c.Gravity <= 0 {
		c.Gravity = d.Gravity
	}
	if c.SlideDuration <= 0 {
		c.SlideDuration = d.SlideDuration
	}
	if c.SlamVelocity <= 0 {
		c.SlamVelocity = d.SlamVelocity
	}
	if c.SlamLockout <= 0 {
		c.SlamLockout = d.SlamLockout
	}
	if c.MaxSpeed < c.BaseSpeed {
		c.MaxSpeed = c.BaseSpeed
	}
	return c
}

func (r *Runner) speedFor(z float64) float64 {
	base := r.cfg.BaseSpeed
	maxSpeed := r.cfg.MaxSpeed
	if maxSpeed <= base {
		return base
	}
	t := z / r.cfg.SpeedRampDistance
	if t > 1 {
		t = 1
	}
	return base + (maxSpeed-base)*t
}

func (r *Runner) handleInput(in Input, events *[]Event) {
	if !r.switching {
		if in.Left && r.lane > 0 {
			r.startSwitch(r.lane - 1)
		} else if in.Right && r.lane < 2 {
			r.startSwitch(r.lane + 1)
		}
	}

	if in.Slide {
		if r.air {
			if !r.slamming {
				r.slamming = true
				r.vy = -r.cfg.SlamVelocity
				r.slamLock = r.cfg.SlamLockout
				r.sliding = false
				*events = append(*events, EventSlam)
			}
		} else if r.slamLock <= 0 {
			if !r.sliding {
				r.sliding = true
				*events = append(*events, EventSlide)
			}
			r.slideTimer = r.cfg.SlideDuration
		}
	}

	if in.Jump && !r.air && !r.sliding && r.slamLock <= 0 {
		r.air = true
		r.vy = r.cfg.JumpVelocity
		r.slamming = false
		*events = append(*events, EventJump)
	}
}

func (r *Runner) startSwitch(target float64) {
	r.switching = true
	r.switchFrom = r.lane
	r.switchTarget = target
	r.switchProgress = 0
}

func (r *Runner) stepLane(dt float64) {
	if !r.switching {
		return
	}
	r.switchProgress += dt / r.cfg.LaneSwitchDuration
	if r.switchProgress >= 1 {
		r.lane = r.switchTarget
		r.switching = false
		r.switchProgress = 0
		return
	}
	r.lane = r.switchFrom + (r.switchTarget-r.switchFrom)*r.switchProgress
}

func (r *Runner) stepVertical(dt float64) {
	if !r.air {
		r.height = r.groundHeight()
		return
	}
	r.vy -= r.cfg.Gravity * dt
	r.height += r.vy * dt
	ground := r.groundHeight()
	if r.height <= ground {
		r.height = ground
		r.vy = 0
		r.air = false
		r.slamming = false
	}
}

func (r *Runner) tickTimers(dt float64) {
	if r.slamLock > 0 {
		r.slamLock -= dt
		if r.slamLock < 0 {
			r.slamLock = 0
		}
	}
	if r.sliding {
		r.slideTimer -= dt
		if r.slideTimer <= 0 {
			r.sliding = false
			r.slideTimer = 0
		}
	}
}

func (r *Runner) groundHeight() float64 {
	if r.onContainer {
		return containerHeight
	}
	return 0
}

func (r *Runner) resolve(events *[]Event) {
	r.resolveRampTransition(events)

	for r.cursor < len(r.elems) {
		e := r.elems[r.cursor]
		start, end := spanOf(e)
		if end <= r.z {
			r.cursor++
			continue
		}
		if r.z < start {
			break
		}
		switch e.Kind {
		case levelgen.Pit:
			if !r.air {
				*events = append(*events, EventFellInPit)
				r.stopped = true
				r.speed = 0
			}
			// A pit spans every lane and is cleared by jumping. While the
			// hero is airborne over it, keep the cursor here so a landing
			// inside the pit is caught on a later step.
			return
		case levelgen.LowBarrier, levelgen.HighBarrier:
			r.resolveBarrier(e, events)
			r.cursor++
		case levelgen.Enemy:
			r.resolveEnemy(e, events)
			r.cursor++
		default:
			// Pickups and ramps never collide. Ramp entry and exit are
			// handled by resolveRampTransition.
			r.cursor++
		}
		if r.stopped {
			return
		}
	}
}

func (r *Runner) resolveBarrier(e levelgen.Element, events *[]Event) {
	if laneOf(r.lane) != e.Lane {
		return
	}
	switch e.Kind {
	case levelgen.LowBarrier:
		if r.air && r.height > lowBarrierHeight {
			return
		}
	case levelgen.HighBarrier:
		if r.height+r.profileHeight() < highBarrierClearance {
			return
		}
	}
	*events = append(*events, EventHitBarrier)
	r.stopped = true
	r.speed = 0
}

func (r *Runner) resolveEnemy(e levelgen.Element, events *[]Event) {
	if laneOf(r.lane) != e.Lane {
		return
	}
	if r.air && r.height >= enemyHeight {
		*events = append(*events, EventStomp)
		return
	}
	*events = append(*events, EventHitEnemy)
	r.stopped = true
	r.speed = 0
}

func (r *Runner) resolveRampTransition(events *[]Event) {
	if r.onRamp(laneOf(r.lane), r.z) {
		if !r.onContainer {
			r.onContainer = true
			*events = append(*events, EventRampUp)
		}
		return
	}
	if r.onContainer {
		r.onContainer = false
		*events = append(*events, EventRampDown)
	}
}

func (r *Runner) onRamp(lane int, z float64) bool {
	for _, e := range r.elems {
		if e.Kind == levelgen.Ramp && e.Lane == lane && z >= e.Z && z < e.Z+e.Span {
			return true
		}
	}
	return false
}

func (r *Runner) profileHeight() float64 {
	if r.sliding {
		return slideHeight
	}
	return bodyHeight
}

// laneOf returns the lane the hero currently centers on, used for collisions
// and ramp detection. During a switch it moves to the target lane once the
// hero passes the halfway point.
func laneOf(lane float64) int {
	return int(math.Round(lane))
}

// spanOf returns the half-open [start, end) z interval an element occupies.
func spanOf(e levelgen.Element) (float64, float64) {
	switch e.Kind {
	case levelgen.Pit, levelgen.Ramp:
		return e.Z, e.Z + e.Span
	case levelgen.LowBarrier, levelgen.HighBarrier, levelgen.Enemy:
		return e.Z, e.Z + barrierDepth
	default:
		return e.Z, e.Z
	}
}

// Geometry the runner assumes about the world. The levelgen package keeps its
// own copies private, so these constants state the runner's half of the
// contract. barrierDepth matches levelgen's barrier and enemy footprint.
const (
	bodyHeight           = 1.8
	slideHeight          = 0.8
	lowBarrierHeight     = 0.7
	highBarrierClearance = 1.2
	enemyHeight          = 1.0
	containerHeight      = 1.5
	barrierDepth         = 1.0
)
