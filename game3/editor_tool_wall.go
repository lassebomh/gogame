package game3

import (
	"game/vec3"
	"image/color"
	"math"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolWall struct {
	CellPos   vec3.Value
	Paste     Face
	FaceIndex FaceDirection

	PastingCells map[vec3.Value]bool
}

func (t *ToolWall) Update(e *Editor) {

	ix := math.Floor(e.mouseWorldPosition.X)
	iz := math.Floor(e.mouseWorldPosition.Z)
	fx := e.mouseWorldPosition.X - ix - 0.5
	fz := e.mouseWorldPosition.Z - iz - 0.5

	t.CellPos = vec3.XYZ(ix, e.mouseWorldPosition.Y, iz)

	if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {
		cellRef, _ := e.world.GetCell(t.CellPos)
		t.Paste = cellRef.Faces[t.FaceIndex]
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		if t.PastingCells == nil {
			t.PastingCells = map[vec3.Value]bool{}
		}

		t.PastingCells[t.CellPos] = true
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
		chunks := make(map[*Chunk]bool, 0)
		for pos, _ := range t.PastingCells {
			cell, chunk := e.world.GetCell(pos)
			cell.Faces[t.FaceIndex] = t.Paste
			chunks[chunk] = true
		}
		for chunk, _ := range chunks {
			chunk.Reload()
		}
		t.PastingCells = nil
	}

	if !rl.IsMouseButtonDown(rl.MouseButtonRight) {
		t.FaceIndex = 0
		if math.Abs(fx) > math.Abs(fz) {
			fz = 0
		} else {
			fx = 0
		}

		if math.Abs(fx) > math.Abs(fz) {
			if fx < 0 {
				t.FaceIndex = FaceEast
			}
			if fx >= 0 {
				t.FaceIndex = FaceWest
			}
		} else {
			if fz > 0 {
				t.FaceIndex = FaceNorth
			}
			if fz <= 0 {
				t.FaceIndex = FaceSouth
			}
		}
	}
}

func (t *ToolWall) Draw3D(e *Editor) {

	aa, bb := FaceSolidMeshes[t.FaceIndex].GetAABB()
	center := aa.Lerp(bb, 0.5)
	size := bb.Subtract(aa)

	if t.PastingCells != nil {
		for pos, _ := range t.PastingCells {
			rl.DrawCubeWiresV(pos.Add(center).Raylib(), size.Raylib(), color.RGBA{255, 0, 0, 255})
		}
	} else {
		rl.DrawCubeWiresV(t.CellPos.Add(center).Raylib(), size.Raylib(), rl.White)
	}
}

func (t *ToolWall) DrawHUD(e *Editor) {

	size := float64(30)
	line := NewLineLayout(0, 50, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), t.Paste.Type == FaceNone) {
		t.Paste.Type = FaceNone
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, ""), t.Paste.Type == FaceSolid) {
		t.Paste.Type = FaceSolid
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_DOOR, ""), t.Paste.Type == FaceDoor) {
		t.Paste.Type = FaceDoor
	}

	line.Break(size)

	// for y := range g.Tileset.Tiles {
	// 	for x := range g.Tileset.Tiles {

	// 		rect := line.Next(size)
	// 		aa, bb := g.Tileset.GetAABB(x, y)
	// 		bb = bb.Subtract(aa)

	// 		source := rl.NewRectangle(
	// 			float32(aa.X)*float32(g.Tileset.Texture.Width),
	// 			float32(aa.Y)*float32(g.Tileset.Texture.Height),
	// 			float32(bb.X)*float32(g.Tileset.Texture.Width),
	// 			float32(bb.Y)*float32(g.Tileset.Texture.Height),
	// 		)

	// 		rl.DrawTexturePro(g.Tileset.Texture, source, rect, rl.NewVector2(0, 0), 0, rl.White)

	// 		if t.Paste.TileX == x && t.Paste.TileY == y {
	// 			rl.DrawRectangleLinesEx(rect, 1, rl.White)
	// 		}
	// 		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(g.Editor.MousePosition.Raylib(), rect) {
	// 			t.Paste.TileX = x
	// 			t.Paste.TileY = y
	// 		}
	// 	}

	// 	line.Break(size)
	// }

	// line.Break(size)

}
