// Package render is the POC's only raylib layer. It owns the desktop window,
// the input mapping, the drawing of the cheap film set and the HUD, and the
// synthesized audio. It reads the logic packages' plain state to draw each
// frame and forwards input back to them as plain values; no logic package
// imports render or raylib, and render never calls into the game rules.
package render

import (
	"errors"

	rl "github.com/gen2brain/raylib-go/raylib"

	"copycat/enemy"
	"copycat/leveldata"
	"copycat/move"
	"copycat/scenestate"
)

// The desktop window's size in pixels.
const (
	screenWidth  = 1280
	screenHeight = 720
)

// World bundles one frame's plain state for DrawFrame: the hero, the take, the
// crane with its defeat state, the reel counts, the lives, the scene state and
// the intercom line the scene is speaking.
type World struct {
	Hero       move.Hero
	Take       leveldata.Take
	Crane      enemy.Crane
	ReelsFound int
	ReelsTotal int
	Lives      int
	State      scenestate.State
	Line       string
}

// MapInput turns the held left and right steer keys and the press-frame jump
// key into move.Input. Left steers -1, right steers +1, both cancel to 0, and
// jump maps to Jump. It is pure and never touches raylib.
func MapInput(left, right, jump bool) move.Input {
	steer := 0.0
	if left != right {
		if right {
			steer = 1
		} else {
			steer = -1
		}
	}
	return move.Input{Steer: steer, Jump: jump}
}

// Open opens the desktop window and the audio device and prepares the four
// synthesized sounds. It returns an error when the audio device fails to
// start; a desktop window without audio is not a usable POC.
func Open() error {
	rl.InitWindow(screenWidth, screenHeight, gameName)
	rl.InitAudioDevice()
	if !rl.IsAudioDeviceReady() {
		rl.CloseWindow()
		return errors.New("render: audio device did not start")
	}
	rl.SetTargetFPS(60)
	initSounds()
	return nil
}

// Close frees the synthesized sounds, closes the audio device and closes the
// window. It is safe to call after a failed Open: the sounds are only unloaded
// when they were built.
func Close() {
	if soundsLoaded {
		rl.UnloadSound(jumpSound)
		rl.UnloadSound(pickupSound)
		rl.UnloadSound(thudSound)
		rl.UnloadSound(crackleSound)
		soundsLoaded = false
	}
	rl.CloseAudioDevice()
	rl.CloseWindow()
}

// ShouldClose reports whether the window has asked to close.
func ShouldClose() bool {
	return rl.WindowShouldClose()
}

// PollInput maps the current key state to plain booleans. Arrow keys or A and
// D steer, space jumps, enter starts or confirms, and escape leaves. The steer
// keys are reported as held; jump, start and leave are press-frame reports the
// main loop can edge-detect.
func PollInput() (left, right, jump, start, leave bool) {
	left = rl.IsKeyDown(rl.KeyLeft) || rl.IsKeyDown(rl.KeyA)
	right = rl.IsKeyDown(rl.KeyRight) || rl.IsKeyDown(rl.KeyD)
	jump = rl.IsKeyPressed(rl.KeySpace)
	start = rl.IsKeyPressed(rl.KeyEnter)
	leave = rl.IsKeyPressed(rl.KeyEscape)
	return left, right, jump, start, leave
}

// DrawFrame draws one frame from w's plain state. The five scene states pick
// their screens: the title, the playing set with the HUD, the dying farce, the
// recap and the game-over screen. It is the only place render reads the logic
// state to draw it.
func DrawFrame(w World) {
	rl.BeginDrawing()
	rl.ClearBackground(colorSky)
	switch w.State {
	case scenestate.Title:
		drawTitle()
	case scenestate.Playing:
		drawSet(w, false)
		drawHUD(w)
	case scenestate.Dying:
		drawSet(w, true)
	case scenestate.Recapping:
		drawRecap(w)
	case scenestate.GameOver:
		drawGameOver(w)
	}
	rl.EndDrawing()
}

// cameraPos returns the camera position in meters for hero h: behind and above
// him, locked to his sideways and forward motion down the corridor. It returns
// plain numbers with no raylib types so the placement stays pure.
func cameraPos(h move.Hero) (x, y, z float64) {
	return h.X, h.Y + cameraHeightMeters, h.Z - cameraBackMeters
}

// cameraTarget returns the point in meters the camera looks at for hero h: the
// hero and a little ahead down the corridor. It returns plain numbers.
func cameraTarget(h move.Hero) (x, y, z float64) {
	return h.X, h.Y + cameraTargetHeightMeters, h.Z + cameraLookAheadMeters
}
