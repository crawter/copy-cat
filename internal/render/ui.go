package render

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"

	"copycat/internal/campaign"
	"copycat/internal/font"
	"copycat/internal/gameplay"
)

// ScreenLine is one line of UI text: what it says, where it sits in the
// unscaled screen layout, whether the menu cursor highlights it, and how much
// larger than a normal line it is drawn. Scale defaults to 1 when it is zero
// or negative.
type ScreenLine struct {
	Text        string
	X, Y        int
	Highlighted bool
	Scale       int
}

// UI draws the game's screens from a flow's plain state. It is created once
// with the pixel font and reused every frame.
type UI struct {
	fnt   *font.Font
	scale int

	// anim is the animation clock in seconds, advanced once per frame in
	// Draw and shared by every per-frame animation in the render layer.
	anim float64

	// scene and sceneAt track when the current scene started, so Draw can
	// overlay a short fade on every scene change.
	scene   campaign.Scene
	sceneAt time.Time

	// prevTotalCombs, prevBald, prevLivesLost and prevAirborne are the last
	// run state drawScene observed, so it can spawn captions and particle
	// bursts exactly when a counter changes or the hero lands.
	prevTotalCombs int
	prevBald       bool
	prevLivesLost  int
	prevAirborne   bool
	airTracked     bool

	// caption and captionAt are the short play caption currently on screen.
	caption   []ScreenLine
	captionAt time.Time
	// introLines and introAt are the level intro card shown when a run starts.
	introLines []ScreenLine
	introAt    time.Time
	// bursts are the active particle bursts.
	bursts []burst
}

// burst is one short-lived particle burst spawned at a hero position.
type burst struct {
	cmds   []Cmd
	origin Vec3
	at     time.Time
	life   time.Duration
}

// VolumeBar is the unscaled geometry of one settings volume bar: where it
// sits, how large it is, and the fraction of it that is filled.
type VolumeBar struct {
	X, Y          int
	Width, Height int
	Fill          float64
}

// NewUI returns a UI that draws its text with fnt. The font is read-only and
// safe to share.
func NewUI(fnt *font.Font) *UI {
	return &UI{fnt: fnt, scale: uiScale, sceneAt: time.Now()}
}

// The screen layout is described in the font's unscaled pixel grid; Draw
// scales it to the window.
const (
	uiScale   = 2
	uiMarginX = 8
	uiMarginY = 8
	uiIndent  = 4
	uiLineGap = 10 // matches font.Font's default line height

	// uiLogoScale and uiTaglineScale size the main-menu logo and tagline
	// relative to a normal UI line.
	uiLogoScale    = 4
	uiTaglineScale = 2

	// logoBobSpeed and logoBobAmplitude drive the main-menu logo's subtle
	// vertical bob.
	logoBobSpeed     = 2.0
	logoBobAmplitude = 2.0

	// storyWrapChars is the maximum rune count a story or ending line holds
	// before it wraps to the next line.
	storyWrapChars = 60

	// sceneFadeDuration is how many seconds a scene-change fade takes to
	// clear.
	sceneFadeDuration = 0.35

	// captionDuration is how many seconds a play caption stays on screen.
	captionDuration = 1.6
	// levelIntroDuration is how many seconds the level intro card stays on
	// screen after a run starts.
	levelIntroDuration = 1.6

	// burstLife is how long a particle burst lives before it is dropped.
	burstLife = 400 * time.Millisecond
	// burstRise is how many meters a particle burst rises over its life.
	burstRise = 0.8

	// volumeBarWidth and volumeBarHeight size the settings volume bars in
	// the unscaled screen grid.
	volumeBarWidth  = 40
	volumeBarHeight = 6

	// Copycat trail distance tuning: a base gap grows with the run's speed
	// and shrinks as Chase closes in, never below copycatMinGap.
	copycatBaseGap  = 1.2
	copycatSpeedGap = 0.25
	copycatChaseGap = 2.5
	copycatMinGap   = 1.6
)

// UI colors, drawn behind and over the screen text.
var (
	uiBackground    = Color{R: 24, G: 20, B: 28, A: 255}
	uiTextColor     = Color{R: 236, G: 232, B: 240, A: 255}
	uiHighlightText = Color{R: 40, G: 30, B: 20, A: 255}
	uiHighlightBar  = Color{R: 240, G: 202, B: 112, A: 255}
	uiBarBack       = Color{R: 60, G: 54, B: 70, A: 255}
)

// Render-layer labels for the level-end stats and the play captions. They
// live here rather than in the campaign texts because only the render layer
// uses them.
var (
	uiTotalCombs = campaign.Line{EN: "Total Combs", UA: "Усього гребінців"}
	uiLivesLost  = campaign.Line{EN: "Lives Lost", UA: "Втрачено життів"}
	uiStomps     = campaign.Line{EN: "Stomps", UA: "Стомпів"}

	uiCombCaption = campaign.Line{EN: "Comb!", UA: "Гребінець!"}
	uiSnipCaption = campaign.Line{EN: "Snip! Hair lost.", UA: "Чик! Волосся втрачено."}
	uiLandCaption = campaign.Line{EN: "Landed!", UA: "Приземлення!"}

	uiEndless = campaign.Line{EN: "Endless Run", UA: "Нескінченний забіг"}
)

// cameraPerspective is raylib's CameraPerspective (0). It stays an untyped
// constant so the assignment compiles against either int32 or the binding's
// named CameraProjection field type.
const cameraPerspective = 0

// Draw renders the scene for the flow's current state in a window of w by h
// pixels. It clears the window, draws the 3D corridor or boss arena while
// playing or paused, then draws the matching screen text and overlays a fade
// when the scene just changed.
func (u *UI) Draw(f *campaign.Flow, w, h int) {
	rl.ClearBackground(rlColor(uiBackground))
	u.anim += float64(rl.GetFrameTime())
	u.trackScene(f)
	switch f.Scene() {
	case campaign.ScenePlay:
		u.drawScene(f)
		u.drawOverlay(u.caption, w, h, u.captionAt, captionDuration)
		u.drawOverlay(u.introLines, w, h, u.introAt, levelIntroDuration)
		u.drawLines(HUDLines(f), w, h, false)
	case campaign.ScenePause:
		u.drawScene(f)
		u.drawLines(MenuLines(f), w, h, true)
	case campaign.SceneStory, campaign.SceneEnding, campaign.SceneCredits:
		u.drawLines(StoryLines(f), w, h, true)
	case campaign.SceneSettings:
		lines := MenuLines(f)
		u.drawLines(lines, w, h, true)
		u.drawVolumeBars(f, lines, w, h)
	case campaign.SceneMainMenu:
		lines := MenuLines(f)
		if len(lines) > 0 {
			lines[0].Y += logoOffset(u.anim)
		}
		u.drawLines(lines, w, h, true)
	default:
		u.drawLines(MenuLines(f), w, h, true)
	}
	u.drawFade(w, h)
}

// trackScene records when the flow changes scene so the fade can restart. A
// transition into play that is not a pause resume also resets the one-shot
// play overlays, because that is when a new run or boss fight begins.
func (u *UI) trackScene(f *campaign.Flow) {
	if f.Scene() == u.scene {
		return
	}
	prev := u.scene
	u.scene = f.Scene()
	u.sceneAt = time.Now()
	if f.Scene() == campaign.ScenePlay && prev != campaign.ScenePause {
		u.beginPlay(f)
	}
}

// beginPlay resets the captions, particles and level intro when a new run or
// boss fight starts. Resuming from pause does not call it.
func (u *UI) beginPlay(f *campaign.Flow) {
	u.introLines = LevelIntroLines(f)
	u.introAt = time.Now()
	u.caption = nil
	u.captionAt = time.Time{}
	u.bursts = nil
	u.airTracked = false
	if r := f.Run(); r != nil {
		u.prevTotalCombs = r.TotalCombs
		u.prevBald = r.Bald
		u.prevLivesLost = r.LivesLost
		u.prevAirborne = r.Runner.Airborne()
		u.airTracked = true
	}
}

// drawOverlay draws lines centered on the window behind a translucent
// backdrop for duration seconds after at. It draws nothing once the overlay
// has expired or when there are no lines.
func (u *UI) drawOverlay(lines []ScreenLine, w, h int, at time.Time, duration float64) {
	if len(lines) == 0 || duration <= 0 {
		return
	}
	if time.Since(at).Seconds() >= duration {
		return
	}
	scale := u.scale
	if scale < 1 {
		scale = 1
	}
	bw := u.blockWidth(lines) * scale
	bh := u.blockHeight(lines) * scale
	ox := (w - bw) / 2
	oy := (h - bh) / 2
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}
	pad := 2 * scale
	rl.DrawRectangle(
		int32(ox-pad),
		int32(oy-pad),
		int32(bw+2*pad),
		int32(bh+2*pad),
		rlColor(Color{R: 0, G: 0, B: 0, A: 150}),
	)
	u.drawLines(lines, w, h, true)
}

// drawFade overlays the scene-change fade on the whole window. It draws
// nothing once the fade has cleared.
func (u *UI) drawFade(w, h int) {
	alpha := fadeAlpha(time.Since(u.sceneAt).Seconds(), sceneFadeDuration)
	if alpha <= 0 {
		return
	}
	rl.DrawRectangle(
		0,
		0,
		int32(w),
		int32(h),
		rlColor(Color{R: uiBackground.R, G: uiBackground.G, B: uiBackground.B, A: uint8(alpha * 255)}),
	)
}

// fadeAlpha maps the elapsed time since a scene change to the overlay alpha
// in [0, 1]. It starts opaque and fades linearly to transparent after
// duration seconds. A non-positive duration reports a cleared fade.
func fadeAlpha(elapsed, duration float64) float64 {
	if duration <= 0 {
		return 0
	}
	if elapsed <= 0 {
		return 1
	}
	if elapsed >= duration {
		return 0
	}
	return 1 - elapsed/duration
}

// logoOffset returns the subtle vertical offset, in unscaled pixels, of the
// main-menu logo at animation clock seconds: a gentle sine bob. It is pure so
// tests can pin the motion.
func logoOffset(clock float64) int {
	return int(math.Round(math.Sin(clock*logoBobSpeed) * logoBobAmplitude))
}

// MenuLines returns the plain text of the current menu screen: the language
// pick, main menu, New Game confirmation, map, settings, pause, level end or
// game over. It picks every string by the flow's language setting and marks
// the flow's current menu selection. It returns nil for scenes without a menu.
func MenuLines(f *campaign.Flow) []ScreenLine {
	switch f.Scene() {
	case campaign.SceneLanguagePick:
		return languagePickLines(f)
	case campaign.SceneMainMenu:
		return mainMenuLines(f)
	case campaign.SceneNewGameConfirm:
		return newGameConfirmLines(f)
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

// LevelIntroLines returns the short level intro card shown when a run starts:
// the current level number and name, or the endless banner.
func LevelIntroLines(f *campaign.Flow) []ScreenLine {
	if f.Endless() {
		return []ScreenLine{uiLine(textFor(f, uiEndless), uiMarginX, uiMarginY, false)}
	}
	lv, ok := f.CurrentLevel()
	if !ok {
		return nil
	}
	name := textFor(f, campaign.LevelNames[lv.ID-1])
	return []ScreenLine{
		uiLineScaled(fmt.Sprintf("%s %d", textFor(f, uiLevel), lv.ID), uiMarginX, uiMarginY, false, uiTaglineScale),
		uiLineScaled(name, uiMarginX, uiMarginY+uiLineGap*uiTaglineScale, false, uiLogoScale),
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
	logoY := uiMarginY
	taglineY := logoY + uiLineGap*uiLogoScale + 2*logoBobAmplitude
	menuY := taglineY + uiLineGap*uiTaglineScale + uiLineGap
	lines := []ScreenLine{
		uiLineScaled(textFor(f, campaign.GameTitle), uiMarginX, logoY, false, uiLogoScale),
		uiLineScaled(textFor(f, campaign.TitleTagline), uiMarginX+uiIndent, taglineY, false, uiTaglineScale),
	}
	for i, item := range campaign.MainMenuItems {
		lines = append(lines, uiLine(textFor(f, item), uiMarginX+uiIndent, menuY+i*uiLineGap, i == f.MenuIndex()))
	}
	return lines
}

func newGameConfirmLines(f *campaign.Flow) []ScreenLine {
	lines := []ScreenLine{uiLine(textFor(f, campaign.NewGameConfirmTitle), uiMarginX, uiMarginY, false)}
	for i, item := range campaign.NewGameConfirmItems {
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
	y := uiMarginY
	for i, label := range campaign.SettingsItems {
		y += uiLineGap
		lines = append(lines, uiLine(fmt.Sprintf("%s: %s", textFor(f, label), values[i]), uiMarginX+uiIndent, y, i == f.MenuIndex()))
		if i < 2 {
			y += uiLineGap // the volume rows reserve a line for their bar
		}
	}
	return lines
}

func pauseLines(f *campaign.Flow) []ScreenLine {
	lines := []ScreenLine{
		uiLine(textFor(f, campaign.PauseTitle), uiMarginX, uiMarginY, false),
		uiLine(textFor(f, campaign.PauseHint), uiMarginX+uiIndent, uiMarginY+uiLineGap, false),
	}
	for i, item := range campaign.PauseItems {
		lines = append(lines, uiLine(textFor(f, item), uiMarginX+uiIndent, uiMarginY+(i+2)*uiLineGap, i == f.MenuIndex()))
	}
	return lines
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
	if r := f.Run(); r != nil {
		lines = append(lines,
			uiLine(fmt.Sprintf("%s: %d", textFor(f, uiTotalCombs), r.TotalCombs), uiMarginX+uiIndent, y, false),
			uiLine(fmt.Sprintf("%s: %d", textFor(f, uiLivesLost), r.LivesLost), uiMarginX+uiIndent, y+uiLineGap, false),
			uiLine(fmt.Sprintf("%s: %d", textFor(f, uiStomps), r.Stomps), uiMarginX+uiIndent, y+2*uiLineGap, false),
			uiLine(fmt.Sprintf("%s: %.0f m", hudText(f, gameplay.HUDDistance), r.Runner.Z()), uiMarginX+uiIndent, y+3*uiLineGap, false),
		)
		y += 4 * uiLineGap
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

// CombCaption returns the play caption for a comb pickup, showing the run's
// total comb count.
func CombCaption(f *campaign.Flow, totalCombs int) []ScreenLine {
	return []ScreenLine{uiLine(fmt.Sprintf("%s: %d", textFor(f, uiCombCaption), totalCombs), 0, 0, false)}
}

// SnipCaption returns the play caption for Copycat's snip. When bald is
// false there is nothing to caption.
func SnipCaption(f *campaign.Flow, bald bool, livesLost int) []ScreenLine {
	if !bald {
		return nil
	}
	return []ScreenLine{uiLine(fmt.Sprintf("%s %s: %d", textFor(f, uiSnipCaption), textFor(f, uiLivesLost), livesLost), 0, 0, false)}
}

// LandingCaption returns the play caption for a hero landing.
func LandingCaption(f *campaign.Flow) []ScreenLine {
	return []ScreenLine{uiLine(textFor(f, uiLandCaption), 0, 0, false)}
}

func uiLine(text string, x, y int, highlighted bool) ScreenLine {
	return uiLineScaled(text, x, y, highlighted, 1)
}

func uiLineScaled(text string, x, y int, highlighted bool, scale int) ScreenLine {
	return ScreenLine{Text: text, X: x, Y: y, Highlighted: highlighted, Scale: scale}
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
	return fmt.Sprintf("%d%%", int(math.Round(v*100)))
}

// volumeBar returns the unscaled geometry of a settings volume bar on the row
// whose text sits at y, filled to fraction of its width. It is pure so tests
// can pin the layout.
func volumeBar(y int, fraction float64) VolumeBar {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return VolumeBar{
		X:      uiMarginX + uiIndent,
		Y:      y + uiLineGap,
		Width:  volumeBarWidth,
		Height: volumeBarHeight,
		Fill:   fraction,
	}
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

// drawScene draws the 3D scene behind play and pause: the generated corridor
// with the hero and Copycat, or the boss arena with the hero and Copycat's
// boss form. The hero is animated by the UI's per-frame animation clock.
// During a run it also detects comb, snip and landing events and draws the
// matching particle bursts.
func (u *UI) drawScene(f *campaign.Flow) {
	cam := u.cameraFor(f)
	rl.BeginMode3D(cam)
	defer rl.EndMode3D()

	t := u.anim

	if b := f.Boss(); b != nil {
		DrawCmds(BuildBossArena(Cam{Lane: float64(b.HeroLane()), Z: 0, Height: 0}))
		if !b.Vulnerable() {
			DrawCmds(BuildArenaTelegraph(b.AttackLane(), b.AttackKind()))
		}
		heroPos := Vec3{X: f32(laneX(float64(b.HeroLane()))), Y: 0, Z: 0}
		u.drawModelAt(BuildHeroAt(BossPose(b.Jumping(), b.Sliding()), false, t), heroPos)
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
	DrawCmds(ApplyFog(BuildWorld(r.Level, theme, cam3), cam3, Palette(theme)))
	pos := Vec3{X: f32(laneX(r.Runner.Lane())), Y: f32(r.Runner.Height()), Z: f32(r.Runner.Z())}
	u.drawModelAt(BuildHeroAt(heroPoseFor(r), r.Bald, t), pos)
	chase := CopycatDistance(r.Runner.Speed(), r.Chase)
	u.drawModelAt(BuildCopycat(chase), Vec3{X: pos.X, Y: 0, Z: pos.Z})

	u.updateRunOverlays(f, r, pos)
	u.drawBursts()
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

// RunnerPose maps the runner's movement state to the hero pose the model
// draws. It is pure so tests can cover every move without constructing a run.
func RunnerPose(slamming, sliding, airborne bool) HeroPose {
	switch {
	case slamming:
		return PoseSlam
	case sliding:
		return PoseSlide
	case airborne:
		return PoseJump
	default:
		return PoseRun
	}
}

// BossPose maps the boss fight's hero moves to the hero pose the model draws:
// a jump, a grounded slide, or the run stance.
func BossPose(jumping, sliding bool) HeroPose {
	if jumping {
		return PoseJump
	}
	if sliding {
		return PoseSlide
	}
	return PoseRun
}

func heroPoseFor(r *gameplay.Run) HeroPose {
	return RunnerPose(r.Runner.Slamming(), r.Runner.Sliding(), r.Runner.Airborne())
}

// CopycatDistance returns how far behind the hero Copycat trails: a base gap
// grows with the run's speed and shrinks as Chase closes in, but never drops
// below the minimum that keeps Copycat clear of the hero.
func CopycatDistance(speed, chase float64) float64 {
	gap := copycatBaseGap + copycatSpeedGap*speed - copycatChaseGap*chase
	if gap < copycatMinGap {
		gap = copycatMinGap
	}
	return gap
}

// updateRunOverlays tracks the run counters and the hero's airborne state,
// spawning the matching caption and particle burst on comb pickups, snips and
// landings.
func (u *UI) updateRunOverlays(f *campaign.Flow, r *gameplay.Run, pos Vec3) {
	air := r.Runner.Airborne()
	if !u.airTracked {
		u.prevAirborne = air
		u.airTracked = true
	}

	if r.TotalCombs > u.prevTotalCombs {
		u.showCaption(CombCaption(f, r.TotalCombs))
		u.spawnBurst(BuildCombBurst(), pos)
	}
	if r.Bald && !u.prevBald {
		u.showCaption(SnipCaption(f, r.Bald, r.LivesLost))
		u.spawnBurst(BuildSnipBurst(), pos)
	}
	if u.prevAirborne && !air {
		u.showCaption(LandingCaption(f))
		u.spawnBurst(BuildLandingBurst(), pos)
	}

	u.prevTotalCombs = r.TotalCombs
	u.prevBald = r.Bald
	u.prevLivesLost = r.LivesLost
	u.prevAirborne = air
}

func (u *UI) showCaption(lines []ScreenLine) {
	u.caption = lines
	u.captionAt = time.Now()
}

func (u *UI) spawnBurst(cs []Cmd, at Vec3) {
	u.bursts = append(u.bursts, burst{cmds: cs, origin: at, at: time.Now(), life: burstLife})
}

// drawBursts draws and ages the active particle bursts, dropping expired
// ones. The commands rise and fade as they age.
func (u *UI) drawBursts() {
	now := time.Now()
	kept := u.bursts[:0]
	for _, b := range u.bursts {
		age := now.Sub(b.at)
		if age >= b.life {
			continue
		}
		p := age.Seconds() / b.life.Seconds()
		rise := float32(p * burstRise)
		alpha := uint8(255 * (1 - p))
		for _, c := range b.cmds {
			cc := c
			cc.Pos = Vec3{
				X: c.Pos.X + b.origin.X,
				Y: c.Pos.Y + b.origin.Y + rise,
				Z: c.Pos.Z + b.origin.Z,
			}
			cc.Color.A = uint8(float64(c.Color.A) * float64(alpha) / 255)
			drawCmd(cc)
		}
		kept = append(kept, b)
	}
	u.bursts = kept
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
	ox, oy := u.centeredOrigin(lines, w, h, center)
	for _, l := range lines {
		u.drawLine(l, ox+l.X*scale, oy+l.Y*scale)
	}
}

// centeredOrigin returns the window offset that centers the lines block, or
// the zero offset when center is false or there are no lines.
func (u *UI) centeredOrigin(lines []ScreenLine, w, h int, center bool) (int, int) {
	if !center || len(lines) == 0 {
		return 0, 0
	}
	scale := u.scale
	if scale < 1 {
		scale = 1
	}
	ox := (w - u.blockWidth(lines)*scale) / 2
	oy := (h - u.blockHeight(lines)*scale) / 2
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}
	return ox, oy
}

// drawVolumeBars draws the filled settings volume bars aligned with the
// centered settings block.
func (u *UI) drawVolumeBars(f *campaign.Flow, lines []ScreenLine, w, h int) {
	if len(lines) < 3 {
		return
	}
	s := f.Settings()
	scale := u.scale
	if scale < 1 {
		scale = 1
	}
	ox, oy := u.centeredOrigin(lines, w, h, true)
	bars := []VolumeBar{
		volumeBar(lines[1].Y, s.MusicVolume),
		volumeBar(lines[2].Y, s.SoundVolume),
	}
	for _, b := range bars {
		x := ox + b.X*scale
		y := oy + b.Y*scale
		bw := b.Width * scale
		bh := b.Height * scale
		rl.DrawRectangle(int32(x), int32(y), int32(bw), int32(bh), rlColor(uiBarBack))
		if b.Fill > 0 {
			rl.DrawRectangle(int32(x), int32(y), int32(float64(bw)*b.Fill), int32(bh), rlColor(uiHighlightBar))
		}
	}
}

func (u *UI) blockWidth(lines []ScreenLine) int {
	width := 0
	for _, l := range lines {
		w, _ := u.fnt.Measure(l.Text)
		ls := l.Scale
		if ls < 1 {
			ls = 1
		}
		if l.X+w*ls > width {
			width = l.X + w*ls
		}
	}
	return width
}

func (u *UI) blockHeight(lines []ScreenLine) int {
	height := 0
	for _, l := range lines {
		ls := l.Scale
		if ls < 1 {
			ls = 1
		}
		if l.Y+u.fnt.LineHeight()*ls > height {
			height = l.Y + u.fnt.LineHeight()*ls
		}
	}
	return height
}

func (u *UI) drawLine(l ScreenLine, x, y int) {
	lineScale := l.Scale
	if lineScale < 1 {
		lineScale = 1
	}
	baseScale := u.scale
	if baseScale < 1 {
		baseScale = 1
	}
	pixelScale := baseScale * lineScale
	col := rlColor(uiTextColor)
	if l.Highlighted {
		col = rlColor(uiHighlightText)
		w, _ := u.fnt.Measure(l.Text)
		rl.DrawRectangle(
			int32(x-2*pixelScale),
			int32(y-pixelScale),
			int32(w*pixelScale+4*pixelScale),
			int32(u.fnt.LineHeight()*pixelScale+2*pixelScale),
			rlColor(uiHighlightBar),
		)
	}
	for _, r := range l.Text {
		g, ok := u.fnt.Glyph(r)
		if !ok {
			x += u.fnt.SpaceWidth() * pixelScale
			continue
		}
		u.drawGlyph(g, x, y, col, pixelScale)
		x += g.Advance * pixelScale
	}
}

func (u *UI) drawGlyph(g font.Glyph, x, y int, col rl.Color, pixelScale int) {
	for row := 0; row < g.Height; row++ {
		for colIdx := 0; colIdx < g.Width; colIdx++ {
			if !g.Pixels[row*g.Width+colIdx] {
				continue
			}
			rl.DrawRectangle(
				int32(x+colIdx*pixelScale),
				int32(y+row*pixelScale),
				int32(pixelScale),
				int32(pixelScale),
				col,
			)
		}
	}
}
