package game3

import (
	"encoding/gob"
	"game/vec2"
	"game/vec3"
	"log"
	"math"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

var game *Game

type Game struct {
	Time time.Duration
	Day  float64

	Earth   *World
	Station *World
}

func LoadGame(path string) {
	game = &Game{}

	file, err := os.Open(path)
	defer file.Close()

	if err == nil {
		decoder := gob.NewDecoder(file)
		err := decoder.Decode(game)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err != nil {
		log.Println("Save not found.")
	}

	game.Earth.Upsert(WorldEarth)
	game.Station.Upsert(WorldStation)

	if game.Station.Player != nil {
		game.Station.Player.Spawn(game.Station)
	} else if game.Earth.Player != nil {
		game.Earth.Player.Spawn(game.Earth)
	} else {
		SpawnNewPlayer(game.Earth)
	}

	if game.Station.Monster != nil {
		game.Station.Monster.Spawn(game.Station)
	} else if game.Earth.Monster != nil {
		game.Earth.Monster.Spawn(game.Earth)
	} else {
		SpawnNewMonster(game.Earth)
	}
}

func (g *Game) Update(dt time.Duration) {
	g.Time += dt
	g.Day = g.Time.Seconds() / (60 * 1)

	if g.Station.Player != nil {
		g.Station.Update(dt)
	} else {
		g.Earth.Update(dt)
	}
}

func (g *Game) Draw() {
	if g.Station.Player != nil {
		g.Station.Draw()
	} else {
		g.Earth.Draw()
	}
}

func (g *Game) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(g); err != nil {
		return err
	}

	return nil
}

type WorldType int32

const (
	WorldEarth = WorldType(iota)
	WorldStation
)

type World struct {
	TimeStep               time.Duration
	TimePhysicsAccumulator time.Duration

	Type   WorldType
	Chunks map[ChunkPos]*Chunk
	space  *cp.Space

	Player  *Player
	Monster *Monster
	Camera  Camera3D

	EditorActive bool
	Editor       *Editor

	MousePosition     vec2.Value
	MouseRayOrigin    vec3.Value
	MouseRayDirection vec3.Value

	renderTexture rl.RenderTexture2D
	shader        *MainShader

	planetShader         *PlanetShader
	planetTexture        rl.Texture2D
	planetOrganicTexture rl.Texture2D
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
		w.planetOrganicTexture = rl.LoadTexture("./models/organic.png")
		w.planetTexture = rl.LoadTexture("./models/earth_elevation.png")
		w.planetShader = NewShader(&PlanetShader{}, "", "./glsl330/planet2.fs")
	}
	w.space = cp.NewSpace()
	w.space.SetCollisionSlop(0.01)
	w.shader = NewShader(&MainShader{}, "./glsl330/lighting.vs", "./glsl330/lighting.fs")
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
		w.space.Step(PhysicsTickrate.Seconds())
		w.TimePhysicsAccumulator -= PhysicsTickrate
	}

	mousePos := rl.GetMousePosition()
	mouseRay := rl.GetScreenToWorldRay(mousePos, w.Camera.Raylib())

	w.MouseRayOrigin = vec3.FromRaylib(mouseRay.Position)
	w.MouseRayDirection = vec3.FromRaylib(mouseRay.Direction)

	if w.Player != nil {

		w.Player.Update()

		cameraDistance := 8.
		cameraDirection := vec3.XYZ(0, -5, 0.5).Normalize()

		w.Camera.Target = w.Camera.Target.Lerp(w.Player.Position, w.TimeStep.Seconds()*10)

		w.Camera = Camera3D{
			Position:   w.Camera.Target.Subtract(cameraDirection.Scale(cameraDistance)),
			Target:     w.Camera.Target,
			Up:         vec3.Y(1),
			Fovy:       45,
			Projection: rl.CameraPerspective,
		}
	}

	if w.Monster != nil {
		w.Monster.Update()
	}
}

func (w *World) Draw() {
	w.Player.UpdateView()

	BeginDrawing(func() {
		BeginTextureMode(w.renderTexture, func() {

			rl.ClearBackground(rl.Black)

			if w.Type == WorldStation {
				BeginShaderMode(w.planetShader, func() {
					w.planetShader.Time.Set(game.Day)
					w.planetShader.Fov.Set(30)
					w.planetShader.Channel0.Set(w.planetOrganicTexture)
					w.planetShader.Channel1.Set(w.planetTexture)
					w.planetShader.Resolution.Set(float64(w.renderTexture.Texture.Width), float64(w.renderTexture.Texture.Height))

					rl.DrawRectangle(0, 0, w.renderTexture.Texture.Width, w.renderTexture.Texture.Height, rl.White)
				})
			}

			BeginMode3D(w.Camera, func() {
				w.shader.Visibility.Set(1)

				if w.Player != nil {

					w.shader.ShadowMap.Set(w.Player.viewTexture.Texture)
					w.shader.PlayerPosition.SetVec3(w.Player.Position)
					w.shader.LightSpot(w.Player.Position.Add(vec3.Y(0.5)), w.Player.lookPosition.Add(vec3.Y(0.5)), 35, 40, rl.White, 1.5)
				}

				w.shader.LightDirectional(vec3.XYZ(0.3, -1, 0), rl.White, 1)

				w.shader.UpdateValues()

				targetChunkPos, targetLocalPos := WorldToChunk(w.Camera.Target)

				maxY := targetLocalPos.Y

				if w.Camera.Target.Y-math.Trunc(w.Camera.Target.Y) > 0.1 {
					maxY++
				}

				for y := 0; y <= maxY; y++ {

					visibility := 1.
					if w.Player != nil {
						visibility = w.Player.Position.Y - float64(y-1)
					}

					w.shader.Visibility.Set(visibility)

					for dx := -1; dx <= 1; dx++ {
						for dz := -1; dz <= 1; dz++ {
							chunkPos := ChunkPos{targetChunkPos.X - dx, targetChunkPos.Z - dz}
							chunk, ok := w.Chunks[chunkPos]
							if !ok {
								continue
							}

							if !rl.IsModelValid(chunk.models[y]) {
								continue
							}
							worldPos := ChunkToWorld(chunkPos, LocalPos{0, y, 0})
							rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
						}
					}
				}

				if w.Player != nil {
					w.Player.Draw()
				}
				if w.Monster != nil {
					w.Monster.Draw()

					// BeginOverlayMode(func() {
					// 	for _, arm := range w.Monster.arms {
					// 		if len(arm.path) >= 2 {
					// 			for i := range arm.path[:len(arm.path)-1] {
					// 				rl.DrawLine3D(arm.path[i].Raylib(), arm.path[i+1].Raylib(), rl.Yellow)
					// 			}
					// 		}

					// 	}
					// })
				}

			})
		})

		rl.DrawTexturePro(
			w.renderTexture.Texture,
			rl.Rectangle{X: 0, Y: 0, Width: float32(w.renderTexture.Texture.Width), Height: -float32(w.renderTexture.Texture.Height)},
			rl.Rectangle{X: 0, Y: 0, Width: float32(rl.GetScreenWidth()), Height: float32(rl.GetScreenHeight())},
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)
	})

}

func (w *World) UpsertCellChunk(pos vec3.Value) (*Cell, *Chunk) {
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

func (w *World) UpsertCell(pos vec3.Value) *Cell {
	cell, _ := w.UpsertCellChunk(pos)
	return cell

}

func (w *World) GetCell(pos vec3.Value) (*Cell, bool) {
	cpos, lpos := WorldToChunk(pos)
	chunk, ok := w.Chunks[cpos]

	if !ok {
		return nil, false
	}

	cell := &chunk.Cells[lpos.Y][lpos.X][lpos.Z]

	return cell, true
}
