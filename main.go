package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	. "game/game2"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const DEBUG = false

var SAVES_PATH, _ = filepath.Abs("./saves/")

func main() {

	screenWidth := int32(1700)
	screenHeight := int32(800)

	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowUnfocused)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetTargetFPS(144)
	rl.InitWindow(screenWidth, screenHeight, "raylib")
	rl.SetWindowPosition(0, 0)
	defer rl.CloseWindow()

	var g *Game

	saveDir, _ := os.ReadDir(SAVES_PATH)
	if len(saveDir) > 0 {
		save := GameSave{}

		filename := saveDir[len(saveDir)-1].Name()
		err := LoadSaveFromFile(filepath.Join(SAVES_PATH, filename), &save)
		if err != nil {
			log.Fatal(err)
		}
		g = save.Load()
		log.Printf("loaded \"%v\"", filename)
	} else {
		log.Println("no saves exist. using default.")
		g = NewGameSave().Load()
	}

	t0 := rl.GetTime()

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()

		if rl.IsKeyDown(rl.KeyLeftControl) {
			if rl.IsKeyReleased(rl.KeyOne) {
				g.Mode = MODE_DEFAULT
				log.Println("MODE_DEFAULT")
			}
			if rl.IsKeyReleased(rl.KeyTwo) {
				g.Mode = MODE_FREE
				log.Println("MODE_FREE")
			}
			// if rl.IsKeyReleased(rl.KeyS) {
			// 	if err := g.ToSave().WriteToFile(filepath.Join(SAVES_PATH, time.Now().Format("20060102_150405")+".json")); err != nil {
			// 		log.Fatal(err)
			// 	}
			// }
		}

		dt := time.Duration((t1 - t0) * float64(time.Second))

		g.ModeUpdate(g.Mode, dt)

		BeginDrawing(func() {
			g.ModeDraw(g.Mode)
		})

		t0 = t1
	}

	if err := g.ToSave().WriteToFile(filepath.Join(SAVES_PATH, time.Now().Format("20060102_150405")+".json")); err != nil {
		log.Fatal(err)
	}
}
