package engine

// import (
// 	"math"

// 	rl "github.com/gen2brain/raylib-go/raylib"
// )

// // ---------------------------
// // Body and Shapes
// // ---------------------------

// type Body struct {
// 	Position     rl.Vector2
// 	PrevPosition rl.Vector2
// 	Velocity     rl.Vector2
// 	Acceleration rl.Vector2
// 	Mass         float32 // 0 = static
// 	Shape        Shape
// }

// type Shape interface {
// 	// now just a marker interface; collisions handled externally
// }

// type Circle struct {
// 	Radius float32
// }

// type AABB struct {
// 	Min, Max rl.Vector2
// }

// // ---------------------------
// // Contact
// // ---------------------------

// type Contact struct {
// 	A, B        *Body
// 	Normal      rl.Vector2
// 	Penetration float32
// }

// // ---------------------------
// // World
// // ---------------------------

// type Constraint interface {
// 	Solve(dt float32)
// }

// type World struct {
// 	Bodies      []*Body
// 	Constraints []Constraint
// }

// // Add/Remove bodies and constraints
// func (w *World) AddBody(body *Body) *Body {
// 	w.Bodies = append(w.Bodies, body)
// 	return body
// }

// func (w *World) AddConstraint(c Constraint) {
// 	w.Constraints = append(w.Constraints, c)
// }

// // ---------------------------
// // Step function
// // ---------------------------

// func (w *World) Step(dt float32, iterations int) {
// 	// 1. Integrate positions
// 	for _, b := range w.Bodies {
// 		if b.Mass == 0 {
// 			continue
// 		}
// 		b.Velocity = rl.Vector2Add(b.Velocity, rl.Vector2Scale(b.Acceleration, dt))
// 		b.PrevPosition = b.Position
// 		b.Position = rl.Vector2Add(b.Position, rl.Vector2Scale(b.Velocity, dt))
// 	}

// 	// 2. Solve constraints multiple times
// 	for i := 0; i < iterations; i++ {
// 		for _, c := range w.Constraints {
// 			c.Solve(dt)
// 		}
// 	}

// 	// 3. Update velocities
// 	for _, b := range w.Bodies {
// 		if b.Mass == 0 {
// 			continue
// 		}
// 		b.Velocity = rl.Vector2Scale(
// 			rl.Vector2Subtract(b.Position, b.PrevPosition),
// 			1/dt,
// 		)
// 	}
// }

// // ---------------------------
// // Collision Functions
// // ---------------------------

// // Circle vs AABB
// func CircleVsAABB(circle, aabb *Body) (bool, Contact) {
// 	cPos := circle.Position
// 	radius := circle.Shape.(Circle).Radius
// 	min := rl.Vector2Add(aabb.Position, aabb.Shape.(AABB).Min)
// 	max := rl.Vector2Add(aabb.Position, aabb.Shape.(AABB).Max)

// 	closestX := clamp(cPos.X, min.X, max.X)
// 	closestY := clamp(cPos.Y, min.Y, max.Y)
// 	closest := rl.Vector2{X: closestX, Y: closestY}

// 	delta := rl.Vector2Subtract(cPos, closest)
// 	distSq := delta.X*delta.X + delta.Y*delta.Y

// 	if distSq < radius*radius {
// 		dist := float32(math.Sqrt(float64(distSq)))
// 		normal := rl.Vector2{X: 0, Y: 0}
// 		if dist != 0 {
// 			normal = rl.Vector2Scale(delta, 1/dist)
// 		}
// 		penetration := radius - dist

// 		return true, Contact{
// 			A:           circle,
// 			B:           aabb,
// 			Normal:      normal,
// 			Penetration: penetration,
// 		}
// 	}
// 	return false, Contact{}
// }

// // Circle vs Circle
// func CircleVsCircle(a, b *Body) (bool, Contact) {
// 	aPos := a.Position
// 	bPos := b.Position
// 	radiusA := a.Shape.(Circle).Radius
// 	radiusB := b.Shape.(Circle).Radius

// 	delta := rl.Vector2Subtract(bPos, aPos)
// 	distSq := delta.X*delta.X + delta.Y*delta.Y
// 	radiusSum := radiusA + radiusB

// 	if distSq < radiusSum*radiusSum {
// 		dist := float32(math.Sqrt(float64(distSq)))
// 		normal := rl.Vector2{X: 0, Y: 0}
// 		if dist != 0 {
// 			normal = rl.Vector2Scale(delta, 1/dist)
// 		}
// 		penetration := radiusSum - dist
// 		return true, Contact{
// 			A:           a,
// 			B:           b,
// 			Normal:      normal,
// 			Penetration: penetration,
// 		}
// 	}
// 	return false, Contact{}
// }

// // AABB vs AABB
// func AABBvsAABB(a, b *Body) (bool, Contact) {
// 	aMin := rl.Vector2Add(a.Position, a.Shape.(AABB).Min)
// 	aMax := rl.Vector2Add(a.Position, a.Shape.(AABB).Max)
// 	bMin := rl.Vector2Add(b.Position, b.Shape.(AABB).Min)
// 	bMax := rl.Vector2Add(b.Position, b.Shape.(AABB).Max)

// 	overlapX := math.Min(float64(aMax.X), float64(bMax.X)) - math.Max(float64(aMin.X), float64(bMin.X))
// 	overlapY := math.Min(float64(aMax.Y), float64(bMax.Y)) - math.Max(float64(aMin.Y), float64(bMin.Y))

// 	if overlapX > 0 && overlapY > 0 {
// 		var normal rl.Vector2
// 		var penetration float32
// 		if overlapX < overlapY {
// 			if a.Position.X < b.Position.X {
// 				normal = rl.Vector2{X: -1, Y: 0}
// 			} else {
// 				normal = rl.Vector2{X: 1, Y: 0}
// 			}
// 			penetration = float32(overlapX)
// 		} else {
// 			if a.Position.Y < b.Position.Y {
// 				normal = rl.Vector2{X: 0, Y: -1}
// 			} else {
// 				normal = rl.Vector2{X: 0, Y: 1}
// 			}
// 			penetration = float32(overlapY)
// 		}
// 		return true, Contact{
// 			A:           a,
// 			B:           b,
// 			Normal:      normal,
// 			Penetration: penetration,
// 		}
// 	}
// 	return false, Contact{}
// }

// // ---------------------------
// // Utility
// // ---------------------------

// func clamp(x, min, max float32) float32 {
// 	if x < min {
// 		return min
// 	}
// 	if x > max {
// 		return max
// 	}
// 	return x
// }
