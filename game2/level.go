package game2

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const WALL_WIDTH = float64(0.1)

var WALL_NORTH_VERTS = []cp.Vector{
	{1, 1},
	{0, 1},
	{0, 1 - WALL_WIDTH},
	{1, 1 - WALL_WIDTH},
}

var WALL_SOUTH_VERTS = []cp.Vector{
	{1, 0},
	{0, 0},
	{0, WALL_WIDTH},
	{1, WALL_WIDTH},
}

var WALL_EAST_VERTS = []cp.Vector{
	{0, 1},
	{0, 0},
	{WALL_WIDTH, 0},
	{WALL_WIDTH, 1},
}

var WALL_WEST_VERTS = []cp.Vector{
	{1, 1},
	{1, 0},
	{1 - WALL_WIDTH, 0},
	{1 - WALL_WIDTH, 1},
}

type FaceType = uint8

const (
	FaceEmpty = FaceType(iota)
	FaceDoor
	FaceWall
)

type GroundType = uint8

const (
	GroundNone = GroundType(iota)
	GroundSolid
	GroundStair
)

type Ground struct {
	Rotation float64
	TileX    int
	TileY    int
	Type     GroundType
}

type Face struct {
	Type  FaceType
	TileX int
	TileY int

	shape *cp.Shape
}

type Cell struct {
	North  Face
	South  Face
	East   Face
	West   Face
	Ground Ground

	Level    *Level
	Position Vec3
}

func (c *Cell) Wake(g *Game) {
	transform := cp.NewTransformTranslate(cp.Vector{c.Position.X, c.Position.Z})
	if c.West.Type == FaceWall && c.West.shape == nil {
		shape := cp.NewPolyShape(g.Space.StaticBody, 4, WALL_WEST_VERTS, transform, 0)
		shape.Filter.Categories = 1 << uint(c.Position.Y)
		c.West.shape = g.Space.AddShape(shape)
	}
	if c.East.Type == FaceWall && c.East.shape == nil {
		shape := cp.NewPolyShape(g.Space.StaticBody, 4, WALL_EAST_VERTS, transform, 0)
		shape.Filter.Categories = 1 << uint(c.Position.Y)
		c.East.shape = g.Space.AddShape(shape)
	}
	if c.North.Type == FaceWall && c.North.shape == nil {
		shape := cp.NewPolyShape(g.Space.StaticBody, 4, WALL_NORTH_VERTS, transform, 0)
		shape.Filter.Categories = 1 << uint(c.Position.Y)
		c.North.shape = g.Space.AddShape(shape)
	}
	if c.South.Type == FaceWall && c.South.shape == nil {
		shape := cp.NewPolyShape(g.Space.StaticBody, 4, WALL_SOUTH_VERTS, transform, 0)
		shape.Filter.Categories = 1 << uint(c.Position.Y)
		c.South.shape = g.Space.AddShape(shape)
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

	cellx := ((int(math.Floor(pos.X))%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
	cellz := ((int(math.Floor(pos.Z))%CHUNK_WIDTH + CHUNK_WIDTH) % CHUNK_WIDTH)
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

func (l *Level) Draw(g *Game, maxY int) {
	for pos, chunk := range l.Chunks {

		pos := NewVec3(float64(pos.X)*float64(CHUNK_WIDTH), 0, float64(pos.Y)*float64(CHUNK_WIDTH))

		for x := range CHUNK_WIDTH {
			for z := range CHUNK_WIDTH {
				for y := range CHUNK_HEIGHT {
					cell := &chunk[x][z][y]

					if y > maxY {
						continue
					}

					if !(cell.North.Type == FaceWall ||
						cell.East.Type == FaceWall ||
						cell.South.Type == FaceWall ||
						cell.West.Type == FaceWall ||
						cell.Ground.Type != GroundNone) {
						continue
					}

					cellPos := pos.Add(NewVec3(float64(x)+0.5, float64(y)+0.5-WALL_WIDTH, float64(z)+0.5))

					cell.North.Draw(g, cellPos, Y, 270)
					cell.East.Draw(g, cellPos, Y, 180)
					cell.South.Draw(g, cellPos, Y, 90)
					cell.West.Draw(g, cellPos, Y, 0)
					cell.Ground.Draw(g, cellPos)
				}
			}
		}
	}
}

func (f *Face) Draw(g *Game, cellPos Vec3, rotationAxis Vec3, rotationDegrees float32) {
	if f.Type == FaceWall {
		aa, bb := g.Tileset.GetAABB(f.TileX, f.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		rl.DrawModelEx(g.GetModel("wall"), cellPos.Raylib(), rotationAxis.Raylib(), rotationDegrees, XYZ.Raylib(), rl.White)
	}
}

func (gr *Ground) Draw(g *Game, cellPos Vec3) {
	if gr.Type == GroundSolid {
		aa, bb := g.Tileset.GetAABB(gr.TileX, gr.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		rl.DrawModelEx(g.GetModel("wall"), cellPos.Raylib(), Z.Raylib(), float32(-90), XYZ.Raylib(), rl.White)
	} else if gr.Type == GroundStair {
		aa, bb := g.Tileset.GetAABB(gr.TileX, gr.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		rl.DrawModelEx(g.GetModel("stair"), (cellPos.Subtract(NewVec3(0, 0.5, 0))).Raylib(), Y.Raylib(), float32(gr.Rotation), XYZ.Scale(0.99).Raylib(), rl.White)
	}
}
