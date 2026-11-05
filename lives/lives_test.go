package lives

import (
	"testing"

	"copycat/leveldata"
)

func TestNewStartsAtStart(t *testing.T) {
	l := New()

	if got := l.Lives(); got != startLives {
		t.Fatalf("Lives = %d, want %d", got, startLives)
	}
	if l.GameOver() {
		t.Fatal("GameOver = true on a new session")
	}
	var origin leveldata.Point
	if got := l.Respawn(); got != origin {
		t.Fatalf("Respawn = %+v, want the take's start %+v", got, origin)
	}
}

func TestOneLifePerDeath(t *testing.T) {
	l := New()
	for death := 1; death <= startLives; death++ {
		before := l.Lives()
		next, _ := l.Die()
		if got := next.Lives(); got != before-1 {
			t.Fatalf("death %d: Lives = %d, want %d", death, got, before-1)
		}
		l = next
	}
}

func TestGameOverAtZero(t *testing.T) {
	l := New()
	for death := 1; death <= 8; death++ {
		var ok bool
		l, ok = l.Die()
		if !ok {
			t.Fatalf("death %d: respawn = false, want true", death)
		}
	}
	if got := l.Lives(); got != 1 {
		t.Fatalf("after eight deaths Lives = %d, want 1", got)
	}
	if l.GameOver() {
		t.Fatal("after eight deaths GameOver = true, want false")
	}

	next, ok := l.Die()
	if ok {
		t.Fatal("ninth death respawn = true, want false")
	}
	if got := next.Lives(); got != 0 {
		t.Fatalf("after ninth death Lives = %d, want 0", got)
	}
	if !next.GameOver() {
		t.Fatal("after ninth death GameOver = false, want true")
	}
}

func TestRespawnAtCheckpoint(t *testing.T) {
	l := New()
	first := leveldata.Point{X: 4, Y: 1, Z: 100}
	l = l.AtCheckpoint(first)

	if got := l.Respawn(); got != first {
		t.Fatalf("Respawn = %+v, want checkpoint %+v", got, first)
	}

	var ok bool
	l, ok = l.Die()
	if !ok {
		t.Fatal("Die = false, want true while lives remain")
	}
	if got := l.Respawn(); got != first {
		t.Fatalf("Respawn after death = %+v, want checkpoint %+v", got, first)
	}

	second := leveldata.Point{X: 4, Y: 1, Z: 1133}
	l = l.AtCheckpoint(second)
	if got := l.Respawn(); got != second {
		t.Fatalf("Respawn = %+v, want later checkpoint %+v", got, second)
	}
	if l.Respawn() == first {
		t.Fatal("later checkpoint did not replace the earlier one")
	}
}

func TestNewAfterGameOverIsTheRetry(t *testing.T) {
	l := New()
	for l.Lives() > 0 {
		l, _ = l.Die()
	}
	if !l.GameOver() {
		t.Fatal("expected game over after spending every life")
	}

	retry := New()
	if got := retry.Lives(); got != startLives {
		t.Fatalf("retry Lives = %d, want %d", got, startLives)
	}
	if retry.GameOver() {
		t.Fatal("retry GameOver = true")
	}
	var origin leveldata.Point
	if got := retry.Respawn(); got != origin {
		t.Fatalf("retry Respawn = %+v, want the take's start %+v", got, origin)
	}
}

// TestValuesArePlainCopies pins down that a session is a plain Go value: no
// file or device is involved, and mutating one copy never leaks into another.
func TestValuesArePlainCopies(t *testing.T) {
	l := New()
	l = l.AtCheckpoint(leveldata.Point{X: 4, Y: 1, Z: 566})

	snapshot := l
	next, ok := snapshot.Die()
	if !ok {
		t.Fatal("snapshot Die = false, want true")
	}
	if got := l.Lives(); got != startLives {
		t.Fatalf("original Lives changed by a copy's death: %d, want %d", got, startLives)
	}
	if got := next.Lives(); got != startLives-1 {
		t.Fatalf("copy after death Lives = %d, want %d", got, startLives-1)
	}
}
