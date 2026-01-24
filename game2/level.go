package game2

import (
	"fmt"
	"math"
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

					cellPos := pos.Add(NewVec3(float64(x)+0.5, float64(y)+0.5-WALL_WIDTH, float64(z)+0.5))

					cell.Ground.Draw(g, cellPos)
					for FACE := range FACES {
						face := &cell.Faces[FACE]
						face.Draw(g, cellPos, Y.Negate(), float32(FACE_DEGREE[FACE]))
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
