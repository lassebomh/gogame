package game3

import (
	"game/vec2"
	"game/vec3"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Editor struct {
	TimeStep time.Duration

	Position         vec3.Value
	PositionVelocity vec3.Value

	Pitch           float64
	Yaw             float64
	ScrollY         float64
	ScrollYVelocity float64
	Scale           float64

	mousePosition      vec2.Value
	mouseWorldPosition vec3.Value

	Tool      Tool
	ToolFloor ToolFloor
	ToolWall  ToolWall

	world *World
}

type Tool = int32

const (
	TOOL_FLOOR = Tool(iota)
	TOOL_WALL
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
	if rl.IsKeyPressed(rl.KeyQ) {
		e.Position.Y -= 1
	}
	if rl.IsKeyPressed(rl.KeyE) {
		e.Position.Y += 1
	}

	friction := math.Pow(0.5, timeStep.Seconds()*20)

	e.ScrollYVelocity += float64(rl.GetMouseWheelMoveV().Y)
	e.ScrollYVelocity *= friction
	e.ScrollY += e.ScrollYVelocity * timeStep.Seconds() * 120

	e.Scale = math.Pow(2, -e.ScrollY/50)

	if targetVelocity.Length() > 0 {
		targetVelocity = targetVelocity.Normalize().Scale(e.Scale * timeStep.Seconds() * 1.5)

	}

	e.PositionVelocity = e.PositionVelocity.Lerp(targetVelocity, 1.0-friction)
	e.Position = e.Position.Add(e.PositionVelocity)

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

	mouseRay := rl.GetScreenToWorldRay(rl.Vector2{float32(currentMousePos.X), float32(currentMousePos.Y)}, e.GetCamera().Raylib())

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

	if rl.IsKeyPressed(rl.KeyOne) {
		e.Tool = TOOL_FLOOR
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		e.Tool = TOOL_WALL
	}

	switch e.Tool {
	case TOOL_FLOOR:
		e.ToolFloor.Update(e)
	case TOOL_WALL:
		e.ToolWall.Update(e)
	}
}

func (e *Editor) GetCamera() Camera3D {
	return Camera3D{
		Position: e.Position.Subtract(vec3.XYZ(
			math.Cos(e.Pitch)*math.Cos(e.Yaw),
			math.Sin(e.Pitch),
			math.Cos(e.Pitch)*math.Sin(e.Yaw),
		).Scale(e.Scale)),
		Target:     e.Position,
		Up:         vec3.Y(1),
		Fovy:       70,
		Projection: rl.CameraPerspective,
	}
}

func (e *Editor) Draw() {
	rl.ClearBackground(rl.DarkGray)

	BeginMode3D(e.GetCamera(), func() {
		rl.DrawGrid(3, 3)

		for pos, chunk := range e.world.Chunks {
			for y := range min(ChunkHeight, int(e.Position.Y)+1) {
				worldPos := ChunkToWorld(pos, LocalPos{0, y, 0})
				rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
			}
		}

		BeginOverlayMode(func() {
			rl.DrawCube(e.mouseWorldPosition.Raylib(), 0.05, 0.05, 0.05, rl.Black)
			rl.DrawCube(e.mouseWorldPosition.Add(vec3.X(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Red)
			rl.DrawCube(e.mouseWorldPosition.Add(vec3.Y(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Green)
			rl.DrawCube(e.mouseWorldPosition.Add(vec3.Z(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Blue)

			phys := NewPhysicsDrawer(e.Position.Y, true, true, true)
			cp.DrawSpace(e.world.space, &phys)

			switch e.Tool {
			case TOOL_FLOOR:
				e.ToolFloor.Draw3D(e)
			case TOOL_WALL:
				e.ToolWall.Draw3D(e)
			}
		})

	})

	switch e.Tool {
	case TOOL_FLOOR:
		e.ToolFloor.DrawHUD(e)
	case TOOL_WALL:
		e.ToolWall.DrawHUD(e)
	}

}
