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

	shader := rl.LoadShader("./glsl330/lighting.vs", "./glsl330/lighting.fs")
	defer rl.UnloadShader(shader)

	render := NewRender(shader)
	defer render.Unload()

	renderTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(renderTexture)
	rl.SetTextureFilter(renderTexture.Texture, rl.FilterPoint)

	backgroundShader := rl.LoadShader("./glsl330/station.vs", "./glsl330/planet2.fs")
	defer rl.UnloadShader(backgroundShader)

	background := rl.LoadModel("./models/plane.glb")
	defer rl.UnloadModel(background)
	background.Materials.Shader = backgroundShader

	iChannel0 := rl.LoadTexture("./models/organic.png")
	rl.SetTextureWrap(iChannel0, rl.WrapRepeat)
	defer rl.UnloadTexture(iChannel0)

	iChannel1 := rl.LoadTexture("./models/earth_elevation.png")
	rl.SetTextureWrap(iChannel1, rl.WrapRepeat)
	defer rl.UnloadTexture(iChannel1)

	iChannel0Location := GetUniform(backgroundShader, "iChannel0")
	iChannel1Location := GetUniform(backgroundShader, "iChannel1")

	backgroundShaderTime := GetUniform(backgroundShader, "iTime")
	backgroundShaderFov := GetUniform(backgroundShader, "iFov")
	backgroundShaderResolution := GetUniform(backgroundShader, "iResolution")
	backgroundShaderResolution.SetVec2(float32(renderWidth), float32(renderHeight))

	rl.SetTargetFPS(60)
	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w := game.Update(float32(dt))

		rl.BeginTextureMode(renderTexture)
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(w.Camera)

		w.RenderEarth(render)

		backgroundShaderTime.SetFloat(float32(game.Day))
		backgroundShaderFov.SetFloat(float32(game.TeleportTransition * 30.0))
		iChannel0Location.SetTexture(iChannel0)
		iChannel1Location.SetTexture(iChannel1)

		cameraPos := Vec{Vector3: w.Camera.Position}
		cameraTarget := Vec{Vector3: w.Camera.Target}
		cameraDir := cameraTarget.Subtract(cameraPos).Normalize()

		angle := float32(math.Acos(float64(Z.DotProduct(cameraDir))))
		axis := Z.Normalize().CrossProduct(cameraDir).Normalize()
		scale := X.Scale(w.Camera.Fovy * float32(renderWidth) / float32(renderHeight)).Add(Z.Scale(w.Camera.Fovy)).Add(Y)
		rl.DrawModelEx(background, cameraPos.Add(cameraDir.Scale(w.Camera.Fovy*2)).Vector3, axis.Vector3, angle, scale.Vector3, rl.White)

		if DEBUG {
			rl.DrawRenderBatchActive()
			rl.DisableDepthTest()
			cp.DrawSpace(w.Space, game.PhysicsDrawer)
			if w.Monster != nil {
				DrawLine(rl.Red, w.Monster.Path...)
				if w.Monster != nil {
					for _, arm := range w.Monster.Arms {
						DrawLine(rl.Blue, arm.Segments[len(arm.Segments)-1].Body.Position(), arm.TipTarget)
					}
					DrawLine(rl.Green, w.Monster.Body.Position(), w.Monster.Target)
				}
			}
			rl.DrawRenderBatchActive()
			rl.EnableDepthTest()
		}
		rl.EndMode3D()
		rl.EndTextureMode()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawTexturePro(
			renderTexture.Texture,
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
