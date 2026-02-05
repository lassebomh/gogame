package game3

import (
	"fmt"
	. "game/model"
	"game/vec3"
	"math"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	FloorWidth = 0.2
	WallWidth  = 0.2
)

var FaceSolidMeshes [5]Mesh
var FaceStairMeshes [5]Mesh

func init() {
	wall := UnitCube.Transform(vec3.NewMatrix().TranslateXYZ(-1, -0.5, -0.5).Scale(WallWidth, 1, 1).TranslateXYZ(0.5, 0, 0))

	for i := range FaceDown {
		angle := math.Pi * 2 * (float64(i) / 4)
		mat := vec3.NewMatrix().RotateY(angle).TranslateXYZ(0.5, 0.5, 0.5)
		mesh := wall.Transform(mat)
		FaceSolidMeshes[i] = mesh
	}

	floor := wall.Transform(vec3.NewMatrix().TranslateXYZ(-0.5+WallWidth, 0, 0).RotateZ(math.Pi/2).TranslateXYZ(0.5, 0, 0.5))
	FaceSolidMeshes[FaceDown] = floor

	stair := NewMesh()
	steps := 4.
	baseStep := UnitCube.Transform(vec3.NewMatrix().TranslateXYZ(-0.5, 0, 0).Scale(1, 1/steps, 1))

	for i := 0.; i < steps; i += 1 {
		y := i / steps
		mat := vec3.NewMatrix().Scale(1, 1, 1-y).TranslateXYZ(0, y, -0.5)
		stair.Combine(baseStep.Transform(mat))
	}

	for i := range FaceDown {
		angle := math.Pi * 2 * (float64(i+1) / 4)
		mat := vec3.NewMatrix().TranslateXYZ(0, 0, 0).RotateY(angle).TranslateXYZ(0.5, 0, 0.5)
		mesh := stair.Transform(mat)
		FaceStairMeshes[i] = mesh
	}

}

func GenerateChunkYModel(c *Chunk, y int) rl.Model {
	path := filepath.Join(CHUNKS_PATH, fmt.Sprintf("chunk_%d.obj", rl.GetRandomValue(0, 1000)))
	os.Remove(path)
	edit := NewModel(5, path, "./chunks/chunk.mtl", "Material")

	for x := range ChunkWidth {
		for z := range ChunkWidth {
			cell := &c.Cells[y][x][z]
			pos := vec3.XYZ(float64(x), 0, float64(z))
			transform := vec3.MatrixTranslate(pos)

			for i, face := range cell.Faces {
				faceDir := FaceDirection(i)

				switch face.Type {
				case FaceSolid:
					if faceDir == FaceDown {
						edit.AddMesh(FaceSolidMeshes[i].Transform(transform), 5, 5)
					} else {
						edit.AddMesh(FaceSolidMeshes[i].Transform(transform), 3, 4)
					}
				case FaceStair:
					edit.AddMesh(FaceStairMeshes[face.Direction].Transform(transform), 3, 4)
				}
			}
		}
	}

	// stair := NewMesh()
	// steps := 4.
	// baseStep := UnitCube.Transform(vec3.NewMatrix().TranslateXYZ(-0.5, 0, 0).Scale(1, 1/steps, 1))

	// for i := 0.; i < steps; i += 1 {
	// 	y := i / steps
	// 	mat := vec3.NewMatrix().Scale(1, 1, 1-y).TranslateXYZ(0, y, -0.5)
	// 	stair.Combine(baseStep.Transform(mat))
	// }

	// // 1, 0.75, 0.5, 0.25

	// edit.AddMesh(stair, 4, 2)

	edit.Export()
	mdl := rl.LoadModel(path)
	mdl.Materials.Shader = c.world.shader.shader

	os.Remove(path)

	return mdl
}
