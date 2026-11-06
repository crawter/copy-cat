package render

// CameraBehind is how far behind the player the camera sits, in world units.
const CameraBehind = 10.0

// CameraHeight is how far above the player the camera sits, in world units.
const CameraHeight = 6.0

// LookAhead is how far ahead of the player the camera aims, in world units.
const LookAhead = 20.0

// Vec3 is a plain three-component position or direction. It carries no
// raylib type so the camera math stays pure and testable headless.
type Vec3 struct {
	X float64
	Y float64
	Z float64
}

// CameraPlacement returns the camera's position, target and up vectors for
// the given player position. The camera sits CameraBehind behind and
// CameraHeight above the player, keeps the player's sideways X, and aims
// LookAhead ahead at the player's height. Up points straight up.
func CameraPlacement(px, py, pz float64) (position, target, up Vec3) {
	position = Vec3{X: px, Y: py + CameraHeight, Z: pz - CameraBehind}
	target = Vec3{X: px, Y: py, Z: pz + LookAhead}
	up = Vec3{X: 0, Y: 1, Z: 0}
	return position, target, up
}
