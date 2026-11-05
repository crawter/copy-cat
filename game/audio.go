package game

import (
	"math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/v5/raylib"
)

const (
	audioSampleRate = 44100
	audioSampleSize = 16
	audioChannels   = 1
)

// Audio owns the six synthesized game sounds and the previous-frame state used
// to detect the transitions that trigger them.
type Audio struct {
	jump   rl.Sound
	pickup rl.Sound
	attack rl.Sound
	hit    rl.Sound
	death  rl.Sound
	end    rl.Sound

	prevVY     float32
	prevStamps int
	prevScene  SceneKind
}

// NewAudio synthesizes every sound from PCM waves generated in code and loaded
// through raylib Wave and LoadSoundFromWave.
func NewAudio() *Audio {
	return &Audio{
		jump:      loadSound(0.15, jumpSample),
		pickup:    loadSound(0.24, pickupSample),
		attack:    loadSound(0.22, attackSample),
		hit:       loadSound(0.16, hitSample),
		death:     loadSound(0.60, deathSample),
		end:       loadSound(0.54, endSample),
		prevScene: ScenePlay,
	}
}

// Update detects each sound's trigger by comparing state against the previous
// frame and plays the matching sound only on the transition.
func (a *Audio) Update(dt float32, g *Game) {
	if g == nil {
		return
	}

	if g.Player.VY > 0 && a.prevVY <= 0 {
		rl.PlaySound(a.jump)
	}
	a.prevVY = g.Player.VY

	if g.Run.Stamps > a.prevStamps {
		rl.PlaySound(a.pickup)
	}
	a.prevStamps = g.Run.Stamps

	if g.Scene == SceneDeath && a.prevScene != SceneDeath {
		rl.PlaySound(a.death)
	}
	if g.Scene == SceneEnd && a.prevScene != SceneEnd {
		rl.PlaySound(a.end)
	}
	a.prevScene = g.Scene

	if rl.IsKeyPressed(rl.KeyX) || rl.IsKeyPressed(rl.KeyEnter) {
		rl.PlaySound(a.attack)
		if a.enemyAhead(g) {
			rl.PlaySound(a.hit)
		}
	}
}

// Draw is a no-op; sound has no visuals.
func (a *Audio) Draw(g *Game) {}

// enemyAhead reports whether an attack press lands on a layout enemy spawn:
// within 2.5 meters ahead and 1.5 meters sideways of the player.
func (a *Audio) enemyAhead(g *Game) bool {
	for i := range g.Layout.Segments {
		spawn := g.Layout.Segments[i].Enemy
		if spawn == nil {
			continue
		}
		dz := spawn.Z - g.Player.Z
		if dz < 0 || dz > 2.5 {
			continue
		}
		dx := g.Layout.LaneX(spawn.Lane) - g.Player.X
		if dx < 0 {
			dx = -dx
		}
		if dx <= 1.5 {
			return true
		}
	}
	return false
}

// loadSound synthesizes a mono 16-bit PCM wave from sample and loads it as a
// raylib sound. The wave buffer is Go memory pinned for the cgo call; raylib
// copies the samples into the sound, so the buffer must not be unloaded and is
// left for the garbage collector.
func loadSound(duration float32, sample func(i int, t float32) float32) rl.Sound {
	frameCount := int(duration * audioSampleRate)
	if frameCount < 1 {
		frameCount = 1
	}

	data := make([]int16, frameCount)
	for i := range data {
		t := float32(i) / audioSampleRate
		v := sample(i, t)
		switch {
		case v > 1:
			v = 1
		case v < -1:
			v = -1
		}
		data[i] = int16(v * math.MaxInt16)
	}

	wave := rl.Wave{
		FrameCount: uint32(frameCount),
		SampleRate: audioSampleRate,
		SampleSize: audioSampleSize,
		Channels:   audioChannels,
		Data:       unsafe.Pointer(&data[0]),
	}

	return rl.LoadSoundFromWave(wave)
}

// sweepPhase returns the phase in radians at time t of a linear frequency sweep
// from f0 to f1 over duration, starting at zero phase.
func sweepPhase(f0, f1, t, duration float32) float64 {
	return 2 * math.Pi * float64(f0*t+(f1-f0)*t*t/(2*duration))
}

// hashNoise returns a deterministic pseudo-random value in [-1, 1).
func hashNoise(i int) float32 {
	x := uint32(i) * 2654435761
	x ^= x >> 13
	x *= 0x5bd1e995
	x ^= x >> 15
	return float32(x%20001)/10000 - 1
}

// jumpSample is a short rising blip.
func jumpSample(i int, t float32) float32 {
	const duration = 0.15
	if t > duration {
		return 0
	}
	env := 1 - t/duration
	return float32(math.Sin(sweepPhase(300, 760, t, duration))) * env
}

// pickupSample is a bright two-note chime.
func pickupSample(i int, t float32) float32 {
	const note = 0.12
	if t > 2*note {
		return 0
	}
	var freq float32
	var local float32
	if t < note {
		freq = 880
		local = t
	} else {
		freq = 1320
		local = t - note
	}
	env := 1 - local/note
	return float32(math.Sin(2*math.Pi*float64(freq)*float64(local))) * env
}

// attackSample is a noise whoosh with a falling pitch.
func attackSample(i int, t float32) float32 {
	const duration = 0.22
	if t > duration {
		return 0
	}
	env := float32(math.Sin(math.Pi * float64(t/duration)))
	sweep := float32(math.Sin(sweepPhase(600, 180, t, duration)))
	return 0.6*sweep*env + 0.5*hashNoise(i)*env
}

// hitSample is a low thump with a punchy noise attack.
func hitSample(i int, t float32) float32 {
	const duration = 0.16
	if t > duration {
		return 0
	}
	env := 1 - t/duration
	thump := float32(math.Sin(2 * math.Pi * 90 * float64(t)))
	return thump*env + 0.4*hashNoise(i)*env
}

// deathSample is a slow descending slide.
func deathSample(i int, t float32) float32 {
	const duration = 0.60
	if t > duration {
		return 0
	}
	env := 1 - t/duration
	return float32(math.Sin(sweepPhase(520, 80, t, duration))) * env
}

// endSample is a little ascending fanfare.
func endSample(i int, t float32) float32 {
	const note = 0.18
	if t > 3*note {
		return 0
	}
	freqs := [3]float32{523.25, 659.25, 783.99}
	idx := int(t / note)
	if idx > 2 {
		idx = 2
	}
	freq := freqs[idx]
	local := t - float32(idx)*note
	env := 1 - local/note
	return float32(math.Sin(2*math.Pi*float64(freq)*float64(local))) * env
}
