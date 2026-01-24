package game2

import (
	"image/color"
	"math"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolFloor struct {
	CellPos Vec3
	Paste   Ground
}

func (t *ToolFloor) Update(g *Game, e *Editor) {

	ix := math.Floor(e.HitPos.X)
	iz := math.Floor(e.HitPos.Z)
	fx := e.HitPos.X - ix - 0.5
	fz := e.HitPos.Z - iz - 0.5

	t.CellPos = NewVec3(ix, e.HitPos.Y, iz)

	if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {
		cellRef := g.Level.GetCell(t.CellPos)
		t.Paste = cellRef.Ground
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		cellRef := g.Level.GetCell(t.CellPos)

		if math.Abs(fx) > math.Abs(fz) {
			if fx < 0 {
				t.Paste.StairDirection = FACE_EAST
			}
			if fx >= 0 {
				t.Paste.StairDirection = FACE_WEST
			}
		} else {
			if fz > 0 {
				t.Paste.StairDirection = FACE_NORTH
			}
			if fz <= 0 {
				t.Paste.StairDirection = FACE_SOUTH
			}
		}

		cellRef.Ground = t.Paste
	}
}

func (t *ToolFloor) Draw3D(g *Game, e *Editor) {
	cellPos := (t.CellPos.Add(NewVec3(0.5, 0.4, 0.5)))

	col := rl.White
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		col = color.RGBA{255, 0, 0, 255}
	}
	rl.SetLineWidth(3)
	rl.DrawModelWiresEx(g.GetModel("wallDebug"), cellPos.Raylib(), Z.Raylib(), -90, XYZ.Raylib(), col)
	rl.SetLineWidth(1)
}

func (t *ToolFloor) DrawHUD(g *Game, e *Editor) {

	size := float64(30)
	line := NewLineLayout(0, 50, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), t.Paste.Type == GroundEmpty) {
		t.Paste.Type = GroundEmpty
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, ""), t.Paste.Type == GroundFloor) {
		t.Paste.Type = GroundFloor
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_VERTICAL_BARS, ""), t.Paste.Type == GroundStair) {
		t.Paste.Type = GroundStair
	}

	line.Break(size)

	for y := range g.Tileset.Tiles {
		for x := range g.Tileset.Tiles {

			rect := line.Next(size)
			aa, bb := g.Tileset.GetAABB(x, y)
			bb = bb.Subtract(aa)

			source := rl.NewRectangle(
				float32(aa.X)*float32(g.Tileset.Texture.Width),
				float32(aa.Y)*float32(g.Tileset.Texture.Height),
				float32(bb.X)*float32(g.Tileset.Texture.Width),
				float32(bb.Y)*float32(g.Tileset.Texture.Height),
			)

			rl.DrawTexturePro(g.Tileset.Texture, source, rect, rl.NewVector2(0, 0), 0, rl.White)

			if t.Paste.TileX == x && t.Paste.TileY == y {
				rl.DrawRectangleLinesEx(rect, 1, rl.White)
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(g.Editor.MousePosition.Raylib(), rect) {
				t.Paste.TileX = x
				t.Paste.TileY = y
			}
		}

		line.Break(size)
	}

	line.Break(size)

}
