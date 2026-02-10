package game3

import (
	"game/vec3"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const (
	ShapeHeight float32 = 0.01 // The vertical thickness of the wireframes
)

type PhysicsDrawer struct {
	flags uint
	y     float32
}

func NewPhysicsDrawer(y float64, shapes, constraints, collisionPoints bool) PhysicsDrawer {
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
	return PhysicsDrawer{
		flags: flags,
		y:     float32(math.Floor(y)),
	}
}

// cv3 converts 2D physics vectors to 3D positions
func (d *PhysicsDrawer) cv3(v cp.Vector, offset float32) rl.Vector3 {
	return rl.Vector3{X: float32(v.X), Y: d.y + offset, Z: float32(v.Y)}
}

func (d *PhysicsDrawer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data interface{}) {
	color := fColorToRaylib(fill)
	// Draw a wireframe cylinder to represent the volume
	rl.DrawCylinderEx(d.cv3(pos, 0), d.cv3(pos, ShapeHeight), float32(radius), float32(radius), 16, color)
}

func (d *PhysicsDrawer) DrawSegment(a, b cp.Vector, fill cp.FColor, data interface{}) {
	color := fColorToRaylib(fill)
	rl.DrawLine3D(d.cv3(a, 0), d.cv3(b, 0), color)
	rl.DrawLine3D(d.cv3(a, ShapeHeight), d.cv3(b, ShapeHeight), color)
	rl.DrawLine3D(d.cv3(a, 0), d.cv3(a, ShapeHeight), color) // Vertical connector
}

func (d *PhysicsDrawer) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	// Draw wireframe capsule
	rl.DrawCapsule(d.cv3(a, 0), d.cv3(b, 0), float32(radius), 8, 8, fColorToRaylib(fill))
}

func (d *PhysicsDrawer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	color := fColorToRaylib(fill)

	if count < 3 {
		return // need at least 3 vertices for a polygon
	}

	// Triangulate using a fan from the first vertex
	for i := 1; i < count-1; i++ {
		v0 := d.cv3(verts[0], 0)
		v1 := d.cv3(verts[i], 0)
		v2 := d.cv3(verts[i+1], 0)

		rl.DrawTriangle3D(v2, v1, v0, color)
	}
}

func (d *PhysicsDrawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {
	rl.DrawSphere(d.cv3(pos, 0), float32(size)/200, fColorToRaylib(fill))
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

type DynamicPhysicsObject struct {
	shape    *cp.Shape
	Position vec3.Value
}

func UpdatePhysicsY(w *World, shape *cp.Shape, pos vec3.Value, yVelocity float64) (vec3.Value, float64) {
	bodyPos := shape.Body().Position()
	pos.X = bodyPos.X
	pos.Z = bodyPos.Y

	cell, _ := w.UpsertCellChunk(pos)

	groundY := math.Floor(pos.Y)

	switch cell.Faces[FaceDown].Type {
	case FaceStair:
		x := math.Ceil(pos.X) - pos.X
		z := pos.Z - math.Floor(pos.Z)

		switch cell.Faces[FaceDown].Rotation {
		case FaceEast:
			groundY += x
		case FaceNorth:
			groundY += z
		case FaceWest:
			groundY += 1 - x
		case FaceSouth:
			groundY += 1 - z
		}

	case FaceNone:
		groundY = 0
	}

	if pos.Y > groundY {
		yVelocity -= w.TimeStep.Seconds() / 5
	}
	if pos.Y-0.2 < groundY && cell.Faces[FaceDown].Type == FaceStair {
		yVelocity *= 10
	}

	if pos.Y+yVelocity < groundY {
		pos.Y = groundY
		yVelocity = 0
	}

	pos.Y += yVelocity

	nextCell, _ := w.UpsertCellChunk(pos.Add(vec3.Y(0.1)))

	if nextCell != cell && (nextCell.Faces[FaceDown].Type == FaceSolid || nextCell.Faces[FaceDown].Type == FaceStair) {
		pos.Y = math.Ceil(pos.Y)
	}

	shape.Filter.Categories = Category(pos.Y, false, true)
	shape.Filter.Mask = Category(pos.Y, true, true)

	return pos, yVelocity
}

const PhysicsTickrate = time.Second / 60

const (
	GroupStatic = uint(1 << iota)
	GroupPlayer
	GroupMonster
)

func Category(y float64, level bool, entity bool) uint {
	category := uint(0)
	yCategory := uint(1 << uint(math.Floor(y+0.4)))

	if level {
		category |= yCategory
	}
	if entity {
		category |= yCategory << uint(ChunkHeight)
	}
	return category
}
