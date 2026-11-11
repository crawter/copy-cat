package render

// Particle burst builders. Each burst is a plain []Cmd centered on the
// hero's feet at the origin; the draw step translates and ages the commands
// so a burst fires once and then fades.

var (
	combParticle = Color{R: 240, G: 202, B: 112, A: 255} // golden comb sparkle
	landParticle = Color{R: 150, G: 140, B: 120, A: 255} // kicked-up dust
	snipParticle = heroHair                              // snipped hair strands

	stompParticle   = Color{R: 255, G: 190, B: 70, A: 255}  // stomp impact flash
	checkpointShine = Color{R: 190, G: 230, B: 255, A: 255} // mirror glint
	bossHitParticle = Color{R: 235, G: 60, B: 60, A: 255}   // hit sparks
	slideParticle   = Color{R: 255, G: 240, B: 150, A: 255} // friction sparks
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

// BuildStompBurst returns the ground-level impact flash shown when the hero
// stomps an enemy.
func BuildStompBurst() []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: v3(-0.55, 0.08, 0.25), Size: Size{X: 0.1}, Color: stompParticle},
		{Shape: ShapeSphere, Pos: v3(0.55, 0.08, 0.25), Size: Size{X: 0.1}, Color: stompParticle},
		{Shape: ShapeSphere, Pos: v3(-0.35, 0.12, -0.2), Size: Size{X: 0.08}, Color: stompParticle},
		{Shape: ShapeSphere, Pos: v3(0.35, 0.12, -0.2), Size: Size{X: 0.08}, Color: stompParticle},
		{Shape: ShapeSphere, Pos: v3(0, 0.06, 0.35), Size: Size{X: 0.09}, Color: stompParticle},
		{Shape: ShapeBox, Pos: v3(0, 0.14, 0.1), Size: Size{X: 0.5, Y: 0.08, Z: 0.5}, Color: stompParticle},
	}
}

// BuildCheckpointShine returns the rising glint shown when the hero reaches a
// checkpoint mirror.
func BuildCheckpointShine() []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: v3(0, 0.9, 0.1), Size: Size{X: 0.06}, Color: checkpointShine},
		{Shape: ShapeSphere, Pos: v3(-0.28, 1.1, -0.05), Size: Size{X: 0.05}, Color: checkpointShine},
		{Shape: ShapeSphere, Pos: v3(0.28, 1.1, -0.05), Size: Size{X: 0.05}, Color: checkpointShine},
		{Shape: ShapeSphere, Pos: v3(0, 1.35, 0.05), Size: Size{X: 0.07}, Color: checkpointShine},
		{Shape: ShapeBox, Pos: v3(-0.16, 1.55, 0.05), Size: Size{X: 0.04, Y: 0.3, Z: 0.04}, Color: checkpointShine},
		{Shape: ShapeBox, Pos: v3(0.16, 1.55, 0.05), Size: Size{X: 0.04, Y: 0.3, Z: 0.04}, Color: checkpointShine},
	}
}

// BuildBossHitBurst returns the red sparks shown when the Copycat boss lands
// an attack.
func BuildBossHitBurst() []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: v3(0.5, 1.2, 0.4), Size: Size{X: 0.1}, Color: bossHitParticle},
		{Shape: ShapeSphere, Pos: v3(-0.5, 1.2, 0.4), Size: Size{X: 0.1}, Color: bossHitParticle},
		{Shape: ShapeSphere, Pos: v3(0.3, 1.6, -0.2), Size: Size{X: 0.08}, Color: bossHitParticle},
		{Shape: ShapeSphere, Pos: v3(-0.3, 1.6, -0.2), Size: Size{X: 0.08}, Color: bossHitParticle},
		{Shape: ShapeCone, Pos: v3(0, 1.05, 0.55), Size: Size{X: 0.08, Y: 0.3}, Color: bossHitParticle},
		{Shape: ShapeBox, Pos: v3(0, 0.9, 0.5), Size: Size{X: 0.25, Y: 0.1, Z: 0.25}, Color: bossHitParticle},
	}
}

// BuildSlideSparks returns the low trailing sparks shown when the hero starts
// a grounded slide.
func BuildSlideSparks() []Cmd {
	return []Cmd{
		{Shape: ShapeSphere, Pos: v3(0.4, 0.08, -0.3), Size: Size{X: 0.05}, Color: slideParticle},
		{Shape: ShapeSphere, Pos: v3(-0.4, 0.08, -0.3), Size: Size{X: 0.05}, Color: slideParticle},
		{Shape: ShapeSphere, Pos: v3(0.25, 0.1, -0.55), Size: Size{X: 0.04}, Color: slideParticle},
		{Shape: ShapeSphere, Pos: v3(-0.25, 0.1, -0.55), Size: Size{X: 0.04}, Color: slideParticle},
		{Shape: ShapeBox, Pos: v3(0, 0.12, -0.7), Size: Size{X: 0.16, Y: 0.06, Z: 0.16}, Color: slideParticle},
		{Shape: ShapeCone, Pos: v3(0, 0.2, -0.9), Size: Size{X: 0.06, Y: 0.25}, Color: slideParticle},
	}
}
