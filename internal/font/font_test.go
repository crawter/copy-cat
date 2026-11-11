package font

import (
	"reflect"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
)

// everyLine returns every exported Line value the three text packages own,
// as plain values so reflect can walk each struct and slice uniformly.
func everyLine() []any {
	return []any{
		// internal/campaign.
		campaign.GameTitle, campaign.TitleTagline, campaign.MainMenuItems, campaign.MapTitle, campaign.MapLocked,
		campaign.LevelNames, campaign.SettingsTitle, campaign.SettingsItems,
		campaign.LanguagePickTitle, campaign.LanguageEnglishName, campaign.LanguageUkrainianName,
		campaign.PauseTitle, campaign.PauseHint, campaign.LevelEndTitle, campaign.LevelEndHint,
		campaign.EndlessLevelEndTitle, campaign.GameOverTitle, campaign.GameOverEpitaph,
		campaign.GameOverItems, campaign.StoryCards, campaign.LoadingTips, campaign.MedalsOfShame,
		campaign.Epitaphs, campaign.Credits, campaign.CreditsClosing,
		// internal/gameplay.
		gameplay.CauseLines, gameplay.StompedMudskipper, gameplay.SnipLine, gameplay.ChaseLine,
		gameplay.ShieldLine, gameplay.BossPhaseLines, gameplay.HUDCombs, gameplay.HUDLives,
		gameplay.HUDShield, gameplay.HUDChase, gameplay.HUDDistance, gameplay.HUDPause,
		// internal/levelgen.
		levelgen.WorldNames, levelgen.EndlessName,
	}
}

// TestEveryTextRuneHasGlyph walks every exported Line value of the text
// packages by reflection and pins that the font covers every rune of every
// English and Ukrainian string.
func TestEveryTextRuneHasGlyph(t *testing.T) {
	f := New()
	for _, v := range everyLine() {
		walkLines(t, f, reflect.ValueOf(v))
	}
}

// walkLines walks a Line value by reflection: a struct contributes every
// string field, and a slice or array contributes each element.
func walkLines(t *testing.T, f *Font, v reflect.Value) {
	t.Helper()
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.Kind() == reflect.String {
				checkRunes(t, f, field.String())
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkLines(t, f, v.Index(i))
		}
	}
}

func checkRunes(t *testing.T, f *Font, s string) {
	t.Helper()
	for _, r := range s {
		if _, ok := f.Glyph(r); !ok {
			t.Errorf("font has no glyph for %q (U+%04X) in %q", r, r, s)
		}
	}
}

// TestGlyphBounds checks that every stored glyph fits its declared size and
// has at least one set pixel.
func TestGlyphBounds(t *testing.T) {
	f := New()
	for r, g := range f.glyphs {
		if g.Rune != r {
			t.Errorf("glyph %q stores rune %q", r, g.Rune)
		}
		if g.Width <= 0 {
			t.Errorf("glyph %q has non-positive width %d", r, g.Width)
		}
		if g.Height != gridHeight {
			t.Errorf("glyph %q height = %d, want %d", r, g.Height, gridHeight)
		}
		if g.Advance < g.Width {
			t.Errorf("glyph %q advance %d < width %d", r, g.Advance, g.Width)
		}
		if len(g.Pixels) != g.Width*g.Height {
			t.Errorf("glyph %q has %d pixels, want %d", r, len(g.Pixels), g.Width*g.Height)
		}
		set := 0
		for _, p := range g.Pixels {
			if p {
				set++
			}
		}
		if set == 0 {
			t.Errorf("glyph %q has no set pixels", r)
		}
	}
}

// TestMeasureMatchesAdvanceSum pins that Measure of a known string equals
// the sum of its glyph advances.
func TestMeasureMatchesAdvanceSum(t *testing.T) {
	f := New()
	const s = "Downald D — носач, бар'єр: Ґ Є І Ї 0123"
	w, h := f.Measure(s)
	want := 0
	for _, r := range s {
		g, ok := f.Glyph(r)
		if !ok {
			t.Fatalf("test string uses an uncovered rune %q", r)
		}
		want += g.Advance
	}
	if w != want {
		t.Errorf("Measure(%q) width = %d, want %d", s, w, want)
	}
	if h != f.LineHeight() {
		t.Errorf("Measure(%q) height = %d, want line height %d", s, h, f.LineHeight())
	}
}

// TestMeasureSpacesAndEmpty pins the empty string and space handling.
func TestMeasureSpacesAndEmpty(t *testing.T) {
	f := New()
	if w, _ := f.Measure(""); w != 0 {
		t.Errorf("Measure(\"\") width = %d, want 0", w)
	}
	if w, _ := f.Measure("  "); w != 2*f.SpaceWidth() {
		t.Errorf("Measure(\"  \") width = %d, want %d", w, 2*f.SpaceWidth())
	}
}

// TestNewMetricsAndSpace pins that a new font has usable metrics and covers
// the space rune.
func TestNewMetricsAndSpace(t *testing.T) {
	f := New()
	if f.SpaceWidth() <= 0 {
		t.Errorf("SpaceWidth = %d, want > 0", f.SpaceWidth())
	}
	if f.LineHeight() <= 0 {
		t.Errorf("LineHeight = %d, want > 0", f.LineHeight())
	}
	g, ok := f.Glyph(' ')
	if !ok {
		t.Fatal("Glyph(' ') not covered")
	}
	if g.Advance != f.SpaceWidth() {
		t.Errorf("space advance = %d, want %d", g.Advance, f.SpaceWidth())
	}
}

// TestRequiredRunes pins the Ukrainian letters and punctuation the task
// calls out explicitly, in both cases, plus the apostrophe as U+0027 and
// the orthographic U+02BC.
func TestRequiredRunes(t *testing.T) {
	f := New()
	for _, r := range []rune{'Ґ', 'ґ', 'Є', 'є', 'І', 'і', 'Ї', 'ї', '\'', '\u02BC', '—', '«', '»'} {
		if _, ok := f.Glyph(r); !ok {
			t.Errorf("font missing required rune %q (U+%04X)", r, r)
		}
	}
}

// TestAlphabetAndDigitsCovered pins full Latin and digit coverage.
func TestAlphabetAndDigitsCovered(t *testing.T) {
	f := New()
	for r := 'A'; r <= 'Z'; r++ {
		if _, ok := f.Glyph(r); !ok {
			t.Errorf("font missing Latin rune %q", r)
		}
	}
	for r := 'a'; r <= 'z'; r++ {
		if _, ok := f.Glyph(r); !ok {
			t.Errorf("font missing Latin rune %q", r)
		}
	}
	for r := '0'; r <= '9'; r++ {
		if _, ok := f.Glyph(r); !ok {
			t.Errorf("font missing digit %q", r)
		}
	}
}
