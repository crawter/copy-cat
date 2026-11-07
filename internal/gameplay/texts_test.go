package gameplay

import (
	"strings"
	"testing"
)

// allLines returns every player-facing Line the package owns.
func allLines() []Line {
	lines := []Line{
		StompedMudskipper,
		SnipLine,
		ChaseLine,
		ShieldLine,
		HUDCombs,
		HUDLives,
		HUDShield,
		HUDChase,
		HUDDistance,
		HUDPause,
	}
	lines = append(lines, CauseLines...)
	lines = append(lines, BossPhaseLines...)
	return lines
}

func TestEveryLineHasBothLanguages(t *testing.T) {
	for _, l := range allLines() {
		if l.EN == "" {
			t.Errorf("line has empty EN")
		}
		if l.UA == "" {
			t.Errorf("line %q has empty UA", l.EN)
		}
	}
}

func TestUkrainianHasNoForbiddenRunes(t *testing.T) {
	for _, l := range allLines() {
		if strings.ContainsAny(l.UA, "ъыэё") {
			t.Errorf("UA line %q contains a forbidden rune", l.UA)
		}
	}
}

func TestNamesDownaldDInLatin(t *testing.T) {
	en, ua := false, false
	for _, l := range allLines() {
		if strings.Contains(l.EN, "Downald D") {
			en = true
		}
		if strings.Contains(l.UA, "Downald D") {
			ua = true
		}
		if strings.Contains(l.EN, "Даунальд") || strings.Contains(l.UA, "Даунальд") {
			t.Errorf("line transliterates the hero's name into Cyrillic: %q / %q", l.EN, l.UA)
		}
	}
	if !en || !ua {
		t.Fatalf("gameplay texts do not name Downald D in Latin in both languages (EN=%v, UA=%v)", en, ua)
	}
}

func TestCauseLinesCoverEveryCause(t *testing.T) {
	if len(CauseLines) != 5 {
		t.Fatalf("len(CauseLines) = %d, want 5", len(CauseLines))
	}
	for _, c := range []Cause{CauseNone, CausePit, CauseBarrier, CauseEnemy, CauseSnip} {
		if CauseLines[c].EN == "" || CauseLines[c].UA == "" {
			t.Errorf("CauseLines[%d] is not filled", c)
		}
	}
}

func TestBossPhaseLinesCoverThreePhases(t *testing.T) {
	if len(BossPhaseLines) != 3 {
		t.Fatalf("len(BossPhaseLines) = %d, want 3", len(BossPhaseLines))
	}
}
