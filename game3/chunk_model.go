package game3

import (
	"fmt"
	"game/model"
	"game/vec3"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	FloorWidth = 0.2
	WallWidth  = 0.2
)

var FaceSolidPoint = []vec3.Value{
	vec3.XYZ(0, 0, 0),
	vec3.XYZ(1, -FloorWidth, 1),

	vec3.XYZ(1, 0, 0),
	vec3.XYZ(1-WallWidth, 1, 1),

	vec3.XYZ(0, 0, 1),
	vec3.XYZ(1, 1, 1-WallWidth),

	vec3.XYZ(0, 0, 0),
	vec3.XYZ(WallWidth, 1, 1),

	vec3.XYZ(0, 0, 0),
	vec3.XYZ(1, 1, WallWidth),
}

func GenerateChunkYModel(c *Chunk, y int) rl.Model {
	path := filepath.Join(CHUNKS_PATH, fmt.Sprintf("chunk_%d.obj", y))
	edit := model.NewModel(5, path, "./chunks/chunk.mtl", "Material")

	for x := range ChunkWidth {
		for z := range ChunkWidth {
			cell := &c.Cells[y][x][z]
			pos := vec3.XYZ(float64(x), 0, float64(z))

			for i := range cell.Faces {
				if cell.Faces[i].Type != FaceSolid {
					continue
				}
				aa := pos.Add(FaceSolidPoint[i*2])
				bb := pos.Add(FaceSolidPoint[i*2+1])
				edit.Cube(aa, bb, 0, 2)
			}
		}
	}

	edit.Export()
	mdl := rl.LoadModel(path)

	os.Remove(path)

	return mdl
}
