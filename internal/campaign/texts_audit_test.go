package campaign

import (
	"reflect"
	"strings"
	"testing"

	"copycat/internal/gameplay"
	"copycat/internal/levelgen"
)

// textsAudit walks every exported Line value of one texts file by reflection
// and checks both languages, the hero's Latin name and the 2019 Ukrainian
// orthography smoke rules.
type textsAudit struct {
	t       *testing.T
	name    string
	lines   int
	enNamed bool
	uaNamed bool
}

func newTextsAudit(t *testing.T, name string) *textsAudit {
	t.Helper()
	return &textsAudit{t: t, name: name}
}

// walk reflects over every value it is given, descending into slices, arrays
// and pointers, and checks each Line it finds.
func (a *textsAudit) walk(values ...any) {
	for _, v := range values {
		a.walkValue(reflect.ValueOf(v))
	}
}

func (a *textsAudit) walkValue(rv reflect.Value) {
	if !rv.IsValid() {
		return
	}
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return
		}
		a.walkValue(rv.Elem())
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			a.walkValue(rv.Index(i))
		}
	case reflect.Struct:
		en := rv.FieldByName("EN")
		ua := rv.FieldByName("UA")
		if en.IsValid() && ua.IsValid() && en.Kind() == reflect.String && ua.Kind() == reflect.String {
			a.check(en.String(), ua.String())
		}
	}
}

// check audits one Line's English and Ukrainian strings.
func (a *textsAudit) check(en, ua string) {
	a.lines++
	if en == "" {
		a.t.Errorf("%s: line %d has empty EN", a.name, a.lines)
	}
	if ua == "" {
		a.t.Errorf("%s: line %d %q has empty UA", a.name, a.lines, en)
	}
	if strings.Contains(en, "Downald D") {
		a.enNamed = true
	}
	if strings.Contains(ua, "Downald D") {
		a.uaNamed = true
	}
	for _, s := range []string{en, ua} {
		if strings.Contains(s, "Даунальд") {
			a.t.Errorf("%s: line %q transliterates Downald D into Cyrillic", a.name, s)
		}
	}
	if strings.ContainsAny(ua, "ъыэё") {
		a.t.Errorf("%s: UA line %q contains a forbidden rune", a.name, ua)
	}
	for _, phrase := range []string{"на протязі", "приймати участь"} {
		if strings.Contains(ua, phrase) {
			a.t.Errorf("%s: UA line %q uses %q, want the 2019-orthography form", a.name, ua, phrase)
		}
	}
}

// finish asserts the whole-file properties of the walked Lines.
func (a *textsAudit) finish() {
	if a.lines == 0 {
		a.t.Errorf("%s: no Lines walked", a.name)
	}
	if !a.enNamed || !a.uaNamed {
		a.t.Errorf("%s: does not name Downald D in Latin in both languages (EN=%v, UA=%v)", a.name, a.enNamed, a.uaNamed)
	}
}

func TestTextsAuditAcrossPackages(t *testing.T) {
	tests := []struct {
		name   string
		values []any
	}{
		{
			name: "internal/levelgen/texts.go",
			values: []any{
				levelgen.WorldNames,
				levelgen.EndlessName,
			},
		},
		{
			name: "internal/gameplay/texts.go",
			values: []any{
				gameplay.CauseLines,
				gameplay.StompedMudskipper,
				gameplay.SnipLine,
				gameplay.ChaseLine,
				gameplay.ShieldLine,
				gameplay.BossPhaseLines,
				gameplay.HUDCombs,
				gameplay.HUDLives,
				gameplay.HUDShield,
				gameplay.HUDChase,
				gameplay.HUDDistance,
				gameplay.HUDPause,
			},
		},
		{
			name: "internal/campaign/texts.go",
			values: []any{
				GameTitle,
				MainMenuItems,
				NewGameConfirmTitle,
				NewGameConfirmItems,
				MapTitle,
				MapLocked,
				LevelNames,
				SettingsTitle,
				SettingsItems,
				LanguagePickTitle,
				LanguageEnglishName,
				LanguageUkrainianName,
				PauseTitle,
				PauseHint,
				PauseItems,
				LevelEndTitle,
				LevelEndHint,
				EndlessLevelEndTitle,
				GameOverTitle,
				GameOverEpitaph,
				GameOverItems,
			},
		},
		{
			name: "internal/campaign/story_texts.go",
			values: []any{
				StoryCards,
				LoadingTips,
				MedalsOfShame,
				Epitaphs,
				Credits,
				CreditsClosing,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTextsAudit(t, tc.name)
			a.walk(tc.values...)
			a.finish()
		})
	}
}
