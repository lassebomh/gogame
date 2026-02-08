package game3

import (
	"game/vec2"
	"game/vec3"
	"image/color"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolCell struct {
	// CellPos vec3.Value

	Delete        bool
	FaceType      FaceType
	FaceRotation  FaceDirection
	FaceModelType FaceModelType
	FaceDirection FaceDirection

	pastingCells map[vec3.Value]bool

	palletePosition vec2.Value
}

func (t *ToolCell) Update(e *Editor) {

	if rl.IsMouseButtonPressed(rl.MouseMiddleButton) {
		face := e.world.GetCell(e.mouseCellPos).Faces[t.FaceDirection]

		faceModel := FaceModelsMap[face.ModelType]
		t.FaceModelType = faceModel.Id
		t.FaceType = faceModel.FaceType
		t.FaceDirection = faceModel.FaceDirection
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		if t.pastingCells == nil {
			t.pastingCells = map[vec3.Value]bool{}
		}

		t.pastingCells[e.mouseCellPos] = true
	}

	if rl.IsKeyPressed(rl.KeyQ) {
		t.palletePosition = e.mousePosition
	}
	if rl.IsKeyReleased(rl.KeyQ) {
		t.palletePosition = vec2.Zero
	}

	t.Delete = rl.IsKeyDown(rl.KeyX)

	if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
		chunks := make(map[*Chunk]bool, 0)
		for pos, _ := range t.pastingCells {
			cell, chunk := e.world.Get(pos)

			face := &cell.Faces[t.FaceDirection]
			if t.Delete {
				face.ModelType = 0
				face.Rotation = 0
				face.Type = 0
			} else {
				face.ModelType = t.FaceModelType
				face.Rotation = t.FaceRotation
				face.Type = t.FaceType
			}
			chunks[chunk] = true
		}
		for chunk, _ := range chunks {
			chunk.ReloadBodies()
			chunk.ReloadModel(int(e.mouseCellPos.Y))
		}
		t.pastingCells = nil
	}

	if !rl.IsMouseButtonDown(rl.MouseButtonRight) {
		if t.FaceDirection != FaceDown {
			t.FaceDirection = e.mouseCellDirection
		}
		t.FaceRotation = e.mouseCellDirection

	}

}

func (t *ToolCell) Draw3D(e *Editor) {
	aa, bb := FaceSolidMeshes[t.FaceDirection].GetAABB()

	center := aa.Lerp(bb, 0.5)
	size := bb.Subtract(aa)

	var alpha uint8

	if t.pastingCells != nil {
		alpha = 255
	} else {
		alpha = 150
	}

	var col color.RGBA
	if t.Delete {
		col = color.RGBA{255, 100, 100, alpha}
	} else {
		col = color.RGBA{255, 255, 255, alpha}
	}

	if t.pastingCells != nil {
		for pos, _ := range t.pastingCells {
			rl.DrawCubeWiresV(pos.Add(center).Raylib(), size.Raylib(), col)
		}
	} else {
		rl.DrawCubeWiresV(e.mouseCellPos.Add(center).Raylib(), size.Raylib(), col)
	}
}

func (t *ToolCell) DrawHUD(e *Editor) {

	if t.palletePosition != vec2.Zero {

		line := NewLineLayout(t.palletePosition.X, t.palletePosition.Y, 24)

		raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT, int64(raygui.TEXT_ALIGN_LEFT))
		raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_PADDING, 4)
		raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, 10)

		for _, faceModel := range FaceModels {
			if raygui.Toggle(line.Next(120), faceModel.Name, t.FaceModelType == faceModel.Id) {
				t.FaceModelType = faceModel.Id
				t.FaceType = faceModel.FaceType
				t.FaceDirection = faceModel.FaceDirection
			}
			line.Break()
		}
	}
}
