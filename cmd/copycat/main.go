// Command copycat runs the finished game: it owns the raylib window and audio
// device, polls input, drives the campaign flow, plays the synthesized audio
// and draws every scene. All game rules live in the logic packages; this file
// only wires them together and is one of the two raylib layers.
package main

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"

	"copycat/internal/audio"
	"copycat/internal/campaign"
	"copycat/internal/font"
	"copycat/internal/gameplay"
	"copycat/internal/input"
	"copycat/internal/render"
	"copycat/internal/runner"
	"copycat/internal/save"
)

// Window, audio and loop tuning. Each number is named with its unit nearby.
const (
	windowWidth  = 1280 // window width, in pixels
	windowHeight = 720  // window height, in pixels
	windowTitle  = "Hair Raid"
	targetFPS    = 60 // fixed update and draw rate

	sampleRate         = 44100 // mixer and stream sample rate, in hertz
	streamChannels     = 2     // interleaved stereo
	streamSampleSize   = 32    // one float32 sample, in bits
	streamBufferFrames = 2048  // samples per channel held in the refill buffer
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "copycat:", err)
		os.Exit(1)
	}
}

func run() error {
	path, err := save.DefaultPath()
	if err != nil {
		return err
	}
	_, statErr := os.Stat(path)
	firstRun := os.IsNotExist(statErr)

	file, err := save.Load(path)
	if err != nil {
		return err
	}

	camp := campaign.Default()
	flow := campaign.NewFlow(camp, file.Settings, file.Progress, firstRun)

	mixer := audio.New(sampleRate)
	mixer.SetMusicVolume(float32(file.Settings.MusicVolume))
	mixer.SetSoundVolume(float32(file.Settings.SoundVolume))

	ui := render.NewUI(font.New())
	mapper := input.NewMapper()

	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	rl.SetExitKey(0)
	defer rl.CloseWindow()
	rl.SetTargetFPS(targetFPS)
	if file.Settings.Fullscreen && !rl.IsWindowFullscreen() {
		rl.ToggleFullscreen()
	}
	prevFullscreen := rl.IsWindowFullscreen()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
	rl.SetAudioStreamBufferSizeDefault(streamBufferFrames * streamChannels)
	stream := rl.LoadAudioStream(sampleRate, streamSampleSize, streamChannels)
	defer func() {
		rl.StopAudioStream(stream)
		rl.UnloadAudioStream(stream)
	}()
	rl.PlayAudioStream(stream)

	var (
		prevState    input.State
		currentTrack = audio.Track(-1) // forces the first track selection below
		lastProgress = file.Progress
		lastSettings = file.Settings
		prevSettings = file.Settings
	)
	buf := make([]float32, streamBufferFrames*streamChannels)

	playTrack(mixer, &currentTrack, trackForScene(flow))

	for !rl.WindowShouldClose() {
		st := input.Poll(mapper)
		dt := float64(rl.GetFrameTime())

		applyActions(flow, mixer, st, prevState)

		if flow.Quit() || st.Quit {
			break
		}

		stepAndPlaySounds(flow, mixer, st, prevState, dt)

		prevState = st

		applySettings(mixer, flow, &prevSettings, &prevFullscreen)
		playTrack(mixer, &currentTrack, trackForScene(flow))
		saveIfChanged(path, flow, &lastProgress, &lastSettings)
		feedStream(stream, mixer, buf)

		rl.BeginDrawing()
		ui.Draw(flow, rl.GetScreenWidth(), rl.GetScreenHeight())
		rl.EndDrawing()
	}

	if err := writeSave(path, flow); err != nil {
		fmt.Fprintln(os.Stderr, "copycat: save:", err)
	}
	return nil
}

// applyActions turns the frame's pressed edges into flow actions. Left and
// right are held in the input state, so their menu edges are derived here.
// In a selectable menu the directional edges play the move blip and confirm
// plays the select chime.
func applyActions(f *campaign.Flow, mixer *audio.Mixer, st, prev input.State) {
	menu := isMenuScene(f)
	if st.Up {
		f.Handle(campaign.ActionUp)
		if menu {
			mixer.Play(audio.SoundMenuMove)
		}
	}
	if st.Down {
		f.Handle(campaign.ActionDown)
		if menu {
			mixer.Play(audio.SoundMenuMove)
		}
	}
	if st.Left && !prev.Left {
		f.Handle(campaign.ActionLeft)
		if menu {
			mixer.Play(audio.SoundMenuMove)
		}
	}
	if st.Right && !prev.Right {
		f.Handle(campaign.ActionRight)
		if menu {
			mixer.Play(audio.SoundMenuMove)
		}
	}
	if st.Confirm {
		f.Handle(campaign.ActionConfirm)
		if menu {
			mixer.Play(audio.SoundMenuSelect)
		}
	}
	if st.Back {
		f.Handle(campaign.ActionBack)
	}
	if st.Pause {
		f.Handle(campaign.ActionPause)
	}
}

// isMenuScene reports whether the flow shows a scene whose directional edges
// move a selection and whose confirm selects it.
func isMenuScene(f *campaign.Flow) bool {
	switch f.Scene() {
	case campaign.SceneLanguagePick, campaign.SceneMainMenu, campaign.SceneMap,
		campaign.SceneSettings, campaign.SceneGameOver:
		return true
	default:
		return false
	}
}

// stepAndPlaySounds advances a run or boss fight and plays the sounds the
// frame produced: the jump, slide and slam input sounds, then one sound per
// run outcome and one per boss transition.
func stepAndPlaySounds(f *campaign.Flow, mixer *audio.Mixer, st, prev input.State, dt float64) {
	if f.Scene() != campaign.ScenePlay {
		return
	}
	if st.Jump && !prev.Jump {
		mixer.Play(audio.SoundJump)
	}
	if st.Slide && !prev.Slide {
		mixer.Play(slideSound(airborne(runRunner(f), f.Boss())))
	}
	beforeLives, beforePhase := bossState(f)
	outcomes := f.StepRun(runner.Input{Left: st.Left, Right: st.Right, Jump: st.Jump, Slide: st.Slide}, dt)
	for _, o := range outcomes {
		for _, s := range outcomeSounds(o) {
			mixer.Play(s)
		}
	}
	afterLives, afterPhase := bossState(f)
	for _, s := range bossTransitionSounds(beforeLives, beforePhase, afterLives, afterPhase) {
		mixer.Play(s)
	}
}

// bossState returns the active boss's lives and phase, or -1 for each outside
// a boss fight.
func bossState(f *campaign.Flow) (lives, phase int) {
	if b := f.Boss(); b != nil {
		return b.Lives(), b.Phase()
	}
	return -1, -1
}

// runRunner returns the active run's runner, or nil outside a run.
func runRunner(f *campaign.Flow) *runner.Runner {
	if r := f.Run(); r != nil {
		return r.Runner
	}
	return nil
}

// airborne reports whether the active hero is in the air. A boss uses its own
// jump flag; a run is airborne only above flat ground, never while riding a
// container ramp.
func airborne(r *runner.Runner, b *gameplay.BossFight) bool {
	if b != nil {
		return b.Jumping()
	}
	if r != nil {
		return r.Height() > 0 && !r.OnContainer()
	}
	return false
}

// slideSound returns the one-shot a slide plays: a slam while airborne, a
// slide while grounded or riding a container ramp.
func slideSound(airborne bool) audio.Sound {
	if airborne {
		return audio.SoundSlam
	}
	return audio.SoundSlide
}

// outcomeSounds maps one run outcome to the one-shot sounds it plays.
func outcomeSounds(o gameplay.Outcome) []audio.Sound {
	switch o {
	case gameplay.OutcomeStumble:
		return []audio.Sound{audio.SoundHit}
	case gameplay.OutcomeSnip:
		return []audio.Sound{audio.SoundSnip, audio.SoundMeow}
	case gameplay.OutcomePit:
		return []audio.Sound{audio.SoundFall, audio.SoundHonk}
	case gameplay.OutcomeStomp:
		return []audio.Sound{audio.SoundStomp}
	case gameplay.OutcomeFinished:
		return []audio.Sound{audio.SoundFanfare}
	case gameplay.OutcomeDied:
		return []audio.Sound{audio.SoundGameOver}
	case gameplay.OutcomeComb:
		return []audio.Sound{audio.SoundComb}
	case gameplay.OutcomeShieldSaved:
		return []audio.Sound{audio.SoundHairspray}
	case gameplay.OutcomeHairspray:
		return []audio.Sound{audio.SoundHairspray}
	case gameplay.OutcomeCheckpoint:
		return []audio.Sound{audio.SoundCheckpoint}
	case gameplay.OutcomeLife:
		return []audio.Sound{audio.SoundExtraLife}
	default:
		return nil
	}
}

// bossTransitionSounds returns the one-shots for the transitions between the
// two boss snapshots: a boss hit when lives dropped, a meow when the phase
// rose.
func bossTransitionSounds(beforeLives, beforePhase, afterLives, afterPhase int) []audio.Sound {
	var out []audio.Sound
	if afterLives < beforeLives {
		out = append(out, audio.SoundBossHit)
	}
	if afterPhase > beforePhase {
		out = append(out, audio.SoundMeow)
	}
	return out
}

// trackForScene returns the looping music track for the flow's current scene.
func trackForScene(f *campaign.Flow) audio.Track {
	switch f.Scene() {
	case campaign.ScenePlay, campaign.ScenePause:
		if f.Boss() != nil {
			return audio.TrackBoss
		}
		if f.Endless() {
			return audio.TrackNeon
		}
		if lv, ok := f.CurrentLevel(); ok {
			switch lv.World {
			case 2:
				return audio.TrackHarbor
			case 3:
				return audio.TrackNeon
			default:
				return audio.TrackMangroves
			}
		}
		return audio.TrackMangroves
	case campaign.SceneEnding, campaign.SceneCredits:
		return audio.TrackEnding
	default:
		return audio.TrackMenu
	}
}

// playTrack starts want once when it differs from the track already playing.
func playTrack(mixer *audio.Mixer, current *audio.Track, want audio.Track) {
	if want == *current {
		return
	}
	*current = want
	mixer.PlayMusic(want)
}

// applySettings pushes the flow's settings into the mixer and the window, and
// plays a preview one-shot once when a volume changes.
func applySettings(mixer *audio.Mixer, f *campaign.Flow, prevSettings *campaign.Settings, prevFullscreen *bool) {
	s := f.Settings()
	mixer.SetMusicVolume(float32(s.MusicVolume))
	mixer.SetSoundVolume(float32(s.SoundVolume))
	if s.Fullscreen != *prevFullscreen {
		rl.ToggleFullscreen()
		*prevFullscreen = s.Fullscreen
	}
	if sound, ok := volumePreviewSound(*prevSettings, s); ok {
		mixer.Play(sound)
	}
	*prevSettings = s
}

// volumePreviewSound returns the one-shot to play when the music or sound
// volume changes between prev and cur, and whether a preview is needed.
func volumePreviewSound(prev, cur campaign.Settings) (audio.Sound, bool) {
	if prev.MusicVolume == cur.MusicVolume && prev.SoundVolume == cur.SoundVolume {
		return 0, false
	}
	return audio.SoundMenuSelect, true
}

// saveIfChanged writes the save whenever progress or settings changed.
func saveIfChanged(path string, f *campaign.Flow, lastProgress *campaign.Progress, lastSettings *campaign.Settings) {
	if f.Progress() == *lastProgress && f.Settings() == *lastSettings {
		return
	}
	if err := writeSave(path, f); err != nil {
		fmt.Fprintln(os.Stderr, "copycat: save:", err)
		return
	}
	*lastProgress = f.Progress()
	*lastSettings = f.Settings()
}

// writeSave persists the flow's progress and settings atomically.
func writeSave(path string, f *campaign.Flow) error {
	file := save.Defaults()
	file.Progress = f.Progress()
	file.Settings = f.Settings()
	return file.Write(path)
}

// feedStream fills the mixer's samples into the audio stream when it asks.
func feedStream(stream rl.AudioStream, mixer *audio.Mixer, buf []float32) {
	if !rl.IsAudioStreamProcessed(stream) {
		return
	}
	mixer.Fill(buf)
	rl.UpdateAudioStream(stream, buf)
}
