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
	WallWidth  = 0.25
	FloorWidth = 0.1
)

var FaceSolidMeshes [5]Mesh
var FaceStairMeshes [5]Mesh

type FaceModelType int32
type FaceModel struct {
	Id            FaceModelType
	Name          string
	TileX         int
	TileY         int
	FaceDirection FaceDirection
	FaceType      FaceType
}

var FaceModels = []FaceModel{
	{
		Id:            face_model_debug,
		Name:          "debug",
		TileX:         6,
		TileY:         1,
		FaceDirection: FaceDown,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_wall_brick,
		Name:          "wall brick",
		TileX:         2,
		TileY:         4,
		FaceDirection: FaceWest,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_sidewalk,
		Name:          "floor sidewalk",
		TileX:         6,
		TileY:         2,
		FaceDirection: FaceDown,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_floor_light_tiles,
		Name:          "floor light tiles",
		TileX:         3,
		TileY:         5,
		FaceDirection: FaceDown,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_stair_metal,
		Name:          "stair metal",
		TileX:         4,
		TileY:         5,
		FaceDirection: FaceDown,
		FaceType:      FaceStair,
	},
}

var FaceModelsMap = make(map[FaceModelType]FaceModel)

var (
	face_model_debug             = FaceModelType(0)
	face_model_wall_brick        = FaceModelType(1)
	face_model_sidewalk          = FaceModelType(2)
	face_model_road              = FaceModelType(3)
	face_model_floor_light_tiles = FaceModelType(4)
	face_model_stair_metal       = FaceModelType(5)
)

func init() {
	for _, faceModel := range FaceModels {
		FaceModelsMap[faceModel.Id] = faceModel
	}

	wall := UnitCube.Transform(vec3.NewMatrix().TranslateXYZ(-1, -0.5, -0.5).Scale(WallWidth, 1, 1).TranslateXYZ(0.5, 0, 0))

	for i := range FaceDown {
		angle := math.Pi * 2 * (float64(i) / 4)
		mat := vec3.NewMatrix().RotateY(angle).TranslateXYZ(0.5, 0.5, 0.5)
		mesh := wall.Transform(mat)
		FaceSolidMeshes[i] = mesh
	}

	floor := wall.Transform(vec3.NewMatrix().TranslateXYZ(-0.5+WallWidth, 0, 0).RotateZ(math.Pi/2).TranslateXYZ(0.5, 0.002, 0.5))
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
	edit := NewModel(15, path, "./chunks/chunk.mtl", "Material")

	for x := range ChunkWidth {
		for z := range ChunkWidth {
			cell := &c.Cells[y][x][z]
			transform := vec3.MatrixTranslateXYZ(float64(x), 0, float64(z))

			for i, face := range cell.Faces {
				if face.Type == FaceNone {
					continue
				}

				var mesh Mesh

				faceTransform := transform

				switch FaceDirection(i) {
				case FaceWest:
					faceTransform = faceTransform.TranslateXYZ(0.001, 0, 0)
				case FaceEast:
					faceTransform = faceTransform.TranslateXYZ(-0.001, 0, 0)
				case FaceNorth:
					faceTransform = faceTransform.TranslateXYZ(0, 0.001, 0.001)
				case FaceSouth:
					faceTransform = faceTransform.TranslateXYZ(0, 0.001, -0.001)
				}

				switch face.Type {
				case FaceSolid:
					mesh = FaceSolidMeshes[i]
				case FaceStair:
					mesh = FaceStairMeshes[face.Rotation]
				}
				mesh = mesh.Transform(faceTransform)

				faceModel := FaceModelsMap[face.ModelType]
				edit.AddMesh(mesh, faceModel.TileX, faceModel.TileY)

				// switch face.ModelType {
				// case face_model_debug:
				// 	edit.AddMesh(mesh, 3, 1)
				// case face_model_wall_brick:
				// 	edit.AddMesh(mesh, 3, 1)
				// }
			}
		}
	}

	edit.Export()
	mdl := rl.LoadModel(path)
	mdl.Materials.Shader = c.world.shader.shader

	os.Remove(path)

	return mdl
}
