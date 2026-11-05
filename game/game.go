package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// Camera wraps raylib's 3D camera. Follow behavior arrives later in camera.go;
// for the scaffold it stays a fixed camera behind the spawn.
type Camera struct {
	rl.Camera3D
}

type SceneKind int

const (
	ScenePlay SceneKind = iota
	SceneDeath
	SceneEnd
)

type RunStats struct {
	Stamps       int
	TotalStamps  int
	Distance     float32
	BestDistance float32
	Deaths       int
	Completed    bool
}

type Game struct {
	Cam       *Camera
	Layout    *Layout
	Player    *Player
	World     *World
	Obstacles *Obstacles
	Pickups   *Pickups
	Enemies   *Enemies
	Audio     *Audio
	Save      *Save
	Scenes    *Scenes
	HUD       *HUD
	Run       RunStats
	Scene     SceneKind
}

// New assembles the game and wires every subsystem.
func New() *Game {
	layout := NewLayout()

	return &Game{
		Layout: layout,
		Cam: &Camera{Camera3D: rl.Camera3D{
			Position:   rl.Vector3{X: 0, Y: 4, Z: -8},
			Target:     rl.Vector3{X: 0, Y: 1.5, Z: 6},
			Up:         rl.Vector3{X: 0, Y: 1, Z: 0},
			Fovy:       60,
			Projection: rl.CameraPerspective,
		}},
		Player:    NewPlayer(),
		World:     NewWorld(layout),
		Obstacles: NewObstacles(layout),
		Pickups:   NewPickups(layout),
		Enemies:   NewEnemies(layout),
		Audio:     NewAudio(),
		Save:      NewSave(),
		Scenes:    NewScenes(),
		HUD:       NewHUD(),
		Scene:     ScenePlay,
	}
}

// Run opens the window and audio device, runs the loop at 60 FPS with vsync,
// and writes the save on exit before closing.
func (g *Game) Run() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.InitWindow(1280, 720, Title)
	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() && !rl.IsKeyPressed(rl.KeyEscape) {
		g.Update()
		g.Draw()
	}

	g.Save.Write(g)
	rl.CloseAudioDevice()
	rl.CloseWindow()
}

// Update advances one frame in the fixed subsystem order.
func (g *Game) Update() {
	dt := rl.GetFrameTime()

	g.Player.Update(dt, g)
	g.World.Update(dt, g)
	g.Obstacles.Update(dt, g)
	g.Pickups.Update(dt, g)
	g.Enemies.Update(dt, g)
	g.Scenes.Update(dt, g)
	g.HUD.Update(dt, g)
	g.Save.Update(dt, g)
	g.Audio.Update(dt, g)
}

// Draw renders one frame: the 3D world first, then the 2D scenes and HUD.
func (g *Game) Draw() {
	rl.BeginDrawing()
	defer rl.EndDrawing()

	rl.ClearBackground(rl.Black)

	g.World.Draw(g)
	g.Player.Draw(g)
	g.Obstacles.Draw(g)
	g.Pickups.Draw(g)
	g.Enemies.Draw(g)

	g.Scenes.Draw(g)
	g.HUD.Draw(g)
}
