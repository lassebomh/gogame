package game3

import (
	v2 "game/vec2"
	v3 "game/vec3"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolCell struct {
	Delete        bool
	FaceType      FaceType
	FaceRotation  FaceDirection
	FaceModelType FaceModelType
	FaceDirection FaceDirection

	pastingCells map[v3.Value]bool

	palletePosition v2.Value
}

func (t *ToolCell) Update(e *Editor) {

	if rl.IsKeyDown(rl.KeySpace) {
		if t.pastingCells == nil {
			t.pastingCells = map[v3.Value]bool{}
		}
		t.pastingCells[e.mouseCellPos] = true
	}

	if rl.IsKeyPressed(rl.KeyQ) {
		t.palletePosition = e.mousePosition
	}
	if rl.IsKeyReleased(rl.KeyQ) {
		t.palletePosition = v2.Zero
	}

	t.Delete = rl.IsKeyDown(rl.KeyX)

	if rl.IsKeyReleased(rl.KeySpace) {
		chunks := make(map[*Chunk]bool, 0)
		for pos := range t.pastingCells {
			cell, chunk := e.world.UpsertCellChunk(pos)

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

	if !rl.IsKeyDown(rl.KeySpace) {
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
}
