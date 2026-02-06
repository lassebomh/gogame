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
	rl.SetTargetFPS(60)
	rl.InitWindow(1900, 1100, "raylib")
	rl.SetWindowPosition(0, 10)
	defer rl.CloseWindow()

	LoadGame("./save.gob")

	t0 := rl.GetTime()

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()
		dt := time.Duration((t1 - t0) * float64(time.Second))

		if rl.IsKeyPressed(rl.KeyTab) {

		switch game.activeEditor {
			case game.Earth.Editor:
				game.activeEditor = game.Station.Editor
				fmt.Println("edit station")
			case game.Station.Editor:
				game.activeEditor = nil
				fmt.Println("disable editor")
			case nil:
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

		rl.DrawFPS(10, 10)

		rl.EndDrawing()
		t0 = t1
	}

	err := game.WriteToFile("./save.gob")
	if err != nil {
		log.Fatalln(err)
	}
}
