package game3

import (
	"fmt"
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Start() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogError)
	rl.SetTargetFPS(144)
	rl.InitWindow(1600, 800, "raylib")
	rl.SetWindowPosition(0, 0)
	defer rl.CloseWindow()

	LoadGame("./save.gob")

	// chunk.Cells[0][0][0].Faces[FaceDown].Type = FaceSolid
	// chunk.Cells[0][0][0].Faces[FaceWest].Type = FaceSolid
	// chunk.Cells[0][0][0].Faces[FaceNorth].Type = FaceSolid
	// chunk.Cells[0][0][0].Faces[FaceEast].Type = FaceSolid
	// chunk.Cells[0][0][0].Faces[FaceSouth].Type = FaceSolid

	t0 := rl.GetTime()

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()
		dt := time.Duration((t1 - t0) * float64(time.Second))

		if rl.IsKeyPressed(rl.KeyTab) {

			if game.activeEditor == game.Earth.Editor {
				game.activeEditor = game.Station.Editor
				fmt.Println("edit station")
			} else if game.activeEditor == game.Station.Editor {
				game.activeEditor = nil
				fmt.Println("disable editor")
			} else if game.activeEditor == nil {
				game.activeEditor = game.Earth.Editor
				fmt.Println("edit earth")
			}
		}

		if game.activeEditor != nil {
			game.activeEditor.Update(dt)
		} else {
			game.Update(dt)
		}

		rl.BeginDrawing()

		if game.activeEditor != nil {
			game.activeEditor.Draw()
		} else {
			game.Draw()
		}

		rl.EndDrawing()
		t0 = t1
	}

	err := game.WriteToFile("./save.gob")
	if err != nil {
		log.Fatalln(err)
	}
}
