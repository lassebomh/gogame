package game3

import (
	"math"

	"game/vec3"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const CHUNKS_PATH = "./chunks/"

var WALL_VERTS = [4][]cp.Vector{
	[]cp.Vector{{1, 1}, {1, 0}, {1 - WallWidth, 0}, {1 - WallWidth, 1}},
	[]cp.Vector{{1, 1}, {0, 1}, {0, 1 - WallWidth}, {1, 1 - WallWidth}},
	[]cp.Vector{{0, 1}, {0, 0}, {WallWidth, 0}, {WallWidth, 1}},
	[]cp.Vector{{1, 0}, {0, 0}, {0, WallWidth}, {1, WallWidth}},
}

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

	world  *World
	top    [ChunkWidth][ChunkWidth]int // no reference, because we need to compare with this
	models [ChunkHeight]rl.Model

	body *cp.Body
}

func (c *Chunk) ReloadBodies() {

	if c.body == nil {
		c.body = c.world.space.AddBody(cp.NewStaticBody())
	} else {
		shapes := []*cp.Shape{}

		c.body.EachShape(func(shape *cp.Shape) {
			shapes = append(shapes, shape)
		})

		for _, shape := range shapes {
			c.world.space.RemoveShape(shape)
		}
	}

	for y := range ChunkHeight {
		for x := range ChunkWidth {
			for z := range ChunkWidth {
				cell := &c.Cells[y][x][z]

				worldPos := ChunkToWorld(c.Position, LocalPos{X: x, Y: y, Z: z})

				transform := cp.NewTransformTranslate(cp.Vector{worldPos.X, worldPos.Z})

				for i := range FaceDown {
					face := &cell.Faces[i]

					if face.Type == FaceSolid {
						shape := cp.NewPolyShape(c.body, 4, WALL_VERTS[i], transform, 0)

						shape.Filter.Group = GroupStatic
						shape.Filter.Categories = Category(worldPos.Y, true, false)
						shape.Filter.Mask = Category(worldPos.Y, true, true)

						face.shape = c.world.space.AddShape(shape)
					} else {
						face.shape = nil
					}
				}
			}
		}
	}
}

func (c *Chunk) ReloadModel(y int) {
	if rl.IsModelValid(c.models[y]) {
		rl.UnloadModel(c.models[y])
	}
	c.models[y], _ = GenerateChunkYModel(c, y)
}

func (c *Chunk) Reload() {
	c.ReloadBodies()

	for y := range ChunkHeight {
		c.ReloadModel(y)
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
	FaceWest = FaceDirection(iota)
	FaceNorth
	FaceEast
	FaceSouth
	FaceDown
)

type Face struct {
	Type      FaceType
	Rotation  FaceDirection
	ModelType FaceModelType

	body  *cp.Body
	shape *cp.Shape
}

type Cell struct {
	Faces [5]Face
}
