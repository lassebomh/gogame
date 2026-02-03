package game3

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Start() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetTargetFPS(144)
	rl.InitWindow(1880, 1040, "raylib")
	rl.SetWindowPosition(0, 0)
	defer rl.CloseWindow()

	game.Upsert()

	pos := ChunkPos{0, 0}
	chunk := game.Earth.chunks[pos].Upsert(pos)

	chunk.Cells[0][0][0].Faces[FaceDown].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceWest].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceNorth].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceEast].Type = FaceSolid
	chunk.Cells[0][0][0].Faces[FaceSouth].Type = FaceSolid

	chunk.Reload()

	t0 := rl.GetTime()

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()
		dt := time.Duration((t1 - t0) * float64(time.Second))

		if rl.IsKeyPressed(rl.KeyTab) {

			if game.ActiveEditor == game.Earth.Editor {
				game.ActiveEditor = game.Station.Editor
				fmt.Println("edit station")
			} else if game.ActiveEditor == game.Station.Editor {
				game.ActiveEditor = nil
				fmt.Println("disable editor")
			} else if game.ActiveEditor == nil {
				game.ActiveEditor = game.Earth.Editor
				fmt.Println("edit earth")
			}
		}

		if game.ActiveEditor != nil {
			game.ActiveEditor.Update(dt)
		}

		rl.BeginDrawing()

		if game.ActiveEditor != nil {
			game.ActiveEditor.Draw()
		}

		rl.EndDrawing()
		t0 = t1
	}

}
