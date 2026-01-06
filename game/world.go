package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Game struct {
	Accumulator   float32
	DT            float32
	PhysicsDrawer *RaylibDrawer

	Earth   *World
	Station *World
}

func NewGame() *Game {

	game := &Game{}

	tilemap := NewTilemap(40, 40, 7.5)
	tilemap.Cols[18][19].Wall = 0
	roomWidth := 5
	roomHeight := 5
	rooms := 5
	for x := range rooms {
		for y := range rooms {
			tilemap.CreateRoom(x*roomWidth+3, y*roomHeight+3, roomWidth, roomHeight, WALL_R|WALL_L|WALL_B|WALL_T)
		}
	}
	tilemap.Cols[25][27].Door = WALL_B
	tilemap.Cols[27][25].Door = WALL_R

	game.Earth = NewWorld(tilemap)
	game.Earth.Monster = NewMonster(game.Earth, tilemap.CenterPosition.Add(cp.Vector{Y: 40}))

	tilemap = NewTilemap(10, 7, 7.5)
	tilemap.CreateRoom(0, 0, 10, 7, 0)

	game.Station = NewWorld(tilemap)
	game.Station.Player = NewPlayer(game.Station, tilemap.CenterPosition)

	game.PhysicsDrawer = NewRaylibDrawer(true, false, true)

	return game

}

func (g *Game) Update(dt float32) {
	g.DT = dt
	g.Accumulator += dt
	var world *World

	if g.Earth.Player != nil {
		world = g.Earth
	} else {
		world = g.Station
	}

	for g.Accumulator >= physicsTickrate {
		world.Space.Step(physicsTickrate)
		g.Accumulator -= physicsTickrate
	}

	world.Update(dt)
}

type World struct {
	Player  *Player
	Space   *cp.Space
	Tilemap *Tilemap
	Monster *Monster
	Items   []*PhysicalItem

	Camera rl.Camera3D

	MousePosition      rl.Vector2
	MouseWorldPosition cp.Vector
}

func (w *World) NewPhysicalItem(item Item, pos cp.Vector) *PhysicalItem {

	radius := 1.
	mass := radius * radius / 25.0
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	shape := w.Space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)

	pitem := &PhysicalItem{
		Item:   item,
		Body:   body,
		Radius: radius,
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

	cam := rl.Camera3D{}
	cam.Fovy = 80
	cam.Position = rl.Vector3{X: 0, Y: 2, Z: 0}
	cam.Target = rl.Vector3{X: 0, Y: 0, Z: -0.2}
	cam.Projection = rl.CameraOrthographic
	cam.Up = rl.Vector3{X: 0, Y: 1, Z: 0}

	world := &World{
		Space:  space,
		Camera: cam,
		Items:  make([]*PhysicalItem, 0),
	}

	world.Tilemap = tilemap
	world.Tilemap.GenerateBodies(world)

	return world
}

const physicsTickrate = 1.0 / 60.0

func (w *World) Update(dt float32) {
	w.MousePosition = rl.GetMousePosition()
	mouseRay := rl.GetScreenToWorldRay(w.MousePosition, w.Camera)

	w.MouseWorldPosition = cp.Vector{
		X: float64(mouseRay.Position.X),
		Y: float64(mouseRay.Position.Z - (mouseRay.Position.Y/mouseRay.Direction.Y)*mouseRay.Direction.Z),
	}

	if w.Player != nil {
		playerPos := VecFrom2D(w.Player.Body.Position(), float64(w.Player.Radius))
		w.Camera.Target = playerPos.Vector3
		w.Camera.Position = playerPos.Add(NewVec(0, 50, 20)).Vector3
		w.Player.Update(w)
	}

	if w.Monster != nil {
		w.Monster.Update(w)
	}

	if w.Player != nil && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		info := w.Space.PointQueryNearest(w.MouseWorldPosition, 0, cp.ShapeFilter{
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
