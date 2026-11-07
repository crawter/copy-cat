package campaign

import (
	"strings"
	"testing"
)

// allLines returns every player-facing Line the package owns.
func allLines() []Line {
	lines := []Line{
		GameTitle,
		MapTitle,
		MapLocked,
		SettingsTitle,
		LanguagePickTitle,
		LanguageEnglishName,
		LanguageUkrainianName,
		PauseTitle,
		PauseHint,
		LevelEndTitle,
		LevelEndHint,
		EndlessLevelEndTitle,
		GameOverTitle,
		GameOverEpitaph,
	}
	lines = append(lines, MainMenuItems...)
	lines = append(lines, LevelNames...)
	lines = append(lines, SettingsItems...)
	lines = append(lines, GameOverItems...)
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
		t.Fatalf("campaign texts do not name Downald D in Latin in both languages (EN=%v, UA=%v)", en, ua)
	}
}

func TestLevelNamesCoverTenLevelsWithBossLast(t *testing.T) {
	if len(LevelNames) != 10 {
		t.Fatalf("len(LevelNames) = %d, want 10", len(LevelNames))
	}
	last := LevelNames[len(LevelNames)-1]
	if last.EN != "The Last Haircut" || last.UA != "Остання стрижка" {
		t.Fatalf("boss level name = %+v, want The Last Haircut / Остання стрижка", last)
	}
}

func TestMenuItemCounts(t *testing.T) {
	if len(MainMenuItems) != 6 {
		t.Errorf("len(MainMenuItems) = %d, want 6", len(MainMenuItems))
	}
	if len(SettingsItems) != 4 {
		t.Errorf("len(SettingsItems) = %d, want 4", len(SettingsItems))
	}
	if len(GameOverItems) != 2 {
		t.Errorf("len(GameOverItems) = %d, want 2", len(GameOverItems))
	}
}
