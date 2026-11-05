package render

import "fmt"

// Every player-facing HUD and menu string stands here, render's one texts
// file, so the wording can be swapped or localized together. Each line passes
// the tone map: darkly funny, deadpan, cheap-studio flavor, and nothing banned
// for PEGI 7 or ESRB E10+.

const (
	gameName    = "ШКРЯБ"
	startPrompt = "Press Enter to roll"
	retryPrompt = "Press Enter to roll again"
	leavePrompt = "Press Esc to leave"
)

// TakeText returns the take number shown in the HUD.
func TakeText() string {
	return "Take 9"
}

// ReelsText returns the found-of-total reel count, such as "12/20".
func ReelsText(found, total int) string {
	return fmt.Sprintf("%d/%d", found, total)
}

// RecapText returns the end recap's reel line, such as "Reels: 12 of 20".
func RecapText(found, total int) string {
	return fmt.Sprintf("Reels: %d of %d", found, total)
}
