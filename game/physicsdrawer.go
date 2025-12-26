package game

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type RaylibDrawer struct {
	flags uint
	data  interface{}
}

func NewRaylibDrawer(shapes bool, constraints bool, collision_points bool) *RaylibDrawer {

	var flags uint = 0

	if shapes {
		flags |= cp.DRAW_SHAPES
	}
	if constraints {
		flags |= cp.DRAW_CONSTRAINTS
	}
	if collision_points {
		flags |= cp.DRAW_COLLISION_POINTS
	}

	return &RaylibDrawer{
		flags: flags,
		data:  nil,
	}
}

func (d *RaylibDrawer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data interface{}) {
	rl.DrawCircleV(v(pos), float32(radius), fColorToRaylib(fill))
	if outline.A > 0 {
		rl.DrawCircleLinesV(v(pos), float32(radius), fColorToRaylib(outline))
	}

	// Draw angle indicator line
	if radius > 0 {
		endX := pos.X + math.Cos(angle)*radius
		endY := pos.Y + math.Sin(angle)*radius
		end := cp.Vector{X: endX, Y: endY}
		rl.DrawLineV(v(pos), v(end), fColorToRaylib(outline))
	}
}

func (d *RaylibDrawer) DrawSegment(a, b cp.Vector, fill cp.FColor, data interface{}) {
	rl.DrawLineV(v(a), v(b), fColorToRaylib(fill))
}

func (d *RaylibDrawer) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	rl.DrawLineEx(v(a), v(b), max(1, float32(radius*2)), fColorToRaylib(fill))
	if outline.A > 0 {
		rl.DrawCircleV(v(a), float32(radius), fColorToRaylib(outline))
		rl.DrawCircleV(v(b), float32(radius), fColorToRaylib(outline))
	}
}

func (d *RaylibDrawer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {

	// Draw filled polygon
	if fill.A > 0 {
		for i := 1; i < count-1; i++ {
			rl.DrawTriangle(v(verts[0]), v(verts[i]), v(verts[i+1]), fColorToRaylib(fill))
		}
	}

	// Draw outline
	if outline.A > 0 {
		for i := 0; i < count; i++ {
			nextIdx := (i + 1) % count
			rl.DrawLineEx(v(verts[i]), v(verts[nextIdx]), max(1, float32(radius*2)), fColorToRaylib(outline))
		}
	}
}

func (d *RaylibDrawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {
	rl.DrawCircleV(v(pos), float32(size)/2, fColorToRaylib(fill))
}

func (d *RaylibDrawer) Flags() uint {
	return d.flags
}

func (d *RaylibDrawer) OutlineColor() cp.FColor {
	return cp.FColor{R: 0.2, G: 0.2, B: 0.2, A: 1.0}
}

func (d *RaylibDrawer) ShapeColor(shape *cp.Shape, data interface{}) cp.FColor {
	if shape.Body().IsSleeping() {
		return cp.FColor{R: 0.5, G: 0.5, B: 0.5, A: 0.8}
	}
	return cp.FColor{R: 1.0, G: 0.2, B: 0.2, A: 1.0}
}

func (d *RaylibDrawer) ConstraintColor() cp.FColor {
	return cp.FColor{R: 0.0, G: 1.0, B: 0.0, A: 1.0}
}

func (d *RaylibDrawer) CollisionPointColor() cp.FColor {
	return cp.FColor{R: 1.0, G: 0.0, B: 1.0, A: 1.0}
}

func (d *RaylibDrawer) Data() interface{} {
	return d.data
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
