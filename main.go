package main

import (
	. "game/game"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const DEBUG = false

func main() {

	screenWidth := int32(1700)
	screenHeight := int32(1000)

	renderHeight := int32(250)
	renderWidth := int32(float32(renderHeight) * (float32(screenWidth) / float32(screenHeight)))

	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowUnfocused | rl.FlagWindowUnfocused)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(screenWidth, screenHeight, "raylib")
	defer rl.CloseWindow()

	fadeShader := NewShader[struct {
		Channel0   UniformTexture `glsl:"iChannel0"`
		Channel1   UniformTexture `glsl:"iChannel1"`
		Transition UniformFloat   `glsl:"iTransition"`
		Resolution UniformVec2    `glsl:"iResolution"`
	}]("", "./glsl330/fade.fs")
	defer fadeShader.Unload()

	// shader := NewShader[struct {
	// 	Enabled [4]UniformInt `glsl:"lights[%d].enabled"`
	// }]("./glsl330/lighting.vs", "./glsl330/lighting.fs")

	// shader.Uniform.Enabled[0].Set(1)

	// Debug(shader)

	// return

	if rl.GetMonitorCount() > 1 {
		pos := rl.GetMonitorPosition(1)
		rl.SetWindowPosition(int(pos.X), int(pos.Y))
	}

	game := NewGame()

	stationRender := NewRender(renderWidth, renderHeight)
	defer stationRender.Unload()
	earthRender := NewRender(renderWidth, renderHeight)
	defer earthRender.Unload()

	// return

	earthTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(earthTexture)
	rl.SetTextureFilter(earthTexture.Texture, rl.FilterPoint)
	rl.SetTextureWrap(earthTexture.Texture, rl.WrapClamp)

	stationTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(stationTexture)
	rl.SetTextureFilter(stationTexture.Texture, rl.FilterPoint)
	rl.SetTextureWrap(stationTexture.Texture, rl.WrapClamp)

	hudTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(hudTexture)

	displayTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(displayTexture)
	rl.SetTextureFilter(displayTexture.Texture, rl.FilterPoint)

	font := rl.LoadFont("./fonts/setback.png")
	defer rl.UnloadFont(font)

	rl.SetTargetFPS(144)
	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w := game.Update(float32(dt))

		if w.IsStation {
			BeginTextureMode(stationTexture, func() {
				w.Render(stationRender)
			})
		} else {

			BeginTextureMode(earthTexture, func() {
				w.Render(earthRender)
			})

		}

		BeginTextureMode(displayTexture, func() {

			fadeShader.UseMode(func() {
				if w.IsStation {
					fadeShader.Uniform.Channel0.Set(earthTexture.Texture)
					fadeShader.Uniform.Channel1.Set(stationTexture.Texture)
					fadeShader.Uniform.Transition.Set(game.TeleportTransition)
				} else {
					fadeShader.Uniform.Channel0.Set(stationTexture.Texture)
					fadeShader.Uniform.Channel1.Set(earthTexture.Texture)
					fadeShader.Uniform.Transition.Set(1 - game.TeleportTransition)
				}
				fadeShader.Uniform.Resolution.Set(float64(renderWidth), float64(renderHeight))
				rl.DrawRectangle(0, 0, renderWidth, renderHeight, rl.White)
			})
		})

		BeginTextureMode(hudTexture, func() {
			rl.ClearBackground(color.RGBA{})
			w.Player.RenderHud(w, font)
		})

		BeginDrawing(func() {
			rl.ClearBackground(rl.Black)

			rl.DrawTexturePro(
				displayTexture.Texture,
				rl.Rectangle{X: 0, Y: 0, Width: float32(renderWidth), Height: -float32(renderHeight)},
				rl.Rectangle{X: 0, Y: 0, Width: float32(screenWidth), Height: float32(screenHeight)},
				rl.Vector2{X: 0, Y: 0},
				0,
				rl.White,
			)

			rl.DrawTexturePro(
				hudTexture.Texture,
				rl.Rectangle{X: 0, Y: 0, Width: float32(hudTexture.Texture.Width), Height: -float32(hudTexture.Texture.Height)},
				rl.Rectangle{X: 0, Y: 0, Width: float32(screenWidth), Height: float32(screenHeight)},
				rl.Vector2{X: 0, Y: 0},
				0,
				rl.White,
			)
			rl.DrawFPS(10, 20)
		})

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
