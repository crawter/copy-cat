package game

import rl "github.com/gen2brain/raylib-go/v5/raylib"

// LaneLeft reports whether a move-left input was pressed this frame: the A key
// or the Left arrow.
func LaneLeft() bool {
	return rl.IsKeyPressed(rl.KeyA) || rl.IsKeyPressed(rl.KeyLeft)
}

// LaneRight reports whether a move-right input was pressed this frame: the D
// key or the Right arrow.
func LaneRight() bool {
	return rl.IsKeyPressed(rl.KeyD) || rl.IsKeyPressed(rl.KeyRight)
}

// Jump reports whether a jump input was pressed this frame: Space or the W key.
func Jump() bool {
	return rl.IsKeyPressed(rl.KeySpace) || rl.IsKeyPressed(rl.KeyW)
}

// Attack reports whether an attack input was pressed this frame: X or Enter.
// The attack feature consumes it later.
func Attack() bool {
	return rl.IsKeyPressed(rl.KeyX) || rl.IsKeyPressed(rl.KeyEnter)
}
