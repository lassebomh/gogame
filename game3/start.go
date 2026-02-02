package game3

import (
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Start() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetTargetFPS(144)
	rl.InitWindow(1700, 800, "raylib")
	rl.SetWindowPosition(0, 0)
	defer rl.CloseWindow()

	game.Init()

	pos := ChunkPos{0, 0}
	chunk := game.Level.Chunks[pos].Init(pos)

	chunk.Cells[0][0][0].Faces[FaceDown].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceWest].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceNorth].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceEast].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceSouth].Type = FaceSolid

	chunk.UpsertModels()

	cam := rl.Camera3D{
		Position:   rl.NewVector3(0, 2, -5),
		Target:     rl.NewVector3(0, 0, 0),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       45,
		Projection: rl.CameraPerspective,
	}

	t0 := rl.GetTime()
	angle := float64(0)

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()
		dt := time.Duration((t1 - t0) * float64(time.Second))

		angle += dt.Seconds() / 5

		cam.Position.X = float32(math.Sin(angle) * 5)
		cam.Position.Z = float32(math.Cos(angle) * 5)
		// cam.Position.Y = float32(math.Sin(angle) * 5)

		// fmt.Printf("cam.Position: %v\n", cam.Position)

		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)

		rl.BeginMode3D(cam)

		rl.DrawGrid(3, 3)
		rl.DrawModel(chunk.models[0], rl.NewVector3(0, 0, 0), 1, rl.White)

		rl.EndMode3D()

		rl.EndDrawing()
		t0 = t1
	}

}
