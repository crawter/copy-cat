// Texts for the generated worlds: the three campaign world names and the
// endless run name, in English and modern academic Ukrainian. Every value is
// a string literal so the font step can scan the file.
package levelgen

// Line is one player-facing string in English and Ukrainian. The render
// layer picks a field by the campaign language setting.
type Line struct {
	EN string
	UA string
}

// WorldNames holds the three campaign world names in theme order: the wet
// mangroves, the rusty harbor and the neon Chic-Chic street.
var WorldNames = []Line{
	{EN: "Wet Mangroves", UA: "Мокрі мангри"},
	{EN: "Rusty Harbor", UA: "Іржава гавань"},
	{EN: "Chic-Chic Street", UA: "Вулиця Чик-Чик"},
}

// EndlessName is the name of the endless run on its menu entry and HUD.
var EndlessName = Line{EN: "Downald D's Endless Run", UA: "Нескінченний забіг Downald D"}
