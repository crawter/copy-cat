package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/v5/raylib"
)

const (
	attackCooldown = 0.6
	attackWindow   = 0.25
	attackReach    = 2.5
	attackSide     = 1.5

	patrolRange = 1.0
	patrolSpeed = 1.3
)

// Enemy is one living office machine spawned by the layout.
type Enemy struct {
	Kind     EnemyKind
	Lane     int
	SpawnZ   float32
	Phase    float32
	Z        float32
	Defeated bool
}

// Enemies owns the office-machine enemies, their patrol, and the player's
// melee attack state.
type Enemies struct {
	layout *Layout

	// Enemies holds one enemy per layout spawn, in layout order.
	Enemies []*Enemy

	// AttackTime is the raylib timestamp of the last attack swing, exported
	// so other subsystems can react to the attack.
	AttackTime float64

	elapsed float32
}

// NewEnemies builds one enemy per layout spawn.
func NewEnemies(layout *Layout) *Enemies {
	e := &Enemies{
		layout:     layout,
		AttackTime: -attackCooldown,
	}
	for i := range layout.Segments {
		spawn := layout.Segments[i].Enemy
		if spawn == nil {
			continue
		}
		e.Enemies = append(e.Enemies, &Enemy{
			Kind:   spawn.Kind,
			Lane:   spawn.Lane,
			SpawnZ: spawn.Z,
			Phase:  float32(i) * 0.7,
			Z:      spawn.Z,
		})
	}
	return e
}

// Update patrols living enemies, runs the melee attack, and checks for
// contact death while the game is playing.
func (e *Enemies) Update(dt float32, g *Game) {
	if g.Scene != ScenePlay {
		return
	}
	e.elapsed += dt

	for _, enemy := range e.Enemies {
		if enemy.Defeated {
			continue
		}
		enemy.Z = enemy.SpawnZ + patrolRange*float32(math.Sin(float64(e.elapsed*patrolSpeed+enemy.Phase)))
	}

	now := rl.GetTime()
	if (rl.IsKeyPressed(rl.KeyX) || rl.IsKeyPressed(rl.KeyEnter)) && now-e.AttackTime >= attackCooldown {
		e.AttackTime = now
		e.defeatInRange(g.Player)
	}

	e.checkContact(g)
}

// Draw renders living enemies and the attack arc while its hit window lasts.
func (e *Enemies) Draw(g *Game) {
	rl.BeginMode3D(g.Cam.Camera3D)
	defer rl.EndMode3D()

	for _, enemy := range e.Enemies {
		if enemy.Defeated {
			continue
		}
		x := e.layout.LaneX(enemy.Lane)
		switch enemy.Kind {
		case EnemyCopier:
			e.drawCopier(x, enemy.Z)
		case EnemyCabinet:
			e.drawCabinet(x, enemy.Z)
		case EnemyCoffee:
			e.drawCoffee(x, enemy.Z)
		}
	}

	if rl.GetTime()-e.AttackTime <= attackWindow {
		e.drawAttackArc(g.Player)
	}
}

func (e *Enemies) defeatInRange(player *Player) {
	for _, enemy := range e.Enemies {
		if enemy.Defeated {
			continue
		}
		dz := enemy.Z - player.Z
		dx := math.Abs(float64(player.X - e.layout.LaneX(enemy.Lane)))
		if dz > 0 && dz <= attackReach && dx <= attackSide {
			enemy.Defeated = true
		}
	}
}

func (e *Enemies) checkContact(g *Game) {
	if !g.Player.Grounded {
		return
	}
	for _, enemy := range e.Enemies {
		if enemy.Defeated || enemy.Lane != g.Player.Lane {
			continue
		}
		if math.Abs(float64(g.Player.Z-enemy.Z)) <= 0.9 {
			g.Scene = SceneDeath
			return
		}
	}
}

func (e *Enemies) drawCopier(x, z float32) {
	body := rl.Color{R: 96, G: 100, B: 110, A: 255}
	rl.DrawCube(rl.Vector3{X: x, Y: 0.45, Z: z}, 1.2, 0.9, 0.7, body)

	// A glowing slot on the side facing the approaching player.
	rl.DrawCube(rl.Vector3{X: x, Y: 0.55, Z: z - 0.38}, 0.5, 0.08, 0.06, rl.Color{R: 200, G: 235, B: 255, A: 255})
}

func (e *Enemies) drawCabinet(x, z float32) {
	body := rl.Color{R: 118, G: 82, B: 48, A: 255}
	line := rl.Color{R: 78, G: 52, B: 28, A: 255}
	rl.DrawCube(rl.Vector3{X: x, Y: 0.9, Z: z}, 1.1, 1.8, 0.6, body)

	// Drawer lines on the front face.
	for i := 0; i < 3; i++ {
		lineY := 0.45 + float32(i)*0.45
		rl.DrawCube(rl.Vector3{X: x, Y: lineY, Z: z - 0.31}, 1.0, 0.05, 0.04, line)
	}
}

func (e *Enemies) drawCoffee(x, z float32) {
	body := rl.Color{R: 46, G: 44, B: 50, A: 255}
	pot := rl.Color{R: 170, G: 120, B: 60, A: 255}
	rl.DrawCube(rl.Vector3{X: x, Y: 0.55, Z: z}, 0.9, 1.1, 0.7, body)
	rl.DrawCylinder(rl.Vector3{X: x, Y: 1.2, Z: z}, 0.15, 0.15, 0.5, 12, pot)
}

func (e *Enemies) drawAttackArc(player *Player) {
	const (
		radius   = 0.7
		segments = 12
		spread   = 120.0 * math.Pi / 180.0
	)

	origin := rl.Vector3{X: player.X, Y: 1.2, Z: player.Z}
	var prev rl.Vector3
	for i := 0; i <= segments; i++ {
		angle := -spread/2 + spread*float32(i)/segments
		point := rl.Vector3{
			X: origin.X + radius*float32(math.Sin(float64(angle))),
			Y: origin.Y,
			Z: origin.Z + radius*float32(math.Cos(float64(angle))),
		}
		if i > 0 {
			rl.DrawLine3D(prev, point, rl.White)
		}
		prev = point
	}
}
