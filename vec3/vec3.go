package vec3

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Value struct {
	X float64
	Y float64
	Z float64
}

var Zero = Value{}

func X(x float64) Value {
	return Value{x, 0, 0}
}

func Y(y float64) Value {
	return Value{0, y, 0}
}

func Z(z float64) Value {
	return Value{0, 0, z}
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
func (v1 Value) Add(v2 Value) Value {
	return XYZ(v1.X+v2.X, v1.Y+v2.Y, v1.Z+v2.Z)
}

// Vector3Add - Add two vectors
func (v1 Value) AddXYZ(x, y, z float64) Value {
	return XYZ(v1.X+x, v1.Y+y, v1.Z+z)
}

// Vector3Add - Add two vectors
func (v1 Value) SubtractXYZ(x, y, z float64) Value {
	return XYZ(v1.X-x, v1.Y-y, v1.Z-z)
}

func (v Value) ToColor() color.RGBA {
	return color.RGBA{uint8(v.X * 255), uint8(v.Y * 255), uint8(v.Z * 255), 255}
}

// Vector3AddValue - Add vector and float value
func (v Value) AddValue(add float64) Value {
	return XYZ(v.X+add, v.Y+add, v.Z+add)
}

// Vector3Subtract - Subtract two vectors
func (v1 Value) Subtract(v2 Value) Value {
	return XYZ(v1.X-v2.X, v1.Y-v2.Y, v1.Z-v2.Z)
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
func (v1 Value) Multiply(v2 Value) Value {
	result := Value{}

	result.X = v1.X * v2.X
	result.Y = v1.Y * v2.Y
	result.Z = v1.Z * v2.Z

	return result
}

// Vector3CrossProduct - Calculate two vectors cross product
func (v1 Value) CrossProduct(v2 Value) Value {
	result := Value{}

	result.X = v1.Y*v2.Z - v1.Z*v2.Y
	result.Y = v1.Z*v2.X - v1.X*v2.Z
	result.Z = v1.X*v2.Y - v1.Y*v2.X

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
func (v1 Value) DotProduct(v2 Value) float64 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

// Vector3Distance - Calculate distance between two vectors
func (v1 Value) Distance(v2 Value) float64 {
	dx := v2.X - v1.X
	dy := v2.Y - v1.Y
	dz := v2.Z - v1.Z

	return float64(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Vector3DistanceSqr - Calculate square distance between two vectors
func (v1 Value) DistanceSqr(v2 Value) float64 {
	var result float64

	dx := v2.X - v1.X
	dy := v2.Y - v1.Y
	dz := v2.Z - v1.Z
	result = dx*dx + dy*dy + dz*dz

	return result
}

// Vector3Angle - Calculate angle between two vectors
func (v1 Value) Angle(v2 Value) float64 {
	var result float64

	cross := Value{X: v1.Y*v2.Z - v1.Z*v2.Y, Y: v1.Z*v2.X - v1.X*v2.Z, Z: v1.X*v2.Y - v1.Y*v2.X}
	length := float64(math.Sqrt(float64(cross.X*cross.X + cross.Y*cross.Y + cross.Z*cross.Z)))
	dot := v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
	result = float64(math.Atan2(float64(length), float64(dot)))

	return result
}

// Vector3Negate - Negate provided vector (invert direction)
func (v Value) Negate() Value {
	return XYZ(-v.X, -v.Y, -v.Z)
}

// Vector3Divide - Divide vector by vector
func (v1 Value) Divide(v2 Value) Value {
	return XYZ(v1.X/v2.X, v1.Y/v2.Y, v1.Z/v2.Z)
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
func (v1 Value) Project(v2 Value) Value {
	result := Value{}

	v1dv2 := (v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z)
	v2dv2 := (v2.X*v2.X + v2.Y*v2.Y + v2.Z*v2.Z)

	mag := v1dv2 / v2dv2

	result.X = v2.X * mag
	result.Y = v2.Y * mag
	result.Z = v2.Z * mag

	return result
}

// Vector3Reject - Calculate the rejection of the vector v1 on to v2
func (v1 Value) Reject(v2 Value) Value {
	result := Value{}

	v1dv2 := (v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z)
	v2dv2 := (v2.X*v2.X + v2.Y*v2.Y + v2.Z*v2.Z)

	mag := v1dv2 / v2dv2

	result.X = v1.X - (v2.X * mag)
	result.Y = v1.Y - (v2.Y * mag)
	result.Z = v1.Z - (v2.Z * mag)

	return result
}

// Vector3OrthoNormalize - Orthonormalize provided vectors
// Makes vectors normalized and orthogonal to each other
// Gram-Schmidt function implementation
func (v1 *Value) OrthoNormalize(v2 *Value) {
	(*v1).Normalize()

	vn1 := (*v1).CrossProduct(*v2).Normalize()
	vn2 := vn1.CrossProduct(*v1)
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
func (v1 Value) Lerp(v2 Value, amount float64) Value {
	result := Value{}

	result.X = v1.X + amount*(v2.X-v1.X)
	result.Y = v1.Y + amount*(v2.Y-v1.Y)
	result.Z = v1.Z + amount*(v2.Z-v1.Z)

	return result
}

// Vector3Reflect - Calculate reflected vector to normal
func (vector Value) Reflect(normal Value) Value {
	// I is the original vector
	// N is the normal of the incident plane
	// R = I - (2*N*( DotProduct[ I,N] ))

	result := Value{}

	dotProduct := vector.DotProduct(normal)

	result.X = vector.X - (2.0*normal.X)*dotProduct
	result.Y = vector.Y - (2.0*normal.Y)*dotProduct
	result.Z = vector.Z - (2.0*normal.Z)*dotProduct

	return result
}

// Vector3Min - Return min value for each pair of components
func (vec1 Value) Min(vec2 Value) Value {
	result := Value{}

	result.X = float64(math.Min(float64(vec1.X), float64(vec2.X)))
	result.Y = float64(math.Min(float64(vec1.Y), float64(vec2.Y)))
	result.Z = float64(math.Min(float64(vec1.Z), float64(vec2.Z)))

	return result
}

// Vector3Max - Return max value for each pair of components
func (vec1 Value) Max(vec2 Value) Value {
	result := Value{}

	result.X = float64(math.Max(float64(vec1.X), float64(vec2.X)))
	result.Y = float64(math.Max(float64(vec1.Y), float64(vec2.Y)))
	result.Z = float64(math.Max(float64(vec1.Z), float64(vec2.Z)))

	return result
}

// Vector3Barycenter - Barycenter coords for p in triangle abc
func (p Value) Barycenter(a, b, c Value) Value {
	v0 := b.Subtract(a)
	v1 := c.Subtract(a)
	v2 := p.Subtract(a)
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
func (p Value) Equals(q Value) bool {
	return (math.Abs(float64(p.X-q.X)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(p.X)), math.Abs(float64(q.X)))) &&
		math.Abs(float64(p.Y-q.Y)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(p.Y)), math.Abs(float64(q.Y)))) &&
		math.Abs(float64(p.Z-q.Z)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(p.Z)), math.Abs(float64(q.Z)))))
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
