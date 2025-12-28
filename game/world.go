package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type World struct {
	Player        *Player
	Space         *cp.Space
	Tilemap       *Tilemap
	Monster       *Monster
	Items         []*PhysicalItem
	Accumulator   float32
	DT            float32
	Camera        rl.Camera2D
	PhysicsDrawer *RaylibDrawer
}

func (w *World) NewPhysicalItem(item Item, pos cp.Vector) *PhysicalItem {

	radius := 20.
	mass := radius * radius / 25.0
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	shape := w.Space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)

	pitem := &PhysicalItem{
		Item: item,
		Body: body,
	}

	w.Items = append(w.Items, pitem)

	return pitem
}

func (w *World) RemovePhysicalItem(item *PhysicalItem) {
	for i, pitem := range w.Items {
		if pitem == item {
			w.Items[i] = w.Items[len(w.Items)-1]
			w.Items = w.Items[:len(w.Items)-1]
			break
		}
	}
	item.Body.EachShape(func(shape *cp.Shape) {
		w.Space.RemoveShape(shape)
	})
	w.Space.RemoveBody(item.Body)
	item.Body = nil
}

func NewWorld(tilemap *Tilemap) *World {
	space := cp.NewSpace()
	space.Iterations = 20
	space.SetCollisionSlop(0.5)

	world := &World{
		Space:  space,
		Camera: rl.NewCamera2D(rl.Vector2{}, rl.Vector2{}, 0, 1),
		Items:  make([]*PhysicalItem, 0),
	}

	world.Tilemap = tilemap
	world.Tilemap.GenerateBodies(world)

	world.Player = NewPlayer(world, tilemap.CenterPosition)
	world.Monster = NewMonster(world, tilemap.CenterPosition.Add(cp.Vector{X: 1000, Y: 100}))

	world.PhysicsDrawer = NewRaylibDrawer(true, false, true)

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

	if w.Monster != nil {
		w.Monster.Update(w)
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

			for _, pitem := range w.Items {
				if pitem.Body == clickedBody {
					w.Player.Items = append(w.Player.Items, pitem.Item)
					w.RemovePhysicalItem(pitem)

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

	if w.PhysicsDrawer != nil {
		cp.DrawSpace(w.Space, w.PhysicsDrawer)
	}

	for _, arm := range w.Monster.Arms {
		if arm.Path != nil {
			RenderPath(arm.Path, rl.Blue)
		}
	}

	w.Monster.Render(w)

	// w.Player.Render(w)

	rl.EndMode2D()

	w.Player.RenderHud(w)
	rl.DrawFPS(0, 0)

}

func v(v cp.Vector) rl.Vector2 {
	return rl.Vector2{X: float32(v.X), Y: float32(v.Y)}
}
