package input

import "testing"

func TestKeyboardHeldMovement(t *testing.T) {
	tests := []struct {
		name                string
		k                   Keys
		wantLeft, wantRight bool
		wantJump, wantSlide bool
	}{
		{name: "left arrow", k: Keys{Left: true}, wantLeft: true},
		{name: "right arrow", k: Keys{Right: true}, wantRight: true},
		{name: "space jumps", k: Keys{Space: true}, wantJump: true},
		{name: "up arrow jumps", k: Keys{Up: true}, wantJump: true},
		{name: "down arrow slides", k: Keys{Down: true}, wantSlide: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMapper()
			s := m.Map(tt.k, Pad{})
			if s.Left != tt.wantLeft || s.Right != tt.wantRight || s.Jump != tt.wantJump || s.Slide != tt.wantSlide {
				t.Fatalf("Map(%+v) = %+v", tt.k, s)
			}
		})
	}
}

func TestEdgesFireExactlyOnce(t *testing.T) {
	m := NewMapper()

	first := m.Map(Keys{Up: true}, Pad{})
	if !first.Up {
		t.Fatalf("first press should fire the Up edge, got %+v", first)
	}

	second := m.Map(Keys{Up: true}, Pad{})
	if second.Up {
		t.Fatalf("a held key must not repeat its edge, got %+v", second)
	}

	third := m.Map(Keys{}, Pad{})
	if third.Up {
		t.Fatalf("a release must not fire an edge, got %+v", third)
	}

	fourth := m.Map(Keys{Up: true}, Pad{})
	if !fourth.Up {
		t.Fatalf("a new press should fire the edge again, got %+v", fourth)
	}
}

func TestDisconnectedPadIgnored(t *testing.T) {
	m := NewMapper()
	s := m.Map(Keys{}, Pad{Connected: false, Left: true, Right: true, Up: true, Down: true, A: true, B: true, Start: true, Quit: true})
	if s.Left || s.Right || s.Jump || s.Slide || s.Up || s.Down || s.Confirm || s.Back || s.Pause || s.Quit {
		t.Fatalf("disconnected pad must be ignored, got %+v", s)
	}
}

func TestGamepadButtonsMapToActions(t *testing.T) {
	tests := []struct {
		name  string
		p     Pad
		check func(State) bool
	}{
		{
			name:  "left and A",
			p:     Pad{Connected: true, Left: true, A: true},
			check: func(s State) bool { return s.Left && s.Jump && s.Confirm },
		},
		{
			name:  "B slides and goes back",
			p:     Pad{Connected: true, B: true},
			check: func(s State) bool { return s.Slide && s.Back },
		},
		{
			name:  "up jumps and moves menu up",
			p:     Pad{Connected: true, Up: true},
			check: func(s State) bool { return s.Jump && s.Up },
		},
		{
			name:  "down slides and moves menu down",
			p:     Pad{Connected: true, Down: true},
			check: func(s State) bool { return s.Slide && s.Down },
		},
		{
			name:  "start pauses",
			p:     Pad{Connected: true, Start: true},
			check: func(s State) bool { return s.Pause },
		},
		{
			name:  "select quits",
			p:     Pad{Connected: true, Quit: true},
			check: func(s State) bool { return s.Quit },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMapper()
			s := m.Map(Keys{}, tt.p)
			if !tt.check(s) {
				t.Fatalf("Map(%+v) = %+v", tt.p, s)
			}
		})
	}
}

func TestConfirmBackPauseQuitFromKeyboardAndGamepad(t *testing.T) {
	tests := []struct {
		name  string
		k     Keys
		p     Pad
		check func(State) bool
	}{
		{name: "keyboard confirm", k: Keys{Enter: true}, check: func(s State) bool { return s.Confirm }},
		{name: "keyboard back esc", k: Keys{Esc: true}, check: func(s State) bool { return s.Back }},
		{name: "keyboard back backspace", k: Keys{Backspace: true}, check: func(s State) bool { return s.Back }},
		{name: "keyboard pause", k: Keys{Pause: true}, check: func(s State) bool { return s.Pause }},
		{name: "keyboard quit", k: Keys{Quit: true}, check: func(s State) bool { return s.Quit }},
		{name: "gamepad confirm", p: Pad{Connected: true, A: true}, check: func(s State) bool { return s.Confirm }},
		{name: "gamepad back", p: Pad{Connected: true, B: true}, check: func(s State) bool { return s.Back }},
		{name: "gamepad pause", p: Pad{Connected: true, Start: true}, check: func(s State) bool { return s.Pause }},
		{name: "gamepad quit", p: Pad{Connected: true, Quit: true}, check: func(s State) bool { return s.Quit }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMapper()
			s := m.Map(tt.k, tt.p)
			if !tt.check(s) {
				t.Fatalf("Map(%+v, %+v) = %+v", tt.k, tt.p, s)
			}
		})
	}
}
