package render

import "copycat/internal/campaign"

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

// BuildHero returns Downald D's model as a draw list. The hero stands at the
// origin with his feet at y = 0 and faces +z, so the caller can translate the
// list onto his lane and height. pose picks the pose; bald removes only the
// hairstyle.
func BuildHero(pose HeroPose, bald bool) []Cmd {
	s := heroSpecFor(pose)
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
	armSwing  float32
	armOut    float32
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
			legAhead: 0.22, armSwing: 0.18, armOut: 0.04,
			tailY: 0.72, tailZ: -0.68,
			hairSweep: 0.12, hairDip: 0.05,
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

func heroTorsoCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.bodyY, Z: 0}, Size: Size{X: 0.7, Y: float64(s.bodyH), Z: 0.45}, Color: heroBody},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.bodyY - 0.05, Z: 0.23}, Size: Size{X: 0.5, Y: float64(s.bodyH - 0.25), Z: 0.12}, Color: heroBelly},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.headY, Z: 0.02}, Size: Size{X: 0.44, Y: 0.34, Z: 0.4}, Color: heroBody},
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
		x := float32(-0.18)
		if i == 1 {
			x = 0.18
		}
		cs = append(cs, Cmd{
			Shape: ShapeBox,
			Pos:   Vec3{X: x, Y: legY, Z: dz},
			Size:  Size{X: 0.16, Y: float64(legH), Z: 0.16},
			Color: heroLimb,
		})
	}
	for i, dz := range []float32{-s.armSwing, s.armSwing} {
		x := 0.38 + s.armOut
		if i == 0 {
			x = -x
		}
		cs = append(cs, Cmd{
			Shape: ShapeBox,
			Pos:   Vec3{X: x, Y: 1.0, Z: dz},
			Size:  Size{X: 0.13, Y: 0.45, Z: 0.13},
			Color: heroLimb,
		})
	}
	return cs
}

func heroTailCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.tailY, Z: s.tailZ}, Size: Size{X: 0.1, Y: 0.1, Z: 0.55}, Color: heroTail},
	}
}

func heroNoseCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.headY - 0.28, Z: 0.3}, Size: Size{X: 0.22, Y: 0.55, Z: 0.4}, Color: heroNose},
	}
}

func heroHairCmds(s heroSpec) []Cmd {
	return []Cmd{
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.headY + 0.18, Z: 0}, Size: Size{X: 0.48, Y: 0.12, Z: 0.42}, Color: heroHair},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.headY + 0.13, Z: -0.16 - s.hairSweep}, Size: Size{X: 0.46, Y: 0.1, Z: 0.34}, Color: heroHair},
		{Shape: ShapeBox, Pos: Vec3{X: 0, Y: s.headY + 0.05 - s.hairDip, Z: -0.3 - s.hairSweep}, Size: Size{X: 0.4, Y: 0.09, Z: 0.28}, Color: heroHair},
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
