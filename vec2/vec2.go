package vec2

import (
	"game/vec3"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Value cp.Vector

var Zero = Value{}

func X(x float64) Value {
	return Value{X: x}
}
func Y(y float64) Value {
	return Value{Y: y}
}
func XY(x, y float64) Value {
	return Value{X: x, Y: y}
}
func All(v float64) Value {
	return Value{X: v, Y: v}
}

func (v Value) Chipmunk() cp.Vector {
	return cp.Vector{v.X, v.Y}
}
func (v Value) To3D() vec3.Value {
	return vec3.Value{v.X, 0, v.Y}
}

func FromRaylib(v rl.Vector2) Value {
	return Value{float64(v.X), float64(v.Y)}
}
func FromChipmunk(v cp.Vector) Value {
	return Value{v.X, v.Y}
}
func FromRadians(radians float64) Value {
	return Value{
		math.Cos(radians),
		math.Sin(radians),
	}
}

func (v Value) Raylib() rl.Vector2 {
	return rl.Vector2{float32(v.X), float32(v.Y)}
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

// Vector2Add - Add two vectors (v1 + v2)
func (v1 Value) Add(v2 Value) Value {
	return XY(v1.X+v2.X, v1.Y+v2.Y)
}

// Vector2AddValue - Add vector and float value
func (v Value) AddValue(add float64) Value {
	return XY(v.X+add, v.Y+add)
}

// Vector2Subtract - Subtract two vectors (v1 - v2)
func (v1 Value) Subtract(v2 Value) Value {
	return XY(v1.X-v2.X, v1.Y-v2.Y)
}

// Vector2Subtract - Subtract two vectors (v1 - v2)
func (v1 Value) SubtractXY(x, y float64) Value {
	return XY(v1.X-x, v1.Y-y)
}

// Vector2SubtractValue - Subtract vector by float value
func (v Value) SubtractValue(sub float64) Value {
	return XY(v.X-sub, v.Y-sub)
}

// Vector2Length - Calculate vector length
func (v Value) Length() float64 {
	return float64(math.Sqrt(float64((v.X * v.X) + (v.Y * v.Y))))
}

// Vector2LengthSqr - Calculate vector square length
func (v Value) LengthSqr() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Vector2DotProduct - Calculate two vectors dot product
func (v1 Value) DotProduct(v2 Value) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

// Vector2Distance - Calculate distance between two vectors
func (v1 Value) Distance(v2 Value) float64 {
	return float64(math.Sqrt(float64((v1.X-v2.X)*(v1.X-v2.X) + (v1.Y-v2.Y)*(v1.Y-v2.Y))))
}

// Vector2DistanceSqr - Calculate square distance between two vectors
func (v1 Value) DistanceSqr(v2 Value) float64 {
	return (v1.X-v2.X)*(v1.X-v2.X) + (v1.Y-v2.Y)*(v1.Y-v2.Y)
}

// Vector2Angle - Calculate angle from two vectors in radians
func (v1 Value) Angle(v2 Value) float64 {
	result := math.Atan2(float64(v2.Y), float64(v2.X)) - math.Atan2(float64(v1.Y), float64(v1.X))

	return float64(result)
}

// Vector2LineAngle - Calculate angle defined by a two vectors line
// NOTE: Parameters need to be normalized. Current implementation should be aligned with glm::angle
func (start Value) LineAngle(end Value) float64 {
	return float64(-math.Atan2(float64(end.Y-start.Y), float64(end.X-start.X)))
}

// Vector2Scale - Scale vector (multiply by value)
func (v Value) Scale(scale float64) Value {
	return XY(v.X*scale, v.Y*scale)
}

// Vector2Multiply - Multiply vector by vector
func (v1 Value) Multiply(v2 Value) Value {
	return XY(v1.X*v2.X, v1.Y*v2.Y)
}

// Vector2Negate - Negate vector
func (v Value) Negate() Value {
	return XY(-v.X, -v.Y)
}

// Vector2Divide - Divide vector by vector
func (v1 Value) Divide(v2 Value) Value {
	return XY(v1.X/v2.X, v1.Y/v2.Y)
}

// Vector2Normalize - Normalize provided vector
func (v Value) Normalize() Value {
	if l := (v.Length()); l > 0 {
		return v.Scale(1 / l)
	}
	return v
}

// Vector2Lerp - Calculate linear interpolation between two vectors
func (v1 Value) Lerp(v2 Value, amount float64) Value {
	return XY(v1.X+amount*(v2.X-v1.X), v1.Y+amount*(v2.Y-v1.Y))
}

// Vector2Reflect - Calculate reflected vector to normal
func (v Value) Reflect(normal Value) Value {
	var result = Value{}

	dotProduct := v.X*normal.X + v.Y*normal.Y // Dot product

	result.X = v.X - 2.0*normal.X*dotProduct
	result.Y = v.Y - 2.0*normal.Y*dotProduct

	return result
}

// Vector2Rotate - Rotate vector by angle
func (v Value) Rotate(angle float64) Value {
	var result = Value{}

	cosres := float64(math.Cos(float64(angle)))
	sinres := float64(math.Sin(float64(angle)))

	result.X = v.X*cosres - v.Y*sinres
	result.Y = v.X*sinres + v.Y*cosres

	return result
}

// Vector2MoveTowards - Move Vector towards target
func (v Value) MoveTowards(target Value, maxDistance float64) Value {
	var result = Value{}

	dx := target.X - v.X
	dy := target.Y - v.Y
	value := dx*dx + dy*dy

	if value == 0 || maxDistance >= 0 && value <= maxDistance*maxDistance {
		return target
	}

	dist := float64(math.Sqrt(float64(value)))

	result.X = v.X + dx/dist*maxDistance
	result.Y = v.Y + dy/dist*maxDistance

	return result
}

// Vector2Invert - Invert the given vector
func (v Value) Invert() Value {
	return XY(1.0/v.X, 1.0/v.Y)
}

// Vector2Clamp - Clamp the components of the vector between min and max values specified by the given vectors
func (v Value) Clamp(min Value, max Value) Value {
	var result = Value{}

	result.X = float64(math.Min(float64(max.X), math.Max(float64(min.X), float64(v.X))))
	result.Y = float64(math.Min(float64(max.Y), math.Max(float64(min.Y), float64(v.Y))))

	return result
}

// Vector2ClampValue - Clamp the magnitude of the vector between two min and max values
func (v Value) ClampValue(min float64, max float64) Value {
	var result = v

	length := v.X*v.X + v.Y*v.Y
	if length > 0.0 {
		length = float64(math.Sqrt(float64(length)))

		if length < min {
			scale := min / length
			result.X = v.X * scale
			result.Y = v.Y * scale
		} else if length > max {
			scale := max / length
			result.X = v.X * scale
			result.Y = v.Y * scale
		}
	}

	return result
}

// Vector2Equals - Check whether two given vectors are almost equal
func (p Value) Equals(q Value) bool {
	return (math.Abs(float64(p.X-q.X)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(p.X)), math.Abs(float64(q.X)))) &&
		math.Abs(float64(p.Y-q.Y)) <= 0.000001*math.Max(1.0, math.Max(math.Abs(float64(p.Y)), math.Abs(float64(q.Y)))))
}

// Vector2CrossProduct - Calculate two vectors cross product
func (v1 Value) CrossProduct(v2 Value) float64 {
	return v1.X*v2.Y - v1.Y*v2.X
}

// Vector2Cross - Calculate the cross product of a vector and a value
func Cross(value float64, vector Value) Value {
	return XY(-value*vector.Y, value*vector.X)
}

// Vector2LenSqr - Returns the len square root of a vector
func (vector Value) LenSqr() float64 {
	return vector.X*vector.X + vector.Y*vector.Y
}
