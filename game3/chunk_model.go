package game3

import (
	"fmt"
	. "game/model"
	v3 "game/vec3"
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
		TileX:         8,
		TileY:         6,
		FaceDirection: FaceDown,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_sidewalk_lightpole,
		Name:          "floor sidewalk light pole",
		TileX:         8,
		TileY:         6,
		FaceDirection: FaceDown,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_road,
		Name:          "floor road",
		TileX:         6,
		TileY:         6,
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
	{
		Id:            face_model_wall_station,
		Name:          "wall station",
		TileX:         6,
		TileY:         2,
		FaceDirection: FaceWest,
		FaceType:      FaceSolid,
	},
	{
		Id:            face_model_floor_station,
		Name:          "floor station",
		TileX:         6,
		TileY:         4,
		FaceDirection: FaceDown,
		FaceType:      FaceSolid,
	},
}

var FaceModelsMap = make(map[FaceModelType]FaceModel)

var (
	face_model_debug              = FaceModelType(0)
	face_model_wall_brick         = FaceModelType(1)
	face_model_sidewalk           = FaceModelType(2)
	face_model_sidewalk_lightpole = FaceModelType(6)
	face_model_road               = FaceModelType(3)
	face_model_floor_light_tiles  = FaceModelType(4)
	face_model_stair_metal        = FaceModelType(5)
	face_model_wall_station       = FaceModelType(7)
	face_model_floor_station      = FaceModelType(8)
)

func init() {
	for _, faceModel := range FaceModels {
		FaceModelsMap[faceModel.Id] = faceModel
	}

	wall := UnitCube.Transform(v3.NewMatrix().TranslateXYZ(-1, -0.5, -0.5).Scale(WallWidth, 1, 1).TranslateXYZ(0.5, 0, 0))

	for i := range FaceDown {
		angle := math.Pi * 2 * (float64(i) / 4)
		mat := v3.NewMatrix().RotateY(angle).TranslateXYZ(0.5, 0.5, 0.5)
		mesh := wall.Transform(mat)
		FaceSolidMeshes[i] = mesh
	}

	floor := wall.Transform(v3.NewMatrix().TranslateXYZ(-0.5+WallWidth, 0, 0).RotateZ(math.Pi/2).TranslateXYZ(0.5, 0.01, 0.5))
	FaceSolidMeshes[FaceDown] = floor

	stair := NewMesh()
	steps := 5.
	baseStep := UnitCube.Transform(v3.NewMatrix().TranslateXYZ(-0.5, 0, 0).Scale(1, 1/steps, 1))

	for i := 0.; i < steps; i += 1 {
		y := i / steps
		mat := v3.NewMatrix().Scale(1, 1, 1-y).TranslateXYZ(0, y, -0.5)
		stair.Combine(baseStep.Transform(mat))
	}

	for i := range FaceDown {
		angle := math.Pi * 2 * (float64(i+1) / 4)
		mat := v3.NewMatrix().TranslateXYZ(0, 0, 0).RotateY(angle).TranslateXYZ(0.5, 0, 0.5)
		mesh := stair.Transform(mat)
		FaceStairMeshes[i] = mesh
	}

}

func GenerateChunkYModel(c *Chunk, y int) (rl.Model, bool) {
	path := filepath.Join(CHUNKS_PATH, fmt.Sprintf("chunk_%d.obj", rl.GetRandomValue(0, 2147483647)))
	os.Remove(path)
	edit := NewModel(15, path, "./chunks/chunk.mtl", "Material")

	touched := false

	for x := range ChunkWidth {
		for z := range ChunkWidth {
			cell := &c.Cells[y][x][z]
			worldPos := ChunkToWorld(c.Position, LocalPos{x, y, z})
			translate := v3.MatrixTranslateXYZ(float64(x), 0, float64(z))

			for i, face := range cell.Faces {
				if face.Type == FaceNone {
					continue
				}

				var mesh Mesh

				faceTransform := translate

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

				rotation := 0

				switch face.Rotation {
				case FaceDown:
				case FaceWest:
					rotation = 0
				case FaceNorth:
					rotation = 1
				case FaceEast:
					rotation = 2
				case FaceSouth:
					rotation = 3
				}

				if face.Type == FaceStair {
					rotation = (rotation + 1) % 4
				}

				switch faceModel.Id {
				case face_model_sidewalk_lightpole:
					{
						tileX := faceModel.TileX
						tileY := faceModel.TileY
						rotation := (rotation + 1) % 4

						polePart := UnitCube.Transform(
							v3.NewMatrix().
								TranslateXYZ(-0.5, 0, -0.5).
								Scale(0.08, 1, 0.08).
								RotateY(math.Pi/4).
								TranslateXYZ(0.5, 0, 0.5).
								Translate(FaceForward[FaceRight[FaceDirection(rotation)]].Scale(0.4)),
						).Transform(translate)

						edit.AddMesh(polePart, 3, 4, 0)
						edit.AddMesh(polePart.Transform(v3.MatrixTranslateXYZ(0, 1, 0)), 3, 4, 0)
						edit.AddMesh(polePart.Transform(v3.MatrixTranslateXYZ(0, 2, 0)), 3, 4, 0)

						poleHead := UnitCube.Transform(v3.NewMatrix().
							TranslateXYZ(-0.5, -0.5, -0.1).
							Scale(0.15, 0.1, 0.35).
							RotateY((math.Pi*2.)*float64(rotation)/4).
							TranslateXYZ(0.5, 3, 0.5).
							Translate(FaceForward[FaceRight[FaceDirection(rotation)]].Scale(0.4)),
						).Transform(translate)

						edit.AddMesh(poleHead, 3, 4, 0)
						edit.AddMesh(mesh, tileX, tileY, rotation)
					}
				case face_model_sidewalk:
					{
						up := c.world.UpsertCell(worldPos.Add(v3.Z(1))).Faces[FaceDown].ModelType
						down := c.world.UpsertCell(worldPos.Add(v3.Z(-1))).Faces[FaceDown].ModelType
						left := c.world.UpsertCell(worldPos.Add(v3.X(1))).Faces[FaceDown].ModelType
						right := c.world.UpsertCell(worldPos.Add(v3.X(-1))).Faces[FaceDown].ModelType

						tileX := faceModel.TileX
						tileY := faceModel.TileY

						if left == face_model_sidewalk || left == face_model_sidewalk_lightpole {
							rotation = 0
						}
						if right == face_model_sidewalk || right == face_model_sidewalk_lightpole {
							rotation = 2
						}
						if down == face_model_sidewalk || down == face_model_sidewalk_lightpole {
							rotation = 3
						}
						if up == face_model_sidewalk || up == face_model_sidewalk_lightpole {
							rotation = 1
						}
						if up == face_model_road && left == face_model_road && down != face_model_road && right != face_model_road {
							tileX = 7
							rotation = 2
						}

						if up == face_model_road && right == face_model_road && down != face_model_road && left != face_model_road {
							tileX = 7
							rotation = 3
						}

						if right == face_model_road && down == face_model_road && left != face_model_road && up != face_model_road {
							tileX = 7
							rotation = 0
						}

						if left == face_model_road && down == face_model_road && right != face_model_road && up != face_model_road {
							tileX = 7
							rotation = 1
						}

						edit.AddMesh(mesh, tileX, tileY, rotation)
					}

				default:
					edit.AddMesh(mesh, faceModel.TileX, faceModel.TileY, rotation)
				}

				touched = true
			}
		}
	}

	if !touched {
		return rl.Model{}, false
	}

	edit.Export()
	mdl := rl.LoadModel(path)
	mdl.Materials.Shader = globals.Shaders.Main.shader

	os.Remove(path)

	return mdl, true
}

func (c *Chunk) UpdateLights() {

	for y := range ChunkHeight {
		for x := range ChunkWidth {
			for z := range ChunkWidth {
				cell := &c.Cells[y][x][z]
				worldPos := ChunkToWorld(c.Position, LocalPos{x, y, z})
				to := worldPos.AddXYZ(0.5, 0, 0.5).Add(FaceForward[FaceOpposite[cell.Faces[FaceDown].Rotation]].Scale(0.7))
				from := to.AddXYZ(0, 3, 0)

				if cell.Faces[FaceDown].ModelType == face_model_sidewalk_lightpole {
					globals.Shaders.Main.LightSpot(from, to, 5, 25, rl.NewColor(253, 249, 100, 255), 0.4)
				}
			}
		}
	}
}
