package scene

import (
	"strings"
	"testing"
)

func TestSceneTextsAreClean(t *testing.T) {
	lines := map[string]string{
		"MenuTitle":     MenuTitle,
		"MenuPrompt":    MenuPrompt,
		"RestartPrompt": RestartPrompt,
	}
	for name, line := range lines {
		if line == "" {
			t.Fatalf("%s is empty, want a player-facing scene line", name)
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
