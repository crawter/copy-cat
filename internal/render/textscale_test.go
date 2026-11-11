package render

import (
	"testing"

	"copycat/internal/font"
)

func textScaleUI() *UI {
	return NewUI(font.New())
}

func TestBlockWidthHonorsLineScale(t *testing.T) {
	u := textScaleUI()
	w, _ := u.fnt.Measure("A")
	lines := []ScreenLine{{Text: "A", X: 5, Y: 0, Scale: 3}}
	if got, want := u.blockWidth(lines), 5+w*3; got != want {
		t.Fatalf("blockWidth = %d, want %d", got, want)
	}
}

func TestBlockHeightHonorsLineScale(t *testing.T) {
	u := textScaleUI()
	lines := []ScreenLine{{Text: "A", X: 0, Y: 4, Scale: 2}}
	if got, want := u.blockHeight(lines), 4+u.fnt.LineHeight()*2; got != want {
		t.Fatalf("blockHeight = %d, want %d", got, want)
	}
}

func TestScreenLineZeroScaleDefaultsToOne(t *testing.T) {
	u := textScaleUI()
	w, _ := u.fnt.Measure("A")
	lines := []ScreenLine{{Text: "A", X: 0, Y: 0}} // Scale left zero
	if got, want := u.blockWidth(lines), w; got != want {
		t.Fatalf("blockWidth with zero scale = %d, want %d", got, want)
	}
	if got, want := u.blockHeight(lines), u.fnt.LineHeight(); got != want {
		t.Fatalf("blockHeight with zero scale = %d, want %d", got, want)
	}
}

func TestMixedScaleLayout(t *testing.T) {
	u := textScaleUI()
	titleW, _ := u.fnt.Measure("Hair")
	menuW, _ := u.fnt.Measure("Map")
	lines := []ScreenLine{
		{Text: "Hair", X: 0, Y: 0, Scale: 4},
		{Text: "Map", X: 4, Y: 40, Scale: 1},
	}

	wantWidth := titleW * 4
	if menu := 4 + menuW; menu > wantWidth {
		wantWidth = menu
	}
	if got := u.blockWidth(lines); got != wantWidth {
		t.Fatalf("blockWidth = %d, want %d", got, wantWidth)
	}
	if got, want := u.blockHeight(lines), 40+u.fnt.LineHeight(); got != want {
		t.Fatalf("blockHeight = %d, want %d", got, want)
	}
}
