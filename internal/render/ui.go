package render

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"

	"copycat/internal/campaign"
	"copycat/internal/font"
	"copycat/internal/gameplay"
)

// ScreenLine is one line of UI text: what it says, where it sits in the
// unscaled screen layout, and whether the menu cursor highlights it.
type ScreenLine struct {
	Text        string
	X, Y        int
	Highlighted bool
}

// UI draws the game's screens from a flow's plain state. It is created once
// with the pixel font and reused every frame.
type UI struct {
	fnt   *font.Font
	scale int
	anim  float32 // seconds of animation elapsed, advanced every frame
}

// NewUI returns a UI that draws its text with fnt. The font is read-only and
// safe to share.
func NewUI(fnt *font.Font) *UI {
	return &UI{fnt: fnt, scale: uiScale}
}

// The screen layout is described in the font's unscaled pixel grid; Draw
// scales it to the window.
const (
	uiScale   = 2
	uiMarginX = 8
	uiMarginY = 8
	uiIndent  = 4
	uiLineGap = 10 // matches font.Font's default line height

	// storyWrapChars is the maximum rune count a story or ending line holds
	// before it wraps to the next line.
	storyWrapChars = 60
)

// UI colors, drawn behind and over the screen text.
var (
	uiBackground    = Color{R: 24, G: 20, B: 28, A: 255}
	uiTextColor     = Color{R: 236, G: 232, B: 240, A: 255}
	uiHighlightText = Color{R: 40, G: 30, B: 20, A: 255}
	uiHighlightBar  = Color{R: 240, G: 202, B: 112, A: 255}
)

// cameraPerspective is raylib's CameraPerspective (0). It stays an untyped
// constant so the assignment compiles against either int32 or the binding's
// named CameraProjection field type.
const cameraPerspective = 0

// Draw renders the scene for the flow's current state in a window of w by h
// pixels. It clears the window, draws the 3D corridor or boss arena while
// playing or paused, then draws the matching screen text.
func (u *UI) Draw(f *campaign.Flow, w, h int) {
	rl.ClearBackground(rlColor(uiBackground))
	u.anim += rl.GetFrameTime()
	switch f.Scene() {
	case campaign.ScenePlay:
		u.drawScene(f)
		u.drawLines(HUDLines(f), w, h, false)
	case campaign.ScenePause:
		u.drawScene(f)
		u.drawLines(MenuLines(f), w, h, true)
	case campaign.SceneMainMenu:
		u.drawHeroFacing(PoseRun)
		u.drawLines(MenuLines(f), w, h, true)
	case campaign.SceneGameOver:
		u.drawHeroFacing(PoseIdle)
		u.drawLines(MenuLines(f), w, h, true)
	case campaign.SceneStory, campaign.SceneEnding:
		u.drawHeroFacing(PoseIdle)
		u.drawLines(StoryLines(f), w, h, true)
	case campaign.SceneCredits:
		u.drawLines(StoryLines(f), w, h, true)
	default:
		u.drawLines(MenuLines(f), w, h, true)
	}
}

// MenuLines returns the plain text of the current menu screen: the language
// pick, main menu, map, settings, pause, level end or game over. It picks
// every string by the flow's language setting and marks the flow's current
// menu selection. It returns nil for scenes without a menu.
func MenuLines(f *campaign.Flow) []ScreenLine {
	switch f.Scene() {
	case campaign.SceneLanguagePick:
		return languagePickLines(f)
	case campaign.SceneMainMenu:
		return mainMenuLines(f)
	case campaign.SceneMap:
		return mapLines(f)
	case campaign.SceneSettings:
		return settingsLines(f)
	case campaign.ScenePause:
		return pauseLines(f)
	case campaign.SceneLevelEnd:
		return levelEndLines(f)
	case campaign.SceneGameOver:
		return gameOverLines(f)
	default:
		return nil
	}
}

// HUDLines returns the plain text of the in-play HUD: lives, combs, shield,
// chase, and either campaign level progress or endless distance, plus the
// boss fight's lives and phase. It returns nil outside play.
func HUDLines(f *campaign.Flow) []ScreenLine {
	if f.Scene() != campaign.ScenePlay {
		return nil
	}
	if b := f.Boss(); b != nil {
		return bossHUDLines(f, b)
	}
	if r := f.Run(); r != nil {
		return runHUDLines(f, r)
	}
	return nil
}

// StoryLines returns the plain text of the story cards, the ending card and
// the credits roll. Story and ending paragraphs are wrapped so Draw can show
// them inside the window. It returns nil for the other scenes.
func StoryLines(f *campaign.Flow) []ScreenLine {
	switch f.Scene() {
	case campaign.SceneStory, campaign.SceneEnding:
		return storyWrap(textFor(f, campaign.StoryCards[f.Story()]), uiMarginX+uiIndent, uiMarginY, uiLineGap)
	case campaign.SceneCredits:
		return creditsLines(f)
	default:
		return nil
	}
}

func languagePickLines(f *campaign.Flow) []ScreenLine {
	return []ScreenLine{
		uiLine(textFor(f, campaign.LanguagePickTitle), uiMarginX, uiMarginY, false),
		uiLine(textFor(f, campaign.LanguageEnglishName), uiMarginX+uiIndent, uiMarginY+uiLineGap, f.MenuIndex() == 0),
		uiLine(textFor(f, campaign.LanguageUkrainianName), uiMarginX+uiIndent, uiMarginY+2*uiLineGap, f.MenuIndex() == 1),
	}
}

func mainMenuLines(f *campaign.Flow) []ScreenLine {
	lines := []ScreenLine{uiLine(textFor(f, campaign.GameTitle), uiMarginX, uiMarginY, false)}
	for i, item := range campaign.MainMenuItems {
		lines = append(lines, uiLine(textFor(f, item), uiMarginX+uiIndent, uiMarginY+(i+1)*uiLineGap, i == f.MenuIndex()))
	}
	return lines
}

func mapLines(f *campaign.Flow) []ScreenLine {
	lines := []ScreenLine{uiLine(textFor(f, campaign.MapTitle), uiMarginX, uiMarginY, false)}
	for i, name := range campaign.LevelNames {
		text := fmt.Sprintf("%2d. %s", i+1, textFor(f, name))
		if i+1 > f.Progress().Unlocked {
			text += " — " + textFor(f, campaign.MapLocked)
		}
		lines = append(lines, uiLine(text, uiMarginX+uiIndent, uiMarginY+(i+1)*uiLineGap, i == f.MenuIndex()))
	}
	return lines
}

func settingsLines(f *campaign.Flow) []ScreenLine {
	s := f.Settings()
	values := []string{
		volumeText(s.MusicVolume),
		volumeText(s.SoundVolume),
		onOffText(f, s.Fullscreen),
		languageValueName(s.Language),
	}
	lines := []ScreenLine{uiLine(textFor(f, campaign.SettingsTitle), uiMarginX, uiMarginY, false)}
	for i, label := range campaign.SettingsItems {
		lines = append(lines, uiLine(fmt.Sprintf("%s: %s", textFor(f, label), values[i]), uiMarginX+uiIndent, uiMarginY+(i+1)*uiLineGap, i == f.MenuIndex()))
	}
	return lines
}

func pauseLines(f *campaign.Flow) []ScreenLine {
	return []ScreenLine{
		uiLine(textFor(f, campaign.PauseTitle), uiMarginX, uiMarginY, false),
		uiLine(textFor(f, campaign.PauseHint), uiMarginX+uiIndent, uiMarginY+uiLineGap, false),
	}
}

func levelEndLines(f *campaign.Flow) []ScreenLine {
	title := campaign.LevelEndTitle
	if f.Endless() {
		title = campaign.EndlessLevelEndTitle
	}
	cause := gameplay.CauseNone
	if r := f.Run(); r != nil {
		cause = r.Cause
	}
	lines := []ScreenLine{uiLine(textFor(f, title), uiMarginX, uiMarginY, false)}
	lines = append(lines, uiLine(textFor(f, campaign.MedalsOfShame[cause]), uiMarginX+uiIndent, uiMarginY+uiLineGap, false))
	y := uiMarginY + 2*uiLineGap
	if f.Endless() {
		if r := f.Run(); r != nil {
			lines = append(lines, uiLine(fmt.Sprintf("%s: %.0f m", hudText(f, gameplay.HUDDistance), r.Runner.Z()), uiMarginX+uiIndent, y, false))
			y += uiLineGap
		}
	}
	lines = append(lines, uiLine(textFor(f, campaign.LevelEndHint), uiMarginX+uiIndent, y, false))
	return lines
}

func gameOverLines(f *campaign.Flow) []ScreenLine {
	cause := gameplay.CauseNone
	if f.Boss() != nil {
		cause = gameplay.CauseSnip
	} else if r := f.Run(); r != nil {
		cause = r.Cause
	}
	lines := []ScreenLine{
		uiLine(textFor(f, campaign.GameOverTitle), uiMarginX, uiMarginY, false),
		uiLine(textFor(f, campaign.Epitaphs[cause]), uiMarginX+uiIndent, uiMarginY+uiLineGap, false),
	}
	for i, item := range campaign.GameOverItems {
		lines = append(lines, uiLine(textFor(f, item), uiMarginX+uiIndent, uiMarginY+(i+2)*uiLineGap, i == f.MenuIndex()))
	}
	return lines
}

func runHUDLines(f *campaign.Flow, r *gameplay.Run) []ScreenLine {
	lines := []ScreenLine{
		uiLine(fmt.Sprintf("%s: %d", hudText(f, gameplay.HUDLives), r.Lives), uiMarginX, uiMarginY, false),
		uiLine(fmt.Sprintf("%s: %d", hudText(f, gameplay.HUDCombs), r.Combs), uiMarginX, uiMarginY+uiLineGap, false),
		uiLine(fmt.Sprintf("%s: %s", hudText(f, gameplay.HUDShield), onOffText(f, r.ShieldHits > 0)), uiMarginX, uiMarginY+2*uiLineGap, false),
		uiLine(fmt.Sprintf("%s: %d/10", hudText(f, gameplay.HUDChase), chaseTenths(r.Chase)), uiMarginX, uiMarginY+3*uiLineGap, false),
	}
	if f.Endless() {
		lines = append(lines, uiLine(fmt.Sprintf("%s: %.0f m", hudText(f, gameplay.HUDDistance), r.Runner.Z()), uiMarginX, uiMarginY+4*uiLineGap, false))
	} else if lv, ok := f.CurrentLevel(); ok {
		lines = append(lines, uiLine(fmt.Sprintf("%s: %d/10", textFor(f, uiLevel), lv.ID), uiMarginX, uiMarginY+4*uiLineGap, false))
	}
	return lines
}

func bossHUDLines(f *campaign.Flow, b *gameplay.BossFight) []ScreenLine {
	return []ScreenLine{
		uiLine(fmt.Sprintf("%s: %d", hudText(f, gameplay.HUDLives), b.Lives()), uiMarginX, uiMarginY, false),
		uiLine(fmt.Sprintf("%s: %d/3", textFor(f, uiPhase), b.Phase()), uiMarginX, uiMarginY+uiLineGap, false),
	}
}

func creditsLines(f *campaign.Flow) []ScreenLine {
	lines := make([]ScreenLine, 0, len(campaign.Credits)+1)
	for i, c := range campaign.Credits {
		lines = append(lines, uiLine(textFor(f, c), uiMarginX, uiMarginY+i*uiLineGap, false))
	}
	lines = append(lines, uiLine(textFor(f, campaign.CreditsClosing), uiMarginX, uiMarginY+len(campaign.Credits)*uiLineGap, false))
	return lines
}

func uiLine(text string, x, y int, highlighted bool) ScreenLine {
	return ScreenLine{Text: text, X: x, Y: y, Highlighted: highlighted}
}

func textFor(f *campaign.Flow, l campaign.Line) string {
	if f.Settings().Language == campaign.LanguageUkrainian {
		return l.UA
	}
	return l.EN
}

func hudText(f *campaign.Flow, l gameplay.Line) string {
	if f.Settings().Language == campaign.LanguageUkrainian {
		return l.UA
	}
	return l.EN
}

func onOffText(f *campaign.Flow, on bool) string {
	if on {
		return textFor(f, uiOn)
	}
	return textFor(f, uiOff)
}

func languageValueName(l campaign.Language) string {
	if l == campaign.LanguageUkrainian {
		return campaign.LanguageUkrainianName.UA
	}
	return campaign.LanguageEnglishName.EN
}

func volumeText(v float64) string {
	return fmt.Sprintf("%d", int(math.Round(v*10)))
}

func chaseTenths(chase float64) int {
	n := int(math.Round(chase * 10))
	if n < 0 {
		n = 0
	}
	if n > 10 {
		n = 10
	}
	return n
}

// storyWrap breaks a story or ending paragraph into ScreenLines that stay
// within storyWrapChars runes, splitting only at word boundaries so the text
// remains readable.
func storyWrap(text string, x, y, gap int) []ScreenLine {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []ScreenLine{uiLine(text, x, y, false)}
	}
	lines := make([]ScreenLine, 0, len(words)/4+1)
	cur := words[0]
	for _, w := range words[1:] {
		if utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(w) > storyWrapChars {
			lines = append(lines, uiLine(cur, x, y, false))
			y += gap
			cur = w
			continue
		}
		cur += " " + w
	}
	lines = append(lines, uiLine(cur, x, y, false))
	return lines
}

// drawHeroFacing draws the animated Downald D turned toward the camera. The
// facing model negates every command's Z, so the nose points at -z; the
// camera therefore sits on the -z side and looks back at his face. It is the
// hero shot behind the title, story and game-over screens.
func (u *UI) drawHeroFacing(pose HeroPose) {
	cam := rl.Camera3D{
		Position:   rl.Vector3{X: 0, Y: 1.4, Z: -3.4},
		Target:     rl.Vector3{X: 0, Y: 1.05, Z: 0},
		Up:         rl.Vector3{X: 0, Y: 1, Z: 0},
		Fovy:       60,
		Projection: cameraPerspective,
	}
	rl.BeginMode3D(cam)
	defer rl.EndMode3D()
	u.drawModelAt(BuildHeroAtFacing(pose, false, float64(u.anim)), Vec3{})
}

// drawScene draws the 3D scene behind play and pause: the generated corridor
// with the hero and Copycat, or the boss arena with the hero and Copycat's
// boss form.
func (u *UI) drawScene(f *campaign.Flow) {
	cam := u.cameraFor(f)
	rl.BeginMode3D(cam)
	defer rl.EndMode3D()

	if b := f.Boss(); b != nil {
		DrawCmds(BuildBossArena(Cam{Lane: float64(b.HeroLane()), Z: 0, Height: 0}))
		heroPos := Vec3{X: f32(laneX(float64(b.HeroLane()))), Y: 0, Z: 0}
		u.drawModelAt(BuildHero(PoseRun, false), heroPos)
		bossPos := Vec3{X: f32(laneX(float64(b.AttackLane()))), Y: 0, Z: 8}
		u.drawModelAt(BuildBoss(b.Phase(), b.Vulnerable()), bossPos)
		return
	}

	r := f.Run()
	if r == nil {
		return
	}
	theme := runTheme(f)
	cam3 := Cam{Lane: r.Runner.Lane(), Z: r.Runner.Z(), Height: r.Runner.Height()}
	DrawCmds(BuildWorld(r.Level, theme, cam3))
	pos := Vec3{X: f32(laneX(r.Runner.Lane())), Y: f32(r.Runner.Height()), Z: f32(r.Runner.Z())}
	u.drawModelAt(BuildHero(heroPoseFor(r), r.Bald), pos)
	chase := 2 + r.Chase*3
	u.drawModelAt(BuildCopycat(chase), Vec3{X: pos.X, Y: 0, Z: pos.Z})
}

func (u *UI) cameraFor(f *campaign.Flow) rl.Camera3D {
	cam := rl.Camera3D{
		Up:         rl.Vector3{X: 0, Y: 1, Z: 0},
		Fovy:       60,
		Projection: cameraPerspective,
	}
	if b := f.Boss(); b != nil {
		cam.Position = rl.Vector3{X: 0, Y: 5, Z: -8}
		cam.Target = rl.Vector3{X: 0, Y: 1, Z: 4}
		return cam
	}
	x := float32(0)
	z := float32(0)
	if r := f.Run(); r != nil {
		x = f32(laneX(r.Runner.Lane()))
		z = f32(r.Runner.Z())
	}
	cam.Position = rl.Vector3{X: x, Y: 4.5, Z: z - 9}
	cam.Target = rl.Vector3{X: x, Y: 1.2, Z: z + 6}
	return cam
}

func (u *UI) drawModelAt(cs []Cmd, at Vec3) {
	for _, c := range cs {
		c.Pos = Vec3{X: c.Pos.X + at.X, Y: c.Pos.Y + at.Y, Z: c.Pos.Z + at.Z}
		drawCmd(c)
	}
}

func runTheme(f *campaign.Flow) campaign.Theme {
	if f.Endless() {
		return campaign.ThemeNeon
	}
	if lv, ok := f.CurrentLevel(); ok {
		switch lv.World {
		case 2:
			return campaign.ThemeHarbor
		case 3:
			return campaign.ThemeNeon
		default:
			return campaign.ThemeMangroves
		}
	}
	return campaign.ThemeMangroves
}

func heroPoseFor(r *gameplay.Run) HeroPose {
	if r.Runner.Height() > 0.05 {
		return PoseJump
	}
	return PoseRun
}

// drawLines lays out and draws the screen lines. center centers the block in
// the window; the HUD passes false and keeps its top-left anchor.
func (u *UI) drawLines(lines []ScreenLine, w, h int, center bool) {
	if len(lines) == 0 {
		return
	}
	scale := u.scale
	if scale < 1 {
		scale = 1
	}
	ox, oy := 0, 0
	if center {
		ox = (w - u.blockWidth(lines)*scale) / 2
		oy = (h - u.blockHeight(lines)*scale) / 2
		if ox < 0 {
			ox = 0
		}
		if oy < 0 {
			oy = 0
		}
	}
	for _, l := range lines {
		u.drawLine(l, ox+l.X*scale, oy+l.Y*scale)
	}
}

func (u *UI) blockWidth(lines []ScreenLine) int {
	width := 0
	for _, l := range lines {
		w, _ := u.fnt.Measure(l.Text)
		if l.X+w > width {
			width = l.X + w
		}
	}
	return width
}

func (u *UI) blockHeight(lines []ScreenLine) int {
	height := 0
	for _, l := range lines {
		if l.Y+u.fnt.LineHeight() > height {
			height = l.Y + u.fnt.LineHeight()
		}
	}
	return height
}

func (u *UI) drawLine(l ScreenLine, x, y int) {
	col := rlColor(uiTextColor)
	if l.Highlighted {
		col = rlColor(uiHighlightText)
		w, _ := u.fnt.Measure(l.Text)
		rl.DrawRectangle(
			int32(x-2*u.scale),
			int32(y-u.scale),
			int32(w*u.scale+4*u.scale),
			int32(u.fnt.LineHeight()*u.scale+2*u.scale),
			rlColor(uiHighlightBar),
		)
	}
	for _, r := range l.Text {
		g, ok := u.fnt.Glyph(r)
		if !ok {
			x += u.fnt.SpaceWidth() * u.scale
			continue
		}
		u.drawGlyph(g, x, y, col)
		x += g.Advance * u.scale
	}
}

func (u *UI) drawGlyph(g font.Glyph, x, y int, col rl.Color) {
	for row := 0; row < g.Height; row++ {
		for colIdx := 0; colIdx < g.Width; colIdx++ {
			if !g.Pixels[row*g.Width+colIdx] {
				continue
			}
			rl.DrawRectangle(
				int32(x+colIdx*u.scale),
				int32(y+row*u.scale),
				int32(u.scale),
				int32(u.scale),
				col,
			)
		}
	}
}
