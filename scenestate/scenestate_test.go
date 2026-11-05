package scenestate

import (
	"strings"
	"testing"
)

func TestNextMovesOnlyOnSetEvents(t *testing.T) {
	listed := map[pair]State{
		{state: Title, event: EventStart}:     Playing,
		{state: Playing, event: EventDeath}:   Dying,
		{state: Playing, event: EventExit}:    Recapping,
		{state: Dying, event: EventRespawn}:   Playing,
		{state: Dying, event: EventGameOver}:  GameOver,
		{state: GameOver, event: EventStart}:  Playing,
		{state: GameOver, event: EventLeave}:  Title,
		{state: Recapping, event: EventStart}: Title,
	}

	states := []State{Title, Playing, Dying, Recapping, GameOver}
	events := []Event{EventStart, EventDeath, EventRespawn, EventExit, EventGameOver, EventLeave}

	for _, s := range states {
		for _, e := range events {
			want, isListed := listed[pair{state: s, event: e}]
			if !isListed {
				want = s
			}
			if got := Next(s, e); got != want {
				t.Errorf("Next(%v, %v) = %v, want %v", s, e, got, want)
			}
		}
	}
}

type pair struct {
	state State
	event Event
}

func TestFullSessionWalk(t *testing.T) {
	steps := []struct {
		from  State
		event Event
		want  State
	}{
		{from: Title, event: EventStart, want: Playing},
		{from: Playing, event: EventDeath, want: Dying},
		{from: Dying, event: EventRespawn, want: Playing},
		{from: Playing, event: EventDeath, want: Dying},
		{from: Dying, event: EventGameOver, want: GameOver},
		{from: GameOver, event: EventStart, want: Playing},
		{from: Playing, event: EventExit, want: Recapping},
		{from: Recapping, event: EventStart, want: Title},
	}

	s := Title
	for i, step := range steps {
		if s != step.from {
			t.Fatalf("step %d: state = %v, want walk to begin at %v", i, s, step.from)
		}
		got := Next(s, step.event)
		if got != step.want {
			t.Fatalf("step %d: Next(%v, %v) = %v, want %v", i, s, step.event, got, step.want)
		}
		s = got
	}
}

func TestLineByState(t *testing.T) {
	cases := []struct {
		state State
		want  string
	}{
		{state: Playing, want: "Take nine. Camera rolling. Try not to die before the first reel, Шкряб."},
		{state: Recapping, want: "Cut! Print it. The check is in the mail."},
		{state: GameOver, want: "That's a wrap. Someone sweep up the cat."},
		{state: Title, want: ""},
		{state: Dying, want: ""},
	}
	for _, tc := range cases {
		if got := Line(tc.state); got != tc.want {
			t.Errorf("Line(%v) = %q, want %q", tc.state, got, tc.want)
		}
	}
}

func TestRespawnAndReelLines(t *testing.T) {
	if got := RespawnLine(false); got != "That was a take. Reset the cat." {
		t.Errorf("RespawnLine(false) = %q, want the respawn line", got)
	}
	if got := RespawnLine(true); got != "One life left. The studio's insurance covers exactly nothing." {
		t.Errorf("RespawnLine(true) = %q, want the last-life line", got)
	}
	if got := ReelLine(); got != "Found the film. Now it can lose money twice." {
		t.Errorf("ReelLine() = %q, want the reel line", got)
	}
}

func TestLinesPassToneMap(t *testing.T) {
	banned := []string{
		"blood", "gore", "dismember", "sex", "drug", "hell", "damn", "crap",
	}

	lines := []string{
		Line(Playing),
		Line(Recapping),
		Line(GameOver),
		RespawnLine(false),
		RespawnLine(true),
		ReelLine(),
	}

	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, term := range banned {
			if strings.Contains(lower, term) {
				t.Errorf("line %q contains banned term %q", line, term)
			}
		}
	}
}
