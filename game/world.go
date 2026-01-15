package game

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Game struct {
	Accumulator   float32
	DT            float32
	PhysicsDrawer *RaylibDrawer
	Day           float64

	TeleportTransition float64

	Earth   *World
	Station *World
}

const TIME_SCALE float64 = 1

func NewGame() *Game {

	game := &Game{
		Day: 0.35,

		TeleportTransition: 1,

		Accumulator:   0,
		DT:            0,
		PhysicsDrawer: NewRaylibDrawer(true, false, true),
	}

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

	x := tilemap.Width/2 + 1
	y := tilemap.Height/2 + 1
	size := 2

	for w := range size {
		for h := range size {
			tile := &tilemap.Cols[x+w-size/2][y+h-size/2]
			tile.TextureBaseX = 3
			tile.TextureBaseY = 1
		}
	}

	game.Earth = NewWorld(tilemap, false)
	// game.Earth.Monster = NewMonster(game.Earth, tilemap.CenterPosition.Add(cp.Vector{Y: 40}))

	tilemap = NewTilemap(4, 4, 7.5)
	tilemap.CreateRoom(0, 0, 4, 4, 0)

	// for w := range size {
	// 	for h := range size {
	// 		tile := &tilemap.Cols[x+w-size/2][y+h-size/2]
	// 		tile.TextureBaseX = 3
	// 		tile.TextureBaseY = 1
	// 	}
	// }

	game.Station = NewWorld(tilemap, true)

	game.Station.Player = NewPlayer(game.Station, game.Station.Tilemap.CenterPosition)

	return game

}

func (g *Game) Update(dt float32) *World {
	g.DT = dt
	g.Accumulator += dt

	dayDiff := float64(dt) / (60 * TIME_SCALE)
	g.Day += dayDiff

	g.Earth.Day = g.Day
	g.Station.Day = g.Day

	var otherWorld *World
	var currentWorld *World

	if g.Earth.Player != nil {
		currentWorld = g.Earth
		otherWorld = g.Station
	} else {
		currentWorld = g.Station
		otherWorld = g.Earth
	}

	hours := math.Mod(g.Day, 1)

	teleport := (hours >= HOUR_MORNING/24 && hours-dayDiff < HOUR_MORNING/24.) || (hours >= HOUR_NIGHT/24 && hours-dayDiff < HOUR_NIGHT/24.)

	if teleport && currentWorld.Player.Body.Position().Distance(currentWorld.Tilemap.CenterPosition) < 10 {
		currentWorld.Player.Teleport(currentWorld, otherWorld)
		currentWorld, otherWorld = otherWorld, currentWorld
	}

	for g.Accumulator >= physicsTickrate {
		currentWorld.Space.Step(physicsTickrate)
		g.Accumulator -= physicsTickrate
	}

	if currentWorld.IsStation && g.TeleportTransition < 1 {
		g.TeleportTransition = min(1, g.TeleportTransition+float64(dt)/5)
	}

	if !currentWorld.IsStation {
		if g.TeleportTransition > 0 {
			g.TeleportTransition = max(0, g.TeleportTransition-float64(dt)/5)
		}
	}

	currentWorld.Update()

	return currentWorld
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
var HOUR_MORNING float64 = 9
var HOUR_NIGHT float64 = 21
var HOURS_TRANSITION float64 = 1

func c(x float64) float64 {
	return (1 + math.Tanh(x)) / 2
}

func (w *World) Render(r *Render) {
	rl.ClearBackground(rl.Black)

	if w.IsStation {
		rl.BeginShaderMode(r.BackgroundShader)
		r.BackgroundShaderTime.SetFloat(float32(w.Day))
		r.BackgroundShaderFov.SetFloat(30.0)
		r.IChannel0Location.SetTexture(r.Textures["organic"])
		r.IChannel1Location.SetTexture(r.Textures["earth_elevation"])
		r.BackgroundShaderResolution.SetVec2(float32(r.RenderWidth), float32(r.RenderHeight))
		rl.DrawRectangle(0, 0, r.RenderWidth, r.RenderHeight, rl.White)
		rl.EndShaderMode()
	}

	rl.BeginShaderMode(r.Shader)
	defer rl.EndShaderMode()

	rl.BeginMode3D(w.Camera)
	defer rl.EndMode3D()

	if !w.IsStation {

		hour := math.Mod(w.Day, 1) * 24
		day := c(hour-HOUR_MORNING) - c(hour-HOUR_NIGHT)
		transitionColor := 1 + ((c(2*(hour-HOUR_MORNING-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_MORNING+HOURS_TRANSITION/2))) + (c(2*(hour-HOUR_NIGHT-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_NIGHT+HOURS_TRANSITION/2))))
		transitionAngle := 1 + ((c((hour - HOUR_MORNING - HOURS_TRANSITION/2)) - c((hour - HOUR_MORNING + HOURS_TRANSITION/2))) + (c((hour - HOUR_NIGHT - HOURS_TRANSITION/2)) - c((hour - HOUR_NIGHT + HOURS_TRANSITION/2))))

		sunColor := DAWN.Lerp(NIGHT.Lerp(DAY, float32(day)), float32(transitionColor))

		r.LightDirectional(NewVec(float32(1-transitionAngle), float32(1-day*2), 0).Normalize(), rl.ColorFromHSV(sunColor.X, sunColor.Y, sunColor.Z), 0.5)

	} else {
		r.LightDirectional(NewVec(0, -1, 0).Normalize(), rl.White, 0.25)

	}

	if w.Player != nil {
		playerPos := VecFrom2D(w.Player.Body.Position(), w.Player.Radius*1)
		lookDir := NewVec(float32(math.Cos(w.Player.Body.Angle())), 0, float32(math.Sin(w.Player.Body.Angle())))
		flashlightPos := playerPos.Subtract(lookDir.Scale(float32(w.Player.Radius) * 3))
		flashlightTarget := flashlightPos.Add(lookDir)

		r.LightSpot(flashlightPos, flashlightTarget, 13, 18, rl.NewColor(255, 255, 100, 255), 2)
	}

	r.UpdateValues()

	for _, col := range w.Tilemap.Cols {
		for _, tile := range col {
			scale := float32(w.Tilemap.Scale)
			pos := VecFrom2D(tile.WorldPosition, 0)

			// floorY := float32(0)
			// if tile.TextureBaseX == 0 && tile.TextureBaseY == 3 {
			// 	floorY = -0.1
			// }
			r.DrawModel(r.Models["plane"], tile.TextureBaseX, tile.TextureBaseY, pos.Add(XZ.Scale(scale/2)), NewVec(scale/2, scale/2, scale/2), Y, 0)

			scaleV := XYZ.Scale(scale * 0.5)
			scaleVec := scaleV.Vector3

			if tile.Wall&WALL_L != 0 {
				rl.DrawModel(r.Models["wall"], pos.Add(NewVec(0, 0, scale)).Vector3, scale/2, rl.White)
			}
			if tile.Wall&WALL_T != 0 {
				rl.DrawModelEx(r.Models["wall"], pos.Add(NewVec(scale, 0, scale*w.Tilemap.WallDepthRatio)).Vector3, Y.Vector3, 90, scaleVec, rl.RayWhite)
			}
			if tile.Wall&WALL_R != 0 {
				rl.DrawModelEx(r.Models["wall"], pos.Add(NewVec(scale, 0, 0)).Vector3, Y.Scale(1).Vector3, 180, scaleVec, rl.RayWhite)
			}
			if tile.Wall&WALL_B != 0 {
				rl.DrawModelEx(r.Models["wall"], pos.Add(NewVec(0, 0, scale*(1-w.Tilemap.WallDepthRatio))).Vector3, Y.Vector3, 270, scaleVec, rl.RayWhite)
			}
			if tile.DoorBody != nil {
				r.DrawModel(r.Models["door"], 3, 0, VecFrom2D(tile.DoorBody.Position(), 0), scaleV, Y.Negate(), float32(tile.DoorBody.Angle()*rl.Rad2deg))
				rl.DrawModelEx(r.Models["door"], VecFrom2D(tile.DoorBody.Position(), 0).Vector3, Y.Negate().Vector3, float32(tile.DoorBody.Angle()*rl.Rad2deg), scaleVec, rl.RayWhite)
			}
		}
	}

	if w.Player != nil {
		playerPos := VecFrom2D(w.Player.Body.Position(), w.Player.Radius*1)
		rl.DrawSphereEx(playerPos.Vector3, float32(w.Player.Radius), 12, 12, rl.Red)
	}

	for _, pitem := range w.Items {
		rl.DrawSphereEx(VecFrom2D(pitem.Body.Position(), pitem.Radius).Vector3, float32(pitem.Radius), 12, 12, rl.Red)
	}

	if w.Monster != nil {

		color := color.RGBA{R: 40, G: 40, B: 40, A: 255}

		monsterRadius := float32(w.Monster.Radius)
		rl.DrawModelEx(
			r.Models["monster_body"],
			VecFrom2D(w.Monster.Body.Position(), w.Monster.Radius/2).Vector3,
			Y.Vector3,
			float32(-w.Monster.Body.Angle())*rl.Rad2deg,
			NewVec(monsterRadius*1, monsterRadius, monsterRadius*1).Vector3,
			color,
		)

		for _, arm := range w.Monster.Arms {
			for _, segment := range arm.Segments {
				rl.DrawModelEx(
					r.Models["monster_arm_segment"],
					VecFrom2D(segment.Body.Position(), w.Monster.Radius*2-segment.Width).Vector3,
					Y.Vector3,
					float32(-segment.Body.Angle())*rl.Rad2deg,
					NewVec(float32(segment.Length)*1.1, float32(segment.Width), float32(segment.Width)).Vector3,
					color,
				)
			}
		}
	}

	// if DEBUG {
	// 	rl.DrawRenderBatchActive()
	// 	rl.DisableDepthTest()
	// 	cp.DrawSpace(game.Earth.Space, game.PhysicsDrawer)
	// 	if game.Earth.Monster != nil {
	// 		DrawLine(rl.Red, game.Earth.Monster.Path...)
	// 		if game.Earth.Monster != nil {
	// 			for _, arm := range game.Earth.Monster.Arms {
	// 				DrawLine(rl.Blue, arm.Segments[len(arm.Segments)-1].Body.Position(), arm.TipTarget)
	// 			}
	// 			DrawLine(rl.Green, game.Earth.Monster.Body.Position(), game.Earth.Monster.Target)
	// 		}
	// 	}
	// 	rl.DrawRenderBatchActive()
	// 	rl.EnableDepthTest()
	// }
}
