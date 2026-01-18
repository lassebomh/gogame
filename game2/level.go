package game2

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
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
	Type StairType
	// UV Vec4
}

type Face struct {
	Type  FaceType
	TileX int
	TileY int
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

type CellRef struct {
	Position Vec3
	Cell     *Cell
	Level    *Level
	Body     *cp.Body
}

const CHUNK_WIDTH = int(8)
const CHUNK_HEIGHT = int(16)

type Cells = [CHUNK_WIDTH][CHUNK_WIDTH][CHUNK_HEIGHT]Cell

type Chunk struct {
	X     int
	Z     int
	Cells Cells
}

type Level struct {
	Chunks []*Chunk

	CellRefs      map[Vec3]*CellRef
	CellRefsArray []*CellRef
}

type LevelSave struct {
	Chunks []Chunk
}

func (l *Level) ToSave() LevelSave {
	save := LevelSave{
		Chunks: make([]Chunk, 0),
	}

	for _, chunk := range l.Chunks {
		save.Chunks = append(save.Chunks, *chunk)
	}

	return save
}

func (save LevelSave) Load(g *Game) *Level {
	level := &Level{
		Chunks:        make([]*Chunk, 0),
		CellRefs:      make(map[Vec3]*CellRef, 0),
		CellRefsArray: make([]*CellRef, 0),
	}

	for _, chunk := range save.Chunks {
		level.Chunks = append(level.Chunks, &chunk)
	}

	return level
}

func (l *Level) GetCell(pos Vec3) *CellRef {

	ref := l.CellRefs[pos]

	if ref != nil {
		return ref
	}

	var chunk *Chunk

	x := int(math.Floor(pos.X / float64(CHUNK_WIDTH)))
	z := int(math.Floor(pos.Z / float64(CHUNK_WIDTH)))

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

	cellx := ((int(pos.X)%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	cellz := ((int(pos.Z)%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	celly := int(math.Floor(pos.Y))

	ref = &CellRef{
		Position: pos,
		Cell:     &chunk.Cells[cellx][cellz][celly],
		Level:    l,
		Body:     nil,
	}

	l.CellRefs[pos] = ref
	l.CellRefsArray = append(l.CellRefsArray, ref)

	return ref
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

					if !(cell.North.Type == FaceWall ||
						cell.East.Type == FaceWall ||
						cell.South.Type == FaceWall ||
						cell.West.Type == FaceWall ||
						cell.Top.Type == FaceWall ||
						cell.Bottom.Type == FaceWall) {
						continue
					}

					cellPos := pos.Add(NewVec3(float64(x)+0.5, float64(y)+0.5, float64(z)+0.5))

					cell.North.Draw(g, cellPos, Y, 0)
					cell.West.Draw(g, cellPos, Y, 90)
					cell.South.Draw(g, cellPos, Y, 180)
					cell.East.Draw(g, cellPos, Y, 270)
					cell.Top.Draw(g, cellPos, Z, 90)
					cell.Bottom.Draw(g, cellPos, Z, -90)
				}
			}
		}
	}
}

func (f *Face) Draw(g *Game, cellPos Vec3, rotationAxis Vec3, rotationDegrees float32) {
	if f.Type == FaceWall {
		aa, bb := g.Tileset.GetAABB(f.TileX, f.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		rl.DrawModelEx(g.Models["wall"], cellPos.Raylib(), rotationAxis.Raylib(), rotationDegrees, XYZ.Raylib(), rl.White)
	}
}
