package render

import (
	"math"
	"reflect"
	"testing"

	"copycat/internal/campaign"
	"copycat/internal/gameplay"
)

func TestRunnerPoseMapsMoves(t *testing.T) {
	cases := []struct {
		name     string
		slamming bool
		sliding  bool
		airborne bool
		want     HeroPose
	}{
		{name: "run", want: PoseRun},
		{name: "jump", airborne: true, want: PoseJump},
		{name: "slide", sliding: true, want: PoseSlide},
		{name: "slam", slamming: true, airborne: true, want: PoseSlam},
	}
	for _, tc := range cases {
		if got := RunnerPose(tc.slamming, tc.sliding, tc.airborne); got != tc.want {
			t.Fatalf("%s: pose = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestBossPoseMapsMoves(t *testing.T) {
	cases := []struct {
		name    string
		jumping bool
		sliding bool
		want    HeroPose
	}{
		{name: "run", want: PoseRun},
		{name: "jump", jumping: true, want: PoseJump},
		{name: "slide", sliding: true, want: PoseSlide},
	}
	for _, tc := range cases {
		if got := BossPose(tc.jumping, tc.sliding); got != tc.want {
			t.Fatalf("%s: pose = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestApplyFogBlendsTowardBackdrop(t *testing.T) {
	pal := Palette(campaign.ThemeMangroves)
	cam := Cam{Lane: 1, Z: 10, Height: 0}
	from := Color{R: 255, G: 0, B: 0, A: 255}
	near := Cmd{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 1, Z: 10}, Color: from}
	far := Cmd{Shape: ShapeBox, Pos: Vec3{X: 0, Y: 1, Z: 100}, Color: from}

	out := ApplyFog([]Cmd{near, far}, cam, pal)
	if out[0].Color != from {
		t.Fatalf("near color changed: got %v want %v", out[0].Color, from)
	}
	if out[1].Color != pal.Backdrop {
		t.Fatalf("far color = %v, want backdrop %v", out[1].Color, pal.Backdrop)
	}
}

func TestApplyFogBlendsPartially(t *testing.T) {
	pal := Palette(campaign.ThemeMangroves)
	midZ := float32((fogNear + fogFar) / 2)
	from := Color{R: 0, G: 100, B: 200, A: 255}
	out := ApplyFog([]Cmd{{Shape: ShapeBox, Pos: Vec3{Z: midZ}, Color: from}}, Cam{Z: 0}, pal)
	if got := out[0].Color; got == from || got == pal.Backdrop {
		t.Fatalf("mid-distance color = %v, want a blend between %v and %v", got, from, pal.Backdrop)
	}
}

func TestApplyFogDoesNotMutateInput(t *testing.T) {
	pal := Palette(campaign.ThemeMangroves)
	cs := []Cmd{{Shape: ShapeBox, Pos: Vec3{Z: 50}, Color: Color{R: 255}}}
	orig := append([]Cmd(nil), cs...)
	ApplyFog(cs, Cam{Z: 0}, pal)
	if !reflect.DeepEqual(cs, orig) {
		t.Fatal("ApplyFog mutated its input slice")
	}
}

func TestFadeAlphaFadesOut(t *testing.T) {
	if got := fadeAlpha(0, sceneFadeDuration); got != 1 {
		t.Fatalf("fadeAlpha(0) = %v, want 1", got)
	}
	if got := fadeAlpha(sceneFadeDuration/2, sceneFadeDuration); math.Abs(got-0.5) > 1e-9 {
		t.Fatalf("fadeAlpha(mid) = %v, want 0.5", got)
	}
	if got := fadeAlpha(sceneFadeDuration, sceneFadeDuration); got != 0 {
		t.Fatalf("fadeAlpha(duration) = %v, want 0", got)
	}
	if got := fadeAlpha(sceneFadeDuration*2, sceneFadeDuration); got != 0 {
		t.Fatalf("fadeAlpha(after duration) = %v, want 0", got)
	}
	if got := fadeAlpha(1, 0); got != 0 {
		t.Fatalf("fadeAlpha with zero duration = %v, want 0", got)
	}
}

func TestCopycatDistanceGrowsWithSpeedShrinksWithChase(t *testing.T) {
	slow := CopycatDistance(8, 0.5)
	fast := CopycatDistance(16, 0.5)
	if fast <= slow {
		t.Fatalf("faster run does not push Copycat back: slow=%v fast=%v", slow, fast)
	}

	far := CopycatDistance(8, 0)
	close := CopycatDistance(8, 1)
	if close >= far {
		t.Fatalf("higher chase does not bring Copycat closer: far=%v close=%v", far, close)
	}

	if got := CopycatDistance(0, 1); got < copycatMinGap {
		t.Fatalf("distance = %v, below minimum %v", got, copycatMinGap)
	}
}

func TestBuildArenaTelegraphHighlightsAttackLane(t *testing.T) {
	low := BuildArenaTelegraph(1, gameplay.AttackLow)
	if len(low) == 0 {
		t.Fatal("no telegraph commands")
	}
	high := BuildArenaTelegraph(1, gameplay.AttackHigh)
	if reflect.DeepEqual(low, high) {
		t.Fatal("low and high telegraphs are identical")
	}
	for _, c := range low {
		if c.Shape != ShapePlane {
			t.Fatalf("telegraph uses non-plane shape %v", c.Shape)
		}
		if math.Abs(float64(c.Pos.X)-laneX(1)) > 0.01 {
			t.Fatalf("telegraph lane x = %v, want %v", c.Pos.X, laneX(1))
		}
	}

	left := BuildArenaTelegraph(0, gameplay.AttackLow)
	if left[0].Pos.X == low[0].Pos.X {
		t.Fatal("different lanes share the same telegraph position")
	}
}
