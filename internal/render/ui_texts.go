package render

import "copycat/internal/campaign"

// UI-only labels that belong to the render layer rather than the campaign
// texts: the on/off toggle values and the level and phase counters the HUD
// prints.
var (
	uiOn    = campaign.Line{EN: "On", UA: "Увімкнено"}
	uiOff   = campaign.Line{EN: "Off", UA: "Вимкнено"}
	uiLevel = campaign.Line{EN: "Level", UA: "Рівень"}
	uiPhase = campaign.Line{EN: "Phase", UA: "Фаза"}
)
