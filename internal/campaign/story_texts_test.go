package campaign

import (
	"strings"
	"testing"

	"copycat/internal/gameplay"
)

// storyLines returns every story, tip, medal, epitaph and credit line the
// story texts file owns.
func storyLines() []Line {
	lines := append([]Line{}, StoryCards...)
	lines = append(lines, LoadingTips...)
	lines = append(lines, MedalsOfShame...)
	lines = append(lines, Epitaphs...)
	lines = append(lines, Credits...)
	lines = append(lines, CreditsClosing)
	return lines
}

func TestEveryStoryLineHasBothLanguages(t *testing.T) {
	for _, l := range storyLines() {
		if l.EN == "" {
			t.Errorf("story line has empty EN: %q", l.UA)
		}
		if l.UA == "" {
			t.Errorf("story line %q has empty UA", l.EN)
		}
	}
}

func TestStoryUAHasNoForbiddenRunes(t *testing.T) {
	for _, l := range storyLines() {
		if strings.ContainsAny(l.UA, "ъыэё") {
			t.Errorf("UA story line %q contains a forbidden rune", l.UA)
		}
	}
}

func TestStoryCardsCoverFourKinds(t *testing.T) {
	if len(StoryCards) != 4 {
		t.Fatalf("len(StoryCards) = %d, want 4", len(StoryCards))
	}
	for _, k := range []StoryKind{StoryIntro, StoryWorld2, StoryWorld3, StoryEnding} {
		if StoryCards[k].EN == "" || StoryCards[k].UA == "" {
			t.Errorf("StoryCards[%d] is not filled", k)
		}
	}
}

func TestMedalsAndEpitaphsCoverEveryCause(t *testing.T) {
	if len(MedalsOfShame) != 5 {
		t.Fatalf("len(MedalsOfShame) = %d, want 5", len(MedalsOfShame))
	}
	if len(Epitaphs) != 5 {
		t.Fatalf("len(Epitaphs) = %d, want 5", len(Epitaphs))
	}
	for _, c := range []gameplay.Cause{
		gameplay.CauseNone,
		gameplay.CausePit,
		gameplay.CauseBarrier,
		gameplay.CauseEnemy,
		gameplay.CauseSnip,
	} {
		if MedalsOfShame[c].EN == "" || MedalsOfShame[c].UA == "" {
			t.Errorf("MedalsOfShame[%d] is not filled", c)
		}
		if Epitaphs[c].EN == "" || Epitaphs[c].UA == "" {
			t.Errorf("Epitaphs[%d] is not filled", c)
		}
	}
}

func TestEpitaphsNameDownaldDExactly(t *testing.T) {
	for i, l := range Epitaphs {
		if !strings.Contains(l.EN, "Downald D") || !strings.Contains(l.UA, "Downald D") {
			t.Errorf("Epitaphs[%d] does not name Downald D in Latin in both languages: %q / %q", i, l.EN, l.UA)
		}
		if strings.Count(l.EN, "Downald D") != 1 || strings.Count(l.UA, "Downald D") != 1 {
			t.Errorf("Epitaphs[%d] does not contain exactly one Downald D: %q / %q", i, l.EN, l.UA)
		}
	}
}

func TestStoryCardsNameDownaldDInLatin(t *testing.T) {
	en, ua := false, false
	for _, l := range StoryCards {
		en = en || strings.Contains(l.EN, "Downald D")
		ua = ua || strings.Contains(l.UA, "Downald D")
		checkNoCyrillicName(t, l)
	}
	if !en || !ua {
		t.Fatalf("story cards do not name Downald D in Latin in both languages (EN=%v, UA=%v)", en, ua)
	}
}

func TestCreditsNameDownaldDInLatin(t *testing.T) {
	lines := append([]Line{}, Credits...)
	lines = append(lines, CreditsClosing)
	en, ua := false, false
	for _, l := range lines {
		en = en || strings.Contains(l.EN, "Downald D")
		ua = ua || strings.Contains(l.UA, "Downald D")
		checkNoCyrillicName(t, l)
	}
	if !en || !ua {
		t.Fatalf("credits do not name Downald D in Latin in both languages (EN=%v, UA=%v)", en, ua)
	}
}

func TestLoadingTipsIncludeTheDAndHairsprayGags(t *testing.T) {
	if len(LoadingTips) == 0 {
		t.Fatal("LoadingTips is empty")
	}
	enD, uaD, enSpray, uaSpray := false, false, false, false
	for _, l := range LoadingTips {
		enD = enD || strings.Contains(l.EN, "Dazzling")
		uaD = uaD || strings.Contains(l.UA, "Дивовижний")
		enSpray = enSpray || strings.Contains(l.EN, "Hairspray")
		uaSpray = uaSpray || strings.Contains(l.UA, "Лак")
	}
	if !enD || !uaD {
		t.Errorf("loading tips do not tell the D joke in both languages (EN=%v, UA=%v)", enD, uaD)
	}
	if !enSpray || !uaSpray {
		t.Errorf("loading tips do not warn about hairspray in both languages (EN=%v, UA=%v)", enSpray, uaSpray)
	}
}

func checkNoCyrillicName(t *testing.T, l Line) {
	t.Helper()
	if strings.Contains(l.EN, "Даунальд") || strings.Contains(l.UA, "Даунальд") {
		t.Errorf("line transliterates the hero's name into Cyrillic: %q / %q", l.EN, l.UA)
	}
}
