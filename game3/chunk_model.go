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
	WallWidth  = 0.2
	FloorWidth = 0.1
)

var FaceSolidMeshes [5]Mesh
var FaceStairMeshes [5]Mesh

type FaceModelType int32

type FaceModelHandler struct {
	Id       FaceModelType
	TileX    int
	TileY    int
	FaceType FaceType
	Render   func(c *Chunk, worldPos v3.Value, rotation int, translate v3.Matrix, faceModel FaceModelHandler, edit *Model, mesh Mesh)
}

func NewFaceModel(id FaceModelType, faceType FaceType, atlasX int, atlasY int, render func(c *Chunk, worldPos v3.Value, rotation int, translate v3.Matrix, faceModel FaceModelHandler, edit *Model, mesh Mesh)) FaceModelHandler {
	return FaceModelHandler{
		Id:       id,
		TileX:    atlasX,
		TileY:    atlasY,
		FaceType: faceType,
		Render:   render,
	}
}

var (
	f0 = FaceModelType(0)
	f1 = FaceModelType(1)
	f2 = FaceModelType(2)
	f6 = FaceModelType(6)
	f3 = FaceModelType(3)
	f4 = FaceModelType(4)
	f5 = FaceModelType(5)
	f7 = FaceModelType(7)
	f8 = FaceModelType(8)
)

func FaceModelDefaultRender(c *Chunk, worldPos v3.Value, rotation int, translate v3.Matrix, faceModel FaceModelHandler, edit *Model, mesh Mesh) {
	edit.AddMesh(mesh, faceModel.TileX, faceModel.TileY, rotation)
}

var FaceModels = map[string]FaceModelHandler{
	"face_model_debug":             NewFaceModel(0, FaceSolid, 6, 1, FaceModelDefaultRender),
	"face_model_wall_brick":        NewFaceModel(1, FaceSolid, 2, 4, FaceModelDefaultRender),
	"face_model_stair_metal":       NewFaceModel(5, FaceStair, 4, 5, FaceModelDefaultRender),
	"face_model_wall_station":      NewFaceModel(7, FaceSolid, 6, 2, FaceModelDefaultRender),
	"face_model_floor_road":        NewFaceModel(3, FaceSolid, 6, 6, FaceModelDefaultRender),
	"face_model_floor_light_tiles": NewFaceModel(4, FaceSolid, 3, 5, FaceModelDefaultRender),
	"face_model_floor_station":     NewFaceModel(8, FaceSolid, 6, 4, FaceModelDefaultRender),
	"face_model_floor_teleporter":  NewFaceModel(9, FaceSolid, 6, 5, FaceModelDefaultRender),
	"face_model_floor_sidewalk": NewFaceModel(2, FaceSolid, 8, 6, func(c *Chunk, worldPos v3.Value, rotation int, translate v3.Matrix, faceModel FaceModelHandler, edit *Model, mesh Mesh) {
		up := FaceModelType(0)
		down := FaceModelType(0)
		left := FaceModelType(0)
		right := FaceModelType(0)

		if upCell, ok := c.world.GetCell(worldPos.Add(v3.Z(1))); ok {
			up = upCell.Faces[FaceDown].ModelType
		}
		if downCell, ok := c.world.GetCell(worldPos.Add(v3.Z(-1))); ok {
			down = downCell.Faces[FaceDown].ModelType
		}
		if leftCell, ok := c.world.GetCell(worldPos.Add(v3.X(1))); ok {
			left = leftCell.Faces[FaceDown].ModelType
		}
		if rightCell, ok := c.world.GetCell(worldPos.Add(v3.X(-1))); ok {
			right = rightCell.Faces[FaceDown].ModelType
		}

		tileX := faceModel.TileX
		tileY := faceModel.TileY

		if left == f2 || left == f6 {
			rotation = 0
		}
		if right == f2 || right == f6 {
			rotation = 2
		}
		if down == f2 || down == f6 {
			rotation = 3
		}
		if up == f2 || up == f6 {
			rotation = 1
		}
		if up == f3 && left == f3 && down != f3 && right != f3 {
			tileX = 7
			rotation = 2
		}

		if up == f3 && right == f3 && down != f3 && left != f3 {
			tileX = 7
			rotation = 3
		}

		if right == f3 && down == f3 && left != f3 && up != f3 {
			tileX = 7
			rotation = 0
		}

		if left == f3 && down == f3 && right != f3 && up != f3 {
			tileX = 7
			rotation = 1
		}

		edit.AddMesh(mesh, tileX, tileY, rotation)
	}),
	"face_model_sidewalk_lightpole": NewFaceModel(6, FaceSolid, 8, 6, func(c *Chunk, worldPos v3.Value, rotation int, translate v3.Matrix, faceModel FaceModelHandler, edit *Model, mesh Mesh) {
		tileX := faceModel.TileX
		tileY := faceModel.TileY
		rotation = (rotation + 1) % 4

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
	}),
}

var FaceModelsMap = make(map[FaceModelType]FaceModelHandler)

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
	edit := NewModel(globals.Textures.AtlasTiles, path, "./chunks/chunk.mtl", "Material")

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

				faceModel.Render(c, worldPos, rotation, translate, faceModel, edit, mesh)

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

				if cell.Faces[FaceDown].ModelType == f6 {
					globals.Shaders.Main.LightSpot(from, to, 5, 25, rl.NewColor(253, 249, 100, 255), 0.4)
				}
			}
		}
	}
}
