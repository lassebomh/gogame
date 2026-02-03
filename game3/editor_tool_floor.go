package game3

import (
	"game/vec3"
	"image/color"
	"math"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolFloor struct {
	CellPos vec3.Value
	Paste   Face
}

func (t *ToolFloor) Update(e *Editor) {

	ix := math.Floor(e.mouseWorldPosition.X)
	iz := math.Floor(e.mouseWorldPosition.Z)
	fx := e.mouseWorldPosition.X - ix - 0.5
	fz := e.mouseWorldPosition.Z - iz - 0.5

	t.CellPos = vec3.XYZ(ix, e.mouseWorldPosition.Y, iz)

	if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {
		cellRef := e.world.GetCell(t.CellPos)
		t.Paste = cellRef.Faces[FaceDown]
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		cellRef := e.world.GetCell(t.CellPos)

		if math.Abs(fx) > math.Abs(fz) {
			if fx < 0 {
				t.Paste.Direction = FaceEast
			}
			if fx >= 0 {
				t.Paste.Direction = FaceWest
			}
		} else {
			if fz > 0 {
				t.Paste.Direction = FaceNorth
			}
			if fz <= 0 {
				t.Paste.Direction = FaceSouth
			}
		}

		cellRef.Faces[FaceDown] = t.Paste
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
		cpos, _ := WorldToChunk(t.CellPos)
		chunk, ok := e.world.chunks[cpos]
		if !ok {
			chunk = chunk.Upsert(cpos)
		}
		chunk.Reload()
	}
}

func (t *ToolFloor) Draw3D(e *Editor) {
	cellPos := (t.CellPos.Add(vec3.XYZ(0.5, 0.5-WallWidth, 0.5)))

	col := rl.White
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		col = color.RGBA{255, 0, 0, 255}
	}
	rl.SetLineWidth(3)

	rl.DrawCubeWires(cellPos.Subtract(vec3.Y(FloorWidth*2)).Raylib(), 1, FloorWidth, 1, col)
	rl.SetLineWidth(1)
}

func (t *ToolFloor) DrawHUD(e *Editor) {

	size := float64(30)
	line := NewLineLayout(0, 50, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), t.Paste.Type == FaceNone) {
		t.Paste.Type = FaceNone
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, ""), t.Paste.Type == FaceSolid) {
		t.Paste.Type = FaceSolid
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_VERTICAL_BARS, ""), t.Paste.Type == FaceStair) {
		t.Paste.Type = FaceStair
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
