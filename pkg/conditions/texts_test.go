package conditions

import (
	"strings"
	"testing"
)

func TestSurvivalCompleteIsVerbatim(t *testing.T) {
	const want = "You survived. The district sends its regards, and an invoice for the chase."
	if SurvivalComplete != want {
		t.Fatalf("SurvivalComplete = %q, want %q", SurvivalComplete, want)
	}
}

func TestOutcomeTextsAreClean(t *testing.T) {
	lines := map[string]string{
		"SurvivalComplete":   SurvivalComplete,
		"CollectionComplete": CollectionComplete,
		"Fail":               Fail,
	}
	for name, line := range lines {
		if line == "" {
			t.Fatalf("%s is empty, want a player-facing outcome line", name)
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "puk") {
			t.Fatalf("%s = %q holds the forbidden name 'Puk'", name, line)
		}
		if strings.Contains(lower, "todo") {
			t.Fatalf("%s = %q holds a TODO marker", name, line)
		}
	}
}
