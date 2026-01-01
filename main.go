package main

import (
	. "game/game"
	. "game/lib"
	"image/color"

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

	tilemap := NewTilemap(2, 2, 7)

	// GenerateMaze(tilemap, 10, 10, 10, 10)
	// tilemap.Cols[18][19].Wall = 0

	// roomWidth := 5
	// roomHeight := 5
	// rooms := 5

	// for x := range rooms {
	// 	for y := range rooms {
	// 		tilemap.CreateRoom(x*roomWidth+3, y*roomHeight+3, roomWidth, roomHeight, WALL_R|WALL_L|WALL_B|WALL_T)
	// 	}
	// }

	// tilemap.CreateRoom(1, 1, 4, 4, WALL_R)
	// tilemap.CreateRoom(6, 1, 4, 4, WALL_R|WALL_L)
	// tilemap.CreateRoom(6, 6, 4, 4, WALL_R|WALL_L)
	// tilemap.CreateRoom(1, 6, 4, 4, WALL_R|WALL_L)

	w := NewWorld(tilemap)
	tilemap.GenerateBodies(w)

	// w.Player.Body.SetPosition(cp.Vector{3, 3})

	renderTexture := rl.LoadRenderTexture(renderWidth, renderHeight)
	defer rl.UnloadRenderTexture(renderTexture)
	rl.SetTextureFilter(renderTexture.Texture, rl.FilterPoint)

	shader := rl.LoadShader("./glsl330/pbr.vs", "./glsl330/pbr.fs")

	l := Light{}
	l.SetCombineShader(&shader)
	l.Init(0.0, rl.Vector3{X: 1, Y: 1, Z: 1})
	l1 := l.NewLight(LightTypePoint, rl.Vector3{X: 10, Y: 5, Z: 30}, rl.Vector3{}, rl.Yellow, 10, &l.Shader)
	l2 := l.NewLight(LightTypePoint, rl.Vector3{X: 2, Y: 5, Z: 1}, rl.Vector3{}, rl.Green, 10, &l.Shader)
	l3 := l.NewLight(LightTypePoint, rl.Vector3{X: 30, Y: 5, Z: 12}, rl.Vector3{}, rl.White, 12, &l.Shader)
	l4 := l.NewLight(LightTypePoint, rl.Vector3{X: 10, Y: 5, Z: 30}, rl.Vector3{}, rl.Blue, 10, &l.Shader)

	p := PhysicRender{}
	p.SetCombineShader(&shader)
	p.Init()

	p.UseTexNormal()
	p.UseTexMRA()
	p.UseTexAlbedo()
	p.SetTiling(rl.NewVector2(1, 1))

	plane := rl.LoadModel("./models/plane.glb")
	planeMat := &plane.GetMaterials()[0]
	planeMat.Shader = shader
	p.TextureMapAlbedo(planeMat, rl.LoadTexture("./models/ground_a.png"))
	p.TextureMapNormal(planeMat, rl.LoadTexture("./models/ground_n.png"))
	p.TextureMapMetalness(planeMat, rl.LoadTexture("./models/ground_mra.png"))

	wall := rl.LoadModel("./models/cube.glb")
	wall.Materials.Shader = shader
	wallMat := &wall.GetMaterials()[0]
	wallMat.Shader = shader
	p.TextureMapAlbedo(wallMat, rl.LoadTexture("./models/wall_a.png"))
	p.TextureMapNormal(wallMat, rl.LoadTexture("./models/wall_n.png"))
	p.TextureMapMetalness(wallMat, rl.LoadTexture("./models/wall_mra.png"))

	monsterArm := rl.LoadModel("./models/monster/monster_arm_segment.glb")
	monsterArm.Materials.Shader = shader
	monsterArmMat := &monsterArm.GetMaterials()[0]
	monsterArmMat.Shader = shader
	p.TextureMapAlbedo(monsterArmMat, rl.LoadTexture("./models/monster/Segment.png"))
	p.TextureMapNormal(monsterArmMat, rl.LoadTexture("./models/monster/Segment_normal.png"))
	p.TextureMapMetalness(monsterArmMat, rl.LoadTexture("./models/monster/Segment_mra.png"))

	monsterBody := rl.LoadModel("./models/monster/monster_body.glb")
	monsterBody.Materials.Shader = shader
	monsterBodyMat := &monsterBody.GetMaterials()[0]
	monsterBodyMat.Shader = shader
	p.TextureMapAlbedo(monsterBodyMat, rl.LoadTexture("./models/monster/Segment.png"))
	p.TextureMapNormal(monsterBodyMat, rl.LoadTexture("./models/monster/Segment_normal.png"))
	p.TextureMapMetalness(monsterBodyMat, rl.LoadTexture("./models/monster/Segment_mra.png"))

	house := rl.LoadModel("./models/monster/house.glb")
	house.Materials.Shader = shader

	rl.SetTargetFPS(60)

	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w.Update(float32(dt))

		playerPos := VecFrom2D(w.Player.Body.Position(), w.Player.Radius)

		l3.Position = playerPos.Add(NewVec(0, 10, 0)).Vector3
		// l3.UpdateValues()

		p.UpdateByCamera(w.Camera.Position)

		rl.BeginTextureMode(renderTexture)
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(w.Camera)

		p.AmbientColor(rl.Vector3{X: 1, Y: 1, Z: 1}, 0.1)

		for _, col := range w.Tilemap.Cols {
			for _, tile := range col {

				scale := float32(w.Tilemap.Scale)

				pos := VecFrom2D(tile.WorldPosition, 0)

				rl.DrawModel(plane, pos.Add(XZ.Scale(scale/2)).Vector3, scale/2, rl.White)

				scaleVec := XYZ.Scale(scale * 0.5).Vector3

				if tile.Wall&WALL_L != 0 {
					wallPos := pos.Add(NewVec(0, 0, scale))
					rl.DrawModel(wall, wallPos.Vector3, scale/2, rl.White)
				}
				if tile.Wall&WALL_T != 0 {
					wallPos := pos.Add(NewVec(scale, 0, scale*w.Tilemap.WallDepthRatio))
					rl.DrawModelEx(wall, wallPos.Vector3, Y.Vector3, 90, scaleVec, rl.RayWhite)
				}
				if tile.Wall&WALL_R != 0 {
					wallPos := pos.Add(NewVec(scale, 0, 0))
					rl.DrawModelEx(wall, wallPos.Vector3, Y.Scale(1).Vector3, 180, scaleVec, rl.RayWhite)
				}
				if tile.Wall&WALL_B != 0 {
					wallPos := pos.Add(NewVec(0, 0, scale*(1-w.Tilemap.WallDepthRatio)))
					rl.DrawModelEx(wall, wallPos.Vector3, Y.Vector3, 270, scaleVec, rl.RayWhite)
				}
			}
		}

		l.DrawSpherelight(&l1)
		l.DrawSpherelight(&l2)
		l.DrawSpherelight(&l3)
		l.DrawSpherelight(&l4)

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
				NewVec(monsterRadius*1.3, monsterRadius/2, monsterRadius*1.3).Vector3,
				rl.Gray,
			)

			for _, arm := range w.Monster.Arms {
				for _, segment := range arm.Segments {
					rl.DrawModelEx(
						monsterArm,
						VecFrom2D(segment.Body.Position(), w.Monster.Radius).Vector3,
						Y.Vector3,
						float32(-segment.Body.Angle())*rl.Rad2deg,
						NewVec(float32(segment.Length)*1.1, float32(segment.Width)/2, float32(segment.Width)).Vector3,
						rl.Gray,
					)
				}
			}
		}

		// rl.DrawGrid(5, 5)

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

		rl.DrawModelEx(house, (playerPos.Add(X.Scale(20))).Vector3, YZ.Normalize().Vector3, rl.Pi/4, XYZ.Scale(3).Vector3, rl.White)

		rl.EndMode3D()
		rl.EndTextureMode()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawTexturePro(
			renderTexture.Texture,
			rl.Rectangle{X: 0, Y: 0, Width: float32(renderWidth), Height: -float32(renderHeight)}, // Negative height to flip
			rl.Rectangle{X: 0, Y: 0, Width: float32(screenWidth), Height: float32(screenHeight)},
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)

		w.Player.RenderHud(w)
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
