package audio

import (
	"math"
	"testing"
)

const testSampleRate = 44100

func hasSignal(buf []float32) bool {
	for _, v := range buf {
		if v != 0 {
			return true
		}
	}
	return false
}

// allTracks lists every looping track, ending theme included.
func allTracks() []Track {
	return []Track{TrackMenu, TrackMangroves, TrackHarbor, TrackNeon, TrackBoss, TrackEnding}
}

// allSounds lists every one-shot effect, the ten new ones included.
func allSounds() []Sound {
	return []Sound{
		SoundJump, SoundSlide, SoundSlam, SoundComb, SoundHairspray, SoundHit,
		SoundSnip, SoundCheckpoint, SoundStomp, SoundFall, SoundFanfare,
		SoundGameOver, SoundMenuMove, SoundMenuSelect, SoundExtraLife,
		SoundBossHit, SoundHonk, SoundMeow,
	}
}

func TestNewClampsSampleRate(t *testing.T) {
	for _, rate := range []int{0, -100, 1, 7999, 8000, 44100, 192000, 200000} {
		m := New(rate)
		if m.sampleRate < sampleRateFloor || m.sampleRate > sampleRateCeil {
			t.Fatalf("New(%d) sample rate %d out of range", rate, m.sampleRate)
		}
	}
}

func TestEveryTrackRendersNonSilent(t *testing.T) {
	for _, tr := range allTracks() {
		m := New(testSampleRate)
		m.PlayMusic(tr)
		buf := make([]float32, m.loopFrames[tr]*2)
		m.Fill(buf)
		if !hasSignal(buf) {
			t.Errorf("track %v rendered only silence", tr)
		}
	}
}

func TestEverySoundRendersNonSilent(t *testing.T) {
	for _, snd := range allSounds() {
		m := New(testSampleRate)
		m.Play(snd)
		buf := make([]float32, soundFrames(snd, testSampleRate)*2)
		m.Fill(buf)
		if !hasSignal(buf) {
			t.Errorf("sound %v rendered only silence", snd)
		}
	}
}

func TestEverySoundHasPositiveDuration(t *testing.T) {
	for _, snd := range allSounds() {
		if soundDur[snd] <= 0 {
			t.Errorf("sound %v has non-positive duration %v", snd, soundDur[snd])
		}
		if soundFrames(snd, testSampleRate) <= 0 {
			t.Errorf("sound %v has non-positive frame count", snd)
		}
	}
}

func TestFillAdvancesClockByFramesWritten(t *testing.T) {
	m := New(testSampleRate)
	m.PlayMusic(TrackMenu)
	const frames = 1234
	buf := make([]float32, frames*2)
	m.Fill(buf)
	if m.clock != frames {
		t.Fatalf("clock = %d, want %d", m.clock, frames)
	}
	m.Fill(buf)
	if m.clock != 2*frames {
		t.Fatalf("clock after second fill = %d, want %d", m.clock, 2*frames)
	}
}

func TestFillOddLengthWritesSilenceAtTail(t *testing.T) {
	m := New(testSampleRate)
	m.SetMusicVolume(0)
	m.SetSoundVolume(0)
	buf := make([]float32, 5)
	m.Fill(buf)
	if buf[4] != 0 {
		t.Fatalf("tail sample = %v, want 0", buf[4])
	}
	if m.clock != 2 {
		t.Fatalf("clock = %d, want 2 complete frames", m.clock)
	}
}

func TestMusicVolumeScalesOutput(t *testing.T) {
	m := New(testSampleRate)
	m.PlayMusic(TrackMenu)
	m.SetSoundVolume(0)
	full := make([]float32, testSampleRate*2)
	m.Fill(full)

	quiet := New(testSampleRate)
	quiet.PlayMusic(TrackMenu)
	quiet.SetMusicVolume(0.5)
	quiet.SetSoundVolume(0)
	q := make([]float32, testSampleRate*2)
	quiet.Fill(q)

	for i := range q {
		want := float64(full[i]) * 0.5
		if math.Abs(float64(q[i])-want) > 1e-6 {
			t.Fatalf("sample %d: got %v, want %v", i, q[i], want)
		}
	}
}

func TestSoundVolumeSilencesEffects(t *testing.T) {
	m := New(testSampleRate)
	m.SetMusicVolume(0)
	m.Play(SoundSlam)
	m.SetSoundVolume(1)
	full := make([]float32, soundFrames(SoundSlam, testSampleRate)*2)
	m.Fill(full)
	if !hasSignal(full) {
		t.Fatalf("sound at full volume rendered silence")
	}

	quiet := New(testSampleRate)
	quiet.SetMusicVolume(0)
	quiet.Play(SoundSlam)
	quiet.SetSoundVolume(0)
	q := make([]float32, soundFrames(SoundSlam, testSampleRate)*2)
	quiet.Fill(q)
	if hasSignal(q) {
		t.Fatalf("sound at zero volume still produced signal")
	}
}

func TestMixingClamps(t *testing.T) {
	for _, tr := range allTracks() {
		m := New(testSampleRate)
		m.PlayMusic(tr)
		for _, snd := range allSounds() {
			m.Play(snd)
		}
		buf := make([]float32, testSampleRate*2)
		m.Fill(buf)
		for i, v := range buf {
			if v < -1 || v > 1 || math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				t.Fatalf("track %v sample %d out of range: %v", tr, i, v)
			}
		}
	}
}

func TestFreshMixersRenderIdentically(t *testing.T) {
	for _, tr := range allTracks() {
		a := New(testSampleRate)
		b := New(testSampleRate)
		a.PlayMusic(tr)
		b.PlayMusic(tr)
		for _, snd := range allSounds() {
			a.Play(snd)
			b.Play(snd)
		}
		ba := make([]float32, 2*testSampleRate)
		bb := make([]float32, 2*testSampleRate)
		a.Fill(ba)
		b.Fill(bb)
		for i := range ba {
			if ba[i] != bb[i] {
				t.Fatalf("track %v sample %d differs: %v vs %v", tr, i, ba[i], bb[i])
			}
		}
	}
}

func TestTracksLoopWithoutSeamClick(t *testing.T) {
	for _, tr := range allTracks() {
		m := New(testSampleRate)
		m.PlayMusic(tr)
		loop := m.loops[tr]
		frames := len(loop) / 2
		firstL, firstR := loop[0], loop[1]
		lastL, lastR := loop[(frames-1)*2], loop[(frames-1)*2+1]
		if math.Abs(float64(firstL-lastL)) > 0.05 || math.Abs(float64(firstR-lastR)) > 0.05 {
			t.Errorf("track %v has a seam jump of %v/%v", tr, firstL-lastL, firstR-lastR)
		}
	}
}

func TestFillDoesNotAllocateAfterWarmup(t *testing.T) {
	m := New(testSampleRate)
	m.PlayMusic(TrackBoss)
	for _, snd := range []Sound{SoundJump, SoundSlam, SoundHairspray, SoundCheckpoint} {
		m.Play(snd)
	}
	m.Fill(make([]float32, testSampleRate*2)) // warm up: build the loop, drain effects

	buf := make([]float32, 512*2)
	allocs := testing.AllocsPerRun(20, func() {
		m.Fill(buf)
	})
	if allocs != 0 {
		t.Fatalf("Fill allocated %v times per call", allocs)
	}
}

func TestNoPanicOnUnusualInputs(t *testing.T) {
	m := New(0)
	m.PlayMusic(Track(-3))
	m.PlayMusic(Track(99))
	m.Play(Sound(-5))
	m.Play(Sound(99))
	m.SetMusicVolume(-1)
	m.SetMusicVolume(2)
	m.SetSoundVolume(-1)
	m.SetSoundVolume(2)
	m.Fill(nil)
	m.Fill(make([]float32, 0))
	m.Fill(make([]float32, 1))
	m.Fill(make([]float32, 3))

	buf := make([]float32, 64)
	for i := 0; i < 20000; i++ {
		m.Play(Sound(i % 8))
		m.Fill(buf)
	}
}
