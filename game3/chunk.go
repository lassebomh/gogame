package game3

import (
	"math"

	"game/vec3"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const CHUNKS_PATH = "./chunks/"

const (
	ChunkWidth  = 8
	ChunkHeight = 8
)

type ChunkPos struct {
	X int
	Z int
}

type LocalPos struct {
	X int
	Y int
	Z int
}

func WorldToChunk(v vec3.Value) (ChunkPos, LocalPos) {
	xi := math.Floor(v.X / float64(ChunkWidth))
	zi := math.Floor(v.Z / float64(ChunkWidth))
	xf := v.X - xi*ChunkWidth
	zf := v.Z - zi*ChunkWidth
	chunkPos := ChunkPos{X: int(xi), Z: int(zi)}
	localPos := LocalPos{X: int(xf), Y: int(v.Y), Z: int(zf)}
	return chunkPos, localPos
}

func ChunkToWorld(c ChunkPos, l LocalPos) vec3.Value {
	return vec3.XYZ(
		float64(c.X*ChunkWidth)+float64(l.X),
		float64(l.Y),
		float64(c.Z*ChunkWidth)+float64(l.Z),
	)
}

type Chunk struct {
	Cells    [ChunkHeight][ChunkWidth][ChunkWidth]Cell
	Position ChunkPos

	top    [ChunkWidth][ChunkWidth]*Cell
	models [ChunkHeight]rl.Model
}

func (c *Chunk) Reload() {
	for y := range ChunkHeight {
		if rl.IsModelValid(c.models[y]) {
			rl.UnloadModel(c.models[y])
		}
		c.models[y] = GenerateChunkYModel(c, y)
	}
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
