package audio

import "math"

// wave is the oscillator shape used by a track's notes.
type wave int

const (
	waveSquare wave = iota
	waveTriangle
	waveSine
)

// note is one melodic event: a pitch held for a number of beats.
type note struct {
	start float64 // beat the note begins on
	dur   float64 // length in beats
	semi  int     // pitch as semitones above C4
	wave  wave
}

// percKind distinguishes the two percussion voices.
type percKind int

const (
	percKick percKind = iota
	percHat
)

// perc is one percussion hit.
type perc struct {
	start float64
	dur   float64
	kind  percKind
}

// trackDef is the musical content of one looping track.
type trackDef struct {
	bpm   float64
	beats float64
	lead  []note
	bass  []note
	percs []perc
}

const (
	// baseFreq is middle C (C4) in hertz; semitones are measured from it.
	baseFreq = 261.6255653

	minSemitone = -24 // C2
	maxSemitone = 36  // C7
)

// semitoneFreqs maps semitones above C4 to frequencies in hertz.
var semitoneFreqs [maxSemitone - minSemitone + 1]float64

func init() {
	for i := range semitoneFreqs {
		semitoneFreqs[i] = baseFreq * math.Pow(2, float64(i+minSemitone)/12)
	}
}

// freqFor returns the frequency, in hertz, of the pitch semi semitones above
// C4. Out-of-range pitches are clamped to the supported table.
func freqFor(semi int) float64 {
	if semi < minSemitone {
		semi = minSemitone
	}
	if semi > maxSemitone {
		semi = maxSemitone
	}
	return semitoneFreqs[semi-minSemitone]
}

// waveSample returns the oscillator value in [-1, 1] for the given phase.
func waveSample(w wave, phase float64) float64 {
	switch w {
	case waveSquare:
		if math.Sin(phase) >= 0 {
			return 0.6
		}
		return -0.6
	case waveTriangle:
		return triangleWave(phase)
	default: // waveSine
		return math.Sin(phase)
	}
}

// triangleWave maps phase into a triangle wave in [-1, 1].
func triangleWave(phase float64) float64 {
	const twoPi = 2 * math.Pi
	p := math.Mod(phase, twoPi)
	if p < 0 {
		p += twoPi
	}
	v := p / twoPi // 0..1
	switch {
	case v < 0.25:
		return 4 * v
	case v < 0.75:
		return 2 - 4*v
	default:
		return 4*v - 4
	}
}

// noteEnvelope fades a note in and out so every note starts and ends at zero.
// progress is 0 at the note's start and 1 at its end.
func noteEnvelope(progress float64) float64 {
	const attack = 0.06
	const release = 0.18
	if progress < attack {
		return progress / attack
	}
	if progress > 1-release {
		return (1 - progress) / release
	}
	return 1
}

// percEnvelope fades a percussion hit in and out; the release dominates so
// each hit ends at silence before the next beat.
func percEnvelope(progress float64) float64 {
	const attack = 0.005
	const release = 0.5
	if progress < attack {
		return progress / attack
	}
	if progress > 1-release {
		return (1 - progress) / release
	}
	return 1
}

// chirpPhase returns the number of cycles elapsed for a pitch that moves
// linearly from f0 to f1 hertz over dur seconds, at time t seconds.
func chirpPhase(t, dur, f0, f1 float64) float64 {
	if dur <= 0 {
		return 0
	}
	return f0*t + 0.5*(f1-f0)*t*t/dur
}

// noiseHash turns n into deterministic white noise in [-1, 1].
func noiseHash(n uint32) float64 {
	x := n
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	return float64(x&0xffff)/32767.5 - 1
}

// renderTrack synthesizes one complete loop of def, stereo interleaved.
func renderTrack(sampleRate int, def trackDef) []float32 {
	secondsPerBeat := 60 / def.bpm
	totalFrames := int(math.Round(def.beats * secondsPerBeat * float64(sampleRate)))
	buf := make([]float32, totalFrames*2)
	for frame := 0; frame < totalFrames; frame++ {
		t := float64(frame) / float64(sampleRate)
		beat := t / secondsPerBeat

		lead := renderVoice(def.lead, beat, secondsPerBeat)
		bass := renderVoice(def.bass, beat, secondsPerBeat)
		perc := renderPerc(def.percs, beat, secondsPerBeat, frame)

		// Pan the voices slightly apart for width and keep the mix centered.
		left := 0.45*lead + 0.35*bass + 0.5*perc
		right := 0.35*lead + 0.45*bass + 0.5*perc
		buf[frame*2] = clampSample(left)
		buf[frame*2+1] = clampSample(right)
	}
	return buf
}

// renderVoice sums every note sounding at beat, in the voice's own time base.
func renderVoice(notes []note, beat, secondsPerBeat float64) float64 {
	var out float64
	for _, n := range notes {
		end := n.start + n.dur
		if beat < n.start || beat >= end {
			continue
		}
		progress := (beat - n.start) / n.dur
		env := noteEnvelope(progress)
		t := (beat - n.start) * secondsPerBeat
		durSec := n.dur * secondsPerBeat
		freq := freqFor(n.semi)
		phase := 2 * math.Pi * chirpPhase(t, durSec, freq, freq)
		out += waveSample(n.wave, phase) * env * 0.5
	}
	return out
}

// renderPerc sums every percussion hit sounding at beat.
func renderPerc(percs []perc, beat, secondsPerBeat float64, frame int) float64 {
	var out float64
	for _, p := range percs {
		end := p.start + p.dur
		if beat < p.start || beat >= end {
			continue
		}
		progress := (beat - p.start) / p.dur
		env := percEnvelope(progress)
		t := (beat - p.start) * secondsPerBeat
		durSec := p.dur * secondsPerBeat
		if p.kind == percKick {
			phase := 2 * math.Pi * chirpPhase(t, durSec, 110, 40)
			out += math.Sin(phase) * env
		} else {
			out += noiseHash(uint32(frame)) * env * 0.5
		}
	}
	return out
}

// note builders keep the track tables compact.
func sq(start, dur float64, semi int) note { return note{start: start, dur: dur, semi: semi, wave: waveSquare} }
func tri(start, dur float64, semi int) note { return note{start: start, dur: dur, semi: semi, wave: waveTriangle} }
func kick(start, dur float64) perc         { return perc{start: start, dur: dur, kind: percKick} }
func hat(start, dur float64) perc          { return perc{start: start, dur: dur, kind: percHat} }

// trackDefs holds the musical content of every looping track.
var trackDefs = [trackCount]trackDef{
	TrackMenu: trackDef{
		bpm:   100,
		beats: 16,
		lead: []note{
			sq(0, 1, 9), sq(1, 1, 12), sq(2, 1, 16), sq(3, 1, 9),
			sq(4, 1, 7), sq(5, 1, 11), sq(6, 1, 4), sq(7, 1, 9),
			sq(8, 1, 5), sq(9, 1, 9), sq(10, 1, 14), sq(11, 1, 12),
			sq(12, 1, 11), sq(13, 1, 7), sq(14, 1, 4), sq(15, 1, 9),
		},
		bass: []note{
			tri(0, 2, -15), tri(2, 2, -19), tri(4, 2, -12), tri(6, 2, -17),
			tri(8, 2, -19), tri(10, 2, -15), tri(12, 2, -10), tri(14, 2, -8),
		},
		percs: []perc{
			kick(0, 0.5), kick(4, 0.5), kick(8, 0.5), kick(12, 0.5),
			hat(2, 0.25), hat(6, 0.25), hat(10, 0.25), hat(14, 0.25),
		},
	},
	TrackMangroves: trackDef{
		bpm:   80,
		beats: 16,
		lead: []note{
			tri(0, 2, -3), tri(2, 1, 0), tri(3, 1, 4),
			tri(4, 2, 2), tri(6, 2, -3),
			tri(8, 2, -5), tri(10, 2, -1),
			tri(12, 1, 4), tri(13, 1, 2), tri(14, 2, 0),
		},
		bass: []note{
			tri(0, 4, -15), tri(4, 4, -19), tri(8, 4, -12), tri(12, 4, -17),
		},
		percs: []perc{
			kick(0, 0.5), kick(8, 0.5),
			hat(4, 0.25), hat(12, 0.25),
		},
	},
	TrackHarbor: trackDef{
		bpm:   100,
		beats: 16,
		lead: []note{
			sq(0, 1, 4), sq(1, 1, 7), sq(2, 1, 9), sq(3, 1, 4),
			sq(4, 1, 2), sq(5, 1, 4), sq(6, 1, 7), sq(7, 1, 2),
			sq(8, 1, 0), sq(9, 1, 4), sq(10, 1, 7), sq(11, 1, 12),
			sq(12, 1, 11), sq(13, 1, 7), sq(14, 1, 4), sq(15, 1, 2),
		},
		bass: []note{
			tri(0, 2, -20), tri(2, 2, -15), tri(4, 2, -17), tri(6, 2, -10),
			tri(8, 2, -20), tri(10, 2, -15), tri(12, 2, -14), tri(14, 2, -17),
		},
		percs: []perc{
			kick(0, 0.4), kick(2, 0.4), kick(4, 0.4), kick(6, 0.4),
			kick(8, 0.4), kick(10, 0.4), kick(12, 0.4), kick(14, 0.4),
			hat(1, 0.2), hat(3, 0.2), hat(5, 0.2), hat(7, 0.2),
			hat(9, 0.2), hat(11, 0.2), hat(13, 0.2), hat(15, 0.2),
		},
	},
	TrackNeon: trackDef{
		bpm:   132,
		beats: 16,
		lead: []note{
			sq(0, 0.5, 9), sq(0.5, 0.5, 12), sq(1, 0.5, 16), sq(1.5, 0.5, 21),
			sq(2, 0.5, 16), sq(2.5, 0.5, 12), sq(3, 0.5, 9), sq(3.5, 0.5, 12),
			sq(4, 0.5, 5), sq(4.5, 0.5, 9), sq(5, 0.5, 12), sq(5.5, 0.5, 17),
			sq(6, 0.5, 12), sq(6.5, 0.5, 9), sq(7, 0.5, 5), sq(7.5, 0.5, 9),
			sq(8, 0.5, 4), sq(8.5, 0.5, 7), sq(9, 0.5, 12), sq(9.5, 0.5, 16),
			sq(10, 0.5, 12), sq(10.5, 0.5, 7), sq(11, 0.5, 4), sq(11.5, 0.5, 7),
			sq(12, 0.5, 2), sq(12.5, 0.5, 7), sq(13, 0.5, 11), sq(13.5, 0.5, 14),
			sq(14, 0.5, 11), sq(14.5, 0.5, 7), sq(15, 0.5, 2), sq(15.5, 0.5, 7),
		},
		bass: []note{
			tri(0, 2, -15), tri(2, 2, -15), tri(4, 2, -19), tri(6, 2, -19),
			tri(8, 2, -12), tri(10, 2, -12), tri(12, 2, -17), tri(14, 2, -17),
		},
		percs: []perc{
			kick(0, 0.35), kick(2, 0.35), kick(4, 0.35), kick(6, 0.35),
			kick(8, 0.35), kick(10, 0.35), kick(12, 0.35), kick(14, 0.35),
			hat(0.5, 0.15), hat(1.5, 0.15), hat(2.5, 0.15), hat(3.5, 0.15),
			hat(4.5, 0.15), hat(5.5, 0.15), hat(6.5, 0.15), hat(7.5, 0.15),
			hat(8.5, 0.15), hat(9.5, 0.15), hat(10.5, 0.15), hat(11.5, 0.15),
			hat(12.5, 0.15), hat(13.5, 0.15), hat(14.5, 0.15), hat(15.5, 0.15),
		},
	},
	TrackBoss: trackDef{
		bpm:   150,
		beats: 16,
		lead: []note{
			sq(0, 0.5, 16), sq(0.5, 0.5, 17), sq(1, 0.5, 16), sq(1.5, 0.5, 17),
			sq(2, 0.5, 19), sq(2.5, 0.5, 17), sq(3, 0.5, 16), sq(3.5, 0.5, 17),
			sq(4, 0.5, 14), sq(4.5, 0.5, 16), sq(5, 0.5, 14), sq(5.5, 0.5, 16),
			sq(6, 0.5, 17), sq(6.5, 0.5, 16), sq(7, 0.5, 14), sq(7.5, 0.5, 12),
			sq(8, 0.5, 16), sq(8.5, 0.5, 17), sq(9, 0.5, 16), sq(9.5, 0.5, 17),
			sq(10, 0.5, 19), sq(10.5, 0.5, 17), sq(11, 0.5, 16), sq(11.5, 0.5, 17),
			sq(12, 0.5, 19), sq(12.5, 0.5, 22), sq(13, 0.5, 19), sq(13.5, 0.5, 22),
			sq(14, 0.5, 21), sq(14.5, 0.5, 19), sq(15, 0.5, 17), sq(15.5, 0.5, 16),
		},
		bass: []note{
			tri(0, 0.5, -20), tri(0.5, 0.5, -20), tri(1, 0.5, -8), tri(1.5, 0.5, -20),
			tri(2, 0.5, -20), tri(2.5, 0.5, -8), tri(3, 0.5, -20), tri(3.5, 0.5, -20),
			tri(4, 0.5, -20), tri(4.5, 0.5, -20), tri(5, 0.5, -8), tri(5.5, 0.5, -20),
			tri(6, 0.5, -20), tri(6.5, 0.5, -8), tri(7, 0.5, -20), tri(7.5, 0.5, -20),
			tri(8, 0.5, -20), tri(8.5, 0.5, -20), tri(9, 0.5, -8), tri(9.5, 0.5, -20),
			tri(10, 0.5, -20), tri(10.5, 0.5, -8), tri(11, 0.5, -20), tri(11.5, 0.5, -20),
			tri(12, 0.5, -14), tri(12.5, 0.5, -14), tri(13, 0.5, -2), tri(13.5, 0.5, -14),
			tri(14, 0.5, -14), tri(14.5, 0.5, -2), tri(15, 0.5, -14), tri(15.5, 0.5, -14),
		},
		percs: []perc{
			kick(0, 0.3), kick(2, 0.3), kick(4, 0.3), kick(6, 0.3),
			kick(8, 0.3), kick(10, 0.3), kick(12, 0.3), kick(14, 0.3),
			hat(0.5, 0.15), hat(1.5, 0.15), hat(2.5, 0.15), hat(3.5, 0.15),
			hat(4.5, 0.15), hat(5.5, 0.15), hat(6.5, 0.15), hat(7.5, 0.15),
			hat(8.5, 0.15), hat(9.5, 0.15), hat(10.5, 0.15), hat(11.5, 0.15),
			hat(12.5, 0.15), hat(13.5, 0.15), hat(14.5, 0.15), hat(15.5, 0.15),
		},
	},
	TrackEnding: trackDef{
		bpm:   72,
		beats: 16,
		lead: []note{
			tri(0, 2, 12), tri(2, 2, 16), tri(4, 2, 19), tri(6, 2, 16),
			tri(8, 2, 14), tri(10, 2, 16), tri(12, 4, 12),
		},
		bass: []note{
			tri(0, 4, -12), tri(4, 4, -15), tri(8, 4, -10), tri(12, 4, -8),
		},
		percs: []perc{
			kick(0, 0.6), kick(8, 0.6),
			hat(4, 0.3), hat(12, 0.3),
		},
	},
}

// soundDur holds each effect's length in seconds.
var soundDur = [...]float32{
	SoundJump:        0.18,
	SoundSlide:       0.32,
	SoundSlam:        0.28,
	SoundComb:        0.16,
	SoundHairspray:   0.45,
	SoundHit:         0.15,
	SoundSnip:        0.22,
	SoundCheckpoint:  0.55,
	SoundStomp:       0.25,
	SoundFall:        0.60,
	SoundFanfare:     1.20,
	SoundGameOver:    1.50,
	SoundMenuMove:    0.08,
	SoundMenuSelect:  0.25,
	SoundExtraLife:   0.90,
	SoundBossHit:     0.20,
	SoundHonk:        0.50,
	SoundMeow:        0.60,
}

// soundFrames returns the effect's length in stereo frames at sampleRate.
func soundFrames(s Sound, sampleRate int) int {
	return int(soundDur[s] * float32(sampleRate))
}

var (
	combTeeth        = [3]float64{2200, 1800, 1400}
	snipPitches      = [2]float64{3200, 2600}
	checkpointNotes  = [4]int{12, 16, 19, 24}
	fanfareNotes     = []int{12, 16, 19, 24, 28}
	gameOverNotes    = []int{16, 12, 9, 5, 0}
	menuSelectNotes  = []int{9, 16}
	extraLifeNotes   = []int{0, 4, 7, 12, 16}
)

// fxEnvelope ramps a one-shot in and out so it starts and ends at zero.
func fxEnvelope(progress, attack, release float64) float64 {
	if progress < attack {
		return progress / attack
	}
	if progress > 1-release {
		return (1 - progress) / release
	}
	return 1
}

// seqEffect renders one frame of a sequence of pitches spread evenly over the
// effect, each note fading in and out. It returns silence for an empty
// sequence.
func seqEffect(notes []int, progress, t, D float64, w wave) float32 {
	n := len(notes)
	if n == 0 {
		return 0
	}
	seg := int(progress * float64(n))
	if seg > n-1 {
		seg = n - 1
	}
	segStart := float64(seg) / float64(n)
	p := (progress - segStart) * float64(n)
	phase := 2 * math.Pi * freqFor(notes[seg]) * (t - segStart*D)
	out := waveSample(w, phase) * 0.6
	return clampSample(out * fxEnvelope(p, 0.25, 0.45))
}

// renderEffect synthesizes one frame of a sound effect, deterministic in pos.
func renderEffect(s Sound, pos, dur, sampleRate int) float32 {
	if dur <= 0 || pos < 0 || pos >= dur {
		return 0
	}
	t := float64(pos) / float64(sampleRate)
	D := float64(dur) / float64(sampleRate)
	progress := float64(pos) / float64(dur)

	switch s {
	case SoundJump:
		phase := 2 * math.Pi * chirpPhase(t, D, 250, 700)
		out := waveSample(waveSquare, phase) * 0.6
		return clampSample(out * fxEnvelope(progress, 0.05, 0.4))
	case SoundSlide:
		phase := 2 * math.Pi * chirpPhase(t, D, 900, 200)
		out := waveSample(waveTriangle, phase)*0.6 + noiseHash(uint32(pos))*0.45
		return clampSample(out * fxEnvelope(progress, 0.05, 0.55))
	case SoundSlam:
		phase := 2 * math.Pi * chirpPhase(t, D, 130, 45)
		out := math.Sin(phase)*0.9 + noiseHash(uint32(pos))*0.35
		return clampSample(out * fxEnvelope(progress, 0.01, 0.6))
	case SoundComb:
		seg := int(progress * 3)
		if seg > 2 {
			seg = 2
		}
		segStart := float64(seg) / 3
		p := (progress - segStart) * 3
		phase := 2 * math.Pi * combTeeth[seg] * (t - segStart*D)
		out := waveSample(waveSquare, phase) * 0.5
		return clampSample(out * fxEnvelope(p, 0.15, 0.7))
	case SoundHairspray:
		out := noiseHash(uint32(pos)) * 0.55
		return clampSample(out * fxEnvelope(progress, 0.35, 0.5))
	case SoundHit:
		phase := 2 * math.Pi * chirpPhase(t, D, 110, 70)
		out := waveSample(waveSquare, phase)*0.7 + noiseHash(uint32(pos))*0.35
		return clampSample(out * fxEnvelope(progress, 0.01, 0.55))
	case SoundSnip:
		seg := 0
		if progress >= 0.5 {
			seg = 1
		}
		segStart := float64(seg) / 2
		p := (progress - segStart) * 2
		phase := 2 * math.Pi * snipPitches[seg] * (t - segStart*D)
		out := waveSample(waveSquare, phase)*0.5 + noiseHash(uint32(pos))*0.35
		return clampSample(out * fxEnvelope(p, 0.1, 0.8))
	case SoundCheckpoint:
		seg := int(progress * 4)
		if seg > 3 {
			seg = 3
		}
		segStart := float64(seg) / 4
		p := (progress - segStart) * 4
		phase := 2 * math.Pi * freqFor(checkpointNotes[seg]) * (t - segStart*D)
		out := waveSample(waveTriangle, phase) * 0.6
		return clampSample(out * fxEnvelope(p, 0.25, 0.45))
	case SoundStomp:
		phase := 2 * math.Pi * chirpPhase(t, D, 110, 45)
		out := math.Sin(phase)*0.95 + noiseHash(uint32(pos))*0.4
		return clampSample(out * fxEnvelope(progress, 0.01, 0.55))
	case SoundFall:
		phase := 2 * math.Pi * chirpPhase(t, D, 700, 150)
		out := waveSample(waveTriangle, phase)*0.55 + noiseHash(uint32(pos))*0.25
		if progress > 0.65 {
			p := (progress - 0.65) / 0.35
			thudPhase := 2 * math.Pi * chirpPhase(t-0.65*D, 0.35*D, 100, 45)
			out += math.Sin(thudPhase) * 0.6 * fxEnvelope(p, 0.02, 0.5)
		}
		return clampSample(out * fxEnvelope(progress, 0.1, 0.4))
	case SoundFanfare:
		return seqEffect(fanfareNotes, progress, t, D, waveSquare)
	case SoundGameOver:
		return seqEffect(gameOverNotes, progress, t, D, waveTriangle)
	case SoundMenuMove:
		phase := 2 * math.Pi * chirpPhase(t, D, 900, 1100)
		out := waveSample(waveSquare, phase) * 0.4
		return clampSample(out * fxEnvelope(progress, 0.1, 0.6))
	case SoundMenuSelect:
		return seqEffect(menuSelectNotes, progress, t, D, waveSquare)
	case SoundExtraLife:
		return seqEffect(extraLifeNotes, progress, t, D, waveTriangle)
	case SoundBossHit:
		phase := 2 * math.Pi * chirpPhase(t, D, 180, 90)
		out := waveSample(waveSquare, phase)*0.6 + noiseHash(uint32(pos))*0.4
		return clampSample(out * fxEnvelope(progress, 0.01, 0.5))
	case SoundHonk:
		freqPhase := 2 * math.Pi * chirpPhase(t, D, 320, 270)
		out := waveSample(waveSquare, freqPhase)*0.5 + waveSample(waveSquare, 2*freqPhase)*0.3
		return clampSample(out * fxEnvelope(progress, 0.08, 0.35))
	case SoundMeow:
		var freqPhase float64
		if progress < 0.5 {
			freqPhase = 2 * math.Pi * chirpPhase(t, 0.5*D, 420, 720)
		} else {
			freqPhase = 2 * math.Pi * chirpPhase(t-0.5*D, 0.5*D, 720, 380)
		}
		vibrato := 1 + 0.04*math.Sin(2*math.Pi*6*t)
		out := waveSample(waveTriangle, freqPhase*vibrato) * 0.6
		return clampSample(out * fxEnvelope(progress, 0.15, 0.4))
	}
	return 0
}
