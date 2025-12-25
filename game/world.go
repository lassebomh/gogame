package game

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

var grabbableMaskBit uint = 1 << 31
var grabFilter = cp.ShapeFilter{
	Group: cp.NO_GROUP, Categories: grabbableMaskBit, Mask: grabbableMaskBit,
}

type World struct {
	Player      *Player
	Space       *cp.Space
	Tilemap     *Tilemap
	Items       []Item
	Accumulator float32
	DT          float32
	Camera      rl.Camera2D
}

func NewWorld(tilemap *Tilemap) *World {
	space := cp.NewSpace()
	space.Iterations = 20
	space.SetCollisionSlop(0.5)

	world := &World{
		Space:  space,
		Camera: rl.NewCamera2D(rl.Vector2{}, rl.Vector2{}, 0, 1),
		Items:  make([]Item, 0),
	}

	world.Player = NewPlayer(world)

	world.Tilemap = tilemap
	world.Tilemap.GenerateBodies(world)

	return world
}

const physicsTickrate = 1.0 / 60.0

func (w *World) Update(dt float32) {
	w.DT = dt
	w.Accumulator += dt
	for w.Accumulator >= physicsTickrate {
		w.Space.Step(physicsTickrate)
		w.Accumulator -= physicsTickrate
	}

	if w.Player != nil {
		w.Player.Update(w)
	}

	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		mousePos := rl.GetMousePosition()
		worldPos := rl.GetScreenToWorld2D(mousePos, w.Camera)

		cpWorldPos := cp.Vector{
			X: float64(worldPos.X),
			Y: float64(worldPos.Y),
		}

		info := w.Space.PointQueryNearest(cpWorldPos, 0, cp.ShapeFilter{
			Group:      cp.NO_GROUP,
			Categories: cp.ALL_CATEGORIES,
			Mask:       cp.ALL_CATEGORIES,
		})

		if info.Shape != nil && info.Point.Distance(w.Player.Body.Position()) < 100 {
			clickedBody := info.Shape.Body()

			for i, item := range w.Items {
				if item.GetBody() == clickedBody {
					w.Player.Inventory = append(w.Player.Inventory, item)
					w.Items[i] = w.Items[len(w.Items)-1]
					w.Items = w.Items[:len(w.Items)-1]
					ItemDespawnBody(item, w)

					break
				}
			}
		}

	}

	for _, item := range w.Items {
		item.Update(w)
	}
}

func (w *World) Render() {

	w.Camera.Offset = rl.Vector2{X: float32(rl.GetRenderWidth()) / 2, Y: float32(rl.GetRenderHeight()) / 2}
	cameraTarget := cp.Vector{X: float64(w.Camera.Target.X), Y: float64(w.Camera.Target.Y)}
	cameraTarget = cameraTarget.Lerp(w.Player.Body.Position(), 0.05)
	w.Camera.Target = v(cameraTarget)

	rl.BeginMode2D(w.Camera)

	w.Tilemap.Render()
	w.Space.EachShape(func(s *cp.Shape) {
		switch shape := s.Class.(type) {

		case *cp.Segment:
			a := shape.A()
			b := shape.B()
			r := float32(shape.Radius())
			rl.DrawLineEx(v(a), v(b), r*2, rl.DarkGray)
			rl.DrawCircleV(v(a), r, rl.Blue)
			rl.DrawCircleV(v(b), r, rl.Blue)
			mid := cp.Vector{X: (a.X + b.X) * 0.5, Y: (a.Y + b.Y) * 0.5}
			n := shape.Normal()
			normEnd := cp.Vector{
				X: mid.X + n.X*float64(r*4),
				Y: mid.Y + n.Y*float64(r*4),
			}
			rl.DrawLineV(v(mid), v(normEnd), rl.Green)

		case *cp.Circle:
			pos := shape.Body().Position()
			r := float32(shape.Radius())
			rl.DrawCircleV(v(pos), r, rl.Red)

		case *cp.PolyShape:
			count := shape.Count()
			body := shape.Body()
			for i := 0; i < count; i++ {
				vert := shape.Vert(i)
				nextVert := shape.Vert((i + 1) % count)
				worldVert := body.LocalToWorld(vert)
				worldNextVert := body.LocalToWorld(nextVert)
				rl.DrawLineV(v(worldVert), v(worldNextVert), rl.Red)
			}
		default:
			fmt.Println("unexpected shape", s.Class)
		}
	})

	// w.Player.Render(w)

	rl.EndMode2D()

	w.Player.RenderHud(w)
	rl.DrawFPS(0, 0)

}

func v(v cp.Vector) rl.Vector2 {
	return rl.Vector2{X: float32(v.X), Y: float32(v.Y)}
}
