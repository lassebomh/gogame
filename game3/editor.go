package game3

import (
	"game/vec2"
	"game/vec3"
	"image/color"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Editor struct {
	TimeStep time.Duration

	Position     vec3.Value
	PositionSoft vec3.Value
	// PositionVelocity vec3.Value

	Pitch           float64
	Yaw             float64
	ScrollY         float64
	ScrollYVelocity float64
	Scale           float64

	Camera Camera3D

	mousePosition      vec2.Value
	mouseWorldPosition vec3.Value

	mouseCellPos       vec3.Value
	mouseCellDirection FaceDirection

	Tool      Tool
	ToolFloor ToolFloor
	ToolWall  ToolWall
	ToolCell  ToolCell

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

			Position: vec3.XYZ(0, 0, 0),
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

	forward := vec3.XZ(math.Cos(e.Yaw), math.Sin(e.Yaw))
	right := forward.RotateByAxisAngle(vec3.Y(-1), rl.Pi/2)

	targetVelocity := vec3.Zero

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
		Position: e.PositionSoft.Subtract(vec3.XYZ(
			math.Cos(e.Pitch)*math.Cos(e.Yaw),
			math.Sin(e.Pitch),
			math.Cos(e.Pitch)*math.Sin(e.Yaw),
		).Scale(e.Scale)),
		Target:     e.PositionSoft,
		Up:         vec3.Y(1),
		Fovy:       70,
		Projection: rl.CameraPerspective,
	}

	currentMousePos := vec2.FromRaylib(rl.GetMousePosition())

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

	origin := vec3.FromRaylib(mouseRay.Position)
	dir := vec3.FromRaylib(mouseRay.Direction)
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

	e.mouseCellPos = vec3.XYZ(ix, e.mouseWorldPosition.Y, iz)

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
	case TOOL_FLOOR:
		e.ToolFloor.Update(e)
	case TOOL_WALL:
		e.ToolWall.Update(e)
	case TOOL_CELL:
		e.ToolCell.Update(e)
	}
}

func (e *Editor) Draw() {
	e.world.shader.FullBright.Set(1)
	rl.ClearBackground(rl.DarkGray)

	BeginMode3D(e.Camera, func() {

		for pos, chunk := range e.world.Chunks {
			for y := range min(ChunkHeight, int(e.Position.Y)+1) {
				worldPos := ChunkToWorld(pos, LocalPos{0, y, 0})
				rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
			}
		}

		BeginOverlayMode(func() {
			// rl.DrawCube(e.mouseWorldPosition.Raylib(), 0.05, 0.05, 0.05, rl.Black)
			// rl.DrawCube(e.mouseWorldPosition.Add(vec3.X(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Red)
			// rl.DrawCube(e.mouseWorldPosition.Add(vec3.Y(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Green)
			// rl.DrawCube(e.mouseWorldPosition.Add(vec3.Z(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Blue)

			// phys := NewPhysicsDrawer(e.Position.Y, true, true, true)
			// cp.DrawSpace(e.world.space, &phys)

			rl.DrawCubeWiresV(e.mouseCellPos.AddXYZ(0.5, 0, 0.5).Raylib(), vec3.XYZ(1, 0, 1).Raylib(), color.RGBA{255, 255, 255, 255})
			rl.DrawSphere(e.PositionSoft.Raylib(), float32(e.Scale/400), color.RGBA{0, 255, 0, 255})

			switch e.Tool {
			case TOOL_FLOOR:
				e.ToolFloor.Draw3D(e)
			case TOOL_WALL:
				e.ToolWall.Draw3D(e)
			case TOOL_CELL:
				e.ToolCell.Draw3D(e)
			}
		})

	})
	e.world.shader.FullBright.Set(0)

	switch e.Tool {
	case TOOL_FLOOR:
		e.ToolFloor.DrawHUD(e)
	case TOOL_WALL:
		e.ToolWall.DrawHUD(e)
	case TOOL_CELL:
		e.ToolCell.DrawHUD(e)
	}

}
