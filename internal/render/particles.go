package render

// Particle is one short-lived effect drawn over the world: a comb sparkle, a
// landing dust puff or a snipped hair sliver. Every field is a plain value so
// the particle simulation is deterministic and testable without raylib.
type Particle struct {
	Pos   Vec3
	Vel   Vec3
	Life  float64 // seconds remaining; the particle dies at <= 0
	TTL   float64 // the particle's starting life, kept for tests and tuning
	Size  float64 // meters: sphere radius, or box side
	Color Color
	Shape Shape
}

// Particle lifetimes, in seconds, per effect.
const (
	sparkleTTL = 0.6
	dustTTL    = 0.8
	hairTTL    = 1.4
)

// Particle effect colors, kept fixed so the effects always read the same.
var (
	sparkleColor = Color{R: 250, G: 214, B: 90, A: 255}  // golden comb sparkle
	dustColor    = Color{R: 228, G: 222, B: 208, A: 210} // pale landing dust
)

// spawnSparkle returns a few small golden particles around the hero's body,
// thrown outward and up as a collected comb bursts.
func spawnSparkle(pos Vec3) []Particle {
	y := pos.Y + 1.25
	return []Particle{
		{Pos: Vec3{X: pos.X - 0.45, Y: y + 0.05, Z: pos.Z + 0.05}, Vel: Vec3{X: -0.9, Y: 0.8, Z: 0.25}, Life: sparkleTTL, TTL: sparkleTTL, Size: 0.07, Color: sparkleColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X - 0.25, Y: y + 0.3, Z: pos.Z - 0.1}, Vel: Vec3{X: -0.5, Y: 1.1, Z: -0.3}, Life: sparkleTTL, TTL: sparkleTTL, Size: 0.06, Color: sparkleColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X, Y: y + 0.4, Z: pos.Z}, Vel: Vec3{X: 0, Y: 1.2, Z: 0}, Life: sparkleTTL, TTL: sparkleTTL, Size: 0.08, Color: sparkleColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X + 0.25, Y: y + 0.3, Z: pos.Z - 0.1}, Vel: Vec3{X: 0.5, Y: 1.1, Z: -0.3}, Life: sparkleTTL, TTL: sparkleTTL, Size: 0.06, Color: sparkleColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X + 0.45, Y: y + 0.05, Z: pos.Z + 0.05}, Vel: Vec3{X: 0.9, Y: 0.8, Z: 0.25}, Life: sparkleTTL, TTL: sparkleTTL, Size: 0.07, Color: sparkleColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X, Y: y + 0.2, Z: pos.Z + 0.35}, Vel: Vec3{X: 0, Y: 0.9, Z: 0.7}, Life: sparkleTTL, TTL: sparkleTTL, Size: 0.05, Color: sparkleColor, Shape: ShapeSphere},
	}
}

// spawnDust returns pale dust puffs at the hero's feet, kicked outward and
// backward as a landing settles.
func spawnDust(pos Vec3) []Particle {
	y := pos.Y + 0.06
	return []Particle{
		{Pos: Vec3{X: pos.X - 0.3, Y: y, Z: pos.Z - 0.15}, Vel: Vec3{X: -0.7, Y: 0.4, Z: -0.5}, Life: dustTTL, TTL: dustTTL, Size: 0.11, Color: dustColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X - 0.1, Y: y + 0.05, Z: pos.Z - 0.3}, Vel: Vec3{X: -0.25, Y: 0.5, Z: -0.8}, Life: dustTTL, TTL: dustTTL, Size: 0.09, Color: dustColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X + 0.15, Y: y, Z: pos.Z - 0.2}, Vel: Vec3{X: 0.3, Y: 0.35, Z: -0.6}, Life: dustTTL, TTL: dustTTL, Size: 0.1, Color: dustColor, Shape: ShapeSphere},
		{Pos: Vec3{X: pos.X + 0.32, Y: y + 0.04, Z: pos.Z - 0.05}, Vel: Vec3{X: 0.75, Y: 0.3, Z: -0.3}, Life: dustTTL, TTL: dustTTL, Size: 0.12, Color: dustColor, Shape: ShapeSphere},
	}
}

// spawnHair returns heroHair-colored slivers clipped from the hero's quiff,
// fluttering sideways as they fall.
func spawnHair(pos Vec3) []Particle {
	y := pos.Y + 1.5
	return []Particle{
		{Pos: Vec3{X: pos.X - 0.1, Y: y + 0.2, Z: pos.Z - 0.1}, Vel: Vec3{X: -0.35, Y: -0.7, Z: -0.2}, Life: hairTTL, TTL: hairTTL, Size: 0.05, Color: heroHair, Shape: ShapeBox},
		{Pos: Vec3{X: pos.X + 0.12, Y: y + 0.28, Z: pos.Z - 0.2}, Vel: Vec3{X: 0.3, Y: -0.9, Z: -0.35}, Life: hairTTL, TTL: hairTTL, Size: 0.04, Color: heroHair, Shape: ShapeBox},
		{Pos: Vec3{X: pos.X, Y: y + 0.34, Z: pos.Z - 0.15}, Vel: Vec3{X: -0.1, Y: -1.1, Z: -0.15}, Life: hairTTL, TTL: hairTTL, Size: 0.05, Color: heroHair, Shape: ShapeBox},
		{Pos: Vec3{X: pos.X - 0.04, Y: y + 0.16, Z: pos.Z}, Vel: Vec3{X: 0.15, Y: -0.6, Z: -0.1}, Life: hairTTL, TTL: hairTTL, Size: 0.04, Color: heroHair, Shape: ShapeBox},
	}
}

// stepParticles advances every particle by its velocity for dt seconds, ages
// it by dt, and returns only the particles still alive. It reuses the input
// slice's backing array, so repeated calls on the same list do not allocate.
func stepParticles(ps []Particle, dt float64) []Particle {
	if dt < 0 {
		dt = 0
	}
	alive := 0
	for i := range ps {
		p := &ps[i]
		p.Pos.X += p.Vel.X * float32(dt)
		p.Pos.Y += p.Vel.Y * float32(dt)
		p.Pos.Z += p.Vel.Z * float32(dt)
		p.Life -= dt
		if p.Life > 0 {
			ps[alive] = ps[i]
			alive++
		}
	}
	return ps[:alive]
}
