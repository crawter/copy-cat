// Package render turns the game's plain state into raylib draw commands.
// BuildWorld and BuildBossArena are pure Go and return a []Cmd whose every
// field is a plain value, so the world layout is testable headlessly;
// DrawCmds is the only function in the package that calls raylib.
package render

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Vec3 is a plain position or size in meters. Its fields are float32 so
// positions flow straight to and from raylib's Vector3.
type Vec3 struct {
	X, Y, Z float32
}

// Size is a plain size in meters. Which fields a shape reads is documented on
// the Shape constants.
type Size struct {
	X, Y, Z float64
}

// Color is a plain RGBA color with eight bits per channel. It mirrors
// raylib's color type so the build functions never need to import raylib.
type Color struct {
	R, G, B, A uint8
}

// Shape names one raylib primitive a command draws.
type Shape int

// The primitive shapes a command can draw.
const (
	// ShapeBox is a box centered at Pos, Size.X wide, Size.Y tall and Size.Z
	// deep.
	ShapeBox Shape = iota
	// ShapeCylinder is a vertical cylinder centered at Pos, of radius Size.X
	// and height Size.Y. Size.Z is ignored.
	ShapeCylinder
	// ShapeSphere is a sphere centered at Pos, of radius Size.X. Size.Y and
	// Size.Z are ignored.
	ShapeSphere
	// ShapeCone is a vertical cone centered at Pos, of radius Size.X and
	// height Size.Y. Size.Z is ignored.
	ShapeCone
	// ShapePlane is an XZ plane centered at Pos, Size.X wide and Size.Z deep.
	// Size.Y is ignored.
	ShapePlane
)

// Cmd is one draw command: a shape, where to draw it, how big it is and in
// which color. Every field is a plain value.
type Cmd struct {
	Shape Shape
	Pos   Vec3
	Size  Size
	Color Color
}

// Cam is the point the camera follows: the hero's lane index, its Z position
// along the corridor and its feet height. BuildWorld culls everything farther
// than a fixed distance behind and ahead of Cam.Z.
type Cam struct {
	Lane   float64
	Z      float64
	Height float64
}

// v3 builds a Vec3 from float64 coordinates. The corridor and arena builders
// keep their layout math in float64 and convert to the command's float32
// position only at the edge.
func v3(x, y, z float64) Vec3 {
	return Vec3{X: float32(x), Y: float32(y), Z: float32(z)}
}

// DrawCmds draws cs with raylib primitives in list order.
func DrawCmds(cs []Cmd) {
	for _, c := range cs {
		drawCmd(c)
	}
}

func drawCmd(c Cmd) {
	switch c.Shape {
	case ShapeBox:
		rl.DrawCube(rlVec3(c.Pos), f32(c.Size.X), f32(c.Size.Y), f32(c.Size.Z), rlColor(c.Color))
	case ShapeCylinder:
		base := rlVec3(Vec3{X: c.Pos.X, Y: c.Pos.Y - f32(c.Size.Y/2), Z: c.Pos.Z})
		rl.DrawCylinder(base, f32(c.Size.X), f32(c.Size.X), f32(c.Size.Y), 12, rlColor(c.Color))
	case ShapeSphere:
		rl.DrawSphere(rlVec3(c.Pos), f32(c.Size.X), rlColor(c.Color))
	case ShapeCone:
		base := rlVec3(Vec3{X: c.Pos.X, Y: c.Pos.Y - f32(c.Size.Y/2), Z: c.Pos.Z})
		rl.DrawCylinder(base, 0, f32(c.Size.X), f32(c.Size.Y), 12, rlColor(c.Color))
	case ShapePlane:
		rl.DrawPlane(rlVec3(c.Pos), rl.Vector2{X: f32(c.Size.X), Y: f32(c.Size.Z)}, rlColor(c.Color))
	}
}

func rlVec3(v Vec3) rl.Vector3 {
	return rl.Vector3{X: v.X, Y: v.Y, Z: v.Z}
}

func rlColor(c Color) rl.Color {
	return rl.Color{R: c.R, G: c.G, B: c.B, A: c.A}
}

func f32(v float64) float32 { return float32(v) }

// blendColor returns c blended toward to by amount, clamped to [0, 1]. An
// amount of 0 returns c unchanged; an amount of 1 returns to. Every channel,
// including alpha, blends linearly.
func blendColor(c, to Color, amount float64) Color {
	if amount <= 0 {
		return c
	}
	if amount >= 1 {
		return to
	}
	lerp := func(a, b uint8) uint8 {
		return uint8(float64(a) + (float64(b)-float64(a))*amount)
	}
	return Color{
		R: lerp(c.R, to.R),
		G: lerp(c.G, to.G),
		B: lerp(c.B, to.B),
		A: lerp(c.A, to.A),
	}
}
