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

type Face struct {
	Type  FaceType
	TileX int
	TileY int

	body *cp.Body
}

type Cell struct {
	North  Face
	South  Face
	East   Face
	West   Face
	Top    Face
	Bottom Face

	Level    *Level
	Position Vec3
}

func (c *Cell) Wake(g *Game) {
	if c.West.Type == FaceWall && c.West.body == nil {
		c.West.body = cp.NewStaticBody()
		c.West.body.SetPosition(cp.Vector{c.Position.X + 0.95, c.Position.Z + 0.5})
		shape := cp.NewBox(c.West.body, 0.1, 1, 0)
		g.Space.AddShape(shape)
	}
	if c.East.Type == FaceWall && c.East.body == nil {
		c.East.body = cp.NewStaticBody()
		c.East.body.SetPosition(cp.Vector{c.Position.X + 0.05, c.Position.Z + 0.5})
		shape := cp.NewBox(c.East.body, 0.1, 1, 0)
		g.Space.AddShape(shape)
	}
	if c.North.Type == FaceWall && c.North.body == nil {
		c.North.body = cp.NewStaticBody()
		c.North.body.SetPosition(cp.Vector{c.Position.X + 0.5, c.Position.Z + 0.95})
		shape := cp.NewBox(c.North.body, 1, .1, 0)
		g.Space.AddShape(shape)
	}
	if c.South.Type == FaceWall && c.South.body == nil {
		c.South.body = cp.NewStaticBody()
		c.South.body.SetPosition(cp.Vector{c.Position.X + 0.5, c.Position.Z + 0.05})
		shape := cp.NewBox(c.South.body, 1, .1, 0)
		g.Space.AddShape(shape)
	}
}

const CHUNK_WIDTH = int(8)
const CHUNK_HEIGHT = int(16)

type Chunk = [CHUNK_WIDTH][CHUNK_WIDTH][CHUNK_HEIGHT]Cell

type Level struct {
	Chunks map[Vec2]*Chunk

	refs map[Vec3]*Cell
}

func (l *Level) Init() *Level {
	if l.refs == nil {
		l.refs = make(map[Vec3]*Cell, 0)
	}
	if l.Chunks == nil {
		l.Chunks = make(map[Vec2]*Chunk, 0)
	}
	return l
}

func (l *Level) GetCell(pos Vec3) *Cell {

	ref := l.refs[pos]

	if ref != nil {
		return ref
	}

	chunkPos := NewVec2(pos.X/float64(CHUNK_WIDTH), pos.Z/float64(CHUNK_WIDTH)).Floor()
	chunk := l.Chunks[chunkPos]

	if chunk == nil {
		chunk = &[CHUNK_WIDTH][CHUNK_WIDTH][CHUNK_HEIGHT]Cell{}

		l.Chunks[chunkPos] = chunk

		fmt.Printf("New chunk %+v\n", chunkPos)
	}

	cellx := ((int(pos.X)%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	cellz := ((int(pos.Z)%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	celly := int(math.Floor(pos.Y))

	ref = &chunk[cellx][cellz][celly]
	ref.Position = NewVec3(
		(chunkPos.X*float64(CHUNK_WIDTH))+float64(cellx),
		float64(celly),
		(chunkPos.Y*float64(CHUNK_WIDTH))+float64(cellz),
	)

	l.refs[pos] = ref

	return ref
}

func (l *Level) Draw(g *Game) {
	for pos, chunk := range l.Chunks {

		pos := NewVec3(
			float64(pos.X)*float64(CHUNK_WIDTH),
			0,
			float64(pos.Y)*float64(CHUNK_WIDTH),
		)

		for x := range CHUNK_WIDTH {
			for z := range CHUNK_WIDTH {
				for y := range CHUNK_HEIGHT {
					cell := &chunk[x][z][y]

					if !(cell.North.Type == FaceWall ||
						cell.East.Type == FaceWall ||
						cell.South.Type == FaceWall ||
						cell.West.Type == FaceWall ||
						cell.Top.Type == FaceWall ||
						cell.Bottom.Type == FaceWall) {
						continue
					}

					cellPos := pos.Add(NewVec3(float64(x)+0.5, float64(y)+0.5, float64(z)+0.5))

					cell.North.Draw(g, cellPos, Y, 270)
					cell.East.Draw(g, cellPos, Y, 180)
					cell.South.Draw(g, cellPos, Y, 90)
					cell.West.Draw(g, cellPos, Y, 0)
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
