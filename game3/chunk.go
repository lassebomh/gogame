package game3

import (
	"fmt"
	"math"
	"path/filepath"

	"game/model"
	"game/vec3"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const CHUNKS_PATH = "./chunks/"

const ChunkWidth = 8
const ChunkHeight = 8
const FloorWidth = 0.15

type ChunkPos struct {
	X int
	Z int
}

type LocalPos struct {
	Z uint
	X uint
	Y uint
}

func WorldToChunk(v vec3.Value) (ChunkPos, LocalPos) {
	xi, xf := math.Modf(v.X / ChunkWidth)
	zi, zf := math.Modf(v.Z / ChunkWidth)
	chunkPos := ChunkPos{X: int(xi), Z: int(zi)}
	localPos := LocalPos{X: uint(xf * ChunkWidth), Y: uint(v.Y), Z: uint(zf * ChunkWidth)}
	return chunkPos, localPos
}

func ChunkToWorld(c ChunkPos, l LocalPos) vec3.Value {
	return vec3.XYZ(
		float64(c.X*ChunkWidth)+float64(l.X),
		float64(l.Y),
		float64(c.X*ChunkWidth)+float64(l.Z),
	)
}

type Chunk struct {
	Cells [ChunkHeight][ChunkWidth][ChunkWidth]Cell

	top      [ChunkWidth][ChunkWidth]*Cell
	models   [ChunkHeight]rl.Model
	position ChunkPos
}

func (c *Chunk) Init(position ChunkPos) *Chunk {
	if c == nil {
		c = &Chunk{}
	}
	c.position = position

	game.Level.Chunks[position] = c
	return c
}

func (c *Chunk) UpsertModels() {
	path := filepath.Join(CHUNKS_PATH, fmt.Sprintf("chunk_%d_%d__%d.obj", c.position.X, c.position.Z, 0))
	model := model.NewModel(5, path, "./chunks/chunk.mtl", "Material")

	y := 0

	for x := range ChunkWidth {
		for z := range ChunkWidth {
			cell := &c.Cells[y][x][z]
			pos := vec3.XYZ(float64(x), 0, float64(z))
			if cell.Faces[FaceDown].Type == FaceSolid {
				model.Cube(pos, pos.Add(vec3.XYZ(1, -FloorWidth, 1)), 0, 2)
			}
		}
	}

	model.Export()

	c.models[0] = rl.LoadModel(path)

}

type FaceType int32
type FaceDirection int32

const (
	FaceNone = FaceType(iota)
	FaceSolid
	FaceStair
	FaceDoor
)

const (
	FaceDown = FaceDirection(iota)
	FaceWest
	FaceNorth
	FaceEast
	FaceSouth
)

type Face struct {
	Type      FaceType
	Direction FaceDirection
}

type Cell struct {
	Faces [5]Face
}

const AtlasTiles = 5.
