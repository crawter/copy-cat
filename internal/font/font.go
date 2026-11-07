// Package font draws a pixel font in code for Latin and Ukrainian. Every
// glyph is a small bitmap built from one shared grid and baseline, so the
// render layers can draw every player-facing text without an asset file or
// a Cyrillic-capable system font.
package font

import "fmt"

// Glyph is one glyph's bitmap and metrics. Pixels is row-major: the first
// Width entries are the top row, and there are Width*Height entries in all.
type Glyph struct {
	Rune    rune
	Width   int
	Height  int
	Advance int
	Pixels  []bool
}

// Font is a set of glyphs with the metrics shared by every glyph in the
// set: one grid height and one baseline.
type Font struct {
	glyphs     map[rune]Glyph
	spaceWidth int
	lineHeight int
}

// The shared pixel grid. Row 0 holds diacritics, rows 1..6 hold the body,
// row 6 is the baseline, and row 7 holds descenders.
const (
	gridHeight = 8

	// defaultSpaceWidth is the advance of a space.
	defaultSpaceWidth = 4
	// defaultLineHeight is the distance between successive lines, a little
	// taller than the grid so consecutive lines do not touch.
	defaultLineHeight = 10
)

// New returns a Font covering the Latin alphabet, digits, common
// punctuation and the Ukrainian alphabet, including Ґ, Є, І, Ї and the
// apostrophe. The returned font is read-only and safe to share.
func New() *Font {
	f := &Font{
		glyphs:     make(map[rune]Glyph, len(glyphSpecs)),
		spaceWidth: defaultSpaceWidth,
		lineHeight: defaultLineHeight,
	}
	for _, s := range glyphSpecs {
		f.glyphs[s.r] = Glyph{
			Rune:    s.r,
			Width:   s.width,
			Height:  gridHeight,
			Advance: s.width + 1,
			Pixels:  buildPixels(s),
		}
	}
	return f
}

// Glyph returns the glyph for r and whether the font covers it. A space
// reports true with a blank glyph whose advance is SpaceWidth. The returned
// Pixels slice is shared by every call and must be treated as read-only.
func (f *Font) Glyph(r rune) (Glyph, bool) {
	if r == ' ' {
		return Glyph{Rune: ' ', Height: gridHeight, Advance: f.spaceWidth}, true
	}
	g, ok := f.glyphs[r]
	return g, ok
}

// SpaceWidth returns the advance, in pixels, of a space.
func (f *Font) SpaceWidth() int { return f.spaceWidth }

// LineHeight returns the distance, in pixels, between successive lines.
func (f *Font) LineHeight() int { return f.lineHeight }

// Measure returns the exact pixel size of s drawn on one line: the sum of
// each rune's advance and the font's line height. A rune the font does not
// cover advances by the space width so text never collapses silently.
func (f *Font) Measure(s string) (int, int) {
	w := 0
	for _, r := range s {
		if g, ok := f.Glyph(r); ok {
			w += g.Advance
		} else {
			w += f.spaceWidth
		}
	}
	return w, f.lineHeight
}

// buildPixels turns a glyph spec's '#'/'.' rows into a row-major bitmap. It
// panics on a malformed spec, which is a programmer error in the table.
func buildPixels(s glyphSpec) []bool {
	if len(s.rows) != gridHeight {
		panic(fmt.Sprintf("font: glyph %q has %d rows, want %d", s.r, len(s.rows), gridHeight))
	}
	pixels := make([]bool, s.width*gridHeight)
	for row, line := range s.rows {
		if len(line) != s.width {
			panic(fmt.Sprintf("font: glyph %q row %d has width %d, want %d", s.r, row, len(line), s.width))
		}
		for col := 0; col < s.width; col++ {
			pixels[row*s.width+col] = line[col] == '#'
		}
	}
	return pixels
}
