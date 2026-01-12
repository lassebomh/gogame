package main

import (
	. "game/game"
	"image/color"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const DEBUG = false

func main() {
	screenWidth := int32(1700)
	screenHeight := int32(1000)

	var pixelScale int32

	if DEBUG {
		pixelScale = 1
	} else {
		pixelScale = 5
	}

	renderWidth := screenWidth / pixelScale
	renderHeight := screenHeight / pixelScale

	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowUnfocused | rl.FlagWindowUnfocused)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(screenWidth, screenHeight, "raylib")
	defer rl.CloseWindow()

	if rl.GetMonitorCount() > 1 {
		pos := rl.GetMonitorPosition(1)
		rl.SetWindowPosition(int(pos.X), int(pos.Y))
	}

	game := NewGame()

	render := NewRender(renderWidth, renderHeight)
	defer render.Unload()

	earthTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(earthTexture)
	rl.SetTextureFilter(earthTexture.Texture, rl.FilterPoint)
	rl.SetTextureWrap(earthTexture.Texture, rl.WrapClamp)

	stationTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(stationTexture)
	rl.SetTextureFilter(stationTexture.Texture, rl.FilterPoint)
	rl.SetTextureWrap(stationTexture.Texture, rl.WrapClamp)

	// iChannel0 := rl.LoadTexture("./models/organic.png")
	// rl.SetTextureWrap(iChannel0, rl.WrapRepeat)
	// defer rl.UnloadTexture(iChannel0)
	// iChannel1 := rl.LoadTexture("./models/earth_elevation.png")
	// rl.SetTextureWrap(iChannel1, rl.WrapRepeat)
	// defer rl.UnloadTexture(iChannel1)

	backgroundShader := rl.LoadShader("", "./glsl330/planet2.fs")
	defer rl.UnloadShader(backgroundShader)

	displayTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(displayTexture)
	rl.SetTextureFilter(displayTexture.Texture, rl.FilterPoint)
	fadeShader := rl.LoadShader("", "./glsl330/fade.fs")
	defer rl.UnloadShader(fadeShader)

	fadeiChannel0Location := GetUniform(fadeShader, "iChannel0")
	fadeiChannel1Location := GetUniform(fadeShader, "iChannel1")
	// fadeiChannel0LocationDepth := GetUniform(fadeShader, "iChannel0Depth")
	// fadeiChannel1LocationDepth := GetUniform(fadeShader, "iChannel1Depth")
	fadeiChannelPrevLocation := GetUniform(fadeShader, "iChannelPrev")
	fadebackgroundShaderTime := GetUniform(fadeShader, "iTime")
	fadebackgroundShaderTransition := GetUniform(fadeShader, "iTransition")
	fadebackgroundShaderResolution := GetUniform(fadeShader, "iResolution")

	rl.BeginTextureMode(earthTexture)
	game.Earth.Render(render)
	rl.EndTextureMode()
	rl.BeginTextureMode(stationTexture)
	game.Station.Render(render)
	rl.EndTextureMode()

	rl.SetTargetFPS(60)
	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w := game.Update(float32(dt))

		if w.IsStation {
			rl.BeginTextureMode(stationTexture)
			w.Render(render)
			rl.EndTextureMode()
		} else {
			rl.BeginTextureMode(earthTexture)
			w.Render(render)
			rl.EndTextureMode()
		}

		rl.BeginTextureMode(displayTexture)
		rl.BeginShaderMode(fadeShader)
		if w.IsStation {
			fadeiChannel0Location.SetTexture(earthTexture.Texture)
			fadeiChannel1Location.SetTexture(stationTexture.Texture)
			fadebackgroundShaderTransition.SetFloat(game.TeleportTransition)
		} else {
			fadeiChannel0Location.SetTexture(stationTexture.Texture)
			fadeiChannel1Location.SetTexture(earthTexture.Texture)
			fadebackgroundShaderTransition.SetFloat(1 - game.TeleportTransition)
		}
		fadeiChannelPrevLocation.SetTexture(displayTexture.Texture)
		fadebackgroundShaderTime.SetFloat(float32(t))
		fadebackgroundShaderResolution.SetVec2(float32(renderWidth), float32(renderHeight))
		rl.DrawRectangle(0, 0, renderWidth, renderHeight, rl.White)

		rl.EndShaderMode()
		rl.EndTextureMode()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.DrawTexturePro(
			displayTexture.Texture,
			rl.Rectangle{X: 0, Y: 0, Width: float32(renderWidth), Height: -float32(renderHeight)},
			rl.Rectangle{X: 0, Y: 0, Width: float32(screenWidth), Height: float32(screenHeight)},
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)

		rl.DrawRectangle(0, screenHeight-20, int32(math.Mod(game.Day, 1)*float64(screenWidth)), 20, rl.White)
		rl.DrawRectangle(int32(screenWidth*8./24.), screenHeight-40, 20, 20, rl.Red)
		rl.DrawRectangle(int32(screenWidth*20./24.), screenHeight-40, 20, 20, rl.Red)

		clock := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(int64(game.Day * 24 * float64(time.Hour))))
		rl.DrawTextEx(rl.GetFontDefault(), clock.Format("15:04"), rl.NewVector2(float32(screenWidth)/2-45, 10), 40, 2, rl.White)

		// w.Player.RenderHud(w)
		rl.DrawFPS(10, 20)

		rl.EndDrawing()

	}
}

func DrawLine(col color.RGBA, ps ...cp.Vector) {
	if len(ps) != 0 {

		rl.DrawSphere(VecFrom2D(ps[0], 2).Vector3, 0.3, col)

	}

	for i := 0; i < len(ps)-1; i++ {
		p1 := VecFrom2D(ps[i], 2)
		p2 := VecFrom2D(ps[i+1], 2)
		rl.DrawSphere(p2.Vector3, 0.3, col)
		rl.DrawLine3D(p1.Vector3, p2.Vector3, col)
	}

}
