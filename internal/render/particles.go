package render

// Particle burst builders. Each burst is a plain []Cmd centered on the
// hero's feet at the origin; the draw step translates and ages the commands
// so a burst fires once and then fades.

var (
	combParticle = Color{R: 240, G: 202, B: 112, A: 255} // golden comb sparkle
	landParticle = Color{R: 150, G: 140, B: 120, A: 255} // kicked-up dust
	snipParticle = heroHair                              // snipped hair strands
)

// BuildCombBurst returns the golden sparkle burst shown when a comb is
// collected.
func BuildCombBurst() []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: v3(-0.3, 0.5, 0.2), Size: Size{X: 0.06}, Color: combParticle},
		{Shape: ShapeSphere, Pos: v3(0.3, 0.5, 0.2), Size: Size{X: 0.06}, Color: combParticle},
		{Shape: ShapeSphere, Pos: v3(-0.2, 0.8, -0.1), Size: Size{X: 0.05}, Color: combParticle},
		{Shape: ShapeSphere, Pos: v3(0.2, 0.8, -0.1), Size: Size{X: 0.05}, Color: combParticle},
		{Shape: ShapeSphere, Pos: v3(0, 1.0, 0.1), Size: Size{X: 0.06}, Color: combParticle},
		{Shape: ShapeBox, Pos: v3(-0.1, 0.65, 0.3), Size: Size{X: 0.04, Y: 0.1, Z: 0.04}, Color: combParticle},
		{Shape: ShapeBox, Pos: v3(0.1, 0.65, 0.3), Size: Size{X: 0.04, Y: 0.1, Z: 0.04}, Color: combParticle},
	}
}

// BuildLandingBurst returns the dust puff shown when the hero lands on the
// ground.
func BuildLandingBurst() []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: v3(-0.4, 0.08, 0.1), Size: Size{X: 0.1}, Color: landParticle},
		{Shape: ShapeSphere, Pos: v3(0.4, 0.08, 0.1), Size: Size{X: 0.1}, Color: landParticle},
		{Shape: ShapeSphere, Pos: v3(-0.25, 0.12, -0.15), Size: Size{X: 0.08}, Color: landParticle},
		{Shape: ShapeSphere, Pos: v3(0.25, 0.12, -0.15), Size: Size{X: 0.08}, Color: landParticle},
		{Shape: ShapeSphere, Pos: v3(0, 0.06, 0.25), Size: Size{X: 0.09}, Color: landParticle},
		{Shape: ShapeSphere, Pos: v3(-0.1, 0.05, -0.25), Size: Size{X: 0.07}, Color: landParticle},
		{Shape: ShapeSphere, Pos: v3(0.1, 0.05, -0.25), Size: Size{X: 0.07}, Color: landParticle},
	}
}

// BuildSnipBurst returns the falling hair strands shown when Copycat snips
// the hero.
func BuildSnipBurst() []Cmd {
	return []Cmd{
		{Shape: ShapeCone, Pos: v3(-0.15, 1.2, 0.1), Size: Size{X: 0.04, Y: 0.22}, Color: snipParticle},
		{Shape: ShapeCone, Pos: v3(0.15, 1.2, 0.1), Size: Size{X: 0.04, Y: 0.22}, Color: snipParticle},
		{Shape: ShapeCone, Pos: v3(-0.2, 1.05, -0.1), Size: Size{X: 0.04, Y: 0.2}, Color: snipParticle},
		{Shape: ShapeCone, Pos: v3(0.2, 1.05, -0.1), Size: Size{X: 0.04, Y: 0.2}, Color: snipParticle},
		{Shape: ShapeCone, Pos: v3(0, 0.9, 0.15), Size: Size{X: 0.04, Y: 0.24}, Color: snipParticle},
		{Shape: ShapeSphere, Pos: v3(-0.08, 0.75, 0.2), Size: Size{X: 0.05}, Color: snipParticle},
		{Shape: ShapeSphere, Pos: v3(0.08, 0.75, 0.2), Size: Size{X: 0.05}, Color: snipParticle},
	}
}
