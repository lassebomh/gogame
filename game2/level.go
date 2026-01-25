package game2

import (
	"fmt"
	"math"

	"github.com/beefsack/go-astar"
	rl "github.com/gen2brain/raylib-go/raylib"
)

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
	for _, chunk := range l.Chunks {
		l.ChunkInit(chunk)
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
		l.ChunkInit(chunk)

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

					cellPos := pos.AddXYZ(float64(x), float64(y), float64(z))

					cell.Ground.Draw(g, cellPos)

					if cell.Ground.Type == GroundStair {
						offset := FACE_DIRECTION[cell.Ground.StairDirection].Add(Y)
						forwardUp := l.GetCell(cell.Position.Add(offset))
						forwardUp.Ground.Draw(g, cellPos.Add(offset))
					}

					for FACE := range FACES {
						face := &cell.Faces[FACE]

						aa, bb := g.Tileset.GetAABB(face.TileX, face.TileY)
						g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
						center := cellPos.AddXYZ(0.5, 0.5, 0.5)

						switch face.Type {
						case FaceWall:
							rl.DrawModelEx(g.GetModel("wall"), center.Raylib(), Y.Negate().Raylib(), float32(FACE_DEGREE[FACE]), XYZ.Raylib(), rl.White)
						case FaceDoor:
							if face.body != nil {
								pos := face.body.Position()
								origin := NewVec3(pos.X, cellPos.Y, pos.Y)
								angle := face.body.Angle() * rl.Rad2deg
								rl.DrawModelEx(g.GetModel("door"), origin.Raylib(), Y.Negate().Raylib(), float32(angle), XYZ.Raylib(), rl.White)
							}

						}
					}
				}
			}
		}
	}
}

func (l *Level) ChunkInit(c *Chunk) {

	for x := range CHUNK_WIDTH {
		for z := range CHUNK_WIDTH {
			for y := range CHUNK_HEIGHT {
				cell := &c[x][z][y]
				cell.level = l
			}
		}
	}
}

func (l *Level) FindPath(from Vec3, to Vec3) ([]*Cell, float64, bool) {

	start := l.GetCell(from)
	end := l.GetCell(to.Floor())

	pathers, length, found := astar.Path(end, start)

	cells := make([]*Cell, len(pathers))

	for i, p := range pathers {
		cell := p.(*Cell)
		cells[i] = cell
	}

	return cells, length, found
}
