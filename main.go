package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	// "net/http"
	// _ "net/http/pprof"

	. "game/game2"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const DEBUG = false

var SAVES_PATH, _ = filepath.Abs("./saves/")

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	screenWidth := int32(1700)
	screenHeight := int32(800)

	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetTargetFPS(144)
	rl.InitWindow(screenWidth, screenHeight, "raylib")
	rl.SetWindowPosition(0, 0)
	defer rl.CloseWindow()

	var g *Game

	os.MkdirAll(SAVES_PATH, 0755)

	saveFileName := filepath.Join(SAVES_PATH, time.Now().Format("20060102_150405")+".gob")

	saveDir, _ := os.ReadDir(SAVES_PATH)
	if len(saveDir) > 0 {
		save := GameSave{}

		saveFileName = filepath.Join(SAVES_PATH, saveDir[len(saveDir)-1].Name())
		err := LoadSaveFromFile(saveFileName, &save)
		if err != nil {
			log.Fatal(err)
		}
		g = save.Load()
		log.Printf("loaded \"%v\"", saveFileName)
	} else {
		log.Println("no saves exist. using default.")
		g = NewGameSave().Load()
	}

	g.Update(0)
	t0 := rl.GetTime()

	for !rl.WindowShouldClose() {
		t1 := rl.GetTime()

		if rl.IsKeyReleased(rl.KeyTab) {

			g.EditorEnabled = !g.EditorEnabled

		}

		dt := time.Duration((t1 - t0) * float64(time.Second))

		if g.EditorEnabled {
			g.Editor.Update(g, dt)
		} else {
			g.Update(dt)
		}

		BeginDrawing(func() {

			if g.EditorEnabled {
				g.Editor.Draw(g)
			} else {
				g.Draw()
			}
			rl.DrawFPS(5, 5)

		})

		t0 = t1
	}

	t0 = rl.GetTime()
	log.Println("saving...")

	if err := g.ToSave().WriteToFile(saveFileName); err != nil {
		log.Fatal(err)
	}
	log.Printf("saved to %v. took %.2f seconds", saveFileName, rl.GetTime()-t0)

}
