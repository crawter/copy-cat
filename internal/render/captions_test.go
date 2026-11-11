package render

import (
	"strings"
	"testing"

	"copycat/internal/campaign"
)

func captionFlow(lang campaign.Language) *campaign.Flow {
	s := campaign.DefaultSettings()
	s.Language = lang
	return campaign.NewFlow(campaign.Default(), s, campaign.NewProgress(), false)
}

func TestCombCaptionShowsTotal(t *testing.T) {
	f := captionFlow(campaign.LanguageEnglish)
	lines := CombCaption(f, 7)
	if len(lines) != 1 {
		t.Fatalf("comb caption lines = %d, want 1", len(lines))
	}
	if !strings.Contains(lines[0].Text, "7") {
		t.Fatalf("comb caption %q missing total", lines[0].Text)
	}
	if !strings.Contains(lines[0].Text, textFor(f, uiCombCaption)) {
		t.Fatalf("comb caption %q missing label %q", lines[0].Text, textFor(f, uiCombCaption))
	}
}

func TestSnipCaptionOnlyWhenBald(t *testing.T) {
	f := captionFlow(campaign.LanguageEnglish)
	if got := SnipCaption(f, false, 2); len(got) != 0 {
		t.Fatalf("snip caption when not bald = %v, want none", got)
	}

	lines := SnipCaption(f, true, 2)
	if len(lines) != 1 {
		t.Fatalf("snip caption lines = %d, want 1", len(lines))
	}
	if !strings.Contains(lines[0].Text, "2") {
		t.Fatalf("snip caption %q missing lives lost", lines[0].Text)
	}
	if !strings.Contains(lines[0].Text, textFor(f, uiSnipCaption)) {
		t.Fatalf("snip caption %q missing label %q", lines[0].Text, textFor(f, uiSnipCaption))
	}
}

func TestLandingCaptionLocalized(t *testing.T) {
	en := captionFlow(campaign.LanguageEnglish)
	ua := captionFlow(campaign.LanguageUkrainian)

	enLines := LandingCaption(en)
	uaLines := LandingCaption(ua)
	if len(enLines) != 1 || len(uaLines) != 1 {
		t.Fatalf("landing caption lines = %d/%d, want 1/1", len(enLines), len(uaLines))
	}
	if enLines[0].Text == uaLines[0].Text {
		t.Fatalf("landing caption is not localized: %q", enLines[0].Text)
	}
	if got, want := enLines[0].Text, textFor(en, uiLandCaption); got != want {
		t.Fatalf("EN landing caption = %q, want %q", got, want)
	}
}
