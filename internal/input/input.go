// Package input turns raylib's keyboard and first gamepad into the plain
// State the game loop consumes. Poll is the only function that calls raylib;
// Mapper.Map is a pure mapping of the raw Keys and Pad plus its own previous
// frame, so the mapping is testable headlessly.
//
// The task left the pause and quit inputs unnamed, so this package pins them
// down: P pauses and Q quits on the keyboard, the gamepad Start button pauses
// and the gamepad back/select button quits. Esc and Backspace both go back,
// as does the gamepad B button.
package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Keys is the keyboard state Poll reads from raylib. Every field reports
// whether that key is currently held down.
type Keys struct {
	Left, Right bool
	Up, Down    bool
	Enter       bool
	Space       bool
	Esc         bool
	Backspace   bool
	Pause       bool
	Quit        bool
}

// Pad is the first gamepad's state. Every field reports whether that button
// is currently held down; Connected reports whether a gamepad is present.
type Pad struct {
	Connected   bool
	Left, Right bool
	Up, Down    bool
	A, B        bool
	Start       bool
	Quit        bool
}

// State is the plain input the game loop consumes. Left, Right, Jump and
// Slide are held states; Up, Down, Confirm, Back, Pause and Quit are
// one-frame pressed edges.
type State struct {
	Left, Right, Jump, Slide bool
	Up, Down                 bool
	Confirm, Back            bool
	Pause, Quit              bool
}

// Mapper converts raw Keys and Pad into State. It remembers the previous raw
// frame so every pressed edge fires exactly once per press.
type Mapper struct {
	prevKeys Keys
	prevPad  Pad
}

// NewMapper returns a ready-to-use mapper.
func NewMapper() *Mapper { return &Mapper{} }

// Map turns k and p into State. Movement fields stay held; menu fields fire
// only on the rising edge of their key or button. A disconnected pad is
// ignored entirely.
func (m *Mapper) Map(k Keys, p Pad) State {
	s := State{
		Left:    k.Left,
		Right:   k.Right,
		Jump:    k.Space || k.Up,
		Slide:   k.Down,
		Up:      pressed(k.Up, m.prevKeys.Up),
		Down:    pressed(k.Down, m.prevKeys.Down),
		Confirm: pressed(k.Enter, m.prevKeys.Enter),
		Back:    pressed(k.Esc, m.prevKeys.Esc) || pressed(k.Backspace, m.prevKeys.Backspace),
		Pause:   pressed(k.Pause, m.prevKeys.Pause),
		Quit:    pressed(k.Quit, m.prevKeys.Quit),
	}
	if p.Connected {
		s.Left = s.Left || p.Left
		s.Right = s.Right || p.Right
		s.Jump = s.Jump || p.A || p.Up
		s.Slide = s.Slide || p.B || p.Down
		s.Up = s.Up || pressed(p.Up, m.prevPad.Up)
		s.Down = s.Down || pressed(p.Down, m.prevPad.Down)
		s.Confirm = s.Confirm || pressed(p.A, m.prevPad.A)
		s.Back = s.Back || pressed(p.B, m.prevPad.B)
		s.Pause = s.Pause || pressed(p.Start, m.prevPad.Start)
		s.Quit = s.Quit || pressed(p.Quit, m.prevPad.Quit)
	}
	m.prevKeys = k
	m.prevPad = p
	return s
}

// pressed reports whether now is held and before was not.
func pressed(now, before bool) bool { return now && !before }

// stickDeadzone is the stick deflection below which Poll ignores the axis.
const stickDeadzone = 0.5

// Poll reads raylib's keyboard and the first gamepad, then maps them. It is
// the only function in the package that calls raylib.
func Poll(m *Mapper) State {
	k := Keys{
		Left:      rl.IsKeyDown(rl.KeyLeft),
		Right:     rl.IsKeyDown(rl.KeyRight),
		Up:        rl.IsKeyDown(rl.KeyUp),
		Down:      rl.IsKeyDown(rl.KeyDown),
		Enter:     rl.IsKeyDown(rl.KeyEnter),
		Space:     rl.IsKeyDown(rl.KeySpace),
		Esc:       rl.IsKeyDown(rl.KeyEscape),
		Backspace: rl.IsKeyDown(rl.KeyBackspace),
		Pause:     rl.IsKeyDown(rl.KeyP),
		Quit:      rl.IsKeyDown(rl.KeyQ),
	}

	p := Pad{Connected: rl.IsGamepadAvailable(0)}
	if p.Connected {
		stickX := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftX)
		stickY := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftY)
		p.Left = rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceLeft) || stickX < -stickDeadzone
		p.Right = rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceRight) || stickX > stickDeadzone
		p.Up = rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceUp) || stickY < -stickDeadzone
		p.Down = rl.IsGamepadButtonDown(0, rl.GamepadButtonLeftFaceDown) || stickY > stickDeadzone
		p.A = rl.IsGamepadButtonDown(0, rl.GamepadButtonRightFaceDown)
		p.B = rl.IsGamepadButtonDown(0, rl.GamepadButtonRightFaceRight)
		p.Start = rl.IsGamepadButtonDown(0, rl.GamepadButtonMiddleRight)
		p.Quit = rl.IsGamepadButtonDown(0, rl.GamepadButtonMiddleLeft)
	}

	return m.Map(k, p)
}
