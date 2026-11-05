package render

import (
	"encoding/binary"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The four synthesized sounds, built in memory after Open and freed in Close.
var (
	jumpSound    rl.Sound
	pickupSound  rl.Sound
	thudSound    rl.Sound
	crackleSound rl.Sound
	soundsLoaded bool
)

// sampleRate is the rate in hertz every synthesized sound shares.
const sampleRate = 22050

// PlayJump triggers the jump blip.
func PlayJump() {
	play(jumpSound)
}

// PlayPickup triggers the pickup chime.
func PlayPickup() {
	play(pickupSound)
}

// PlayThud triggers the farce thud.
func PlayThud() {
	play(thudSound)
}

// PlayCrackle triggers the intercom crackle.
func PlayCrackle() {
	play(crackleSound)
}

// play restarts s so a rapid repeat always retriggers the sound. It does
// nothing before Open has built the sounds.
func play(s rl.Sound) {
	if !soundsLoaded {
		return
	}
	rl.StopSound(s)
	rl.PlaySound(s)
}

// initSounds synthesizes the four sounds from generated wave samples and loads
// them for playback. No sound comes from a file.
func initSounds() {
	jumpSound = buildSound(0.15, jumpGen)
	pickupSound = buildSound(0.4, pickupGen)
	thudSound = buildSound(0.3, thudGen)
	crackleSound = buildSound(0.2, crackleGen)
	soundsLoaded = true
}

// buildSound builds a mono 16-bit wave of durationSec seconds by sampling gen
// at sampleRate, then loads it as a raylib Sound and frees the wave. The
// samples are little-endian PCM bytes, never read from a file.
func buildSound(durationSec float64, gen func(i int, t float64) float64) rl.Sound {
	frames := int(durationSec * sampleRate)
	data := make([]byte, frames*2)
	for i := 0; i < frames; i++ {
		v := gen(i, float64(i)/sampleRate)
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		binary.LittleEndian.PutUint16(data[i*2:], uint16(int16(v*32767)))
	}
	wave := rl.NewWave(uint32(frames), sampleRate, 16, 1, data)
	sound := rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)
	return sound
}

// jumpGen is a short rising blip.
func jumpGen(_ int, t float64) float64 {
	return math.Sin(2*math.Pi*(320+480*(t/0.15))*t) * fadeOut(t, 0.15)
}

// pickupGen is a two-note chime.
func pickupGen(_ int, t float64) float64 {
	if t < 0.16 {
		return 0.7 * math.Sin(2*math.Pi*660*t) * fadeOut(t, 0.16)
	}
	local := t - 0.16
	return 0.7 * math.Sin(2*math.Pi*990*local) * fadeOut(local, 0.24)
}

// thudGen is a low, quickly decaying farce thud.
func thudGen(_ int, t float64) float64 {
	return (math.Sin(2*math.Pi*90*t) + 0.4*math.Sin(2*math.Pi*45*t)) * math.Exp(-14*t)
}

// crackleGen is a burst of deterministic noise for the intercom.
func crackleGen(i int, t float64) float64 {
	noise := float64((i*1103515245+12345)%1000)/500 - 1
	return 0.7 * noise * fadeOut(t, 0.2)
}

// fadeOut fades a sample linearly to silence by t == dur.
func fadeOut(t, dur float64) float64 {
	if dur <= 0 {
		return 0
	}
	x := t / dur
	if x < 0 {
		x = 0
	} else if x > 1 {
		x = 1
	}
	return 1 - x
}
