package game3

import (
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Start() {
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogError)
	rl.SetTargetFPS(144)
	rl.InitWindow(1400, 700, "raylib")
	rl.SetWindowPosition(0, 10)
	defer rl.CloseWindow()

	LoadGame("./save.gob")

	t0 := rl.GetTime()

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()
		dt := time.Duration((t1 - t0) * float64(time.Second))

		var ActiveEditor *Editor

		if game.Earth.EditorActive {
			if rl.IsKeyPressed(rl.KeyTab) {
				game.Earth.EditorActive = false
				game.Station.EditorActive = true
			} else {
				ActiveEditor = game.Earth.Editor
			}
		} else if game.Station.EditorActive {
			if rl.IsKeyPressed(rl.KeyTab) {
				game.Earth.EditorActive = false
				game.Station.EditorActive = false

			} else {
				ActiveEditor = game.Station.Editor
			}
		} else {
			if rl.IsKeyPressed(rl.KeyTab) {
				game.Earth.EditorActive = true
				game.Station.EditorActive = false
			}
			ActiveEditor = nil
		}

		if ActiveEditor != nil {
			ActiveEditor.Update(dt)
		} else {
			game.Update(dt)
		}

		BeginDrawing(func() {
			if ActiveEditor != nil {

				ActiveEditor.Draw()

			} else {
				game.Draw()
			}
			rl.DrawFPS(0, 0)
		})

		t0 = t1
	}

	err := game.WriteToFile("./save.gob")
	if err != nil {
		log.Fatalln(err)
	}
}
