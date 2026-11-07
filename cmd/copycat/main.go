package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	const (
		windowWidth  = 1280
		windowHeight = 720
		windowTitle  = "copycat"
		targetFPS    = 60
	)

	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(targetFPS)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 28, A: 255})
		rl.EndDrawing()
	}
}
