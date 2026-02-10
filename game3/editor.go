package game3

import (
	"fmt"
	v2 "game/vec2"
	v3 "game/vec3"
	"image/color"
	"math"
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type EditorDrawFlags int32

const (
	EditorDrawFlagPhysics = EditorDrawFlags(1 << iota)
)

type Editor struct {
	TimeStep time.Duration

	Position     v3.Value
	PositionSoft v3.Value

	Pitch           float64
	Yaw             float64
	ScrollY         float64
	ScrollYVelocity float64
	Scale           float64

	Camera          Camera3D
	EditorDrawFlags EditorDrawFlags

	mousePosition      v2.Value
	mouseWorldPosition v3.Value

	mouseCellPos       v3.Value
	mouseCellDirection FaceDirection

	Tool     Tool
	ToolCell ToolCell

	world *World
}

type Tool = int32

const (
	TOOL_FLOOR = Tool(iota)
	TOOL_WALL
	TOOL_CELL
	TOOL_PLAY
)

func (e *Editor) Upsert(world *World) {
	if e == nil {
		e = &Editor{
			TimeStep: 0,

			Position: v3.XYZ(0, 0, 0),
			ScrollY:  -30,

			Pitch: -0.8,
			Yaw:   0,
		}
	}
	e.world = world
	world.Editor = e
	e.PositionSoft = e.Position
}

func (e *Editor) Update(timeStep time.Duration) {
	e.TimeStep = timeStep

	forward := v3.XZ(math.Cos(e.Yaw), math.Sin(e.Yaw))
	right := forward.RotateByAxisAngle(v3.Y(-1), rl.Pi/2)

	targetVelocity := v3.Zero

	if rl.IsKeyDown(rl.KeyW) {
		targetVelocity = targetVelocity.Add(forward)
	}
	if rl.IsKeyDown(rl.KeyS) {
		targetVelocity = targetVelocity.Subtract(forward)
	}
	if rl.IsKeyDown(rl.KeyD) {
		targetVelocity = targetVelocity.Add(right)
	}
	if rl.IsKeyDown(rl.KeyA) {
		targetVelocity = targetVelocity.Subtract(right)
	}

	friction := math.Pow(0.5, timeStep.Seconds()*20)
	scroll := float64(rl.GetMouseWheelMoveV().Y)

	if rl.IsKeyDown(rl.KeyLeftShift) {
		if scroll > 0 {
			e.Position.Y += 1
		} else if scroll < 0 {
			e.Position.Y -= 1
		}
	} else {
		e.ScrollYVelocity += scroll
	}

	e.ScrollYVelocity *= friction
	e.ScrollY += e.ScrollYVelocity * timeStep.Seconds() * 120

	e.Scale = math.Pow(2, -e.ScrollY/50)

	if targetVelocity.Length() > 0 {
		targetVelocity = targetVelocity.Normalize().Scale(e.Scale * timeStep.Seconds() * 1.5)

	}

	e.Position = e.Position.Add(targetVelocity)
	e.PositionSoft = e.PositionSoft.Lerp(e.Position, 1-friction)

	e.Camera = Camera3D{
		Position: e.PositionSoft.Subtract(v3.XYZ(
			math.Cos(e.Pitch)*math.Cos(e.Yaw),
			math.Sin(e.Pitch),
			math.Cos(e.Pitch)*math.Sin(e.Yaw),
		).Scale(e.Scale)),
		Target:     e.PositionSoft,
		Up:         v3.Y(1),
		Fovy:       70,
		Projection: rl.CameraPerspective,
	}

	currentMousePos := v2.FromRaylib(rl.GetMousePosition())

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		mouseMove := (currentMousePos.Subtract(e.mousePosition)).Scale(0.007)
		e.Yaw += mouseMove.X
		e.Pitch -= mouseMove.Y

		if e.Pitch < -rl.Pi/2+1e-2 {
			e.Pitch = -rl.Pi/2 + 1e-2
		}

		if e.Pitch >= rl.Pi/2-1e-2 {
			e.Pitch = rl.Pi/2 - 1e-2
		}
	}
	e.mousePosition = currentMousePos

	mouseRay := rl.GetScreenToWorldRay(rl.Vector2{float32(currentMousePos.X), float32(currentMousePos.Y)}, e.Camera.Raylib())

	origin := v3.FromRaylib(mouseRay.Position)
	dir := v3.FromRaylib(mouseRay.Direction)
	ground := math.Floor(e.Position.Y)

	if math.Abs(dir.Y) >= 1e-6 {
		t := (ground - origin.Y) / dir.Y

		if t >= 0 {
			e.mouseWorldPosition = origin.Add(dir.Scale(t))
		}
	}
	e.mouseWorldPosition.Y = ground

	ix := math.Floor(e.mouseWorldPosition.X)
	iz := math.Floor(e.mouseWorldPosition.Z)
	fx := e.mouseWorldPosition.X - ix - 0.5
	fz := e.mouseWorldPosition.Z - iz - 0.5

	e.mouseCellPos = v3.XYZ(ix, e.mouseWorldPosition.Y, iz)

	if math.Abs(fx) > math.Abs(fz) {
		if fx < 0 {
			e.mouseCellDirection = FaceEast
		}
		if fx >= 0 {
			e.mouseCellDirection = FaceWest
		}
	} else {
		if fz > 0 {
			e.mouseCellDirection = FaceNorth
		}
		if fz <= 0 {
			e.mouseCellDirection = FaceSouth
		}
	}

	if rl.IsKeyPressed(rl.KeyOne) {
		e.Tool = TOOL_FLOOR
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		e.Tool = TOOL_WALL
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		e.Tool = TOOL_CELL
	}

	switch e.Tool {
	case TOOL_CELL:
		e.ToolCell.Update(e)
	}
}

func (e *Editor) Draw() {
	globals.Shaders.Main.FullBright.Set(1)
	rl.ClearBackground(rl.DarkGray)

	BeginMode3D(e.Camera, func() {

		for pos, chunk := range e.world.Chunks {
			for y := range min(ChunkHeight, int(e.Position.Y)+1) {
				worldPos := ChunkToWorld(pos, LocalPos{0, y, 0})
				rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
			}
		}

		if e.world.Player != nil && e.world.Player.Position.Y <= e.Position.Y {
			e.world.Player.Draw()
		}
		if e.world.Monster != nil && e.world.Monster.Position.Y <= e.Position.Y {
			e.world.Monster.Draw()
		}

		BeginOverlayMode(func() {
			// rl.DrawCube(e.mouseWorldPosition.Raylib(), 0.05, 0.05, 0.05, rl.Black)
			// rl.DrawCube(e.mouseWorldPosition.Add(v3.X(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Red)
			// rl.DrawCube(e.mouseWorldPosition.Add(v3.Y(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Green)
			// rl.DrawCube(e.mouseWorldPosition.Add(v3.Z(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Blue)

			if e.EditorDrawFlags&EditorDrawFlagPhysics != 0 {
				phys := NewPhysicsDrawer(e.Position.Y, true, true, true)

				e.world.space.BBQuery(
					cp.NewBBForCircle(e.Position.Chipmunk(), 8),
					cp.NewShapeFilter(0, Category(e.Position.Y, true, true), Category(e.Position.Y, true, true)),
					func(shape *cp.Shape, data interface{}) {
						cp.DrawShape(shape, &phys)
					},
					nil,
				)
			}

			rl.DrawCubeWiresV(e.mouseCellPos.AddXYZ(0.5, 0, 0.5).Raylib(), v3.XYZ(1, 0, 1).Raylib(), color.RGBA{255, 255, 255, 255})
			rl.DrawSphere(e.PositionSoft.Raylib(), float32(e.Scale/400), color.RGBA{0, 255, 0, 255})

			for cpos, _ := range e.world.Chunks {

				rl.DrawCubeWires(v3.XYZ(float64(cpos.X)*ChunkWidth+ChunkWidth*0.5, 0, float64(cpos.Z)*ChunkWidth+ChunkWidth*0.5).Raylib(), ChunkWidth, 0, ChunkWidth, color.RGBA{255, 255, 255, 10})
			}

			switch e.Tool {
			case TOOL_CELL:
				e.ToolCell.Draw3D(e)
			}

			if e.world.Player != nil {
				point := NewPathPoint(e.world, e.Position)
				path, _ := point.FindPath(e.world.Player.Position)
				DrawPath(path)
			}
		})

	})
	globals.Shaders.Main.FullBright.Set(0)

	cpos, lpos := WorldToChunk(e.mouseWorldPosition)

	rl.DrawText(fmt.Sprintf("%.1f %.1f %.1f", e.mouseWorldPosition.X, e.mouseWorldPosition.Y, e.mouseWorldPosition.Z), 10, 40, 20, rl.White)
	rl.DrawText(fmt.Sprintf("%+v", cpos), 10, 60, 20, rl.White)
	rl.DrawText(fmt.Sprintf("%+v", lpos), 10, 80, 20, rl.White)
	cell, ok := e.world.GetCell(e.mouseCellPos)

	if ok {
		y := int32(100)

		for i, face := range cell.Faces {

			rl.DrawText(fmt.Sprintf("%+v %+v", FaceDirection(i), face), 10, y, 20, rl.White)

			y += 20
		}
	}

	switch e.Tool {
	case TOOL_CELL:
		e.ToolCell.DrawHUD(e)
	}

	stack := NewStackLayout(0, 30, 100, 30)

	if raygui.Toggle(stack.Down(30), raygui.IconText(raygui.ICON_LASER, "Show Bodies"), e.EditorDrawFlags&EditorDrawFlagPhysics != 0) {
		e.EditorDrawFlags |= EditorDrawFlagPhysics
	} else {
		e.EditorDrawFlags &^= EditorDrawFlagPhysics
	}

	if raygui.Button(stack.Down(30), "TP Player") {
		player := game.Earth.Player
		if player == nil {
			player = game.Station.Player
		}

		player.Spawn(e.world)
		player.body.SetPosition(e.Position.Chipmunk())
		player.Update()
	}
}
