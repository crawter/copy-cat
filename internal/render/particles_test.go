package render

import (
	"reflect"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/gameplay"
)

// near reports whether a is within r meters of b.
func near(a, b Vec3, r float64) bool {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	dz := float64(a.Z - b.Z)
	return dx*dx+dy*dy+dz*dz <= r*r
}

func TestSpawnParticlesReturnNonEmptyNearPos(t *testing.T) {
	pos := Vec3{X: 2, Y: 1, Z: 30}
	cases := []struct {
		name  string
		spawn func(Vec3) []Particle
	}{
		{"sparkle", spawnSparkle},
		{"dust", spawnDust},
		{"hair", spawnHair},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ps := tc.spawn(pos)
			if len(ps) == 0 {
				t.Fatal("spawn returned no particles")
			}
			for i, p := range ps {
				if p.Life <= 0 {
					t.Fatalf("particle %d Life = %v, want > 0", i, p.Life)
				}
				if p.TTL <= 0 {
					t.Fatalf("particle %d TTL = %v, want > 0", i, p.TTL)
				}
				if p.Life > p.TTL {
					t.Fatalf("particle %d Life %v > TTL %v", i, p.Life, p.TTL)
				}
				if !near(pos, p.Pos, 2.0) {
					t.Fatalf("particle %d Pos %+v not near spawn pos %+v", i, p.Pos, pos)
				}
			}
		})
	}
}

func TestSpawnHairUsesHairColor(t *testing.T) {
	for i, p := range spawnHair(Vec3{}) {
		if p.Color != heroHair {
			t.Fatalf("hair particle %d color = %+v, want heroHair %+v", i, p.Color, heroHair)
		}
	}
}

func TestStepParticlesMovesAndAges(t *testing.T) {
	ps := []Particle{{
		Pos:  Vec3{X: 0, Y: 1, Z: 2},
		Vel:  Vec3{X: 2, Y: -1, Z: 3},
		Life: 1,
		TTL:  1,
		Size: 0.1,
	}}
	got := stepParticles(ps, 0.25)
	if len(got) != 1 {
		t.Fatalf("stepParticles kept %d particles, want 1", len(got))
	}
	want := Vec3{X: 0.5, Y: 0.75, Z: 2.75}
	if got[0].Pos != want {
		t.Fatalf("Pos = %+v, want %+v", got[0].Pos, want)
	}
	if wantLife := 0.75; got[0].Life != wantLife {
		t.Fatalf("Life = %v, want %v", got[0].Life, wantLife)
	}
}

func TestStepParticlesRemovesDead(t *testing.T) {
	ps := []Particle{
		{Pos: Vec3{X: 0}, Vel: Vec3{X: 1}, Life: 0.1, TTL: 0.5, Size: 0.1},
		{Pos: Vec3{X: 1}, Vel: Vec3{X: 1}, Life: 1, TTL: 1, Size: 0.1},
		{Pos: Vec3{X: 2}, Vel: Vec3{X: 1}, Life: 0, TTL: 0.5, Size: 0.1},
	}
	got := stepParticles(ps, 0.5)
	if len(got) != 1 {
		t.Fatalf("stepParticles kept %d particles, want 1", len(got))
	}
	if got[0].Pos.X != 1.5 {
		t.Fatalf("survivor Pos.X = %v, want 1.5", got[0].Pos.X)
	}
}

func TestParticlesDeterministic(t *testing.T) {
	pos := Vec3{X: 0, Y: 1, Z: 10}
	cases := []struct {
		name  string
		spawn func(Vec3) []Particle
	}{
		{"sparkle", spawnSparkle},
		{"dust", spawnDust},
		{"hair", spawnHair},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.spawn(pos)
			b := tc.spawn(pos)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("identical spawns differ:\n%+v\n%+v", a, b)
			}
			a = stepParticles(a, 0.1)
			b = stepParticles(b, 0.1)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("identical steps differ:\n%+v\n%+v", a, b)
			}
		})
	}
}

// particleUI returns a UI whose caption snapshot matches a fresh endless run,
// so each updateParticles transition test starts from a real gameplay.Run.
func particleUI(t *testing.T) (*UI, *campaign.Flow) {
	t.Helper()
	f := campaign.NewFlow(campaign.Default(), campaign.DefaultSettings(), campaign.NewProgress(), false)
	f.StartEndless(7)
	if f.Run() == nil {
		t.Fatal("endless run did not start")
	}
	r := f.Run()
	u := &UI{caption: captionSnap{
		run:            r,
		prevAirborne:   r.Runner.Airborne(),
		prevTotalCombs: r.TotalCombs,
		bald:           r.Bald,
	}}
	return u, f
}

// runHeroPos returns the hero's world position the spawn triggers use.
func runHeroPos(r *gameplay.Run) Vec3 {
	return Vec3{X: f32(laneX(r.Runner.Lane())), Y: f32(r.Runner.Height()), Z: f32(r.Runner.Z())}
}

func TestUpdateParticlesSpawnsOnTransitions(t *testing.T) {
	t.Run("comb sparkle", func(t *testing.T) {
		u, f := particleUI(t)
		r := f.Run()
		r.TotalCombs++
		u.updateParticles(f, 0)
		want := spawnSparkle(runHeroPos(r))
		if !reflect.DeepEqual(u.particles, want) {
			t.Fatalf("particles = %+v, want sparkle %+v", u.particles, want)
		}
	})
	t.Run("landing dust", func(t *testing.T) {
		u, f := particleUI(t)
		r := f.Run()
		u.caption.prevAirborne = true
		u.updateParticles(f, 0)
		want := spawnDust(runHeroPos(r))
		if !reflect.DeepEqual(u.particles, want) {
			t.Fatalf("particles = %+v, want dust %+v", u.particles, want)
		}
	})
	t.Run("snipped hair", func(t *testing.T) {
		u, f := particleUI(t)
		r := f.Run()
		r.Bald = true
		u.updateParticles(f, 0)
		want := spawnHair(runHeroPos(r))
		if !reflect.DeepEqual(u.particles, want) {
			t.Fatalf("particles = %+v, want hair %+v", u.particles, want)
		}
	})
}
