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

	if err == nil {
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
}

func (g *Game) Update(dt time.Duration) {
	g.Time += dt
	g.Earth.Update(dt)
}

func (g *Game) Draw() {
	g.Earth.Draw()
}

// func LoadSaveFromFile(path string, g *Game) error {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	decoder := gob.NewDecoder(file)
// 	if err := decoder.Decode(game); err != nil {
// 		return err
// 	}

// 	return nil
// }

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

// Seperate levels for station and world?
type World struct {
	TimeStep               time.Duration
	TimePhysicsAccumulator time.Duration

	Type   WorldType
	Chunks map[ChunkPos]*Chunk
	space  *cp.Space

	Player *Player
	Camera Camera3D

	EditorActive bool
	Editor       *Editor

	MousePosition     vec2.Value
	MouseRayOrigin    vec3.Value
	MouseRayDirection vec3.Value

	renderTexture rl.RenderTexture2D
	shader        *MainShader
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

		w.Camera = Camera3D{
			Position:   w.Player.Position.Subtract(cameraDirection.Scale(cameraDistance)),
			Target:     w.Player.Position,
			Up:         vec3.Y(1),
			Fovy:       45,
			Projection: rl.CameraPerspective,
		}
	}
}

func (w *World) Draw() {
	BeginTextureMode(w.renderTexture, func() {

		rl.ClearBackground(rl.Black)
		BeginMode3D(w.Camera, func() {
			w.shader.Visibility.Set(1)

			w.shader.Ambient.Set(1, 1, 1, 0.05)
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

				visibility := w.Player.Position.Y - float64(y-1)

				w.shader.Visibility.Set(visibility)

				for dx := -1; dx <= 1; dx++ {
					for dz := -1; dz <= 1; dz++ {
						chunkPos := ChunkPos{targetChunkPos.X - dx, targetChunkPos.Z - dz}

						chunk, ok := w.Chunks[chunkPos]

						if ok {
							worldPos := ChunkToWorld(chunkPos, LocalPos{0, y, 0})

							rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
						}
					}
				}

				// if above {
				// 	rl.EndBlendMode()
				// }
			}

			if w.Player != nil {
				w.Player.Draw()
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

}

func (w *World) Get(pos vec3.Value) (*Cell, *Chunk) {
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

func (w *World) GetCell(pos vec3.Value) *Cell {
	cell, _ := w.Get(pos)
	return cell

}
