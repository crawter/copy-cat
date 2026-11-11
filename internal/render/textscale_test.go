package render

import (
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/font"
)

func TestMainMenuLogoAndTagline(t *testing.T) {
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	lines := MenuLines(f)
	if len(lines) < 2 {
		t.Fatalf("main menu lines = %d, want at least 2", len(lines))
	}
	if got, want := lines[0].Text, textFor(f, campaign.GameTitle); got != want {
		t.Fatalf("logo text = %q, want %q", got, want)
	}
	if got, want := lines[0].Scale, uiLogoScale; got != want {
		t.Fatalf("logo scale = %d, want %d", got, want)
	}
	if got, want := lines[1].Text, textFor(f, campaign.TitleTagline); got != want {
		t.Fatalf("tagline text = %q, want %q", got, want)
	}
	if got, want := lines[1].Scale, uiScale; got != want {
		t.Fatalf("tagline scale = %d, want body scale %d", got, want)
	}
}

func TestScreenTitlesUseHeadingScale(t *testing.T) {
	cases := []struct {
		name      string
		flow      func(t *testing.T) *campaign.Flow
		titleLine campaign.Line
	}{
		{"map", mapFlow, campaign.MapTitle},
		{"settings", settingsFlow, campaign.SettingsTitle},
		{"pause", func(t *testing.T) *campaign.Flow {
			f := newGameFlow(t)
			f.Handle(campaign.ActionPause)
			return f
		}, campaign.PauseTitle},
		{"level end", func(t *testing.T) *campaign.Flow {
			f := newGameFlow(t)
			stepToScene(t, f, campaign.SceneLevelEnd)
			return f
		}, campaign.LevelEndTitle},
		{"game over", gameOverFlow, campaign.GameOverTitle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.flow(t)
			checkHeadingTitle(t, MenuLines(f), textFor(f, tc.titleLine))
		})
	}
}

// checkHeadingTitle asserts the first line is the heading at the heading
// scale and every following body line stays at the body scale.
func checkHeadingTitle(t *testing.T, lines []ScreenLine, wantText string) {
	t.Helper()
	if len(lines) == 0 {
		t.Fatal("screen has no lines")
	}
	if got, want := lines[0].Text, wantText; got != want {
		t.Fatalf("title text = %q, want %q", got, want)
	}
	if got, want := lines[0].Scale, uiHeadingScale; got != want {
		t.Fatalf("title scale = %d, want %d", got, want)
	}
	for i := 1; i < len(lines); i++ {
		if got, want := lines[i].Scale, uiScale; got != want {
			t.Fatalf("body line %d scale = %d, want %d", i, got, want)
		}
	}
}

func TestAtLeastThreeDistinctScales(t *testing.T) {
	seen := map[int]bool{}
	add := func(lines []ScreenLine) {
		for _, l := range lines {
			seen[l.Scale] = true
		}
	}

	menu := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	add(MenuLines(menu))
	add(MenuLines(mapFlow(t)))

	if len(seen) < 3 {
		t.Fatalf("distinct scales = %d, want at least 3: %v", len(seen), seen)
	}
	for _, want := range []int{uiScale, uiHeadingScale, uiLogoScale} {
		if !seen[want] {
			t.Fatalf("scales %v missing %d", seen, want)
		}
	}
}

func TestStoryLinesKeepBodyScale(t *testing.T) {
	cases := []struct {
		name string
		flow func(t *testing.T) *campaign.Flow
	}{
		{"story", storyFlow},
		{"ending", endingFlow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := StoryLines(tc.flow(t))
			if len(lines) == 0 {
				t.Fatal("story lines are empty")
			}
			for i, l := range lines {
				if got, want := l.Scale, uiScale; got != want {
					t.Fatalf("line %d scale = %d, want body scale %d", i, got, want)
				}
			}
		})
	}
}

func TestUILineDefaultsToBodyScale(t *testing.T) {
	l := uiLine("hair raid", 0, 0, false)
	if got, want := l.Scale, uiScale; got != want {
		t.Fatalf("uiLine scale = %d, want %d", got, want)
	}
}

func TestBlockWidthGrowsWithScale(t *testing.T) {
	u := NewUI(font.New())
	text := "Hair Raid"
	low := u.blockWidth([]ScreenLine{{Text: text, Scale: 1}})
	high := u.blockWidth([]ScreenLine{{Text: text, Scale: 2}})
	if high <= low {
		t.Fatalf("blockWidth at scale 2 = %d, want > scale 1 = %d", high, low)
	}
	// A zero scale clamps to at least one, so it measures like scale 1.
	if got, want := u.blockWidth([]ScreenLine{{Text: text}}), low; got != want {
		t.Fatalf("blockWidth at scale 0 = %d, want %d (clamped to 1)", got, want)
	}
}
