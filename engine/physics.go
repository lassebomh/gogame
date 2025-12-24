package engine

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Body[T Shape] struct {
	Position rl.Vector2
	Angle    float32
	Shape    T
}

type Contact struct {
	A, B        rl.Vector2
	Normal      rl.Vector2
	Penetration float32
}

type Shape interface{}

type Circle struct {
	Radius float32
}

type Box struct {
	Width  float32
	Height float32
}

func (b *Body[T]) Render() {

	// length of the direction indicator
	const dirLen float32 = 20

	// compute direction vector once
	dir := rl.Vector2{
		X: float32(math.Cos(float64(b.Angle))),
		Y: float32(math.Sin(float64(b.Angle))),
	}

	dirEnd := rl.Vector2{
		X: b.Position.X + dir.X*dirLen,
		Y: b.Position.Y + dir.Y*dirLen,
	}

	switch c := any(b.Shape).(type) {

	case Circle:

		rl.DrawCircleLines(
			int32(b.Position.X),
			int32(b.Position.Y),
			c.Radius,
			rl.Green,
		)

		rl.DrawLineV(b.Position, dirEnd, rl.Red)

	case Box:
		hw := c.Width / 2
		hh := c.Height / 2

		corners := [4]rl.Vector2{
			{X: -hw, Y: -hh},
			{X: hw, Y: -hh},
			{X: hw, Y: hh},
			{X: -hw, Y: hh},
		}

		cos := float32(math.Cos(float64(b.Angle)))
		sin := float32(math.Sin(float64(b.Angle)))

		for i := range corners {
			x := corners[i].X
			y := corners[i].Y

			corners[i].X = b.Position.X + x*cos - y*sin
			corners[i].Y = b.Position.Y + x*sin + y*cos
		}

		for i := 0; i < 4; i++ {
			a := corners[i]
			b := corners[(i+1)%4]
			rl.DrawLineV(a, b, rl.Green)
		}

		rl.DrawLineV(b.Position, dirEnd, rl.Red)
	}
}

func CircleVsCircle(a, b *Body[Circle]) (bool, Contact) {
	delta := rl.Vector2Subtract(b.Position, a.Position)
	distSq := delta.X*delta.X + delta.Y*delta.Y
	radiusSum := a.Shape.Radius + b.Shape.Radius

	if distSq <= radiusSum*radiusSum {
		dist := float32(math.Sqrt(float64(distSq)))
		normal := rl.Vector2{X: 0, Y: 0}

		if dist != 0 {
			normal = rl.Vector2Scale(delta, 1/dist)
		}

		penetration := radiusSum - dist

		// Contact points on each circle's surface
		contactA := rl.Vector2Add(a.Position, rl.Vector2Scale(normal, a.Shape.Radius))
		contactB := rl.Vector2Add(b.Position, rl.Vector2Scale(normal, -b.Shape.Radius))

		return true, Contact{
			A:           contactA,
			B:           contactB,
			Normal:      normal,
			Penetration: penetration,
		}
	}

	return false, Contact{}
}

func CircleVsBox(circle *Body[Circle], box *Body[Box]) (bool, Contact) {
	// Transform circle center into box's local space
	delta := rl.Vector2Subtract(circle.Position, box.Position)

	// Rotate delta by -box.Angle to get local coordinates
	cos := float32(math.Cos(float64(-box.Angle)))
	sin := float32(math.Sin(float64(-box.Angle)))
	localX := delta.X*cos - delta.Y*sin
	localY := delta.X*sin + delta.Y*cos

	// Clamp circle center to box edges (find closest point on box)
	halfW := box.Shape.Width / 2
	halfH := box.Shape.Height / 2
	closestX := clamp(localX, -halfW, halfW)
	closestY := clamp(localY, -halfH, halfH)

	// Check if circle center is inside box
	inside := false
	if localX == closestX && localY == closestY {
		inside = true
		// Find closest edge
		distLeft := localX + halfW
		distRight := halfW - localX
		distTop := localY + halfH
		distBottom := halfH - localY

		minDist := min(distLeft, distRight, distTop, distBottom)
		switch minDist {
		case distLeft:
			closestX = -halfW
		case distRight:
			closestX = halfW
		case distTop:
			closestY = -halfH
		default:
			closestY = halfH
		}
	}

	// Transform closest point back to world space
	// Minimal fix: rotate back by +box.Angle
	cosA := float32(math.Cos(float64(box.Angle)))
	sinA := float32(math.Sin(float64(box.Angle)))
	worldClosestX := closestX*cosA - closestY*sinA
	worldClosestY := closestX*sinA + closestY*cosA
	closestWorld := rl.Vector2Add(box.Position, rl.Vector2{X: worldClosestX, Y: worldClosestY})

	// Vector from closest point to circle center
	normal := rl.Vector2Subtract(circle.Position, closestWorld)
	distSq := normal.X*normal.X + normal.Y*normal.Y
	radius := circle.Shape.Radius

	if distSq > radius*radius && !inside {
		return false, Contact{}
	}

	dist := float32(math.Sqrt(float64(distSq)))

	if inside {
		normal = rl.Vector2Scale(normal, -1)
		dist = -dist
	}

	if dist != 0 {
		normal = rl.Vector2Scale(normal, 1/dist)
	} else {
		// Circle center exactly on box surface, use fallback normal
		normal = rl.Vector2{X: 0, Y: -1}
	}

	penetration := radius - dist

	// Contact points
	contactA := rl.Vector2Add(circle.Position, rl.Vector2Scale(normal, -radius))
	contactB := closestWorld

	return true, Contact{
		A:           contactA,
		B:           contactB,
		Normal:      normal,
		Penetration: penetration,
	}
}

func clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func min(vals ...float32) float32 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
