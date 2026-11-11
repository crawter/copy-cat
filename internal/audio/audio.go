// Package audio synthesizes the game's music and sound effects in code.
//
// It renders PCM float32 samples from Go code only and never imports raylib,
// so its tests run headless. A Mixer keeps its own sample clock, advances it
// by exactly the number of frames written, and renders the same samples for a
// given frame offset on every fresh mixer.
package audio

// Track identifies one looping music track.
type Track int

// The looping music tracks, one per world theme plus menu, boss and ending
// themes.
const (
	// TrackMenu is the looping theme played on the main menu.
	TrackMenu Track = iota
	// TrackMangroves is the looping theme for the wet mangrove world.
	TrackMangroves
	// TrackHarbor is the looping theme for the rusty harbor world.
	TrackHarbor
	// TrackNeon is the looping theme for the neon Chic-Chic street world.
	TrackNeon
	// TrackBoss is the tense, faster looping theme for the boss fight.
	TrackBoss
	// TrackEnding is the gentle theme played on the ending and credits scenes.
	TrackEnding
)

// trackCount is the number of tracks and bounds the mixer's loop caches.
const trackCount = int(TrackEnding) + 1

// Sound identifies one one-shot sound effect.
type Sound int

// The one-shot sound effects the game can play.
const (
	// SoundJump is the short rising blip played when the hero jumps.
	SoundJump Sound = iota
	// SoundSlide is the sliding noise played when the hero slides.
	SoundSlide
	// SoundSlam is the low thud played when the hero slams onto the ground.
	SoundSlam
	// SoundComb is the comb-teeth tick played when a comb is picked up.
	SoundComb
	// SoundHairspray is the hiss played when a hairspray is picked up.
	SoundHairspray
	// SoundHit is the impact played when the hero takes a hit.
	SoundHit
	// SoundSnip is the scissor snip played when the Copycat cuts hair.
	SoundSnip
	// SoundCheckpoint is the chime played when a checkpoint is reached.
	SoundCheckpoint
	// SoundStomp is the low crush played when the hero stomps an enemy.
	SoundStomp
	// SoundFall is the descending whoosh and thud played when the hero falls
	// into a pit.
	SoundFall
	// SoundFanfare is the victory fanfare played when a run is finished.
	SoundFanfare
	// SoundGameOver is the descending phrase played when the hero dies.
	SoundGameOver
	// SoundMenuMove is the short blip played when a menu selection moves.
	SoundMenuMove
	// SoundMenuSelect is the chime played when a menu selection is confirmed.
	SoundMenuSelect
	// SoundExtraLife is the rising arpeggio played when an extra life is won.
	SoundExtraLife
	// SoundBossHit is the impact played when the boss hits the hero.
	SoundBossHit
	// SoundHonk is the honk played when the hero falls into a pit.
	SoundHonk
	// SoundMeow is the meow played when Copycat snips hair or the boss phase
	// rises.
	SoundMeow
)

const (
	// sampleRateFloor and sampleRateCeil bound New's argument so a caller
	// error cannot make the mixer allocate without bound.
	sampleRateFloor = 8000
	sampleRateCeil  = 192000

	// defaultSampleRate is used when New receives an unusable rate.
	defaultSampleRate = 44100

	// maxActiveSounds bounds how many one-shot effects may overlap.
	maxActiveSounds = 64
)

// Mixer renders music and sound effects into interleaved stereo PCM.
//
// It owns its sample clock and never reads the wall clock.
type Mixer struct {
	sampleRate int

	clock int // stereo frames rendered so far

	musicTrack  Track
	musicActive bool
	musicStart  int // clock value when the current track started

	loops      [trackCount][]float32 // synthesized loop, cached per track
	loopFrames [trackCount]int       // loop length in stereo frames

	musicVolume float32
	soundVolume float32

	effects []effect
}

// effect is one playing one-shot sound.
type effect struct {
	sound Sound
	start int // clock value when the sound started
	dur   int // length in stereo frames
}

// New returns a ready-to-use mixer that renders at sampleRate samples per
// second. Rates below the supported floor fall back to the default rate and
// rates above the ceiling are capped, so a caller error cannot cause an
// unbounded allocation.
func New(sampleRate int) *Mixer {
	if sampleRate < sampleRateFloor {
		sampleRate = defaultSampleRate
	}
	if sampleRate > sampleRateCeil {
		sampleRate = sampleRateCeil
	}
	return &Mixer{
		sampleRate:  sampleRate,
		musicVolume: 1,
		soundVolume: 1,
		effects:     make([]effect, 0, maxActiveSounds),
	}
}

// PlayMusic starts t from its beginning, looping seamlessly. Calling it again
// restarts the track from the top.
func (m *Mixer) PlayMusic(t Track) {
	if t < TrackMenu {
		t = TrackMenu
	}
	if t > TrackEnding {
		t = TrackEnding
	}
	m.musicTrack = t
	m.musicStart = m.clock
	m.musicActive = true
	if m.loops[t] == nil {
		m.loops[t] = renderTrack(m.sampleRate, trackDefs[t])
		m.loopFrames[t] = len(m.loops[t]) / 2
	}
}

// StopMusic stops the current music track. Sound effects keep playing.
func (m *Mixer) StopMusic() {
	m.musicActive = false
}

// Play starts a one-shot sound effect mixed over the music. If more than
// maxActiveSounds effects overlap, the oldest one is dropped.
func (m *Mixer) Play(s Sound) {
	if s < SoundJump {
		s = SoundJump
	}
	if s > SoundMeow {
		s = SoundMeow
	}
	if len(m.effects) == maxActiveSounds {
		copy(m.effects, m.effects[1:])
		m.effects = m.effects[:len(m.effects)-1]
	}
	m.effects = append(m.effects, effect{
		sound: s,
		start: m.clock,
		dur:   soundFrames(s, m.sampleRate),
	})
}

// SetMusicVolume sets the music level, clamped to the range 0 through 1.
func (m *Mixer) SetMusicVolume(v float32) {
	m.musicVolume = clamp01(v)
}

// SetSoundVolume sets the one-shot effect level, clamped to 0 through 1.
func (m *Mixer) SetSoundVolume(v float32) {
	m.soundVolume = clamp01(v)
}

// Fill writes exactly len(buf) interleaved stereo samples: one frame of two
// samples, left then right, for a raylib audio stream. It advances the
// mixer's clock by the exact number of complete frames written. The final
// sample of an odd-length buffer is filled with silence.
func (m *Mixer) Fill(buf []float32) {
	frames := len(buf) / 2
	for f := 0; f < frames; f++ {
		var left, right float64

		if m.musicActive {
			pos := (m.clock - m.musicStart) % m.loopFrames[m.musicTrack]
			idx := pos * 2
			mv := float64(m.musicVolume)
			left += float64(m.loops[m.musicTrack][idx]) * mv
			right += float64(m.loops[m.musicTrack][idx+1]) * mv
		}

		sv := float64(m.soundVolume)
		for i := 0; i < len(m.effects); i++ {
			e := m.effects[i]
			pos := m.clock - e.start
			if pos >= e.dur {
				m.effects = append(m.effects[:i], m.effects[i+1:]...)
				i--
				continue
			}
			sample := float64(renderEffect(e.sound, pos, e.dur, m.sampleRate)) * sv
			left += sample
			right += sample
		}

		buf[f*2] = clampSample(left)
		buf[f*2+1] = clampSample(right)
		m.clock++
	}
	if len(buf)%2 == 1 {
		buf[len(buf)-1] = 0
	}
}

// clamp01 returns v clamped to the closed range [0, 1].
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// clampSample returns the PCM sample v clamped to the range [-1, 1].
func clampSample(v float64) float32 {
	if v > 1 {
		return 1
	}
	if v < -1 {
		return -1
	}
	return float32(v)
}
