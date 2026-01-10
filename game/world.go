package game

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Game struct {
	Accumulator   float32
	DT            float32
	PhysicsDrawer *RaylibDrawer
	Day           float64

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

	game.Earth = NewWorld(tilemap, false)
	game.Earth.Monster = NewMonster(game.Earth, tilemap.CenterPosition.Add(cp.Vector{Y: 40}))

	tilemap = NewTilemap(1, 1, 7.5)
	tilemap.CreateRoom(0, 0, 1, 1, 0)

	game.Station = NewWorld(tilemap, true)

	game.Station.Player = NewPlayer(game.Station, game.Station.Tilemap.CenterPosition)

	game.PhysicsDrawer = NewRaylibDrawer(true, false, true)

	return game

}

func (g *Game) Update(dt float32) *World {
	g.DT = dt
	g.Accumulator += dt
	g.Day += float64(dt) / (5 * 1)

	g.Earth.Day = g.Day
	g.Station.Day = g.Day

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

	world.Update()

	return world
}

type World struct {
	Player  *Player
	Space   *cp.Space
	Tilemap *Tilemap
	Monster *Monster
	Items   []*PhysicalItem

	IsStation bool
	Day       float64

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

func NewWorld(tilemap *Tilemap, isStation bool) *World {
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
		Space:     space,
		Camera:    cam,
		Items:     make([]*PhysicalItem, 0),
		IsStation: isStation,
	}

	world.Tilemap = tilemap
	world.Tilemap.GenerateBodies(world)

	return world
}

const physicsTickrate = 1.0 / 60.0

func (w *World) Update() {
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

var NIGHT = NewVec(-115, 0.3, .1)
var DAWN = NewVec(5, 0.5, 1)
var DAY = NewVec(55, 0.1, 1)
var HOUR_MORNING float64 = 8
var HOUR_NIGHT float64 = 20
var HOURS_TRANSITION float64 = 1

// func SkyHSV(hour float64) (h, s, v float64) {
// 	// solar phase: sunrise ≈ 6, noon ≈ 12
// 	x := math.Sin(2 * math.Pi * (hour - 6) / 24)

// 	Debug(x)

// 	// daylight factor
// 	day := 1 + math.Min(0, x)

// 	// twilight factor (strongest near horizon)
// 	// t := math.Exp(-8 * x * x)

// 	// hue: white(day) → orange/red/purple → blue(night)
// 	h = 55

// 	// saturation: minimal at noon, higher at twilight/night
// 	s = 0.5

// 	// value: bright day, dark night
// 	v = 1 // 0.1 + 0.9*d

// 	return
// }

func c(x float64) float64 {
	return (1 + math.Tanh(x)) / 2
}

func (w *World) RenderEarth(r *Render) {

	hour := math.Mod(w.Day, 1) * 24
	day := c(hour-HOUR_MORNING) - c(hour-HOUR_NIGHT)
	transition := 1 + ((c(2*(hour-HOUR_MORNING-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_MORNING+HOURS_TRANSITION/2))) + (c(2*(hour-HOUR_NIGHT-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_NIGHT+HOURS_TRANSITION/2))))

	sunColor := DAWN.Lerp(NIGHT.Lerp(DAY, float32(day)), float32(transition))

	// sunColor :=
	fmt.Printf("day %.2f tra %.2f\n", day, transition)

	// sun.Strength = 0.1
	// sun.Color = color.RGBA{20, 0, 255, 255}
	// sun.Position = NewVec(0, 0, 0).Vector3
	// sun.Target = NewVec(0, -1, 0).Vector3

	// sun.Strength = 0.5
	// sun.Color = color.RGBA{255, 0, 100, 255}
	// sun.Position = NewVec(0, 0, 0).Vector3
	// sun.Target = NewVec(0.5, -1, 0).Vector3

	// strength := min(1+math.Sin(hoursRadians), 0.8) + 0.2
	// sun.Strength = float32(strength)

	r.Light(LIGHT_DIRECTIONAL, NewVec(0, 0, 0), NewVec(0, -1, 0), rl.ColorFromHSV(sunColor.X, sunColor.Y, sunColor.Z), 0.5)

	playerPos := VecFrom2D(w.Player.Body.Position(), w.Player.Radius*1)
	lookDir := NewVec(float32(math.Cos(w.Player.Body.Angle())), 0, float32(math.Sin(w.Player.Body.Angle())))
	flashlightPos := playerPos.Subtract(lookDir.Scale(float32(w.Player.Radius) * 3))
	flashlightTarget := flashlightPos.Add(lookDir)

	r.Light(LIGHT_SPOT, flashlightPos, flashlightTarget, rl.NewColor(255, 255, 100, 255), 2)
}
