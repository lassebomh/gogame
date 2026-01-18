package game2

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const (
	WorldLevel  float32 = 0 // The base Z-height
	ShapeHeight float32 = 1 // The vertical thickness of the wireframes
)

type PhysicsDrawer struct {
	flags uint
}

func NewPhysicsDrawer(shapes, constraints, collisionPoints bool) PhysicsDrawer {
	var flags uint
	if shapes {
		flags |= cp.DRAW_SHAPES
	}
	if constraints {
		flags |= cp.DRAW_CONSTRAINTS
	}
	if collisionPoints {
		flags |= cp.DRAW_COLLISION_POINTS
	}
	return PhysicsDrawer{flags: flags}
}

// cv3 converts 2D physics vectors to 3D positions
func cv3(v cp.Vector, offset float32) rl.Vector3 {
	return rl.Vector3{X: float32(v.X), Y: WorldLevel + offset, Z: float32(v.Y)}
}

func (d *PhysicsDrawer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data interface{}) {
	color := fColorToRaylib(fill)
	// Draw a wireframe cylinder to represent the volume
	rl.DrawCylinderWiresEx(cv3(pos, 0), cv3(pos, ShapeHeight), float32(radius), float32(radius), 16, color)
}

func (d *PhysicsDrawer) DrawSegment(a, b cp.Vector, fill cp.FColor, data interface{}) {
	color := fColorToRaylib(fill)
	rl.DrawLine3D(cv3(a, 0), cv3(b, 0), color)
	rl.DrawLine3D(cv3(a, ShapeHeight), cv3(b, ShapeHeight), color)
	rl.DrawLine3D(cv3(a, 0), cv3(a, ShapeHeight), color) // Vertical connector
}

func (d *PhysicsDrawer) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	// Draw wireframe capsule
	rl.DrawCapsuleWires(cv3(a, 0), cv3(b, 0), float32(radius), 8, 8, fColorToRaylib(fill))
}

func (d *PhysicsDrawer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	color := fColorToRaylib(fill)
	for i := 0; i < count; i++ {
		nextIdx := (i + 1) % count
		currBottom := cv3(verts[i], 0)
		nextBottom := cv3(verts[nextIdx], 0)
		currTop := cv3(verts[i], ShapeHeight)
		nextTop := cv3(verts[nextIdx], ShapeHeight)

		rl.DrawLine3D(currBottom, nextBottom, color) // Bottom ring
		rl.DrawLine3D(currTop, nextTop, color)       // Top ring
		rl.DrawLine3D(currBottom, currTop, color)    // Vertical ribs
	}
}

func (d *PhysicsDrawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {
	rl.DrawSphereWires(cv3(pos, 0), float32(size)/100, 8, 8, fColorToRaylib(fill))
}

func (d *PhysicsDrawer) Flags() uint {
	return d.flags
}

func (d *PhysicsDrawer) OutlineColor() cp.FColor {
	return cp.FColor{R: 0.2, G: 0.2, B: 0.2, A: 1.0}
}

func (d *PhysicsDrawer) ShapeColor(shape *cp.Shape, data interface{}) cp.FColor {
	if shape.Body().IsSleeping() {
		return cp.FColor{R: 0.5, G: 0.5, B: 0.5, A: 0.8}
	}
	return cp.FColor{R: 1.0, G: 0.2, B: 0.2, A: 1.0}
}

func (d *PhysicsDrawer) ConstraintColor() cp.FColor {
	return cp.FColor{R: 0.0, G: 1.0, B: 0.0, A: 1.0}
}

func (d *PhysicsDrawer) CollisionPointColor() cp.FColor {
	return cp.FColor{R: 1.0, G: 0.0, B: 1.0, A: 1.0}
}

func (d *PhysicsDrawer) Data() interface{} {
	return nil
}

// Helper function to convert cp.FColor to rl.Color
func fColorToRaylib(c cp.FColor) rl.Color {
	return rl.Color{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}
