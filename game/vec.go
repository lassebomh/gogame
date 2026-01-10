package game

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Vec struct{ rl.Vector3 }

func NewVec(x, y, z float32) Vec {
	return Vec{rl.Vector3{X: x, Y: y, Z: z}}
}
func VecFrom2D(v cp.Vector, y float64) Vec {
	return Vec{rl.Vector3{X: float32(v.X), Y: float32(y), Z: float32(v.Y)}}
}
func VecFromColor(color color.RGBA) Vec {
	return Vec{rl.Vector3{X: float32(color.R), Y: float32(color.G), Z: float32(color.B)}}
}
func (v Vec) ToRGBA() color.Color {
	return color.RGBA{uint8(v.X), uint8(v.Y), uint8(v.Z), 255}
}

func (v Vec) Add(o Vec) Vec {
	return Vec{rl.Vector3Add(v.Vector3, o.Vector3)}
}
func (v Vec) AddValue(add float32) Vec {
	return Vec{rl.Vector3AddValue(v.Vector3, add)}
}
func (v Vec) Subtract(o Vec) Vec {
	return Vec{rl.Vector3Subtract(v.Vector3, o.Vector3)}
}
func (v Vec) SubtractValue(sub float32) Vec {
	return Vec{rl.Vector3SubtractValue(v.Vector3, sub)}
}
func (v Vec) Scale(scale float32) Vec {
	return Vec{rl.Vector3Scale(v.Vector3, scale)}
}
func (v Vec) Multiply(o Vec) Vec {
	return Vec{rl.Vector3Multiply(v.Vector3, o.Vector3)}
}
func (v Vec) CrossProduct(o Vec) Vec {
	return Vec{rl.Vector3CrossProduct(v.Vector3, o.Vector3)}
}
func (v Vec) Perpendicular() Vec {
	return Vec{rl.Vector3Perpendicular(v.Vector3)}
}
func (v Vec) Length() float32 {
	return rl.Vector3Length(v.Vector3)
}
func (v Vec) LengthSqr() float32 {
	return rl.Vector3LengthSqr(v.Vector3)
}
func (v Vec) DotProduct(o Vec) float32 {
	return rl.Vector3DotProduct(v.Vector3, o.Vector3)
}
func (v Vec) Distance(o Vec) float32 {
	return rl.Vector3Distance(v.Vector3, o.Vector3)
}
func (v Vec) DistanceSqr(o Vec) float32 {
	return rl.Vector3DistanceSqr(v.Vector3, o.Vector3)
}
func (v Vec) Angle(o Vec) float32 {
	return rl.Vector3Angle(v.Vector3, o.Vector3)
}
func (v Vec) Negate() Vec {
	return Vec{rl.Vector3Negate(v.Vector3)}
}
func (v Vec) Divide(o Vec) Vec {
	return Vec{rl.Vector3Divide(v.Vector3, o.Vector3)}
}
func (v Vec) Normalize() Vec {
	return Vec{rl.Vector3Normalize(v.Vector3)}
}
func (v Vec) Project(o Vec) Vec {
	return Vec{rl.Vector3Project(v.Vector3, o.Vector3)}
}
func (v Vec) Reject(o Vec) Vec {
	return Vec{rl.Vector3Reject(v.Vector3, o.Vector3)}
}
func (v *Vec) OrthoNormalize(o *Vec) {
	rl.Vector3OrthoNormalize(&v.Vector3, &o.Vector3)
}
func (v Vec) Transform(mat rl.Matrix) Vec {
	return Vec{rl.Vector3Transform(v.Vector3, mat)}
}
func (v Vec) RotateByQuaternion(q rl.Quaternion) Vec {
	return Vec{rl.Vector3RotateByQuaternion(v.Vector3, q)}
}
func (v Vec) RotateByAxisAngle(axis Vec, angle float32) Vec {
	return Vec{rl.Vector3RotateByAxisAngle(v.Vector3, axis.Vector3, angle)}
}
func (v Vec) Lerp(o Vec, amount float32) Vec {
	return Vec{rl.Vector3Lerp(v.Vector3, o.Vector3, amount)}
}
func (v Vec) Reflect(normal Vec) Vec {
	return Vec{rl.Vector3Reflect(v.Vector3, normal.Vector3)}
}
func (v Vec) Min(o Vec) Vec {
	return Vec{rl.Vector3Min(v.Vector3, o.Vector3)}
}
func (v Vec) Max(o Vec) Vec {
	return Vec{rl.Vector3Max(v.Vector3, o.Vector3)}
}
func (v Vec) Barycenter(a, b, c Vec) Vec {
	return Vec{rl.Vector3Barycenter(v.Vector3, a.Vector3, b.Vector3, c.Vector3)}
}
func (v Vec) Unproject(projection rl.Matrix, view rl.Matrix) Vec {
	return Vec{rl.Vector3Unproject(v.Vector3, projection, view)}
}
func (v Vec) ToFloatV() [3]float32 {
	return rl.Vector3ToFloatV(v.Vector3)
}
func (v Vec) Invert() Vec {
	return Vec{rl.Vector3Invert(v.Vector3)}
}
func (v Vec) Clamp(min Vec, max Vec) Vec {
	return Vec{rl.Vector3Clamp(v.Vector3, min.Vector3, max.Vector3)}
}
func (v Vec) ClampValue(min float32, max float32) Vec {
	return Vec{rl.Vector3ClampValue(v.Vector3, min, max)}
}
func (v Vec) Equals(o Vec) bool {
	return rl.Vector3Equals(v.Vector3, o.Vector3)
}
func (v Vec) Refract(n Vec, r float32) Vec {
	return Vec{rl.Vector3Refract(v.Vector3, n.Vector3, r)}
}

var X Vec = NewVec(1, 0, 0)
var Y Vec = NewVec(0, 1, 0)
var Z Vec = NewVec(0, 0, 1)
var XY Vec = NewVec(1, 1, 0)
var XZ Vec = NewVec(1, 0, 1)
var YZ Vec = NewVec(0, 1, 1)
var XYZ Vec = NewVec(1, 1, 1)
