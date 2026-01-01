package main

import (
	"fmt"
	. "game/game"
	. "game/lib"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

func main() {
	screenWidth := int32(1200)
	screenHeight := int32(800)

	pixelScale := int32(1)
	renderWidth := screenWidth / pixelScale
	renderHeight := screenHeight / pixelScale

	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowUnfocused)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(screenWidth, screenHeight, "raylib")

	if rl.GetMonitorCount() > 1 {
		pos := rl.GetMonitorPosition(1)
		rl.SetWindowPosition(int(pos.X)+50, int(pos.Y)+50)
	}

	tilemap := NewTilemap(40, 40, 7)

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

	cam := rl.Camera3D{}
	cam.Fovy = 70
	cam.Position = rl.Vector3{X: 0, Y: 2, Z: 0}
	cam.Target = rl.Vector3{X: 0, Y: 0, Z: -0.2}
	cam.Projection = rl.CameraOrthographic
	cam.Up = rl.Vector3{X: 0, Y: 1, Z: 0}

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

	monsterArm := rl.LoadModel("./models/monster/cube3d.glb")
	monsterArm.Materials.Shader = shader

	monsterBody := rl.LoadModel("./models/monster/monster_body.glb")
	monsterBody.Materials.Shader = shader

	monsterArmMat := &monsterArm.GetMaterials()[0]
	monsterArmMat.Shader = shader
	p.TextureMapAlbedo(monsterArmMat, rl.LoadTexture("./models/monster/Segment.png"))
	p.TextureMapNormal(monsterArmMat, rl.LoadTexture("./models/monster/Segment_normal.png"))
	p.TextureMapMetalness(monsterArmMat, rl.LoadTexture("./models/monster/Segment_mra.png"))

	monsterBodyMat := &monsterBody.GetMaterials()[0]
	monsterBodyMat.Shader = shader
	p.TextureMapAlbedo(monsterBodyMat, rl.LoadTexture("./models/monster/Segment.png"))
	p.TextureMapNormal(monsterBodyMat, rl.LoadTexture("./models/monster/Segment_normal.png"))
	p.TextureMapMetalness(monsterBodyMat, rl.LoadTexture("./models/monster/Segment_mra.png"))

	rl.SetTargetFPS(60)

	t := rl.GetTime()

	for !rl.WindowShouldClose() {

		dt := rl.GetTime() - t
		t = rl.GetTime()

		w.Update(float32(dt))

		playerPos := w.Player.Body.Position()
		fmt.Printf("%+v\n", playerPos)
		position := rl.Vector3{X: float32(playerPos.X), Y: .1, Z: float32(playerPos.Y)}

		cam.Target = rl.Vector3Add(position, rl.Vector3{Y: 0, Z: 0})
		// cam.Position = rl.Vector3Add(position, rl.Vector3{X: float32(math.Cos(t/5)) * 55, Y: 55, Z: float32(math.Sin(t/5)) * 55})
		cam.Position = rl.Vector3Add(position, rl.Vector3{X: 0, Y: 50, Z: 20})

		l3.Position = rl.Vector3Add(position, rl.Vector3{X: 0, Y: 10, Z: 0})
		l3.UpdateValues()

		p.UpdateByCamera(cam.Position)

		rl.BeginTextureMode(renderTexture)
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(cam)

		p.AmbientColor(rl.Vector3{X: 1, Y: 1, Z: 1}, 0.1)

		for _, col := range w.Tilemap.Cols {
			for _, tile := range col {

				scale := float32(w.Tilemap.Scale)
				pos := rl.Vector3{X: float32(tile.X) * scale, Y: 0, Z: float32(tile.Y) * scale}

				rl.DrawModel(plane, rl.Vector3Add(pos, rl.Vector3{X: scale / 2, Y: 0, Z: scale / 2}), scale/2, rl.White)

				if tile.Wall&WALL_L != 0 {
					wallPos := rl.Vector3Add(pos, rl.NewVector3(0, 0, float32(w.Tilemap.Scale)))
					rl.DrawModel(wall, wallPos, float32(w.Tilemap.Scale)/2, rl.White)
				}
				if tile.Wall&WALL_T != 0 {
					wallPos := rl.Vector3Add(pos, rl.NewVector3(float32(w.Tilemap.Scale), 0, float32(w.Tilemap.Scale)*w.Tilemap.WallDepthRatio))
					rl.DrawModelEx(
						wall,
						wallPos,
						rl.Vector3{
							X: 0,
							Y: 1,
							Z: 0,
						},
						90,
						rl.Vector3{float32(w.Tilemap.Scale) / 2, float32(w.Tilemap.Scale) / 2, float32(w.Tilemap.Scale) / 2},
						rl.RayWhite,
					)
				}

				if tile.Wall&WALL_R != 0 {
					wallPos := rl.Vector3Add(pos, rl.Vector3{float32(w.Tilemap.Scale), 0, 0})
					rl.DrawModelEx(
						wall,
						wallPos,
						rl.Vector3{
							X: 0,
							Y: 1,
							Z: 0,
						},
						180,
						rl.Vector3{float32(w.Tilemap.Scale) / 2, float32(w.Tilemap.Scale) / 2, float32(w.Tilemap.Scale) / 2},
						rl.RayWhite,
					)
				}
				if tile.Wall&WALL_B != 0 {
					wallPos := rl.Vector3Add(pos, rl.Vector3{0, 0, float32(w.Tilemap.Scale) * (1 - w.Tilemap.WallDepthRatio)})
					rl.DrawModelEx(
						wall,
						wallPos,
						rl.Vector3{
							X: 0,
							Y: 1,
							Z: 0,
						},
						270,
						rl.Vector3{float32(w.Tilemap.Scale) / 2, float32(w.Tilemap.Scale) / 2, float32(w.Tilemap.Scale) / 2},
						rl.RayWhite,
					)
				}
			}
		}

		l.DrawSpherelight(&l1)
		l.DrawSpherelight(&l2)
		l.DrawSpherelight(&l3)
		l.DrawSpherelight(&l4)

		rl.DrawSphereEx(position, w.Player.Radius, 12, 12, rl.Red)

		if w.Monster != nil {

			pos := w.Monster.Body.Position()
			rl.DrawModelEx(
				monsterBody,
				rl.NewVector3(float32(pos.X), float32(w.Monster.Radius), float32(pos.Y)),
				rl.NewVector3(0, 1, 0),
				float32(-w.Monster.Body.Angle())*rl.Rad2deg, rl.NewVector3(float32(w.Monster.Radius)*1.3, float32(w.Monster.Radius)/2, float32(w.Monster.Radius)),
				rl.White,
			)

			for _, arm := range w.Monster.Arms {
				for _, segment := range arm.Segments {
					pos := segment.Body.Position()

					rl.DrawModelEx(
						monsterArm,
						rl.NewVector3(float32(pos.X), float32(w.Monster.Radius), float32(pos.Y)),
						rl.NewVector3(0, 1, 0),
						float32(-segment.Body.Angle())*rl.Rad2deg,
						rl.Vector3{X: float32(segment.Length), Y: 1, Z: float32(segment.Width)},
						rl.White,
					)
				}
			}
		}

		rl.DrawGrid(5, 5)

		cp.DrawSpace(w.Space, w.PhysicsDrawer)

		rl.EndMode3D()

		rl.EndTextureMode()

		// Draw scaled-up texture to screen
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

		rl.DrawFPS(10, 20)
		rl.EndDrawing()
	}
	rl.CloseWindow()
}
