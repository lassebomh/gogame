package game3

import (
	"math"

	v3 "game/vec3"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const CHUNKS_PATH = "./chunks/"

const StairBlockerWidth = 0.1

var WALL_VERTS = [4][]cp.Vector{
	[]cp.Vector{{1, 1}, {1, 0}, {1 - WallWidth, 0}, {1 - WallWidth, 1}},
	[]cp.Vector{{1, 1}, {0, 1}, {0, 1 - WallWidth}, {1, 1 - WallWidth}},
	[]cp.Vector{{0, 1}, {0, 0}, {WallWidth, 0}, {WallWidth, 1}},
	[]cp.Vector{{1, 0}, {0, 0}, {0, WallWidth}, {1, WallWidth}},
}

var STAIR_VERTS = [4][]cp.Vector{
	[]cp.Vector{{1, 1}, {1, 0}, {1 - StairBlockerWidth, 0}, {1 - StairBlockerWidth, 1}},
	[]cp.Vector{{1, 1}, {0, 1}, {0, 1 - StairBlockerWidth}, {1, 1 - StairBlockerWidth}},
	[]cp.Vector{{0, 1}, {0, 0}, {StairBlockerWidth, 0}, {StairBlockerWidth, 1}},
	[]cp.Vector{{1, 0}, {0, 0}, {0, StairBlockerWidth}, {1, StairBlockerWidth}},
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

func WorldToChunk(v v3.Value) (ChunkPos, LocalPos) {

	v.X = math.Round(v.X*1e8) / 1e8
	v.Y = math.Round(v.Y*1e8) / 1e8
	v.Z = math.Round(v.Z*1e8) / 1e8

	cx := math.Floor(v.X / float64(ChunkWidth))
	cz := math.Floor(v.Z / float64(ChunkWidth))
	lx := v.X - cx*ChunkWidth
	lz := v.Z - cz*ChunkWidth

	// cx := math.Floor(v.X / ChunkWidth)
	// cz := math.Mod(math.Mod(v.X*ChunkWidth, 1)+1, 1) / ChunkWidth
	// lx := math.Floor(v.Z / ChunkWidth)
	// lz := math.Mod(math.Mod(v.Z*ChunkWidth, 1)+1, 1) / ChunkWidth

	chunkPos := ChunkPos{X: int(cx), Z: int(cz)}
	localPos := LocalPos{X: int(lx), Y: int(v.Y), Z: int(lz)}
	return chunkPos, localPos
}

func ChunkToWorld(c ChunkPos, l LocalPos) v3.Value {
	return v3.XYZ(
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

					switch face.Type {
					case FaceSolid:
						shape := cp.NewPolyShape(c.body, 4, WALL_VERTS[i], transform, 0)
						shape.Filter.Group = GroupStatic
						shape.Filter.Categories = Category(worldPos.Y, true, false)
						shape.Filter.Mask = Category(worldPos.Y, true, true)
						c.world.space.AddShape(shape)
					}
				}

				// down := &cell.Faces[FaceDown]
				// if down.Type == FaceStair {
				// 	left := FaceLeft[down.Rotation]
				// 	right := FaceRight[down.Rotation]
				// 	for i := range 2 {
				// 		shape := cp.NewPolyShape(c.body, 4, STAIR_VERTS[left], transform, 0)
				// 		shape.Filter.Group = GroupStatic
				// 		shape.Filter.Categories = Category(worldPos.Y+float64(i), false, true)
				// 		shape.Filter.Mask = Category(worldPos.Y+float64(i), false, true)
				// 		c.world.space.AddShape(shape)
				// 		shape = cp.NewPolyShape(c.body, 4, STAIR_VERTS[right], transform, 0)
				// 		shape.Filter.Group = GroupStatic
				// 		shape.Filter.Categories = Category(worldPos.Y+float64(i), false, true)
				// 		shape.Filter.Mask = Category(worldPos.Y+float64(i), false, true)
				// 		c.world.space.AddShape(shape)
				// 	}
				// }

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

var FaceForward = []v3.Value{
	v3.X(1),
	v3.Z(1),
	v3.X(-1),
	v3.Z(-1),
	v3.Y(1),
}

var FaceOpposite = []FaceDirection{
	FaceEast,
	FaceSouth,
	FaceWest,
	FaceNorth,
	-1,
}

var FaceRight = []FaceDirection{
	FaceNorth,
	FaceEast,
	FaceSouth,
	FaceWest,
	FaceDown,
}
var FaceLeft = []FaceDirection{
	FaceSouth,
	FaceWest,
	FaceNorth,
	FaceEast,
	FaceDown,
}

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
