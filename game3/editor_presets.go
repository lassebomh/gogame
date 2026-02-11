package game3

import "strings"

type EditorPreset struct {
	Name     string
	TileX    int
	TileY    int
	Active   bool
	Activate func(e *Editor)
}

type EditorPresets struct {
	Wall  []EditorPreset
	Floor []EditorPreset
	Stair []EditorPreset
}

func GetEditorPresets(e *Editor) *EditorPresets {
	presets := &EditorPresets{
		Wall:  []EditorPreset{},
		Floor: []EditorPreset{},
		Stair: []EditorPreset{},
	}

	for name, faceModelHandler := range FaceModels {
		if strings.Contains(name, "wall") {
			presets.Wall = append(presets.Wall, EditorPreset{
				Name:  name,
				TileX: faceModelHandler.TileX,
				TileY: faceModelHandler.TileY,
				Activate: func(e *Editor) {
					e.Tool = TOOL_CELL
					e.ToolCell.FaceDirection = FaceWest
					e.ToolCell.FaceModelType = faceModelHandler.Id
					e.ToolCell.FaceType = FaceSolid
				},
			})
		} else if strings.Contains(name, "floor") {
			presets.Floor = append(presets.Floor, EditorPreset{
				Name:  name,
				TileX: faceModelHandler.TileX,
				TileY: faceModelHandler.TileY,
				Activate: func(e *Editor) {
					e.Tool = TOOL_CELL
					e.ToolCell.FaceDirection = FaceDown
					e.ToolCell.FaceModelType = faceModelHandler.Id
					e.ToolCell.FaceType = FaceSolid
				},
			})
		} else if strings.Contains(name, "stair") {
			presets.Stair = append(presets.Stair, EditorPreset{
				Name:  name,
				TileX: faceModelHandler.TileX,
				TileY: faceModelHandler.TileY,
				Activate: func(e *Editor) {
					e.Tool = TOOL_CELL
					e.ToolCell.FaceDirection = FaceDown
					e.ToolCell.FaceModelType = faceModelHandler.Id
					e.ToolCell.FaceType = FaceStair
				},
			})
		}
	}

	return presets
}
