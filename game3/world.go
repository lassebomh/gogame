package game3

import (
	v2 "game/vec2"
	v3 "game/vec3"
	"image/color"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type WorldType int32

const (
	WorldEarth = WorldType(iota)
	WorldStation
)

type World struct {
	TimeStep               time.Duration
	TimePhysicsAccumulator time.Duration

	other *World

	Type   WorldType
	Chunks map[ChunkPos]*Chunk
	space  *cp.Space

	Player  *Player
	Monster *Monster
	Camera  Camera3D

	EditorActive bool
	Editor       *Editor

	MousePosition     v2.Value
	MouseRayOrigin    v3.Value
	MouseRayDirection v3.Value

	renderTexture rl.RenderTexture2D

	raycastResults []RaycastResult
}

func (w *World) Upsert(worldType WorldType) *World {
	if w == nil {
		w = &World{}
	}
	w.Type = worldType

	if w.Type == WorldEarth {
		game.Earth = w
	} else {
		game.Station = w
	}
	w.raycastResults = make([]RaycastResult, 0)
	w.space = cp.NewSpace()
	w.space.SetCollisionSlop(0.005)
	renderWidth, renderHeight := rl.GetRenderWidth(), rl.GetRenderHeight()
	w.renderTexture = rl.LoadRenderTexture(int32(renderWidth/4), int32(renderHeight/4))

	if w.Chunks == nil {
		w.Chunks = make(map[ChunkPos]*Chunk)
	}
	for _, chunk := range w.Chunks {
		chunk.world = w
		chunk.Reload()
	}
	w.Editor.Upsert(w)

	return w
}

func (w *World) Update(dt time.Duration) {
	w.TimeStep = dt
	w.TimePhysicsAccumulator += dt

	for w.TimePhysicsAccumulator >= PhysicsTickrate {
		w.TimePhysicsAccumulator -= PhysicsTickrate

		w.space.Step(PhysicsTickrate.Seconds())

		w.raycastResults = w.raycastResults[:0]

		mousePos := rl.GetMousePosition()
		mouseRay := rl.GetScreenToWorldRay(mousePos, w.Camera.Raylib())

		w.MouseRayOrigin = v3.FromRaylib(mouseRay.Position)
		w.MouseRayDirection = v3.FromRaylib(mouseRay.Direction)

		if w.Player != nil {

			cameraDistance := 30.
			cameraDirection := v3.XYZ(0, -5, 1).Normalize()

			w.Camera.Target = w.Camera.Target.Lerp(w.Player.Position, w.TimeStep.Seconds()*10)

			w.Camera = Camera3D{
				Position:   w.Camera.Target.Subtract(cameraDirection.Scale(cameraDistance)),
				Target:     w.Camera.Target,
				Up:         v3.Y(1),
				Fovy:       15,
				Projection: rl.CameraPerspective,
			}

			w.Player.Update()
		}

		if w.Monster != nil {
			w.Monster.Update()
		}
	}
}

func (w *World) Draw() {
	w.Player.UpdateView()

	BeginTextureMode(w.renderTexture, func() {

		rl.ClearBackground(rl.Black)

		if w.Type == WorldStation {
			BeginShaderMode(globals.Shaders.Planet, func() {
				globals.Shaders.Planet.Time.Set(game.Day)
				globals.Shaders.Planet.Fov.Set(20) //. + math.Cos(game.Day*math.Pi*2)*25)
				globals.Shaders.Planet.Channel0.Set(globals.Textures.Organic)
				globals.Shaders.Planet.Channel1.Set(globals.Textures.PlanetElevation)
				globals.Shaders.Planet.Resolution.Set(float64(w.renderTexture.Texture.Width), float64(w.renderTexture.Texture.Height))

				rl.DrawRectangle(0, 0, w.renderTexture.Texture.Width, w.renderTexture.Texture.Height, rl.White)
			})
		}

		BeginMode3D(w.Camera, func() {
			globals.Shaders.Main.Visibility.Set(1)

			if w.Player != nil {

				globals.Shaders.Main.ShadowMap.Set(w.Player.viewTexture.Texture)
				globals.Shaders.Main.PlayerPosition.SetVec3(w.Player.Position)
				globals.Shaders.Main.LightSpot(w.Player.Position.Add(v3.Y(0.2)), w.Player.lookPosition.Add(v3.Y(0.2)), 30, 35, rl.White, 1.5)
			} else {
				globals.Shaders.Main.LightSpot(v3.Zero, v3.One, 0, 1, rl.White, 0)
			}

			globals.Shaders.Main.LightDirectional(v3.XYZ(0.3, -1, 0), color.RGBA{180, 190, 255, 255}, 0.7)

			targetChunkPos, targetLocalPos := WorldToChunk(w.Camera.Target)
			chunks := make([]*Chunk, 0, 9)
			for dx := -1; dx <= 1; dx++ {
				for dz := -1; dz <= 1; dz++ {
					chunkPos := ChunkPos{targetChunkPos.X - dx, targetChunkPos.Z - dz}
					chunk, ok := w.Chunks[chunkPos]
					if ok {
						chunks = append(chunks, chunk)
					}
				}
			}

			for _, chunk := range chunks {
				chunk.UpdateLights()
			}

			globals.Shaders.Main.UpdateLights()

			maxY := targetLocalPos.Y

			if w.Camera.Target.Y-math.Trunc(w.Camera.Target.Y) > 0.1 {
				maxY++
			}

			for y := 0; y <= maxY; y++ {

				visibility := 1.
				if w.Player != nil {
					visibility = w.Player.Position.Y - float64(y-1)
				}

				globals.Shaders.Main.Visibility.Set(visibility)

				for _, chunk := range chunks {

					worldPos := ChunkToWorld(chunk.Position, LocalPos{0, y, 0})
					rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
				}
			}

			if w.Player != nil {
				w.Player.Draw()
			}
			if w.Monster != nil {
				w.Monster.Draw()

				// BeginOverlayMode(func() {
				// 	for _, arm := range w.Monster.arms {
				// 		DrawPath(arm.path)
				// 	}
				// })
			}

		})

		t := game.Day * 24 * float64(time.Hour)
		t /= (10 * float64(time.Minute))
		t = math.Floor(t)
		t *= (10 * float64(time.Minute))

		clock := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(int64(t))).Format("15:04")
		rl.DrawText(clock, 12, 12, 10, rl.Green)
	})

	rl.DrawTexturePro(
		w.renderTexture.Texture,
		rl.Rectangle{X: 0, Y: 0, Width: float32(w.renderTexture.Texture.Width), Height: -float32(w.renderTexture.Texture.Height)},
		rl.Rectangle{X: 0, Y: 0, Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())},
		rl.Vector2{X: 0, Y: 0},
		0,
		rl.White,
	)

}

func (w *World) UpsertCellChunk(pos v3.Value) (*Cell, *Chunk) {
	cpos, lpos := WorldToChunk(pos)

	chunk, ok := w.Chunks[cpos]

	if !ok {
		chunk = &Chunk{
			Position: cpos,
			world:    w,
		}
		w.Chunks[cpos] = chunk
		chunk.Reload()
	}

	cell := &chunk.Cells[lpos.Y][lpos.X][lpos.Z]

	return cell, chunk
}

func (w *World) UpsertCell(pos v3.Value) *Cell {
	cell, _ := w.UpsertCellChunk(pos)
	return cell

}

func (w *World) GetCell(pos v3.Value) (*Cell, bool) {
	if pos.Y < 0 || pos.Y >= ChunkHeight {
		return nil, false
	}

	cpos, lpos := WorldToChunk(pos)
	chunk, ok := w.Chunks[cpos]

	if !ok {
		return nil, false
	}

	cell := &chunk.Cells[lpos.Y][lpos.X][lpos.Z]

	return cell, true
}

type RaycastResult struct {
	ID       int
	Start    v3.Value
	End      v3.Value
	Position v3.Value
	Normal   v2.Value
	Alpha    float64
	Hit      bool
}

func (w *World) Raycast(start v3.Value, end v3.Value) (result RaycastResult) {
	result.Start = start
	result.End = end
	result.Position = end

	if int(start.Y+0.4) == int(end.Y+0.4) {
		queryResult := w.space.SegmentQueryFirst(
			start.Chipmunk(),
			end.Chipmunk(),
			0,
			cp.NewShapeFilter(
				0,
				Category(start.Y, true, false),
				Category(start.Y, true, false),
			),
		)
		result.Hit = queryResult.Shape != nil
		result.Position = v3.XYZ(queryResult.Point.X, cp.Lerp(start.Y, end.Y, queryResult.Alpha), queryResult.Point.Y)
		result.Alpha = queryResult.Alpha
		result.Normal = v2.Value(queryResult.Normal)
	} else {
		result.Hit = true
	}

	w.raycastResults = append(w.raycastResults, result)

	return
}

func (w *World) GetPathTarget(from v3.Value, to v3.Value) (v3.Value, float64) {

	pathPoint := NewPathPoint(w, from)
	path, length := pathPoint.FindPath(to)
	if len(path) == 0 {
		return from, 0
	}

	first, path := path[0], path[1:]

	positions := make([]v3.Value, len(path))
	for i, point := range path {
		if i == len(path)-1 {
			positions[i] = to
		} else {
			positions[i] = point.Center
		}
	}

	if len(path) == 0 {
		return from, 0
	}

	if !w.Raycast(first.Center, to).Hit {
		return to, first.Center.Distance(to)
	}

	i := -1
	low, high := 0, 0

	for i := range len(path) - 2 {
		point := path[i]
		if point.Cell.Faces[FaceDown].Type == FaceStair {
			break
		}
		high = i
	}

	for low <= high {
		mid := low + (high-low)/2

		if !w.Raycast(first.Center, positions[mid]).Hit {
			i = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	if i == -1 {
		return positions[0], length
	} else {
		return positions[i], length
	}
}
