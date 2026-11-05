package scenestate

// Every player-facing intercom line stands here, in the package's one texts
// file, so the lines can be swapped or localized together. Each line passes
// the tone map: black-but-decent, darkly funny about the hero's failures and
// the studio's poverty, nothing banned for PEGI 7 or ESRB E10+, and all
// original.

const (
	// textStart is the director's opening line as Playing begins.
	textStart = "Take nine. Camera rolling. Try not to die before the first reel, Шкряб."

	// textRespawn is the director's line as the hero returns at a checkpoint.
	textRespawn = "That was a take. Reset the cat."

	// textReelPickup is the director's line when a film reel is collected.
	textReelPickup = "Found the film. Now it can lose money twice."

	// textLastLife is the director's line when one life remains.
	textLastLife = "One life left. The studio's insurance covers exactly nothing."

	// textGameOver is the director's line when the ninth life is spent.
	textGameOver = "That's a wrap. Someone sweep up the cat."

	// textLevelEnd is the director's closing line after the exit.
	textLevelEnd = "Cut! Print it. The check is in the mail."
)
