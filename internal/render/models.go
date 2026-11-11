package render

import (
	"math"

	"copycat/internal/campaign"
)

// HeroPose is one of Downald D's canonical body poses. The caller maps the
// runner's state onto it; BuildHero turns the pose into a command list.
type HeroPose int

// The poses BuildHero understands.
const (
	// PoseIdle is the standing pose.
	PoseIdle HeroPose = iota
	// PoseRun is the running pose with swinging limbs and tail.
	PoseRun
	// PoseJump is the airborne pose with tucked legs.
	PoseJump
	// PoseSlide is the low, tucked slide.
	PoseSlide
	// PoseSlam is the stretched slam.
	PoseSlam
	// PoseLanding is the crouched landing after a jump or slam.
	PoseLanding
	// PoseStomp is the wide-legged airborne stomp.
	PoseStomp
	// PoseSprint is the fast run with a longer, faster stride.
	PoseSprint
)

// CopycatPose is the stance BuildCopycatAt draws.
type CopycatPose int

// The stances BuildCopycatAt understands.
const (
	// CopycatPoseIdle is the ordinary chase stance.
	CopycatPoseIdle CopycatPose = iota
	// CopycatPoseWindup is the attack wind-up with snapping scissors.
	CopycatPoseWindup
)

// BossStance is the stance BuildBossAt draws.
type BossStance int

// The stances BuildBossAt understands.
const (
	// BossStanceIdle is the ordinary phase stance.
	BossStanceIdle BossStance = iota
	// BossStanceWindup is the low-swipe attack wind-up.
	BossStanceWindup
	// BossStanceJump is the high-swipe jump/rear.
	BossStanceJump
)

// Downald D's part colors, kept fixed so the hero always reads the same.
var (
	heroBody  = Color{R: 190, G: 96, B: 46, A: 255}  // reddish proboscis monkey fur
	heroBelly = Color{R: 244, G: 220, B: 190, A: 255} // pale belly
	heroLimb  = Color{R: 118, G: 118, B: 128, A: 255} // gray arms and legs
	heroTail  = Color{R: 246, G: 246, B: 240, A: 255} // white tail
	heroNose  = Color{R: 210, G: 120, B: 92, A: 255}  // huge dangling nose
	heroHair  = Color{R: 240, G: 202, B: 112, A: 255} // light swept hairstyle
)

// Copycat and his boss form share the barber look: a dark cat, a pale apron
// and silver scissors; the boss adds red anger accents.
var (
	catFur     = Color{R: 70, G: 70, B: 90, A: 255}
	catApron   = Color{R: 236, G: 236, B: 242, A: 255}
	catScissor = Color{R: 198, G: 200, B: 208, A: 255}
	bossFur    = Color{R: 56, G: 48, B: 76, A: 255}
	bossAnger  = Color{R: 210, G: 56, B: 56, A: 255}
)

// Enemy species colors, one per species so the themes and the tests can tell
// them apart.
var (
	mudskipperBody = Color{R: 96, G: 118, B: 64, A: 255}
	mudskipperEye  = Color{R: 250, G: 250, B: 244, A: 255}
	gullBody       = Color{R: 232, G: 232, B: 236, A: 255}
	gullWing       = Color{R: 168, G: 176, B: 188, A: 255}
	gullOrange     = Color{R: 226, G: 162, B: 52, A: 255}
	ratFur         = Color{R: 128, G: 112, B: 100, A: 255}
	ratEar         = Color{R: 178, G: 150, B: 130, A: 255}
	ratNose        = Color{R: 216, G: 130, B: 128, A: 255}
	chairSeat      = Color{R: 200, G: 40, B: 90, A: 255}
	chairMetal     = Color{R: 188, G: 190, B: 198, A: 255}
	vacuumBody     = Color{R: 70, G: 200, B: 190, A: 255}
	vacuumAccent   = Color{R: 30, G: 40, B: 60, A: 255}
)

// Enemy species indices, in the order the campaign worlds meet them.
const (
	enemyMudskipper = iota
	enemyGull
	enemyRat
	enemyChair
	enemyVacuum
)

// BuildHero returns Downald D's rounded model at the first animation frame.
// The hero stands at the origin with his feet at y = 0 and faces +z, so the
// caller can translate the list onto his lane and height. pose picks the
// pose; bald removes only the hairstyle.
func BuildHero(pose HeroPose, bald bool) []Cmd {
	return BuildHeroAt(pose, bald, 0)
}

// BuildHeroAt returns Downald D's rounded, animated model. t is the
// animation phase in seconds; whole seconds repeat, so 0 and 1 draw the same
// frame. pose picks the pose and bald removes only the hairstyle.
func BuildHeroAt(pose HeroPose, bald bool, t float64) []Cmd {
	s := heroSpecFor(pose)
	animateHero(&s, pose, t)
	cs := heroTorsoCmds(s)
	cs = append(cs, heroLimbCmds(s)...)
	cs = append(cs, heroTailCmds(s)...)
	cs = append(cs, heroNoseCmds(s)...)
	if !bald {
		cs = append(cs, heroHairCmds(s)...)
	}
	return cs
}

// heroSpec carries the knobs that make each hero pose read differently.
type heroSpec struct {
	bodyY     float32
	bodyH     float32
	headY     float32
	legAhead  float32
	legUp     float32
	footOut   float32
	footAhead float32
	armSwing  float32
	armOut    float32
	handOut   float32
	handSwing float32
	tailY     float32
	tailZ     float32
	hairSweep float32
	hairDip   float32
}

func heroSpecFor(pose HeroPose) heroSpec {
	switch pose {
	case PoseRun:
		return heroSpec{
			bodyY: 0.95, bodyH: 0.8, headY: 1.5,
			legAhead: 0.22, footAhead: 0.06, armSwing: 0.18, armOut: 0.04, handSwing: 0.06,
			tailY: 0.72, tailZ: -0.68,
			hairSweep: 0.12, hairDip: 0.05,
		}
	case PoseSprint:
		return heroSpec{
			bodyY: 0.9, bodyH: 0.78, headY: 1.46,
			legAhead: 0.3, footAhead: 0.1, armSwing: 0.26, armOut: 0.06, handSwing: 0.1,
			tailY: 0.75, tailZ: -0.72,
			hairSweep: 0.18, hairDip: 0.08,
		}
	case PoseJump:
		return heroSpec{
			bodyY: 0.98, bodyH: 0.72, headY: 1.5,
			legAhead: 0.06, legUp: 0.2, footAhead: 0.08, armSwing: 0.1, armOut: 0.12, handSwing: 0.08,
			tailY: 0.98, tailZ: -0.4,
			hairSweep: 0.18, hairDip: 0.08,
		}
	case PoseSlide:
		return heroSpec{
			bodyY: 0.6, bodyH: 0.5, headY: 1.05,
			legAhead: 0.3, legUp: 0.12, footAhead: 0.14, armSwing: 0.2, armOut: 0.05, handSwing: 0.1,
			tailY: 0.4, tailZ: -0.9,
			hairSweep: 0.22, hairDip: 0.12,
		}
	case PoseSlam:
		return heroSpec{
			bodyY: 1.1, bodyH: 0.95, headY: 1.62,
			legAhead: 0.16, armSwing: 0.3, armOut: 0.18, handSwing: 0.14,
			tailY: 1.05, tailZ: -0.5,
			hairSweep: 0.26, hairDip: 0.16,
		}
	case PoseLanding:
		return heroSpec{
			bodyY: 0.85, bodyH: 0.75, headY: 1.4,
			legAhead: 0.1, legUp: 0.08, footOut: 0.04, footAhead: 0.04,
			armSwing: 0.12, armOut: 0.06, handSwing: 0.06,
			tailY: 0.7, tailZ: -0.6,
			hairSweep: 0.08, hairDip: 0.06,
		}
	case PoseStomp:
		return heroSpec{
			bodyY: 1.15, bodyH: 1.0, headY: 1.65,
			legAhead: 0.24, footOut: 0.06, footAhead: 0.1,
			armSwing: 0.34, armOut: 0.2, handOut: 0.05, handSwing: 0.16,
			tailY: 1.1, tailZ: -0.45,
			hairSweep: 0.3, hairDip: 0.18,
		}
	default: // PoseIdle and any unknown pose.
		return heroSpec{
			bodyY: 0.95, bodyH: 0.8, headY: 1.5,
			tailY: 0.8, tailZ: -0.55,
		}
	}
}

// animateHero oscillates the idle, run and sprint poses with the phase t so
// the hero never freezes. The idle pose breathes; the run and sprint poses
// swing their limbs and tail, with a longer, faster stride when sprinting.
// Other poses keep their fixed silhouettes.
func animateHero(s *heroSpec, pose HeroPose, t float64) {
	phase := t - math.Floor(t)
	if phase < 0 {
		phase++
	}
	switch pose {
	case PoseIdle:
		breath := float32(math.Sin(2 * math.Pi * phase))
		s.bodyY += 0.02 * breath
		s.headY += 0.03 * breath
		s.tailZ += 0.05 * breath
	case PoseRun, PoseSprint:
		swing := float32(math.Sin(2 * math.Pi * phase))
		amp := float32(1)
		if pose == PoseSprint {
			amp = 1.4
		}
		s.legAhead += 0.14 * amp * swing
		s.armSwing += 0.12 * amp * swing
		s.tailZ += 0.06 * amp * swing
		s.hairSweep += 0.04 * amp * swing
	}
}

func heroTorsoCmds(s heroSpec) []Cmd {
	halfH := s.bodyH / 2
	return []Cmd{
		{Shape: ShapeCylinder, Pos: Vec3{X: 0, Y: s.bodyY, Z: 0}, Size: Size{X: 0.34, Y: float64(s.bodyH)}, Color: heroBody},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.bodyY + halfH, Z: 0}, Size: Size{X: 0.3}, Color: heroBody},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.bodyY - halfH + 0.03, Z: 0}, Size: Size{X: 0.3}, Color: heroBody},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.bodyY - 0.02, Z: 0.26}, Size: Size{X: 0.22}, Color: heroBelly},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.headY, Z: 0.02}, Size: Size{X: 0.26}, Color: heroBody},
	}
}

func heroLimbCmds(s heroSpec) []Cmd {
	var cs []Cmd
	legH := 0.55 - s.legUp*2
	if legH < 0.1 {
		legH = 0.1
	}
	legY := 0.28 + s.legUp
	for i, dz := range []float32{-s.legAhead, s.legAhead} {
		x := float32(-0.16)
		dx := float32(-1)
		if i == 1 {
			x = 0.16
			dx = 1
		}
		cs = append(cs, Cmd{
			Shape: ShapeCylinder,
			Pos:   Vec3{X: x, Y: legY, Z: dz},
			Size:  Size{X: 0.07, Y: float64(legH)},
			Color: heroLimb,
		})
		footY := legY - legH/2 + 0.03
		cs = append(cs, Cmd{
			Shape: ShapeSphere,
			Pos:   Vec3{X: x + dx*s.footOut, Y: footY, Z: dz + s.footAhead},
			Size:  Size{X: 0.06},
			Color: heroLimb,
		})
	}
	for i, dz := range []float32{-s.armSwing, s.armSwing} {
		x := 0.34 + s.armOut
		dx := float32(1)
		if i == 0 {
			x = -x
			dx = -1
		}
		cs = append(cs, Cmd{
			Shape: ShapeCylinder,
			Pos:   Vec3{X: x, Y: 1.0, Z: dz},
			Size:  Size{X: 0.06, Y: 0.45},
			Color: heroLimb,
		})
		handY := float32(1.0 - 0.225 + 0.02)
		cs = append(cs, Cmd{
			Shape: ShapeSphere,
			Pos:   Vec3{X: x + dx*s.handOut, Y: handY, Z: dz + s.handSwing},
			Size:  Size{X: 0.05},
			Color: heroLimb,
		})
	}
	return cs
}

func heroTailCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY, Z: s.tailZ}, Size: Size{X: 0.1}, Color: heroTail},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY + 0.02, Z: s.tailZ - 0.18}, Size: Size{X: 0.07}, Color: heroTail},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY + 0.04, Z: s.tailZ - 0.3}, Size: Size{X: 0.05}, Color: heroTail},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY + 0.03, Z: s.tailZ - 0.42}, Size: Size{X: 0.04}, Color: heroTail},
	}
}

func heroNoseCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeCone, Pos: Vec3{X: 0, Y: s.headY - 0.28, Z: 0.32}, Size: Size{X: 0.13, Y: 0.55}, Color: heroNose},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.headY - 0.55, Z: 0.32}, Size: Size{X: 0.06}, Color: heroNose},
	}
}

func heroHairCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.headY + 0.22, Z: 0.02}, Size: Size{X: 0.14}, Color: heroHair},
		{Shape: ShapeCone, Pos: Vec3{X: 0, Y: s.headY + 0.2, Z: -0.08 - s.hairSweep}, Size: Size{X: 0.12, Y: 0.3}, Color: heroHair},
		{Shape: ShapeCone, Pos: Vec3{X: 0, Y: s.headY + 0.16 - s.hairDip, Z: -0.22 - s.hairSweep}, Size: Size{X: 0.1, Y: 0.26}, Color: heroHair},
		{Shape: ShapeCone, Pos: Vec3{X: 0.06 + s.hairSweep, Y: s.headY + 0.12, Z: -0.3 - s.hairSweep}, Size: Size{X: 0.08, Y: 0.2}, Color: heroHair},
	}
}

// BuildEnemy returns the enemy model for theme at pos, with its feet at y = 0.
// Mangroves always send a mudskipper; the harbor alternates gulls and rats and
// the neon street alternates rolling barber chairs and robot vacuums by the
// enemy's z position, so both species of a world show up.
func BuildEnemy(theme campaign.Theme, pos Vec3) []Cmd {
	switch enemySpecies(theme, pos) {
	case enemyGull:
		return gullCmds(pos)
	case enemyRat:
		return ratCmds(pos)
	case enemyChair:
		return chairCmds(pos)
	case enemyVacuum:
		return vacuumCmds(pos)
	default:
		return mudskipperCmds(pos)
	}
}

func enemySpecies(theme campaign.Theme, pos Vec3) int {
	switch theme {
	case campaign.ThemeHarbor:
		if zIndex(pos.Z)%2 == 0 {
			return enemyGull
		}
		return enemyRat
	case campaign.ThemeNeon:
		if zIndex(pos.Z)%2 == 0 {
			return enemyChair
		}
		return enemyVacuum
	default:
		return enemyMudskipper
	}
}

func zIndex(z float32) int {
	i := int(z)
	if i < 0 {
		i = -i
	}
	return i
}

func mudskipperCmds(pos Vec3) []Cmd {
	x, z := pos.X, pos.Z
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.12, Z: z}, Size: Size{X: 0.55, Y: 0.24, Z: 0.85}, Color: mudskipperBody},
		{Shape: ShapeSphere, Pos: Vec3{X: x - 0.14, Y: 0.34, Z: z + 0.2}, Size: Size{X: 0.09}, Color: mudskipperEye},
		{Shape: ShapeSphere, Pos: Vec3{X: x + 0.14, Y: 0.34, Z: z + 0.2}, Size: Size{X: 0.09}, Color: mudskipperEye},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.12, Z: z - 0.5}, Size: Size{X: 0.3, Y: 0.06, Z: 0.3}, Color: mudskipperBody},
	}
}

func gullCmds(pos Vec3) []Cmd {
	x, z := pos.X, pos.Z
	y := float32(0.55)
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: y, Z: z}, Size: Size{X: 0.4, Y: 0.34, Z: 0.7}, Color: gullBody},
		{Shape: ShapeBox, Pos: Vec3{X: x - 0.45, Y: y + 0.05, Z: z}, Size: Size{X: 0.5, Y: 0.06, Z: 0.34}, Color: gullWing},
		{Shape: ShapeBox, Pos: Vec3{X: x + 0.45, Y: y + 0.05, Z: z}, Size: Size{X: 0.5, Y: 0.06, Z: 0.34}, Color: gullWing},
		{Shape: ShapeCone, Pos: Vec3{X: x, Y: y + 0.05, Z: z + 0.4}, Size: Size{X: 0.07, Y: 0.22}, Color: gullOrange},
		{Shape: ShapeBox, Pos: Vec3{X: x - 0.1, Y: 0.15, Z: z}, Size: Size{X: 0.05, Y: 0.3, Z: 0.05}, Color: gullOrange},
		{Shape: ShapeBox, Pos: Vec3{X: x + 0.1, Y: 0.15, Z: z}, Size: Size{X: 0.05, Y: 0.3, Z: 0.05}, Color: gullOrange},
	}
}

func ratCmds(pos Vec3) []Cmd {
	x, z := pos.X, pos.Z
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.18, Z: z}, Size: Size{X: 0.3, Y: 0.28, Z: 0.6}, Color: ratFur},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.3, Z: z + 0.32}, Size: Size{X: 0.2, Y: 0.18, Z: 0.2}, Color: ratFur},
		{Shape: ShapeCone, Pos: Vec3{X: x - 0.12, Y: 0.52, Z: z + 0.34}, Size: Size{X: 0.06, Y: 0.16}, Color: ratEar},
		{Shape: ShapeCone, Pos: Vec3{X: x + 0.12, Y: 0.52, Z: z + 0.34}, Size: Size{X: 0.06, Y: 0.16}, Color: ratEar},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.3, Z: z + 0.44}, Size: Size{X: 0.08, Y: 0.06, Z: 0.06}, Color: ratNose},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.16, Z: z - 0.5}, Size: Size{X: 0.05, Y: 0.04, Z: 0.5}, Color: ratFur},
	}
}

func chairCmds(pos Vec3) []Cmd {
	x, z := pos.X, pos.Z
	return []Cmd{
		{Shape: ShapeCylinder, Pos: Vec3{X: x, Y: 0.28, Z: z}, Size: Size{X: 0.12, Y: 0.56}, Color: chairMetal},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.62, Z: z}, Size: Size{X: 0.6, Y: 0.16, Z: 0.6}, Color: chairSeat},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 1.02, Z: z - 0.28}, Size: Size{X: 0.6, Y: 0.7, Z: 0.12}, Color: chairSeat},
		{Shape: ShapeSphere, Pos: Vec3{X: x - 0.25, Y: 0.08, Z: z + 0.22}, Size: Size{X: 0.08}, Color: chairMetal},
		{Shape: ShapeSphere, Pos: Vec3{X: x + 0.25, Y: 0.08, Z: z + 0.22}, Size: Size{X: 0.08}, Color: chairMetal},
		{Shape: ShapeSphere, Pos: Vec3{X: x - 0.25, Y: 0.08, Z: z - 0.22}, Size: Size{X: 0.08}, Color: chairMetal},
		{Shape: ShapeSphere, Pos: Vec3{X: x + 0.25, Y: 0.08, Z: z - 0.22}, Size: Size{X: 0.08}, Color: chairMetal},
	}
}

func vacuumCmds(pos Vec3) []Cmd {
	x, z := pos.X, pos.Z
	return []Cmd{
		{Shape: ShapeCylinder, Pos: Vec3{X: x, Y: 0.1, Z: z}, Size: Size{X: 0.4, Y: 0.2}, Color: vacuumBody},
		{Shape: ShapeBox, Pos: Vec3{X: x, Y: 0.22, Z: z}, Size: Size{X: 0.5, Y: 0.06, Z: 0.5}, Color: vacuumAccent},
		{Shape: ShapeSphere, Pos: Vec3{X: x, Y: 0.26, Z: z + 0.12}, Size: Size{X: 0.05}, Color: vacuumAccent},
	}
}

// windAmt maps the animation clock to a smooth [0, 1] wind-up phase that
// repeats once per second: 0 at the start, 1 at the peak and 0 again at the
// end of the cycle.
func windAmt(t float64) float32 {
	phase := t - math.Floor(t)
	if phase < 0 {
		phase++
	}
	return float32(math.Sin(math.Pi * phase))
}

// hopAmt maps the animation clock to a signed [-1, 1] vertical hop that
// repeats once per second.
func hopAmt(t float64) float32 {
	phase := t - math.Floor(t)
	if phase < 0 {
		phase++
	}
	return float32(math.Sin(2 * math.Pi * phase))
}

// BuildCopycat returns the Copycat chaser model. It is the same barber cat
// drawn behind the hero: every command sits at z = -chase, so a larger chase
// distance places him farther behind the hero at the origin.
func BuildCopycat(chase float64) []Cmd {
	return BuildCopycatAt(chase, CopycatPoseIdle, 0)
}

// BuildCopycatAt returns the Copycat chaser model at animation clock t in
// seconds. In the wind-up stance the scissors lunge forward and snap with a
// smooth per-second phase.
func BuildCopycatAt(chase float64, pose CopycatPose, t float64) []Cmd {
	z := float32(-chase)
	bodyY := float32(0.7)
	headY := float32(1.15)
	armY := float32(1.0)
	scissorZ := float32(0.15)
	if pose == CopycatPoseWindup {
		wind := windAmt(t)
		bodyY = 0.71 + 0.02*wind
		headY = 1.16 + 0.03*wind
		armY = 1.08 + 0.08*wind
		scissorZ = 0.2 + 0.12*wind
	}
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: bodyY, Z: z}, Size: Size{X: 0.6, Y: 0.7, Z: 0.4}, Color: catFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: bodyY - 0.05, Z: z + 0.21}, Size: Size{X: 0.5, Y: 0.6, Z: 0.1}, Color: catApron},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: headY, Z: z}, Size: Size{X: 0.4, Y: 0.32, Z: 0.36}, Color: catFur},
		{Shape: ShapeCone, Pos: Vec3{X: -0.13, Y: headY + 0.35, Z: z}, Size: Size{X: 0.09, Y: 0.22}, Color: catFur},
		{Shape: ShapeCone, Pos: Vec3{X: 0.13, Y: headY + 0.35, Z: z}, Size: Size{X: 0.09, Y: 0.22}, Color: catFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: bodyY + 0.05, Z: z - 0.35}, Size: Size{X: 0.08, Y: 0.08, Z: 0.5}, Color: catFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0.28, Y: armY, Z: z + scissorZ}, Size: Size{X: 0.06, Y: 0.02, Z: 0.34}, Color: catScissor},
		{Shape: ShapeBox, Pos: Vec3{X: 0.22, Y: armY, Z: z + scissorZ}, Size: Size{X: 0.06, Y: 0.02, Z: 0.34}, Color: catScissor},
		{Shape: ShapeBox, Pos: Vec3{X: 0.25, Y: armY, Z: z + scissorZ + 0.17}, Size: Size{X: 0.08, Y: 0.05, Z: 0.06}, Color: catScissor},
	}
}

// BuildBoss returns the Copycat boss model. It is the same cat drawn larger
// and angrier, with a stance that changes across phases 1 through 3 and a
// crouch when vulnerable is true.
func BuildBoss(phase int, vulnerable bool) []Cmd {
	return BuildBossAt(phase, vulnerable, BossStanceIdle, 0)
}

// BuildBossAt returns the Copycat boss model at animation clock t in seconds.
// stance adds an attack wind-up or a high-swipe jump on top of the phase and
// vulnerable stance.
func BuildBossAt(phase int, vulnerable bool, stance BossStance, t float64) []Cmd {
	bodyY := float32(1.25)
	headY := float32(1.9)
	armY := float32(1.5)
	armZ := float32(0.32)
	switch phase {
	case 2:
		armY = 1.7
		armZ = 0.55
	case 3:
		armY = 1.9
		armZ = 0.8
	}
	if vulnerable {
		bodyY -= 0.45
		headY -= 0.45
		armY -= 0.45
	}
	switch stance {
	case BossStanceWindup:
		wind := windAmt(t)
		armY += 0.05 + 0.05*wind
		armZ += 0.15 + 0.15*wind
	case BossStanceJump:
		hop := hopAmt(t)
		lift := 0.25 + 0.08*hop
		bodyY += lift
		headY += lift
		armY += lift + 0.1
	}
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: bodyY, Z: 0}, Size: Size{X: 0.95, Y: 1.1, Z: 0.65}, Color: bossFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: bodyY - 0.08, Z: 0.34}, Size: Size{X: 0.8, Y: 0.95, Z: 0.16}, Color: catApron},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: headY, Z: 0}, Size: Size{X: 0.64, Y: 0.5, Z: 0.58}, Color: bossFur},
		{Shape: ShapeCone, Pos: Vec3{X: -0.2, Y: headY + 0.34, Z: 0}, Size: Size{X: 0.14, Y: 0.34}, Color: bossFur},
		{Shape: ShapeCone, Pos: Vec3{X: 0.2, Y: headY + 0.34, Z: 0}, Size: Size{X: 0.14, Y: 0.34}, Color: bossFur},
		{Shape: ShapeBox, Pos: Vec3{X: -0.17, Y: headY + 0.05, Z: 0.3}, Size: Size{X: 0.12, Y: 0.08, Z: 0.04}, Color: bossAnger},
		{Shape: ShapeBox, Pos: Vec3{X: 0.17, Y: headY + 0.05, Z: 0.3}, Size: Size{X: 0.12, Y: 0.08, Z: 0.04}, Color: bossAnger},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 0.85, Z: -0.5}, Size: Size{X: 0.12, Y: 0.12, Z: 0.7}, Color: bossFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0.42, Y: armY, Z: armZ}, Size: Size{X: 0.09, Y: 0.04, Z: 0.5}, Color: catScissor},
		{Shape: ShapeBox, Pos: Vec3{X: 0.34, Y: armY, Z: armZ}, Size: Size{X: 0.09, Y: 0.04, Z: 0.5}, Color: catScissor},
		{Shape: ShapeBox, Pos: Vec3{X: 0.38, Y: armY, Z: armZ + 0.28}, Size: Size{X: 0.12, Y: 0.08, Z: 0.09}, Color: catScissor},
	}
}
