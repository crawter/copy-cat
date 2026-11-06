package render

import "testing"

func TestPaletteHoldsThePDDColours(t *testing.T) {
	tests := []struct {
		name   string
		colour Color
	}{
		{name: "MurkGreen", colour: MurkGreen},
		{name: "BogBrown", colour: BogBrown},
		{name: "MossYellow", colour: MossYellow},
		{name: "WarningOrange", colour: WarningOrange},
		{name: "FogWhite", colour: FogWhite},
	}
	for _, tt := range tests {
		if tt.colour.A != 255 {
			t.Fatalf("%s has alpha %d, want full opacity 255", tt.name, tt.colour.A)
		}
	}
}

func TestCameraFollowsBehindPlayer(t *testing.T) {
	const px, py, pz = 1.0, 2.0, 3.0

	position, target, up := CameraPlacement(px, py, pz)

	if position.X != px {
		t.Errorf("position.X = %v, want player X %v", position.X, px)
	}
	if position.Y != py+CameraHeight {
		t.Errorf("position.Y = %v, want CameraHeight above player %v", position.Y, py+CameraHeight)
	}
	if position.Z != pz-CameraBehind {
		t.Errorf("position.Z = %v, want CameraBehind behind player %v", position.Z, pz-CameraBehind)
	}

	if target.X != px {
		t.Errorf("target.X = %v, want player X %v", target.X, px)
	}
	if target.Z != pz+LookAhead {
		t.Errorf("target.Z = %v, want LookAhead ahead of player %v", target.Z, pz+LookAhead)
	}

	wantUp := Vec3{X: 0, Y: 1, Z: 0}
	if up != wantUp {
		t.Errorf("up = %+v, want straight up %+v", up, wantUp)
	}
}
