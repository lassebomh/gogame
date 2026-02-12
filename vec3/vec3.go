package v3

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Value struct {
	X float64
	Y float64
	Z float64
}

var Zero = Value{}
var One = Value{1, 1, 1}

func X(x float64) Value {
	return Value{x, 0, 0}
}

func Y(y float64) Value {
	return Value{0, y, 0}
}

func Z(z float64) Value {
	return Value{0, 0, z}
}

func Fill(v float64) Value {
	return Value{v, v, v}
}

func XY(x, y float64) Value {
	return Value{x, y, 0}
}
func YZ(y, z float64) Value {
	return Value{0, y, z}
}
func XZ(x, z float64) Value {
	return Value{x, 0, z}
}
func XYZ(x, y, z float64) Value {
	return Value{x, y, z}
}
func (v Value) Map(fn func(v float64) float64) Value {
	return XYZ(fn(v.X), fn(v.Y), fn(v.Z))
}

func (v Value) Raylib() rl.Vector3 {
	return rl.Vector3{float32(v.X), float32(v.Y), float32(v.Z)}
}
func (v Value) Chipmunk() cp.Vector {
	return cp.Vector{v.X, v.Z}
}

func FromChipmunk(v cp.Vector) Value {
	return XYZ(v.X, 0, v.Y)
}

func FromRaylib(v rl.Vector3) Value {
	return Value{float64(v.X), float64(v.Y), float64(v.Z)}
}

// Color type, RGBA (32bit)
// TODO remove later, keep type for now to not break code
type Color = color.RGBA

// NewColor - Returns new Color
func NewColor(r, g, b, a uint8) color.RGBA {
	return color.RGBA{r, g, b, a}
}

// Clamp - Clamp float value
func Clamp(value, min, max float64) float64 {
	var res float64
	if value < min {
		res = min
	} else {
		res = value
	}

	if res > max {
		return max
	}

	return res
}

// Lerp - Calculate linear interpolation between two floats
func Lerp(start, end, amount float64) float64 {
	return start + amount*(end-start)
}

// Normalize - Normalize input value within input range
func Normalize(value, start, end float64) float64 {
	return (value - start) / (end - start)
}

// Remap - Remap input value within input range to output range
func Remap(value, inputStart, inputEnd, outputStart, outputEnd float64) float64 {
	return (value-inputStart)/(inputEnd-inputStart)*(outputEnd-outputStart) + outputStart
}

// Wrap - Wrap input value from min to max
func Wrap(value, min, max float64) float64 {
	return ((value) - (max-min)*math.Floor(((value-min)/(max-min))))
}

// FloatEquals - Check whether two given floats are almost equal
func FloatEquals(x, y float64) bool {
	return (math.Abs(float64(x-y)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(x)), math.Abs(float64(y)))))
}

// Vector3Zero - Vector with components value 0.0
func Vector3Zero() Value {
	return XYZ(0.0, 0.0, 0.0)
}

// Vector3One - Vector with components value 1.0
func Vector3One() Value {
	return XYZ(1.0, 1.0, 1.0)
}

// Vector3Add - Add two vectors
func (v Value) Add(v2 Value) Value {
	return XYZ(v.X+v2.X, v.Y+v2.Y, v.Z+v2.Z)
}

// Vector3Add - Add two vectors
func (v Value) AddXYZ(x, y, z float64) Value {
	return XYZ(v.X+x, v.Y+y, v.Z+z)
}

// Vector3Add - Add two vectors
func (v Value) SubtractXYZ(x, y, z float64) Value {
	return XYZ(v.X-x, v.Y-y, v.Z-z)
}

func (v Value) ToColor() color.RGBA {
	return color.RGBA{uint8(v.X * 255), uint8(v.Y * 255), uint8(v.Z * 255), 255}
}

// Vector3AddValue - Add vector and float value
func (v Value) AddValue(o float64) Value {
	return XYZ(v.X+o, v.Y+o, v.Z+o)
}

// Vector3Subtract - Subtract two vectors
func (v Value) Subtract(o Value) Value {
	return XYZ(v.X-o.X, v.Y-o.Y, v.Z-o.Z)
}

// Vector3SubtractValue - Subtract vector by float value
func (v Value) SubtractValue(sub float64) Value {
	return XYZ(v.X-sub, v.Y-sub, v.Z-sub)
}

// Vector3Scale - Scale provided vector
func (v Value) Scale(scale float64) Value {
	return XYZ(v.X*scale, v.Y*scale, v.Z*scale)
}

// Vector3Multiply - Multiply vector by vector
func (v Value) Multiply(v2 Value) Value {
	result := Value{}

	result.X = v.X * v2.X
	result.Y = v.Y * v2.Y
	result.Z = v.Z * v2.Z

	return result
}

// Vector3CrossProduct - Calculate two vectors cross product
func (v Value) CrossProduct(v2 Value) Value {
	result := Value{}

	result.X = v.Y*v2.Z - v.Z*v2.Y
	result.Y = v.Z*v2.X - v.X*v2.Z
	result.Z = v.X*v2.Y - v.Y*v2.X

	return result
}

// Vector3Perpendicular - Calculate one vector perpendicular vector
func (v Value) Perpendicular() Value {
	min := math.Abs(float64(v.X))
	cardinalAxis := XYZ(1.0, 0.0, 0.0)

	if math.Abs(float64(v.Y)) < min {
		min = math.Abs(float64(v.Y))
		cardinalAxis = XYZ(0.0, 1.0, 0.0)
	}

	if math.Abs(float64(v.Z)) < min {
		cardinalAxis = XYZ(0.0, 0.0, 1.0)
	}

	result := v.CrossProduct(cardinalAxis)

	return result
}

// Vector3Length - Calculate vector length
func (v Value) Length() float64 {
	return float64(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

// Vector3LengthSqr - Calculate vector square length
func (v Value) LengthSqr() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// Vector3DotProduct - Calculate two vectors dot product
func (v Value) DotProduct(v2 Value) float64 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

// Vector3Distance - Calculate distance between two vectors
func (v Value) Distance(v2 Value) float64 {
	dx := v2.X - v.X
	dy := v2.Y - v.Y
	dz := v2.Z - v.Z

	return float64(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Vector3DistanceSqr - Calculate square distance between two vectors
func (v Value) DistanceSqr(v2 Value) float64 {
	var result float64

	dx := v2.X - v.X
	dy := v2.Y - v.Y
	dz := v2.Z - v.Z
	result = dx*dx + dy*dy + dz*dz

	return result
}

// Vector3Angle - Calculate angle between two vectors
func (v Value) Angle(v2 Value) float64 {
	var result float64

	cross := Value{X: v.Y*v2.Z - v.Z*v2.Y, Y: v.Z*v2.X - v.X*v2.Z, Z: v.X*v2.Y - v.Y*v2.X}
	length := float64(math.Sqrt(float64(cross.X*cross.X + cross.Y*cross.Y + cross.Z*cross.Z)))
	dot := v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
	result = float64(math.Atan2(float64(length), float64(dot)))

	return result
}

// Vector3Negate - Negate provided vector (invert direction)
func (v Value) Negate() Value {
	return XYZ(-v.X, -v.Y, -v.Z)
}

// Vector3Negate - Negate provided vector (invert direction)
func (v Value) Abs() Value {
	return XYZ(math.Abs(v.X), math.Abs(v.Y), math.Abs(v.Z))
}

// Vector3Divide - Divide vector by vector
func (v Value) Divide(v2 Value) Value {
	return XYZ(v.X/v2.X, v.Y/v2.Y, v.Z/v2.Z)
}

// Vector3Normalize - Normalize provided vector
func (v Value) Normalize() Value {
	result := v

	var length, ilength float64

	length = v.Length()

	if length == 0 {
		length = 1.0
	}

	ilength = 1.0 / length

	result.X *= ilength
	result.Y *= ilength
	result.Z *= ilength

	return result
}

// Vector3Project - Calculate the projection of the vector v1 on to v2
func (v Value) Project(v2 Value) Value {
	result := Value{}

	v1dv2 := (v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z)
	v2dv2 := (v2.X*v2.X + v2.Y*v2.Y + v2.Z*v2.Z)

	mag := v1dv2 / v2dv2

	result.X = v2.X * mag
	result.Y = v2.Y * mag
	result.Z = v2.Z * mag

	return result
}

// Vector3Reject - Calculate the rejection of the vector v1 on to v2
func (v Value) Reject(v2 Value) Value {
	result := Value{}

	v1dv2 := (v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z)
	v2dv2 := (v2.X*v2.X + v2.Y*v2.Y + v2.Z*v2.Z)

	mag := v1dv2 / v2dv2

	result.X = v.X - (v2.X * mag)
	result.Y = v.Y - (v2.Y * mag)
	result.Z = v.Z - (v2.Z * mag)

	return result
}

// Vector3OrthoNormalize - Orthonormalize provided vectors
// Makes vectors normalized and orthogonal to each other
// Gram-Schmidt function implementation
func (v *Value) OrthoNormalize(v2 *Value) {
	(*v).Normalize()

	vn1 := (*v).CrossProduct(*v2).Normalize()
	vn2 := vn1.CrossProduct(*v)
	*v2 = vn2
}

// Vector3RotateByAxisAngle - Rotates a vector around an axis
func (v Value) RotateByAxisAngle(axis Value, angle float64) Value {
	// Using Euler-Rodrigues Formula
	// Ref.: https://en.wikipedia.org/w/index.php?title=Euler%E2%80%93Rodrigues_formula

	result := v

	// Vector3Normalize(axis);
	length := float64(math.Sqrt(float64(axis.X*axis.X + axis.Y*axis.Y + axis.Z*axis.Z)))
	if length == 0.0 {
		length = 1.0
	}
	ilength := 1.0 / length
	axis.X *= ilength
	axis.Y *= ilength
	axis.Z *= ilength

	angle /= 2.0
	a := float64(math.Sin(float64(angle)))
	b := axis.X * a
	c := axis.Y * a
	d := axis.Z * a
	a = float64(math.Cos(float64(angle)))
	w := XYZ(b, c, d)

	// Vector3CrossProduct(w, v)
	wv := XYZ(w.Y*v.Z-w.Z*v.Y, w.Z*v.X-w.X*v.Z, w.X*v.Y-w.Y*v.X)

	// Vector3CrossProduct(w, wv)
	wwv := XYZ(w.Y*wv.Z-w.Z*wv.Y, w.Z*wv.X-w.X*wv.Z, w.X*wv.Y-w.Y*wv.X)

	// Vector3Scale(wv, 2*a)
	a *= 2
	wv.X *= a
	wv.Y *= a
	wv.Z *= a

	// Vector3Scale(wwv, 2)
	wwv.X *= 2
	wwv.Y *= 2
	wwv.Z *= 2

	result.X += wv.X
	result.Y += wv.Y
	result.Z += wv.Z

	result.X += wwv.X
	result.Y += wwv.Y
	result.Z += wwv.Z

	return result
}

// Vector3Lerp - Calculate linear interpolation between two vectors
func (v Value) Lerp(v2 Value, amount float64) Value {
	result := Value{}

	result.X = v.X + amount*(v2.X-v.X)
	result.Y = v.Y + amount*(v2.Y-v.Y)
	result.Z = v.Z + amount*(v2.Z-v.Z)

	return result
}

// Vector3Reflect - Calculate reflected vector to normal
func (v Value) Reflect(normal Value) Value {
	// I is the original vector
	// N is the normal of the incident plane
	// R = I - (2*N*( DotProduct[ I,N] ))

	result := Value{}

	dotProduct := v.DotProduct(normal)

	result.X = v.X - (2.0*normal.X)*dotProduct
	result.Y = v.Y - (2.0*normal.Y)*dotProduct
	result.Z = v.Z - (2.0*normal.Z)*dotProduct

	return result
}

// Vector3Min - Return min value for each pair of components
func (v Value) Min(vec2 Value) Value {
	result := Value{}

	result.X = float64(math.Min(float64(v.X), float64(vec2.X)))
	result.Y = float64(math.Min(float64(v.Y), float64(vec2.Y)))
	result.Z = float64(math.Min(float64(v.Z), float64(vec2.Z)))

	return result
}

// Vector3Max - Return max value for each pair of components
func (v Value) Max(vec2 Value) Value {
	result := Value{}

	result.X = float64(math.Max(float64(v.X), float64(vec2.X)))
	result.Y = float64(math.Max(float64(v.Y), float64(vec2.Y)))
	result.Z = float64(math.Max(float64(v.Z), float64(vec2.Z)))

	return result
}

// Vector3Barycenter - Barycenter coords for p in triangle abc
func (v Value) Barycenter(a, b, c Value) Value {
	v0 := b.Subtract(a)
	v1 := c.Subtract(a)
	v2 := v.Subtract(a)
	d00 := v0.DotProduct(v0)
	d01 := v0.DotProduct(v1)
	d11 := v1.DotProduct(v1)
	d20 := v2.DotProduct(v0)
	d21 := v2.DotProduct(v1)

	denom := d00*d11 - d01*d01

	result := Value{}

	result.Y = (d11*d20 - d01*d21) / denom
	result.Z = (d00*d21 - d01*d20) / denom
	result.X = 1.0 - (result.Z + result.Y)

	return result
}

// Vector3ToFloatV - Get Vector3 as float array
func (v Value) ToFloatV() [3]float64 {
	var result [3]float64

	result[0] = v.X
	result[1] = v.Y
	result[2] = v.Z

	return result
}

// Vector3Invert - Invert the given vector
func (v Value) Invert() Value {
	return XYZ(1.0/v.X, 1.0/v.Y, 1.0/v.Z)
}

// Vector3Clamp - Clamp the components of the vector between min and max values specified by the given vectors
func (v Value) Clamp(min Value, max Value) Value {
	var result = Value{}

	result.X = float64(math.Min(float64(max.X), math.Max(float64(min.X), float64(v.X))))
	result.Y = float64(math.Min(float64(max.Y), math.Max(float64(min.Y), float64(v.Y))))
	result.Z = float64(math.Min(float64(max.Z), math.Max(float64(min.Z), float64(v.Z))))

	return result
}

// Vector3ClampValue - Clamp the magnitude of the vector between two values
func (v Value) ClampValue(min float64, max float64) Value {
	var result = v

	length := v.X*v.X + v.Y*v.Y + v.Z*v.Z
	if length > 0.0 {
		length = float64(math.Sqrt(float64(length)))

		if length < min {
			scale := min / length
			result.X = v.X * scale
			result.Y = v.Y * scale
			result.Z = v.Z * scale
		} else if length > max {
			scale := max / length
			result.X = v.X * scale
			result.Y = v.Y * scale
			result.Z = v.Z * scale
		}
	}

	return result
}

// Vector3Equals - Check whether two given vectors are almost equal
func (v Value) Equals(q Value) bool {
	return (math.Abs(float64(v.X-q.X)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(v.X)), math.Abs(float64(q.X)))) &&
		math.Abs(float64(v.Y-q.Y)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(v.Y)), math.Abs(float64(q.Y)))) &&
		math.Abs(float64(v.Z-q.Z)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(v.Z)), math.Abs(float64(q.Z)))))
}

// Vector3Refract - Compute the direction of a refracted ray
//
// v: normalized direction of the incoming ray
// n: normalized normal vector of the interface of two optical media
// r: ratio of the refractive index of the medium from where the ray comes to the refractive index of the medium on the other side of the surface
func (v Value) Refract(n Value, r float64) Value {
	var result = Value{}

	dot := v.X*n.X + v.Y*n.Y + v.Z*n.Z
	d := 1.0 - r*r*(1.0-dot*dot)

	if d >= 0.0 {
		d = float64(math.Sqrt(float64(d)))
		v.X = r*v.X - (r*dot+d)*n.X
		v.Y = r*v.Y - (r*dot+d)*n.Y
		v.Z = r*v.Z - (r*dot+d)*n.Z

		result = v
	}

	return result
}

// Vector3Transform - Transforms a Vector3 by a given Matrix
func (v Value) Transform(mat Matrix) Value {
	result := Value{}

	x := v.X
	y := v.Y
	z := v.Z

	result.X = mat.M0*x + mat.M4*y + mat.M8*z + mat.M12
	result.Y = mat.M1*x + mat.M5*y + mat.M9*z + mat.M13
	result.Z = mat.M2*x + mat.M6*y + mat.M10*z + mat.M14

	return result
}

const Pi = float64(rl.Pi)

// Matrix type (OpenGL style 4x4 - right handed, column major)
type Matrix struct {
	M0, M4, M8, M12  float64
	M1, M5, M9, M13  float64
	M2, M6, M10, M14 float64
	M3, M7, M11, M15 float64
}

// NewMatrixManual - Returns new Matrix
func NewMatrixManual(m0, m4, m8, m12, m1, m5, m9, m13, m2, m6, m10, m14, m3, m7, m11, m15 float64) Matrix {
	return Matrix{m0, m4, m8, m12, m1, m5, m9, m13, m2, m6, m10, m14, m3, m7, m11, m15}
}

// MatrixDeterminant - Compute matrix determinant
func (mat Matrix) Determinant() float64 {
	var result float64

	a00 := mat.M0
	a01 := mat.M1
	a02 := mat.M2
	a03 := mat.M3
	a10 := mat.M4
	a11 := mat.M5
	a12 := mat.M6
	a13 := mat.M7
	a20 := mat.M8
	a21 := mat.M9
	a22 := mat.M10
	a23 := mat.M11
	a30 := mat.M12
	a31 := mat.M13
	a32 := mat.M14
	a33 := mat.M15

	result = a30*a21*a12*a03 - a20*a31*a12*a03 - a30*a11*a22*a03 + a10*a31*a22*a03 +
		a20*a11*a32*a03 - a10*a21*a32*a03 - a30*a21*a02*a13 + a20*a31*a02*a13 +
		a30*a01*a22*a13 - a00*a31*a22*a13 - a20*a01*a32*a13 + a00*a21*a32*a13 +
		a30*a11*a02*a23 - a10*a31*a02*a23 - a30*a01*a12*a23 + a00*a31*a12*a23 +
		a10*a01*a32*a23 - a00*a11*a32*a23 - a20*a11*a02*a33 + a10*a21*a02*a33 +
		a20*a01*a12*a33 - a00*a21*a12*a33 - a10*a01*a22*a33 + a00*a11*a22*a33

	return result
}

// MatrixTrace - Returns the trace of the matrix (sum of the values along the diagonal)
func (mat Matrix) Trace() float64 {
	return mat.M0 + mat.M5 + mat.M10 + mat.M15
}

// MatrixTranspose - Transposes provided matrix
func (mat Matrix) Transpose() Matrix {
	var result Matrix

	result.M0 = mat.M0
	result.M1 = mat.M4
	result.M2 = mat.M8
	result.M3 = mat.M12
	result.M4 = mat.M1
	result.M5 = mat.M5
	result.M6 = mat.M9
	result.M7 = mat.M13
	result.M8 = mat.M2
	result.M9 = mat.M6
	result.M10 = mat.M10
	result.M11 = mat.M14
	result.M12 = mat.M3
	result.M13 = mat.M7
	result.M14 = mat.M11
	result.M15 = mat.M15

	return result
}

// MatrixInvert - Invert provided matrix
func (mat Matrix) Invert() Matrix {
	var result Matrix

	a00 := mat.M0
	a01 := mat.M1
	a02 := mat.M2
	a03 := mat.M3
	a10 := mat.M4
	a11 := mat.M5
	a12 := mat.M6
	a13 := mat.M7
	a20 := mat.M8
	a21 := mat.M9
	a22 := mat.M10
	a23 := mat.M11
	a30 := mat.M12
	a31 := mat.M13
	a32 := mat.M14
	a33 := mat.M15

	b00 := a00*a11 - a01*a10
	b01 := a00*a12 - a02*a10
	b02 := a00*a13 - a03*a10
	b03 := a01*a12 - a02*a11
	b04 := a01*a13 - a03*a11
	b05 := a02*a13 - a03*a12
	b06 := a20*a31 - a21*a30
	b07 := a20*a32 - a22*a30
	b08 := a20*a33 - a23*a30
	b09 := a21*a32 - a22*a31
	b10 := a21*a33 - a23*a31
	b11 := a22*a33 - a23*a32

	// Calculate the invert determinant (inlined to avoid double-caching)
	invDet := 1.0 / (b00*b11 - b01*b10 + b02*b09 + b03*b08 - b04*b07 + b05*b06)

	result.M0 = (a11*b11 - a12*b10 + a13*b09) * invDet
	result.M1 = (-a01*b11 + a02*b10 - a03*b09) * invDet
	result.M2 = (a31*b05 - a32*b04 + a33*b03) * invDet
	result.M3 = (-a21*b05 + a22*b04 - a23*b03) * invDet
	result.M4 = (-a10*b11 + a12*b08 - a13*b07) * invDet
	result.M5 = (a00*b11 - a02*b08 + a03*b07) * invDet
	result.M6 = (-a30*b05 + a32*b02 - a33*b01) * invDet
	result.M7 = (a20*b05 - a22*b02 + a23*b01) * invDet
	result.M8 = (a10*b10 - a11*b08 + a13*b06) * invDet
	result.M9 = (-a00*b10 + a01*b08 - a03*b06) * invDet
	result.M10 = (a30*b04 - a31*b02 + a33*b00) * invDet
	result.M11 = (-a20*b04 + a21*b02 - a23*b00) * invDet
	result.M12 = (-a10*b09 + a11*b07 - a12*b06) * invDet
	result.M13 = (a00*b09 - a01*b07 + a02*b06) * invDet
	result.M14 = (-a30*b03 + a31*b01 - a32*b00) * invDet
	result.M15 = (a20*b03 - a21*b01 + a22*b00) * invDet

	return result
}

// NewMatrix - Returns identity matrix
func NewMatrix() Matrix {
	return NewMatrixManual(
		1.0, 0.0, 0.0, 0.0,
		0.0, 1.0, 0.0, 0.0,
		0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 1.0)
}

func (mat Matrix) Raylib() rl.Matrix {
	return rl.Matrix{
		M0: float32(mat.M0), M4: float32(mat.M4), M8: float32(mat.M8), M12: float32(mat.M12),
		M1: float32(mat.M1), M5: float32(mat.M5), M9: float32(mat.M9), M13: float32(mat.M13),
		M2: float32(mat.M2), M6: float32(mat.M6), M10: float32(mat.M10), M14: float32(mat.M14),
		M3: float32(mat.M3), M7: float32(mat.M7), M11: float32(mat.M11), M15: float32(mat.M15),
	}
}

// MatrixNormalize - Normalize provided matrix
func (mat Matrix) Normalize() Matrix {
	var result Matrix

	det := mat.Determinant()

	result.M0 /= det
	result.M1 /= det
	result.M2 /= det
	result.M3 /= det
	result.M4 /= det
	result.M5 /= det
	result.M6 /= det
	result.M7 /= det
	result.M8 /= det
	result.M9 /= det
	result.M10 /= det
	result.M11 /= det
	result.M12 /= det
	result.M13 /= det
	result.M14 /= det
	result.M15 /= det

	return result
}

// MatrixAdd - Add two matrices
func (left Matrix) Add(right Matrix) Matrix {
	result := NewMatrix()

	result.M0 = left.M0 + right.M0
	result.M1 = left.M1 + right.M1
	result.M2 = left.M2 + right.M2
	result.M3 = left.M3 + right.M3
	result.M4 = left.M4 + right.M4
	result.M5 = left.M5 + right.M5
	result.M6 = left.M6 + right.M6
	result.M7 = left.M7 + right.M7
	result.M8 = left.M8 + right.M8
	result.M9 = left.M9 + right.M9
	result.M10 = left.M10 + right.M10
	result.M11 = left.M11 + right.M11
	result.M12 = left.M12 + right.M12
	result.M13 = left.M13 + right.M13
	result.M14 = left.M14 + right.M14
	result.M15 = left.M15 + right.M15

	return result
}

// MatrixSubtract - Subtract two matrices (left - right)
func (left Matrix) Subtract(right Matrix) Matrix {
	result := NewMatrix()

	result.M0 = left.M0 - right.M0
	result.M1 = left.M1 - right.M1
	result.M2 = left.M2 - right.M2
	result.M3 = left.M3 - right.M3
	result.M4 = left.M4 - right.M4
	result.M5 = left.M5 - right.M5
	result.M6 = left.M6 - right.M6
	result.M7 = left.M7 - right.M7
	result.M8 = left.M8 - right.M8
	result.M9 = left.M9 - right.M9
	result.M10 = left.M10 - right.M10
	result.M11 = left.M11 - right.M11
	result.M12 = left.M12 - right.M12
	result.M13 = left.M13 - right.M13
	result.M14 = left.M14 - right.M14
	result.M15 = left.M15 - right.M15

	return result
}

// MatrixMultiply - Returns two matrix multiplication
func (left Matrix) Multiply(right Matrix) Matrix {
	var result Matrix

	result.M0 = left.M0*right.M0 + left.M1*right.M4 + left.M2*right.M8 + left.M3*right.M12
	result.M1 = left.M0*right.M1 + left.M1*right.M5 + left.M2*right.M9 + left.M3*right.M13
	result.M2 = left.M0*right.M2 + left.M1*right.M6 + left.M2*right.M10 + left.M3*right.M14
	result.M3 = left.M0*right.M3 + left.M1*right.M7 + left.M2*right.M11 + left.M3*right.M15
	result.M4 = left.M4*right.M0 + left.M5*right.M4 + left.M6*right.M8 + left.M7*right.M12
	result.M5 = left.M4*right.M1 + left.M5*right.M5 + left.M6*right.M9 + left.M7*right.M13
	result.M6 = left.M4*right.M2 + left.M5*right.M6 + left.M6*right.M10 + left.M7*right.M14
	result.M7 = left.M4*right.M3 + left.M5*right.M7 + left.M6*right.M11 + left.M7*right.M15
	result.M8 = left.M8*right.M0 + left.M9*right.M4 + left.M10*right.M8 + left.M11*right.M12
	result.M9 = left.M8*right.M1 + left.M9*right.M5 + left.M10*right.M9 + left.M11*right.M13
	result.M10 = left.M8*right.M2 + left.M9*right.M6 + left.M10*right.M10 + left.M11*right.M14
	result.M11 = left.M8*right.M3 + left.M9*right.M7 + left.M10*right.M11 + left.M11*right.M15
	result.M12 = left.M12*right.M0 + left.M13*right.M4 + left.M14*right.M8 + left.M15*right.M12
	result.M13 = left.M12*right.M1 + left.M13*right.M5 + left.M14*right.M9 + left.M15*right.M13
	result.M14 = left.M12*right.M2 + left.M13*right.M6 + left.M14*right.M10 + left.M15*right.M14
	result.M15 = left.M12*right.M3 + left.M13*right.M7 + left.M14*right.M11 + left.M15*right.M15

	return result
}

// MatrixTranslate - Returns translation matrix
func MatrixTranslateXYZ(x, y, z float64) Matrix {
	return NewMatrixManual(
		1.0, 0.0, 0.0, x,
		0.0, 1.0, 0.0, y,
		0.0, 0.0, 1.0, z,
		0, 0, 0, 1.0)
}
func MatrixTranslate(v Value) Matrix {
	return NewMatrixManual(
		1.0, 0.0, 0.0, v.X,
		0.0, 1.0, 0.0, v.Y,
		0.0, 0.0, 1.0, v.Z,
		0, 0, 0, 1.0)
}

func (m Matrix) Translate(v Value) Matrix {
	return m.Multiply(MatrixTranslate(v))
}

func (m Matrix) TranslateXYZ(x, y, z float64) Matrix {
	return m.Multiply(MatrixTranslateXYZ(x, y, z))
}

func (m Matrix) Rotate(axis Value, angle float64) Matrix {
	return m.Multiply(MatrixRotate(axis, angle))
}

// MatrixRotate - Returns rotation matrix for an angle around an specified axis (angle in radians)
func MatrixRotate(axis Value, angle float64) Matrix {
	var result Matrix

	mat := NewMatrix()

	x := axis.X
	y := axis.Y
	z := axis.Z

	length := float64(math.Sqrt(float64(x*x + y*y + z*z)))

	if length != 1.0 && length != 0.0 {
		length = 1.0 / length
		x *= length
		y *= length
		z *= length
	}

	sinres := float64(math.Sin(float64(angle)))
	cosres := float64(math.Cos(float64(angle)))
	t := 1.0 - cosres

	// Cache some matrix values (speed optimization)
	a00 := mat.M0
	a01 := mat.M1
	a02 := mat.M2
	a03 := mat.M3
	a10 := mat.M4
	a11 := mat.M5
	a12 := mat.M6
	a13 := mat.M7
	a20 := mat.M8
	a21 := mat.M9
	a22 := mat.M10
	a23 := mat.M11

	// Construct the elements of the rotation matrix
	b00 := x*x*t + cosres
	b01 := y*x*t + z*sinres
	b02 := z*x*t - y*sinres
	b10 := x*y*t - z*sinres
	b11 := y*y*t + cosres
	b12 := z*y*t + x*sinres
	b20 := x*z*t + y*sinres
	b21 := y*z*t - x*sinres
	b22 := z*z*t + cosres

	// Perform rotation-specific matrix multiplication
	result.M0 = a00*b00 + a10*b01 + a20*b02
	result.M1 = a01*b00 + a11*b01 + a21*b02
	result.M2 = a02*b00 + a12*b01 + a22*b02
	result.M3 = a03*b00 + a13*b01 + a23*b02
	result.M4 = a00*b10 + a10*b11 + a20*b12
	result.M5 = a01*b10 + a11*b11 + a21*b12
	result.M6 = a02*b10 + a12*b11 + a22*b12
	result.M7 = a03*b10 + a13*b11 + a23*b12
	result.M8 = a00*b20 + a10*b21 + a20*b22
	result.M9 = a01*b20 + a11*b21 + a21*b22
	result.M10 = a02*b20 + a12*b21 + a22*b22
	result.M11 = a03*b20 + a13*b21 + a23*b22
	result.M12 = mat.M12
	result.M13 = mat.M13
	result.M14 = mat.M14
	result.M15 = mat.M15

	return result
}

func (m Matrix) RotateX(angle float64) Matrix {
	return m.Multiply(MatrixRotateX(angle))
}

// MatrixRotateX - Returns x-rotation matrix (angle in radians)
func MatrixRotateX(angle float64) Matrix {
	result := NewMatrix()

	cosres := float64(math.Cos(float64(angle)))
	sinres := float64(math.Sin(float64(angle)))

	result.M5 = cosres
	result.M6 = -sinres
	result.M9 = sinres
	result.M10 = cosres

	return result
}

func (m Matrix) RotateY(angle float64) Matrix {
	return m.Multiply(MatrixRotateY(angle))
}

// MatrixRotateY - Returns y-rotation matrix (angle in radians)
func MatrixRotateY(angle float64) Matrix {
	result := NewMatrix()

	cosres := float64(math.Cos(float64(angle)))
	sinres := float64(math.Sin(float64(angle)))

	result.M0 = cosres
	result.M2 = sinres
	result.M8 = -sinres
	result.M10 = cosres

	return result
}

func (m Matrix) RotateZ(angle float64) Matrix {
	return m.Multiply(MatrixRotateZ(angle))
}

// MatrixRotateZ - Returns z-rotation matrix (angle in radians)
func MatrixRotateZ(angle float64) Matrix {
	result := NewMatrix()

	cosres := float64(math.Cos(float64(angle)))
	sinres := float64(math.Sin(float64(angle)))

	result.M0 = cosres
	result.M1 = -sinres
	result.M4 = sinres
	result.M5 = cosres

	return result
}

func (m Matrix) RotateXYZ(ang Value) Matrix {
	return m.Multiply(MatrixRotateXYZ(ang))
}

// MatrixRotateXYZ - Get xyz-rotation matrix (angles in radians)
func MatrixRotateXYZ(ang Value) Matrix {
	result := NewMatrix()

	cosz := float64(math.Cos(float64(-ang.Z)))
	sinz := float64(math.Sin(float64(-ang.Z)))
	cosy := float64(math.Cos(float64(-ang.Y)))
	siny := float64(math.Sin(float64(-ang.Y)))
	cosx := float64(math.Cos(float64(-ang.X)))
	sinx := float64(math.Sin(float64(-ang.X)))

	result.M0 = cosz * cosy
	result.M4 = (cosz * siny * sinx) - (sinz * cosx)
	result.M8 = (cosz * siny * cosx) + (sinz * sinx)

	result.M1 = sinz * cosy
	result.M5 = (sinz * siny * sinx) + (cosz * cosx)
	result.M9 = (sinz * siny * cosx) - (cosz * sinx)

	result.M2 = -siny
	result.M6 = cosy * sinx
	result.M10 = cosy * cosx

	return result
}

func (m Matrix) RotateZYX(angle Value) Matrix {
	return m.Multiply(MatrixRotateZYX(angle))
}

// MatrixRotateZYX - Get zyx-rotation matrix
// NOTE: Angle must be provided in radians
func MatrixRotateZYX(angle Value) Matrix {
	var result = Matrix{}

	var cz = float64(math.Cos(float64(angle.Z)))
	var sz = float64(math.Sin(float64(angle.Z)))
	var cy = float64(math.Cos(float64(angle.Y)))
	var sy = float64(math.Sin(float64(angle.Y)))
	var cx = float64(math.Cos(float64(angle.X)))
	var sx = float64(math.Sin(float64(angle.X)))

	result.M0 = cz * cy
	result.M4 = cz*sy*sx - cx*sz
	result.M8 = sz*sx + cz*cx*sy
	result.M12 = float64(0)

	result.M1 = cy * sz
	result.M5 = cz*cx + sz*sy*sx
	result.M9 = cx*sz*sy - cz*sx
	result.M13 = float64(0)

	result.M2 = -sy
	result.M6 = cy * sx
	result.M10 = cy * cx
	result.M14 = float64(0)

	result.M3 = float64(0)
	result.M7 = float64(0)
	result.M11 = float64(0)
	result.M15 = float64(1)

	return result
}

func (m Matrix) Scale(x, y, z float64) Matrix {
	return m.Multiply(MatrixScale(x, y, z))
}

// MatrixScale - Returns scaling matrix
func MatrixScale(x, y, z float64) Matrix {
	result := NewMatrixManual(
		x, 0.0, 0.0, 0.0,
		0.0, y, 0.0, 0.0,
		0.0, 0.0, z, 0.0,
		0.0, 0.0, 0.0, 1.0)

	return result
}

// MatrixFrustum - Returns perspective projection matrix
func MatrixFrustum(left, right, bottom, top, near, far float64) Matrix {
	var result Matrix

	rl := right - left
	tb := top - bottom
	fn := far - near

	result.M0 = (near * 2.0) / rl
	result.M1 = 0.0
	result.M2 = 0.0
	result.M3 = 0.0

	result.M4 = 0.0
	result.M5 = (near * 2.0) / tb
	result.M6 = 0.0
	result.M7 = 0.0

	result.M8 = right + left/rl
	result.M9 = top + bottom/tb
	result.M10 = -(far + near) / fn
	result.M11 = -1.0

	result.M12 = 0.0
	result.M13 = 0.0
	result.M14 = -(far * near * 2.0) / fn
	result.M15 = 0.0

	return result
}

// MatrixPerspective - Returns perspective projection matrix
func MatrixPerspective(fovy, aspect, near, far float64) Matrix {
	top := near * float64(math.Tan(float64(fovy*Pi)/360.0))
	right := top * aspect

	return MatrixFrustum(-right, right, -top, top, near, far)
}

// MatrixOrtho - Returns orthographic projection matrix
func MatrixOrtho(left, right, bottom, top, near, far float64) Matrix {
	var result Matrix

	rl := right - left
	tb := top - bottom
	fn := far - near

	result.M0 = 2.0 / rl
	result.M1 = 0.0
	result.M2 = 0.0
	result.M3 = 0.0
	result.M4 = 0.0
	result.M5 = 2.0 / tb
	result.M6 = 0.0
	result.M7 = 0.0
	result.M8 = 0.0
	result.M9 = 0.0
	result.M10 = -2.0 / fn
	result.M11 = 0.0
	result.M12 = -(left + right) / rl
	result.M13 = -(top + bottom) / tb
	result.M14 = -(far + near) / fn
	result.M15 = 1.0

	return result
}

// MatrixLookAt - Returns camera look-at matrix (view matrix)
func MatrixLookAt(eye, target, up Value) Matrix {
	var result Matrix

	z := eye.Subtract(target)
	z = z.Normalize()
	x := up.CrossProduct(z)
	x = x.Normalize()
	y := z.CrossProduct(x)
	y = y.Normalize()

	result.M0 = x.X
	result.M1 = x.Y
	result.M2 = x.Z
	result.M3 = -((x.X * eye.X) + (x.Y * eye.Y) + (x.Z * eye.Z))
	result.M4 = y.X
	result.M5 = y.Y
	result.M6 = y.Z
	result.M7 = -((y.X * eye.X) + (y.Y * eye.Y) + (y.Z * eye.Z))
	result.M8 = z.X
	result.M9 = z.Y
	result.M10 = z.Z
	result.M11 = -((z.X * eye.X) + (z.Y * eye.Y) + (z.Z * eye.Z))
	result.M12 = 0.0
	result.M13 = 0.0
	result.M14 = 0.0
	result.M15 = 1.0

	return result
}

// MatrixToFloatV - Get float array of matrix data
func (mat Matrix) ToFloatV() [16]float64 {
	var result [16]float64

	result[0] = mat.M0
	result[1] = mat.M1
	result[2] = mat.M2
	result[3] = mat.M3
	result[4] = mat.M4
	result[5] = mat.M5
	result[6] = mat.M6
	result[7] = mat.M7
	result[8] = mat.M8
	result[9] = mat.M9
	result[10] = mat.M10
	result[11] = mat.M11
	result[12] = mat.M12
	result[13] = mat.M13
	result[14] = mat.M14
	result[15] = mat.M15

	return result
}
