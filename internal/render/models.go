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

// Hero proportions and animation tuning, in meters and radians per second.
// They are the fixed rounded geometry the poses animate around, so the six
// part colors above stay the only per-model palette.
const (
	heroTorsoR = 0.3  // torso capsule radius
	heroHeadR  = 0.22 // head sphere radius
	heroBellyR = 0.17 // pale belly sphere radius

	heroLegR   = 0.09 // leg capsule radius
	heroLegLen = 0.52 // standing leg length
	heroArmX   = 0.38 // each shoulder's horizontal distance from the spine
	heroArmR   = 0.07 // arm capsule radius
	heroArmLen = 0.45 // arm length
	heroArmY   = 1.0  // arm shoulder height

	runCycleRate     = 10.0 // limb swing rate, radians per second
	noseWobbleRate   = 9.0  // nose side-to-side rate, radians per second
	noseWobble       = 0.07 // nose side-to-side amplitude, meters
	hairFlutterRateY = 11.0 // hair vertical flutter rate, radians per second
	hairFlutterRateZ = 9.0  // hair backward flutter rate, radians per second
	hairFlutterY     = 0.03 // hair vertical flutter amplitude, meters
	hairFlutterZ     = 0.05 // hair backward flutter amplitude, meters
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

// BuildHero returns Downald D's model as a draw list at rest. The hero stands
// at the origin with his feet at y = 0 and faces +z, so the caller can
// translate the list onto his lane and height. pose picks the pose; bald
// removes only the hairstyle.
func BuildHero(pose HeroPose, bald bool) []Cmd {
	return BuildHeroAt(pose, bald, 0)
}

// BuildHeroAt returns Downald D's model as a draw list at animation time t,
// in seconds. It is the same rounded hero as BuildHero with the run cycle,
// nose wobble and hair flutter advanced by t; t = 0 is the rest pose
// BuildHero returns. The result is a pure, deterministic function of its
// arguments.
func BuildHeroAt(pose HeroPose, bald bool, t float64) []Cmd {
	s := heroSpecFor(pose)
	tf := float32(t)
	cs := heroTorsoCmds(s)
	cs = append(cs, heroLimbCmds(s, tf)...)
	cs = append(cs, heroTailCmds(s)...)
	cs = append(cs, heroNoseCmds(s, tf)...)
	if !bald {
		cs = append(cs, heroHairCmds(s, tf)...)
	}
	return cs
}

// BuildHeroAtFacing returns the animated hero turned to face the camera: the
// same draw list as BuildHeroAt with every command's Z negated, so the nose
// that normally points +z points at a camera placed in front of the hero.
func BuildHeroAtFacing(pose HeroPose, bald bool, t float64) []Cmd {
	cs := BuildHeroAt(pose, bald, t)
	for i := range cs {
		cs[i].Pos.Z = -cs[i].Pos.Z
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
	armSwing  float32
	armOut    float32
	tailY     float32
	tailZ     float32
	hairSweep float32
	hairDip   float32
	runCycle  bool
}

func heroSpecFor(pose HeroPose) heroSpec {
	switch pose {
	case PoseRun:
		return heroSpec{
			bodyY: 0.95, bodyH: 0.8, headY: 1.5,
			legAhead: 0.22, armSwing: 0.18, armOut: 0.04,
			tailY: 0.72, tailZ: -0.68,
			hairSweep: 0.12, hairDip: 0.05,
			runCycle:  true,
		}
	case PoseJump:
		return heroSpec{
			bodyY: 0.98, bodyH: 0.72, headY: 1.5,
			legAhead: 0.06, legUp: 0.2, armSwing: 0.1, armOut: 0.12,
			tailY: 0.98, tailZ: -0.4,
			hairSweep: 0.18, hairDip: 0.08,
		}
	case PoseSlide:
		return heroSpec{
			bodyY: 0.6, bodyH: 0.5, headY: 1.05,
			legAhead: 0.3, legUp: 0.12, armSwing: 0.2, armOut: 0.05,
			tailY: 0.4, tailZ: -0.9,
			hairSweep: 0.22, hairDip: 0.12,
		}
	case PoseSlam:
		return heroSpec{
			bodyY: 1.1, bodyH: 0.95, headY: 1.62,
			legAhead: 0.16, armSwing: 0.3, armOut: 0.18,
			tailY: 1.05, tailZ: -0.5,
			hairSweep: 0.26, hairDip: 0.16,
		}
	default: // PoseIdle and any unknown pose.
		return heroSpec{
			bodyY: 0.95, bodyH: 0.8, headY: 1.5,
			tailY: 0.8, tailZ: -0.55,
		}
	}
}

// sin1 and cos1 are the hero's two animation drivers. Every time-varying
// offset derives from them so the model stays a pure function of t.
func sin1(t float32, rate float64) float32 {
	return float32(math.Sin(float64(t) * rate))
}

func cos1(t float32, rate float64) float32 {
	return float32(math.Cos(float64(t) * rate))
}

// capsule returns a rounded limb: a vertical cylinder with a sphere at each
// end, centered at (x, y, z) and length long from end to end. The hero's
// torso, arms and legs are all built from it.
func capsule(x, y, z, length, radius float32, col Color) []Cmd {
	cylH := length - 2*radius
	if cylH < 0 {
		cylH = 0
	}
	half := cylH / 2
	return []Cmd{
		{Shape: ShapeCylinder, Pos: Vec3{X: x, Y: y, Z: z}, Size: Size{X: float64(radius), Y: float64(cylH)}, Color: col},
		{Shape: ShapeSphere, Pos: Vec3{X: x, Y: y - half, Z: z}, Size: Size{X: float64(radius)}, Color: col},
		{Shape: ShapeSphere, Pos: Vec3{X: x, Y: y + half, Z: z}, Size: Size{X: float64(radius)}, Color: col},
	}
}

func heroTorsoCmds(s heroSpec) []Cmd {
	cylH := s.bodyH - 2*heroTorsoR
	if cylH < 0 {
		cylH = 0
	}
	half := cylH / 2
	return []Cmd{
		{Shape: ShapeCylinder, Pos: Vec3{X: 0, Y: s.bodyY, Z: 0}, Size: Size{X: heroTorsoR, Y: float64(cylH)}, Color: heroBody},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.bodyY - half, Z: 0}, Size: Size{X: heroTorsoR}, Color: heroBody},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.bodyY + half, Z: 0}, Size: Size{X: heroTorsoR}, Color: heroBody},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.bodyY - 0.02, Z: 0.22}, Size: Size{X: heroBellyR}, Color: heroBelly},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.headY, Z: 0.02}, Size: Size{X: heroHeadR}, Color: heroBody},
	}
}

func heroLimbCmds(s heroSpec, t float32) []Cmd {
	var cs []Cmd

	legLen := heroLegLen - s.legUp*2
	if legLen < 0.14 {
		legLen = 0.14
	}
	legSwing := s.legAhead
	armSwing := s.armSwing
	if s.runCycle {
		legSwing = s.legAhead * sin1(t, runCycleRate)
		armSwing = s.armSwing * cos1(t, runCycleRate)
	}
	for i, dz := range []float32{-legSwing, legSwing} {
		x := float32(-0.17)
		if i == 1 {
			x = 0.17
		}
		cs = append(cs, capsule(x, legLen/2, dz, legLen, heroLegR, heroLimb)...)
	}
	for i, dz := range []float32{-armSwing, armSwing} {
		x := heroArmX + s.armOut
		if i == 0 {
			x = -x
		}
		cs = append(cs, capsule(x, heroArmY, dz, heroArmLen, heroArmR, heroLimb)...)
	}
	return cs
}

func heroTailCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY, Z: s.tailZ + 0.06}, Size: Size{X: 0.09}, Color: heroTail},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY + 0.02, Z: s.tailZ - 0.06}, Size: Size{X: 0.08}, Color: heroTail},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY + 0.05, Z: s.tailZ - 0.17}, Size: Size{X: 0.07}, Color: heroTail},
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: s.tailY + 0.09, Z: s.tailZ - 0.27}, Size: Size{X: 0.06}, Color: heroTail},
	}
}

func heroNoseCmds(s heroSpec, t float32) []Cmd {
	// The nose is a dangling proboscis: a sphere bridging to the face, a cone
	// bulging forward and down, and a round bulb at the tip. It sways from
	// side to side and bobs forward as Downald moves.
	nx := noseWobble * sin1(t, noseWobbleRate)
	nz := noseWobble * 0.4 * sin1(t, noseWobbleRate+2)
	h := s.headY
	return []Cmd{
		{Shape: ShapeSphere, Pos: Vec3{X: nx, Y: h - 0.13, Z: 0.30 + nz}, Size: Size{X: 0.14}, Color: heroNose},
		{Shape: ShapeCone, Pos: Vec3{X: nx, Y: h - 0.30, Z: 0.36 + nz}, Size: Size{X: 0.18, Y: 0.52}, Color: heroNose},
		{Shape: ShapeSphere, Pos: Vec3{X: nx, Y: h - 0.52, Z: 0.46 + nz}, Size: Size{X: 0.18}, Color: heroNose},
	}
}

func heroHairCmds(s heroSpec, t float32) []Cmd {
	// The quiff is a swept-up blond crest: a cap of spheres on the crown,
	// then cones and a tip sphere sweeping up and backward. Wind makes it bob
	// up and blow farther back.
	fy := hairFlutterY * sin1(t, hairFlutterRateY)
	fz := hairFlutterZ * sin1(t, hairFlutterRateZ)
	h := s.headY
	place := func(baseY, baseZ float32) (float32, float32) {
		return baseY + fy - s.hairDip, baseZ - s.hairSweep - fz
	}
	y, z := place(h+0.2, -0.02)
	cs := []Cmd{
		{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: y, Z: z}, Size: Size{X: 0.19}, Color: heroHair},
	}
	y, z = place(h+0.34, -0.12)
	cs = append(cs, Cmd{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: y, Z: z}, Size: Size{X: 0.14}, Color: heroHair})
	y, z = place(h+0.42, -0.24)
	cs = append(cs, Cmd{Shape: ShapeCone, Pos: Vec3{X: 0, Y: y, Z: z}, Size: Size{X: 0.12, Y: 0.3}, Color: heroHair})
	y, z = place(h+0.34, -0.38)
	cs = append(cs, Cmd{Shape: ShapeCone, Pos: Vec3{X: 0, Y: y, Z: z}, Size: Size{X: 0.09, Y: 0.26}, Color: heroHair})
	y, z = place(h+0.5, -0.34)
	cs = append(cs, Cmd{Shape: ShapeSphere, Pos: Vec3{X: 0, Y: y, Z: z}, Size: Size{X: 0.09}, Color: heroHair})
	return cs
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

// BuildCopycat returns the Copycat chaser model. It is the same barber cat
// drawn behind the hero: every command sits at z = -chase, so a larger chase
// distance places him farther behind the hero at the origin.
func BuildCopycat(chase float64) []Cmd {
	z := float32(-chase)
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 0.7, Z: z}, Size: Size{X: 0.6, Y: 0.7, Z: 0.4}, Color: catFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 0.65, Z: z + 0.21}, Size: Size{X: 0.5, Y: 0.6, Z: 0.1}, Color: catApron},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 1.15, Z: z}, Size: Size{X: 0.4, Y: 0.32, Z: 0.36}, Color: catFur},
		{Shape: ShapeCone, Pos: Vec3{X: -0.13, Y: 1.5, Z: z}, Size: Size{X: 0.09, Y: 0.22}, Color: catFur},
		{Shape: ShapeCone, Pos: Vec3{X: 0.13, Y: 1.5, Z: z}, Size: Size{X: 0.09, Y: 0.22}, Color: catFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 0.75, Z: z - 0.35}, Size: Size{X: 0.08, Y: 0.08, Z: 0.5}, Color: catFur},
		{Shape: ShapeBox, Pos: Vec3{X: 0.28, Y: 1.0, Z: z + 0.15}, Size: Size{X: 0.06, Y: 0.02, Z: 0.34}, Color: catScissor},
		{Shape: ShapeBox, Pos: Vec3{X: 0.22, Y: 1.0, Z: z + 0.15}, Size: Size{X: 0.06, Y: 0.02, Z: 0.34}, Color: catScissor},
		{Shape: ShapeBox, Pos: Vec3{X: 0.25, Y: 1.0, Z: z + 0.32}, Size: Size{X: 0.08, Y: 0.05, Z: 0.06}, Color: catScissor},
	}
}

// BuildBoss returns the Copycat boss model. It is the same cat drawn larger
// and angrier, with a stance that changes across phases 1 through 3 and a
// crouch when vulnerable is true.
func BuildBoss(phase int, vulnerable bool) []Cmd {
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
