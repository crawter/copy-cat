package render

import "copycat/internal/campaign"

// ThemePalette holds the colors one world theme paints with.
type ThemePalette struct {
	Floor       Color
	FloorAlt    Color
	Wall        Color
	Backdrop    Color
	Hazard      Color
	LowBarrier  Color
	HighBarrier Color
	Ramp        Color
	RampTop     Color
	Accent      Color
	Enemy       Color
	Comb        Color
	Hairspray   Color
	Mirror      Color
}

// Palette returns the colors for theme. An unknown theme falls back to the
// wet mangrove palette, the game's first world.
func Palette(theme campaign.Theme) ThemePalette {
	switch theme {
	case campaign.ThemeHarbor:
		return ThemePalette{
			Floor:       Color{R: 70, G: 70, B: 75, A: 255},
			FloorAlt:    Color{R: 85, G: 85, B: 90, A: 255},
			Wall:        Color{R: 95, G: 60, B: 45, A: 255},
			Backdrop:    Color{R: 40, G: 50, B: 60, A: 255},
			Hazard:      Color{R: 10, G: 12, B: 16, A: 255},
			LowBarrier:  Color{R: 140, G: 80, B: 45, A: 255},
			HighBarrier: Color{R: 70, G: 75, B: 80, A: 255},
			Ramp:        Color{R: 150, G: 60, B: 40, A: 255},
			RampTop:     Color{R: 170, G: 85, B: 55, A: 255},
			Accent:      Color{R: 220, G: 170, B: 40, A: 255},
			Enemy:       Color{R: 110, G: 110, B: 120, A: 255},
			Comb:        Color{R: 230, G: 180, B: 40, A: 255},
			Hairspray:   Color{R: 200, G: 200, B: 210, A: 255},
			Mirror:      Color{R: 190, G: 200, B: 210, A: 255},
		}
	case campaign.ThemeNeon:
		return ThemePalette{
			Floor:       Color{R: 25, G: 25, B: 40, A: 255},
			FloorAlt:    Color{R: 35, G: 30, B: 55, A: 255},
			Wall:        Color{R: 30, G: 30, B: 60, A: 255},
			Backdrop:    Color{R: 60, G: 20, B: 80, A: 255},
			Hazard:      Color{R: 30, G: 0, B: 50, A: 255},
			LowBarrier:  Color{R: 230, G: 60, B: 160, A: 255},
			HighBarrier: Color{R: 40, G: 220, B: 220, A: 255},
			Ramp:        Color{R: 90, G: 40, B: 200, A: 255},
			RampTop:     Color{R: 150, G: 90, B: 240, A: 255},
			Accent:      Color{R: 70, G: 240, B: 120, A: 255},
			Enemy:       Color{R: 240, G: 150, B: 40, A: 255},
			Comb:        Color{R: 230, G: 180, B: 40, A: 255},
			Hairspray:   Color{R: 200, G: 200, B: 210, A: 255},
			Mirror:      Color{R: 190, G: 200, B: 210, A: 255},
		}
	default:
		return ThemePalette{
			Floor:       Color{R: 40, G: 70, B: 55, A: 255},
			FloorAlt:    Color{R: 55, G: 90, B: 70, A: 255},
			Wall:        Color{R: 45, G: 80, B: 60, A: 255},
			Backdrop:    Color{R: 30, G: 60, B: 55, A: 255},
			Hazard:      Color{R: 10, G: 30, B: 35, A: 255},
			LowBarrier:  Color{R: 90, G: 60, B: 30, A: 255},
			HighBarrier: Color{R: 60, G: 90, B: 40, A: 255},
			Ramp:        Color{R: 80, G: 55, B: 28, A: 255},
			RampTop:     Color{R: 110, G: 80, B: 40, A: 255},
			Accent:      Color{R: 60, G: 140, B: 80, A: 255},
			Enemy:       Color{R: 70, G: 120, B: 70, A: 255},
			Comb:        Color{R: 230, G: 180, B: 40, A: 255},
			Hairspray:   Color{R: 200, G: 200, B: 210, A: 255},
			Mirror:      Color{R: 190, G: 200, B: 210, A: 255},
		}
	}
}
