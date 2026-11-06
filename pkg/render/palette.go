// Package render holds the game's five-colour palette and pure camera
// placement math as plain values. It imports no raylib, so the later draw
// layer converts these values and the math stays testable headless.
package render

// Color is a plain RGBA colour with one byte per channel. It carries no
// raylib type so the palette stays independent of the drawing layer and can
// be tested headless.
type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

// MurkGreen is the swamp corridor's base green. The RGB values are assumed
// and stay easy to change in a later review.
var MurkGreen = Color{R: 56, G: 87, B: 51, A: 255}

// BogBrown is the swamp corridor's earthy brown, used for banks and crates.
// The RGB values are assumed and stay easy to change in a later review.
var BogBrown = Color{R: 89, G: 66, B: 42, A: 255}

// MossYellow is the jar lids' dull mossy yellow, the one junk colour.
// The RGB values are assumed and stay easy to change in a later review.
var MossYellow = Color{R: 180, G: 165, B: 60, A: 255}

// WarningOrange is the crew's signal orange, used for enemies and warnings.
// The RGB values are assumed and stay easy to change in a later review.
var WarningOrange = Color{R: 220, G: 120, B: 40, A: 255}

// FogWhite is the swamp sky's pale fog white. The RGB values are assumed and
// stay easy to change in a later review.
var FogWhite = Color{R: 235, G: 238, B: 226, A: 255}
