package game2

import (
	"math"

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
	rl.DrawSphere(d.cv3(pos, 0), float32(size)/100, fColorToRaylib(fill))
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

func UpdatePhysicsY(g *Game, shape *cp.Shape, y float64, yVelocity float64) (float64, float64) {
	bodyPosition := shape.Body().Position()
	pos := NewVec3(bodyPosition.X, y, bodyPosition.Y)

	cell := g.Level.GetCell(pos)

	var groundY float64

	switch cell.Ground.Type {
	case GroundStair:
		x := math.Ceil(pos.X) - pos.X
		z := pos.Z - math.Floor(pos.Z)

		switch cell.Ground.StairDirection {
		case FACE_EAST:
			groundY = cell.Position.Y + x
		case FACE_NORTH:
			groundY = cell.Position.Y + z
		case FACE_WEST:
			groundY = cell.Position.Y + 1 - x
		case FACE_SOUTH:
			groundY = cell.Position.Y + 1 - z
		}
	case GroundFloor:
		groundY = cell.Position.Y
	case GroundEmpty:
		groundY = 0
	}

	if y > groundY {
		yVelocity -= g.TimeDelta.Seconds() / 5
	}

	if y+yVelocity < groundY {
		y = groundY
		yVelocity = 0
	}

	y += yVelocity

	nextCell := g.Level.GetCell(pos.Add(Y.Scale(0.1)))

	if nextCell.Position.Y > y && nextCell.Ground.Type == GroundFloor {
		y = math.Ceil(y)
	}

	yLevelCategory := uint(1 << uint(math.Floor(y)))
	shape.Filter.Categories = yLevelCategory
	shape.Filter.Mask = yLevelCategory | (1 << uint(math.Floor(y+0.25)))

	return y, yVelocity
}
