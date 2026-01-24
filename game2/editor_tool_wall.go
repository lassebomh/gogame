package game2

import (
	"image/color"
	"math"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToolWall struct {
	CellPos   Vec3
	Paste     Face
	FaceIndex FaceIndex
}

func (t *ToolWall) Update(g *Game, e *Editor) {

	ix := math.Floor(e.HitPos.X)
	iz := math.Floor(e.HitPos.Z)
	fx := e.HitPos.X - ix - 0.5
	fz := e.HitPos.Z - iz - 0.5

	t.CellPos = NewVec3(ix, e.HitPos.Y, iz)

	if !rl.IsMouseButtonDown(rl.MouseButtonRight) {
		t.FaceIndex = 0
		if math.Abs(fx) > math.Abs(fz) {
			fz = 0
		} else {
			fx = 0
		}

		if math.Abs(fx) > math.Abs(fz) {
			if fx < 0 {
				t.FaceIndex = FACE_EAST
			}
			if fx >= 0 {
				t.FaceIndex = FACE_WEST
			}
		} else {
			if fz > 0 {
				t.FaceIndex = FACE_NORTH
			}
			if fz <= 0 {
				t.FaceIndex = FACE_SOUTH
			}
		}

	}

	if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {
		cellRef := g.Level.GetCell(t.CellPos)

		t.Paste = cellRef.Faces[t.FaceIndex]
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		cellRef := g.Level.GetCell(t.CellPos)
		cellRef.Faces[t.FaceIndex] = t.Paste
	}
}

func (t *ToolWall) Draw3D(g *Game, e *Editor) {
	cellPos := (t.CellPos.Add(NewVec3(0.5, 0.4, 0.5)))

	col := rl.White

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		col = color.RGBA{255, 0, 0, 255}
	}
	rl.SetLineWidth(3)

	rl.DrawModelWiresEx(g.GetModel("wallDebug"), cellPos.Raylib(), Y.Negate().Raylib(), float32(FACE_DEGREE[t.FaceIndex]), XYZ.Raylib(), col)

	rl.SetLineWidth(1)

}

func (t *ToolWall) DrawHUD(g *Game, e *Editor) {

	size := float64(30)
	line := NewLineLayout(0, 50, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), t.Paste.Type == FaceEmpty) {
		t.Paste.Type = FaceEmpty
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, ""), t.Paste.Type == FaceWall) {
		t.Paste.Type = FaceWall
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_DOOR, ""), t.Paste.Type == FaceDoor) {
		t.Paste.Type = FaceDoor
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
