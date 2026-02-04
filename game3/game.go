package game3

import (
	"encoding/gob"
	"game/vec2"
	"game/vec3"
	"log"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

var game *Game

type Game struct {
	TimeStep               time.Duration
	Time                   time.Duration
	TimePhysicsAccumulator time.Duration

	space *cp.Space

	Earth   *World
	Station *World

	activeEditor *Editor
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

	game.space = cp.NewSpace()

	game.Earth.Upsert(WorldEarth)
	game.Station.Upsert(WorldStation)

	if game.Station.Player != nil {
		player := NewPlayer(game.Station.Player)
		player.MoveToWorld(game.Station)
	} else {
		player := NewPlayer(game.Earth.Player)
		player.MoveToWorld(game.Earth)
	}
}

func (g *Game) Update(dt time.Duration) {
	g.TimeStep = dt
	g.Time += dt
	g.TimePhysicsAccumulator += dt

	for g.TimePhysicsAccumulator >= PhysicsTickrate {
		g.space.Step(PhysicsTickrate.Seconds())
		g.TimePhysicsAccumulator -= PhysicsTickrate
	}
	g.Earth.Update()
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
	Type   WorldType
	Chunks map[ChunkPos]*Chunk

	Player *Player
	Camera Camera3D

	Editor *Editor

	MousePosition     vec2.Value
	MouseRayOrigin    vec3.Value
	MouseRayDirection vec3.Value
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

	if w.Chunks == nil {
		w.Chunks = make(map[ChunkPos]*Chunk)
	}
	for _, chunk := range w.Chunks {
		chunk.Reload()
	}
	w.Editor.Upsert(w)

	return w
}

func (w *World) Update() {
	if w.Player != nil {

		w.Player.Update()

		cameraDistance := 5.
		cameraDirection := vec3.XYZ(0, -5, 1).Normalize()

		w.Camera = Camera3D{
			Position:   w.Player.Position.Subtract(cameraDirection.Scale(cameraDistance)),
			Target:     w.Player.Position,
			Up:         vec3.Y(1),
			Fovy:       60,
			Projection: rl.CameraPerspective,
		}
	}
}

func (w *World) Draw() {
	rl.ClearBackground(rl.Black)
	BeginMode3D(w.Camera, func() {

		targetChunkPos, targetLocalPos := WorldToChunk(w.Camera.Target)

		for dx := -1; dx <= 1; dx++ {
			for dz := -1; dz <= 1; dz++ {
				chunkPos := ChunkPos{targetChunkPos.X - dx, targetChunkPos.Z - dz}

				chunk, ok := w.Chunks[chunkPos]

				if ok {
					for y := 0; y < targetLocalPos.Y+1; y++ {
						worldPos := ChunkToWorld(chunkPos, LocalPos{0, y, 0})
						rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
					}
				}
			}
		}

		if w.Player != nil {
			w.Player.Draw()
		}

	})
}

func (w *World) GetCell(pos vec3.Value) (*Cell, *Chunk) {
	cpos, lpos := WorldToChunk(pos)

	chunk, ok := w.Chunks[cpos]

	if !ok {
		chunk = &Chunk{
			Position: cpos,
		}
		w.Chunks[cpos] = chunk
		chunk.Reload()
	}

	cell := &chunk.Cells[lpos.Y][lpos.X][lpos.Z]

	return cell, chunk
}
