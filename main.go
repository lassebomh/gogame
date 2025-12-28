package main

import (
	. "game/game"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

func main() {
	const width, height = 1200, 800

	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowHighdpi)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(width, height, "game")

	tiles := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 8, 1, 1, 1, 1, 1, 2, 0, 0, 8, 1, 1, 1, 1, 1, 2},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 8, 1, 1, 1, 1, 1, 2, 0, 0, 8, 1, 1, 1, 1, 1, 2},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 7, 0, 0, 0, 0, 0, 3, 0, 0, 7, 0, 0, 0, 0, 0, 3},
		{0, 6, 5, 5, 5, 5, 5, 4, 0, 0, 6, 5, 5, 5, 5, 5, 4},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	tilemap := NewTilemap(1000, 1000, 40)

	for y, row := range tiles {
		for x, tile := range row {

			if tile == 0 {
				continue
			}

			X := x + int(tilemap.CenterPosition.X/tilemap.Scale)
			Y := y + int(tilemap.CenterPosition.Y/tilemap.Scale)

			switch tile {
			case 1:
				tilemap.Cols[X][Y].Wall = WALL_T
			case 2:
				tilemap.Cols[X][Y].Wall = WALL_T | WALL_R
			case 3:
				tilemap.Cols[X][Y].Wall = WALL_R
			case 4:
				tilemap.Cols[X][Y].Wall = WALL_B | WALL_R
			case 5:
				tilemap.Cols[X][Y].Wall = WALL_B
			case 6:
				tilemap.Cols[X][Y].Wall = WALL_B | WALL_L
			case 7:
				tilemap.Cols[X][Y].Wall = WALL_L
			case 8:
				tilemap.Cols[X][Y].Wall = WALL_T | WALL_L
			}
		}
	}

	world := NewWorld(tilemap)

	lastTime := rl.GetTime()

	for !rl.WindowShouldClose() {
		now := rl.GetTime()
		dt := float32(now - lastTime)
		lastTime = now

		world.Update(dt)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		world.Render()

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

func v(v cp.Vector) rl.Vector2 {
	return rl.Vector2{X: float32(v.X), Y: float32(v.Y)}
}
