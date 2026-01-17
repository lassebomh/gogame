package game2

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
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

	cellx := ((int(X)%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	cellz := ((int(Z)%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	celly := int(math.Floor(Y))

	return &chunk.Cells[cellx][cellz][celly]
}

func (l *Level) Draw(g *Game) {
	for _, chunk := range l.Chunks {

		pos := NewVec3(
			float64(chunk.X)*float64(CHUNK_WIDTH),
			0,
			float64(chunk.Z)*float64(CHUNK_WIDTH),
		)

		for x := range CHUNK_WIDTH {
			for z := range CHUNK_WIDTH {
				for y := range CHUNK_HEIGHT {
					cell := &chunk.Cells[x][z][y]

					cellPos := pos.Add(NewVec3(float64(x)+0.5, float64(y)+0.5, float64(z)+0.5))

					if cell.North.Type == FaceWall {
						rl.DrawModelEx(g.Models["wall"], cellPos.Raylib(), Y.Raylib(), 0, XYZ.Raylib(), rl.White)
					}

					if cell.East.Type == FaceWall {
						rl.DrawModelEx(g.Models["wall"], cellPos.Raylib(), Y.Raylib(), 270, XYZ.Raylib(), rl.White)
					}

					if cell.South.Type == FaceWall {
						rl.DrawModelEx(g.Models["wall"], cellPos.Raylib(), Y.Raylib(), 180, XYZ.Raylib(), rl.White)
					}

					if cell.West.Type == FaceWall {
						rl.DrawModelEx(g.Models["wall"], cellPos.Raylib(), Y.Raylib(), 90, XYZ.Raylib(), rl.White)
					}

					// cell.Top.Type == FaceWall || cell.Bottom.Type == FaceWall {
					// 	rl.DrawCube(pos.Add(NewVec3(float64(x), float64(y), float64(z))).Raylib(), 1, 1, 1, rl.White)
					// }
				}
			}
		}
	}
}
