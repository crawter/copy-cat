// Package scenestate carries the POC's flow through its five states: Title,
// Playing, Dying, Recapping and GameOver. The state changes only on the set
// events, and the package takes and returns plain values, never importing
// raylib, so the flow is testable headless and reusable unchanged on every
// platform.
package scenestate

// State is one scene of the POC's flow.
type State int

const (
	// Title is the title screen state.
	Title State = iota
	// Playing is the running-take state.
	Playing
	// Dying is the one-second farce-death state.
	Dying
	// Recapping is the end-of-take reel recap state.
	Recapping
	// GameOver is the zero-lives retry state.
	GameOver
)

// Event is a set event that may change the scene's state.
type Event int

const (
	// EventStart is a start press on Title, GameOver or Recapping.
	EventStart Event = iota
	// EventDeath is the hero being hit during the take.
	EventDeath
	// EventRespawn is the wiring's respawn after the farce.
	EventRespawn
	// EventExit is the hero crossing the exit line.
	EventExit
	// EventGameOver is the ninth death with no lives left.
	EventGameOver
	// EventLeave is the leave choice on the game-over screen.
	EventLeave
)

// Next returns the state that follows s on event e, applying the transition
// table. A pair outside the table leaves the state unchanged. The retry reset
// of lives and the take is the main loop's wiring, not this table.
func Next(s State, e Event) State {
	switch s {
	case Title:
		if e == EventStart {
			return Playing
		}
	case Playing:
		switch e {
		case EventDeath:
			return Dying
		case EventExit:
			return Recapping
		}
	case Dying:
		switch e {
		case EventRespawn:
			return Playing
		case EventGameOver:
			return GameOver
		}
	case Recapping:
		if e == EventStart {
			return Title
		}
	case GameOver:
		switch e {
		case EventStart:
			return Playing
		case EventLeave:
			return Title
		}
	}
	return s
}

// Line returns the intercom line that speaks when a state begins. Playing
// picks the start line, Recapping the closing line and GameOver the game-over
// line. Title and Dying speak nothing and return the empty string.
func Line(s State) string {
	switch s {
	case Playing:
		return textStart
	case Recapping:
		return textLevelEnd
	case GameOver:
		return textGameOver
	default:
		return ""
	}
}

// RespawnLine returns the respawn line, or the last-life line when lastLife is
// true.
func RespawnLine(lastLife bool) string {
	if lastLife {
		return textLastLife
	}
	return textRespawn
}

// ReelLine returns the reel-pickup line.
func ReelLine() string {
	return textReelPickup
}
