package game2

import (
	"fmt"
	"math"
)

type FaceType = uint8

const (
	FaceEmpty = FaceType(iota)
	FaceDoor
	FaceWall
)

type StairType = uint8

const (
	StairNone = StairType(iota)
	StairNorth
	StairSouth
	StairEast
	StairWest
)

type Stair struct {
	Type   StairType
	BaseUV Vec2
}

type Face struct {
	Type   FaceType
	BaseUV Vec2
}

type Cell struct {
	North  Face
	South  Face
	East   Face
	West   Face
	Top    Face
	Bottom Face

	Stair Stair
}

const CHUNK_WIDTH = int(64)
const CHUNK_HEIGHT = int(16)

type Cells = [CHUNK_WIDTH][CHUNK_WIDTH][CHUNK_HEIGHT]Cell

type Chunk struct {
	X     int
	Z     int
	Cells Cells
}

type Level struct {
	Chunks []*Chunk
}

func NewLevel() *Level {
	return &Level{
		Chunks: make([]*Chunk, 0),
	}
}

func (l *Level) GetCell(X float64, Y float64, Z float64) *Cell {

	var chunk *Chunk

	x := int(math.Floor(X / float64(CHUNK_WIDTH)))
	z := int(math.Floor(Z / float64(CHUNK_WIDTH)))

	for _, c := range l.Chunks {
		if x == c.X && z == c.Z {
			chunk = c
		}
	}

	if chunk == nil {
		chunk = &Chunk{
			X: x,
			Z: z,
		}
		l.Chunks = append(l.Chunks, chunk)

		fmt.Printf("New chunk %v %v\n", chunk.X, chunk.Z)
	}

	cellx := int(math.Floor(math.Abs(math.Mod(X, float64(CHUNK_WIDTH)))))
	cellz := int(math.Floor(math.Abs(math.Mod(Z, float64(CHUNK_WIDTH)))))
	celly := int(math.Floor(Y))

	return &chunk.Cells[cellx][cellz][celly]
}
