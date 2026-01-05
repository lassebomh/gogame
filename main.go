package main

import (
	. "game/game"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

func main() {
	screenWidth := int32(1200)
	screenHeight := int32(800)

	pixelScale := int32(4)
	renderWidth := screenWidth / pixelScale
	renderHeight := screenHeight / pixelScale

	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowUnfocused)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(screenWidth, screenHeight, "raylib")

	if rl.GetMonitorCount() > 1 {
		pos := rl.GetMonitorPosition(1)
		rl.SetWindowPosition(int(pos.X), int(pos.Y))
	}

	tilemap := NewTilemap(40, 40, 7)

	// GenerateMaze(tilemap, 10, 10, 10, 10)
	// tilemap.Cols[18][19].Wall = 0
	roomWidth := 5
	roomHeight := 5
	rooms := 5
	for x := range rooms {
		for y := range rooms {
			tilemap.CreateRoom(x*roomWidth+3, y*roomHeight+3, roomWidth, roomHeight, WALL_R|WALL_L|WALL_B|WALL_T)
		}
	}

	tilemap.Cols[25][27].Door = WALL_B

	w := NewWorld(tilemap)
	tilemap.GenerateBodies(w)

	shadowWidth, shadowHeight := int32(1000), int32(1000)
	shadowTexture := rl.LoadRenderTexture(shadowWidth, shadowHeight)
	defer rl.UnloadRenderTexture(shadowTexture)

	renderTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(renderTexture)
	rl.SetTextureFilter(renderTexture.Texture, rl.FilterPoint)

	shader := rl.LoadShader("./glsl330/lighting.vs", "./glsl330/lighting.fs")
	// shadowShader := rl.LoadShader("./glsl330/shadow.vs", "./glsl330/shadow.fs")

	render := NewRender(shader)

	// render.NewLight(LIGHT_POINT, rl.NewVector3(-2, 1, -2), rl.NewVector3(0, 0, 0), rl.Yellow, 1)
	// render.NewLight(LIGHT_POINT, rl.NewVector3(2, 1, 2), rl.NewVector3(0, 0, 0), rl.Red, 1)
	flashlight := render.NewLight(LIGHT_SPOT, rl.NewVector3(-2, 1, 2), rl.NewVector3(0, 0, 0), rl.NewColor(255, 255, 100, 255), 2)
	render.NewLight(LIGHT_DIRECTIONAL, rl.NewVector3(2, 1, -2), rl.NewVector3(-0.2, -1, 0), rl.White, 0.5)

	Debug(render)

	*shader.Locs = rl.GetShaderLocation(shader, "viewPos")
	ambientLoc := rl.GetShaderLocation(shader, "ambient")
	shaderValue := []float32{0.1, 0.1, 0.1, 1.0}
	rl.SetShaderValue(shader, ambientLoc, shaderValue, rl.ShaderUniformVec4)

	render.UpdateValues()

	plane := rl.LoadModel("./models/plane.glb")
	plane.Materials.Shader = shader
	wall := rl.LoadModel("./models/cube.glb")
	wall.Materials.Shader = shader
	door := rl.LoadModel("./models/door.glb")
	door.Materials.Shader = shader

	monsterArm := rl.LoadModel("./models/monster/monster_arm_segment.glb")
	monsterArm.Materials.Shader = shader

	monsterBody := rl.LoadModel("./models/monster/monster_body.glb")
	monsterBody.Materials.Shader = shader
	// test := rl.LoadModel("./models/test/test.glb")

	rl.SetTargetFPS(60)

	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w.Update(float32(dt))
		playerPos := VecFrom2D(w.Player.Body.Position(), w.Player.Radius*2)
		lookDir := rl.NewVector3(float32(math.Cos(w.Player.Body.Angle())), 0, float32(math.Sin(w.Player.Body.Angle())))
		flashlight.Position = rl.Vector3Subtract(playerPos.Vector3, rl.Vector3Scale(lookDir, float32(w.Player.Radius)*3))
		flashlight.Target = rl.Vector3Add(flashlight.Position, lookDir)

		render.UpdateValues()

		// lightCam := rl.Camera3D{
		// 	Position:   flashlight.Position,
		// 	Target:     flashlight.Target,
		// 	Up:         Y.Vector3,
		// 	Fovy:       30,
		// 	Projection: rl.CameraPerspective,
		// }

		// rl.BeginTextureMode(shadowTexture)
		// rl.ClearBackground(rl.Black)
		// rl.BeginMode3D(lightCam)
		// rl.BeginShaderMode(shadowShader)
		// oldWallShader := wall.Materials.Shader
		// oldPlaneShader := plane.Materials.Shader
		// wall.Materials.Shader = shadowShader
		// plane.Materials.Shader = shadowShader

		// for _, col := range w.Tilemap.Cols {
		// 	for _, tile := range col {
		// 		scale := float32(w.Tilemap.Scale)
		// 		pos := VecFrom2D(tile.WorldPosition, 0)
		// 		rl.DrawModel(plane, pos.Add(XZ.Scale(scale/2)).Vector3, scale/2, rl.White)
		// 		scaleVec := XYZ.Scale(scale * 0.5).Vector3
		// 		if tile.Wall&WALL_L != 0 {
		// 			rl.DrawModel(wall, pos.Add(NewVec(0, 0, scale)).Vector3, scale/2, rl.White)
		// 		}
		// 		if tile.Wall&WALL_T != 0 {
		// 			rl.DrawModelEx(wall, pos.Add(NewVec(scale, 0, scale*w.Tilemap.WallDepthRatio)).Vector3, Y.Vector3, 90, scaleVec, rl.RayWhite)
		// 		}
		// 		if tile.Wall&WALL_R != 0 {
		// 			rl.DrawModelEx(wall, pos.Add(NewVec(scale, 0, 0)).Vector3, Y.Scale(1).Vector3, 180, scaleVec, rl.RayWhite)
		// 		}
		// 		if tile.Wall&WALL_B != 0 {
		// 			rl.DrawModelEx(wall, pos.Add(NewVec(0, 0, scale*(1-w.Tilemap.WallDepthRatio))).Vector3, Y.Vector3, 270, scaleVec, rl.RayWhite)
		// 		}
		// 	}
		// }
		// wall.Materials.Shader = oldWallShader
		// plane.Materials.Shader = oldPlaneShader

		rl.EndShaderMode()
		rl.EndMode3D()
		rl.EndTextureMode()

		rl.BeginTextureMode(renderTexture)
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(w.Camera)

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
					rl.DrawModelEx(door, VecFrom2D(tile.DoorBody.Position(), 0).Vector3, Y.Vector3, 0, XYZ.Scale(scale/2).Vector3, rl.White)
				}
			}
		}

		// rl.DrawModelEx(taest, playerPos.Add(NewVec(0, 4, 0)).Vector3, Y.Vector3, float32(t)*30, XYZ.Scale(3).Vector3, rl.White)

		rl.DrawSphereEx(playerPos.Vector3, float32(w.Player.Radius), 12, 12, rl.Red)

		for _, pitem := range w.Items {
			rl.DrawSphereEx(VecFrom2D(pitem.Body.Position(), pitem.Radius).Vector3, float32(pitem.Radius), 12, 12, rl.Red)
		}

		if w.Monster != nil {

			monsterRadius := float32(w.Monster.Radius)
			rl.DrawModelEx(
				monsterBody,
				VecFrom2D(w.Monster.Body.Position(), w.Monster.Radius).Vector3,
				Y.Vector3,
				float32(-w.Monster.Body.Angle())*rl.Rad2deg,
				NewVec(monsterRadius*1, monsterRadius, monsterRadius*1).Vector3,
				rl.DarkGray,
			)

			for _, arm := range w.Monster.Arms {
				for _, segment := range arm.Segments {
					rl.DrawModelEx(
						monsterArm,
						VecFrom2D(segment.Body.Position(), w.Monster.Radius).Vector3,
						Y.Vector3,
						float32(-segment.Body.Angle())*rl.Rad2deg,
						NewVec(float32(segment.Length)*1.1, float32(segment.Width), float32(segment.Width)).Vector3,
						rl.DarkGray,
					)
				}
			}
		}

		// rl.DrawRenderBatchActive()
		// rl.DisableDepthTest()
		// // cp.DrawSpace(w.Space, w.PhysicsDrawer)
		// DrawLine(rl.Red, w.Monster.Path...)
		// if w.Monster != nil {
		// 	for _, arm := range w.Monster.Arms {
		// 		DrawLine(rl.Blue, arm.Segments[len(arm.Segments)-1].Body.Position(), arm.TipTarget)
		// 	}
		// 	DrawLine(rl.Green, w.Monster.Body.Position(), w.Monster.Target)
		// }
		// rl.DrawRenderBatchActive()
		// rl.EnableDepthTest()

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
		// rl.DrawTexturePro(
		// 	shadowTexture.Texture,
		// 	rl.Rectangle{X: 0, Y: 0, Width: float32(shadowWidth), Height: -float32(shadowHeight)},
		// 	rl.Rectangle{X: 0, Y: 0, Width: float32(shadowWidth), Height: float32(shadowHeight)},
		// 	rl.Vector2{X: 0, Y: 0},
		// 	0,
		// 	rl.White,
		// )

		// w.Player.RenderHud(w)
		rl.DrawFPS(10, 20)
		rl.EndDrawing()

	}
	rl.CloseWindow()
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
