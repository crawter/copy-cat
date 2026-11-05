package game

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/v5/raylib"
)

// deathJokes is the fixed rotation of reviewed death-scene lines.
var deathJokes = [...]string{DeathJokeOne, DeathJokeTwo, DeathJokeThree}

// deathPause is how long the death scene stays before the run respawns.
const deathPause = 2.5

// Scenes owns the death and level-end overlays and drives the scene state
// machine from g.Scene, which the player, obstacle, and enemy code sets.
type Scenes struct {
	scene      SceneKind
	deathTimer float32
	joke       int // index of the joke shown during the current death
	nextJoke   int // rotating index for the next death
}

func NewScenes() *Scenes {
	return &Scenes{scene: ScenePlay}
}

// Update advances the scene state machine and performs its transitions.
func (s *Scenes) Update(dt float32, g *Game) {
	switch g.Scene {
	case ScenePlay:
		if g.Player.Z >= g.Layout.EndZ {
			g.Scene = SceneEnd
			g.Run.Completed = true
		}
	case SceneDeath:
		if s.scene != SceneDeath {
			s.deathTimer = 0
			s.joke = s.nextJoke
			s.nextJoke = (s.nextJoke + 1) % len(deathJokes)
		}
		s.deathTimer += dt
		if s.deathTimer >= deathPause {
			g.Run.Deaths++
			s.respawn(g)
		}
	case SceneEnd:
		if rl.IsKeyPressed(rl.KeyEnter) {
			s.respawn(g)
		}
	}
	s.scene = g.Scene
}

// respawn restarts the run at lane 1 and clears this run's progress.
// Totals and the death counter keep accumulating.
func (s *Scenes) respawn(g *Game) {
	g.Player.Lane = 1
	g.Player.X = g.Layout.LaneX(1)
	g.Player.Y = 0
	g.Player.Z = 0
	g.Player.VX = 0
	g.Player.VY = 0
	g.Player.Grounded = true
	g.Run.Stamps = 0
	g.Run.Distance = 0
	g.Scene = ScenePlay
}

// Draw renders the current scene overlay, if any.
func (s *Scenes) Draw(g *Game) {
	switch g.Scene {
	case SceneDeath:
		s.drawDeath(g)
	case SceneEnd:
		s.drawEnd(g)
	}
}

func (s *Scenes) drawDeath(g *Game) {
	width := int32(rl.GetScreenWidth())
	height := int32(rl.GetScreenHeight())

	rl.DrawRectangle(0, 0, width, height, rl.Fade(rl.Black, 0.75))

	face := "X_X"
	faceSize := int32(140)
	faceWidth := rl.MeasureText(face, faceSize)
	rl.DrawText(face, (width-faceWidth)/2, height/2-faceSize, faceSize, rl.White)

	joke := deathJokes[s.joke]
	jokeSize := int32(24)
	jokeWidth := rl.MeasureText(joke, jokeSize)
	rl.DrawText(joke, (width-jokeWidth)/2, height/2+jokeSize, jokeSize, rl.White)
}

func (s *Scenes) drawEnd(g *Game) {
	width := int32(rl.GetScreenWidth())
	height := int32(rl.GetScreenHeight())

	rl.DrawRectangle(0, 0, width, height, rl.Fade(rl.Black, 0.75))

	titleSize := int32(56)
	titleWidth := rl.MeasureText(EndTitle, titleSize)
	rl.DrawText(EndTitle, (width-titleWidth)/2, height/6, titleSize, rl.White)

	best := g.Run.BestDistance
	if g.Run.Distance > best {
		best = g.Run.Distance
	}

	stats := []string{
		fmt.Sprintf("%s: %d", EndStamps, g.Run.Stamps),
		fmt.Sprintf("%s: %d", HUDStamps, g.Run.TotalStamps),
		fmt.Sprintf("%s: %.0f", EndDistance, g.Run.Distance),
		fmt.Sprintf("%s: %.0f", HUDBest, best),
		fmt.Sprintf("%s: %d", EndDeaths, g.Run.Deaths),
	}

	statSize := int32(32)
	lineHeight := statSize + 12
	for i, line := range stats {
		lineWidth := rl.MeasureText(line, statSize)
		rl.DrawText(line, (width-lineWidth)/2, height/3+int32(i)*lineHeight, statSize, rl.White)
	}

	hintSize := int32(20)
	hintWidth := rl.MeasureText(ControlsHint, hintSize)
	rl.DrawText(ControlsHint, (width-hintWidth)/2, height-hintSize-40, hintSize, rl.White)
}
