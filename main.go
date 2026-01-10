package main

import (
	"fmt"
	. "game/game"
	"image/color"
	"math"

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

	// shadowWidth, shadowHeight := int32(1000), int32(1000)
	// shadowTexture := rl.LoadRenderTexture(shadowWidth, shadowHeight)
	// defer rl.UnloadRenderTexture(shadowTexture)

	renderTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(renderTexture)
	rl.SetTextureFilter(renderTexture.Texture, rl.FilterPoint)

	shader := rl.LoadShader("./glsl330/lighting.vs", "./glsl330/lighting.fs")
	defer rl.UnloadShader(shader)

	backgroundShader := rl.LoadShader("./glsl330/station.vs", "./glsl330/planet2.fs")
	defer rl.UnloadShader(backgroundShader)

	img := rl.LoadImage("./models/organic.png")
	defer rl.UnloadImage(img)

	iChannel0 := rl.LoadTextureFromImage(img)
	defer rl.UnloadTexture(iChannel0)

	rl.SetTextureWrap(iChannel0, rl.WrapRepeat)
	loc := rl.GetShaderLocation(backgroundShader, "iChannel0")
	backgroundShaderTime := GetUniform(backgroundShader, "iTime")
	backgroundShaderResolution := GetUniform(backgroundShader, "iResolution")
	backgroundShaderResolution.SetVec2(float32(renderWidth), float32(renderHeight))

	render := NewRender(shader)

	// flashlight := render.NewLight(LIGHT_SPOT, rl.NewVector3(-2, 1, 2), rl.NewVector3(0, 0, 0), rl.NewColor(255, 255, 100, 255), 2)
	// sun := render.NewLight(LIGHT_DIRECTIONAL, rl.NewVector3(2, 1, -2), rl.NewVector3(-0.2, -1, 0), color.RGBA{255, 230, 120, 255}, 0.5)

	*shader.Locs = rl.GetShaderLocation(shader, "viewPos")
	rl.SetShaderValue(shader, rl.GetShaderLocation(shader, "ambient"), []float32{0.1, 0.1, 0.1, 1.0}, rl.ShaderUniformVec4)

	render.UpdateValues()

	background := rl.LoadModel("./models/plane.glb")
	defer rl.UnloadModel(background)
	background.Materials.Shader = backgroundShader

	plane := rl.LoadModel("./models/plane.glb")
	defer rl.UnloadModel(plane)
	plane.Materials.Shader = shader
	wall := rl.LoadModel("./models/cube.glb")
	defer rl.UnloadModel(wall)
	wall.Materials.Shader = shader
	door := rl.LoadModel("./models/door.glb")
	defer rl.UnloadModel(door)
	door.GetMaterials()[1].Shader = shader

	monsterArm := rl.LoadModel("./models/monster/monster_arm_segment.glb")
	defer rl.UnloadModel(monsterArm)
	monsterArm.Materials.Shader = shader

	monsterBody := rl.LoadModel("./models/monster/monster_body.glb")
	defer rl.UnloadModel(monsterBody)
	monsterBody.Materials.Shader = shader

	rl.SetTargetFPS(60)
	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w := game.Update(float32(dt))

		playerPos := VecFrom2D(w.Player.Body.Position(), w.Player.Radius*1)

		w.RenderEarth(render)

		render.UpdateValues()

		rl.BeginTextureMode(renderTexture)
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(w.Camera)

		backgroundShaderTime.SetFloat(float32(game.Day))
		rl.SetShaderValueTexture(backgroundShader, loc, iChannel0)

		cameraPos := Vec{Vector3: w.Camera.Position}
		cameraTarget := Vec{Vector3: w.Camera.Target}
		cameraDir := cameraTarget.Subtract(cameraPos).Normalize()

		angle := float32(math.Acos(float64(Z.DotProduct(cameraDir))))
		axis := Z.Normalize().CrossProduct(cameraDir).Normalize()
		scale := X.Scale(w.Camera.Fovy * float32(renderWidth) / float32(renderHeight)).Add(Z.Scale(w.Camera.Fovy)).Add(Y)
		rl.DrawModelEx(background, cameraPos.Add(cameraDir.Scale(w.Camera.Fovy*2)).Vector3, axis.Vector3, angle, scale.Vector3, rl.White)

		for _, col := range w.Tilemap.Cols {
			for _, tile := range col {
				scale := float32(w.Tilemap.Scale)
				pos := VecFrom2D(tile.WorldPosition, 0)
				rl.DrawModel(plane, pos.Add(XZ.Scale(scale/2)).Vector3, scale/2, rl.White)
				scaleVec := XYZ.Scale(scale * 0.5).Vector3
				if tile.Wall&WALL_L != 0 {
					rl.DrawModel(wall, pos.Add(NewVec(0, 0, scale)).Vector3, scale/2, rl.White)
				}
				if tile.Wall&WALL_T != 0 {
					rl.DrawModelEx(wall, pos.Add(NewVec(scale, 0, scale*w.Tilemap.WallDepthRatio)).Vector3, Y.Vector3, 90, scaleVec, rl.RayWhite)
				}
				if tile.Wall&WALL_R != 0 {
					rl.DrawModelEx(wall, pos.Add(NewVec(scale, 0, 0)).Vector3, Y.Scale(1).Vector3, 180, scaleVec, rl.RayWhite)
				}
				if tile.Wall&WALL_B != 0 {
					rl.DrawModelEx(wall, pos.Add(NewVec(0, 0, scale*(1-w.Tilemap.WallDepthRatio))).Vector3, Y.Vector3, 270, scaleVec, rl.RayWhite)
				}
				if tile.DoorBody != nil {
					rl.DrawModelEx(door, VecFrom2D(tile.DoorBody.Position(), 0).Vector3, Y.Negate().Vector3, float32(tile.DoorBody.Angle()*rl.Rad2deg), scaleVec, rl.RayWhite)
				}
			}
		}

		rl.DrawSphereEx(playerPos.Vector3, float32(w.Player.Radius), 12, 12, rl.Red)

		for _, pitem := range w.Items {
			rl.DrawSphereEx(VecFrom2D(pitem.Body.Position(), pitem.Radius).Vector3, float32(pitem.Radius), 12, 12, rl.Red)
		}

		if w.Monster != nil {

			color := color.RGBA{R: 40, G: 40, B: 40, A: 255}

			monsterRadius := float32(w.Monster.Radius)
			rl.DrawModelEx(
				monsterBody,
				VecFrom2D(w.Monster.Body.Position(), w.Monster.Radius/2).Vector3,
				Y.Vector3,
				float32(-w.Monster.Body.Angle())*rl.Rad2deg,
				NewVec(monsterRadius*1, monsterRadius, monsterRadius*1).Vector3,
				color,
			)

			for _, arm := range w.Monster.Arms {
				for _, segment := range arm.Segments {
					rl.DrawModelEx(
						monsterArm,
						VecFrom2D(segment.Body.Position(), w.Monster.Radius*2-segment.Width).Vector3,
						Y.Vector3,
						float32(-segment.Body.Angle())*rl.Rad2deg,
						NewVec(float32(segment.Length)*1.1, float32(segment.Width), float32(segment.Width)).Vector3,
						color,
					)
				}
			}
		}

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

		rl.DrawText(fmt.Sprintf("%.1f", math.Mod(game.Day*24, 24)), 10, 100, 20, rl.RayWhite)

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
