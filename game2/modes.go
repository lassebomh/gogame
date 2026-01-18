package game2

import (
	"fmt"
	"math"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ModeType = int32

const (
	MODE_DEFAULT = ModeType(iota)
	MODE_FREE
)

type ModeFreeTool = int32

const (
	MODE_FREE_TOOL_WALLS = ModeFreeTool(iota)
	MODE_FREE_TOOL_TEXTURE
)

type ModeFreeToolWallsDirection = int32

const (
	TOOL_WALLS_N = ModeFreeToolWallsDirection(1 << iota)
	TOOL_WALLS_E
	TOOL_WALLS_S
	TOOL_WALLS_W
	TOOL_WALLS_T
	TOOL_WALLS_B
)

type ModeFree struct {
	Camera Camera3D
	Pitch  float64
	Yaw    float64

	MousePosition  Vec2
	CurrentY       float64
	CurrentCellPos Vec3

	CellPaste Cell

	Tool ModeFreeTool

	// Tool walls
	ToolWallsFace       Face
	ToolWallsDirections ModeFreeToolWallsDirection

	ToolWallsTop    Face
	ToolWallsSide   Face
	ToolWallsBottom Face
}

func NewModeFree() *ModeFree {
	return (&ModeFree{
		CurrentY: 0,
		Camera: Camera3D{
			Projection: rl.CameraPerspective,
			Position:   NewVec3(-8, 1, 0),
			Target:     NewVec3(0, 0, 0),
			Up:         Y,
			Fovy:       90,
		},
		Pitch:               0,
		Yaw:                 0,
		Tool:                MODE_FREE_TOOL_WALLS,
		ToolWallsDirections: TOOL_WALLS_N,
		ToolWallsTop:        Face{Type: FaceEmpty},
		ToolWallsSide:       Face{Type: FaceWall},
		ToolWallsBottom:     Face{Type: FaceWall},
	})
}

func (d *ModeFree) Update(g *Game) {
	const SPEED = 0.05

	forward := d.Camera.Target.Subtract(d.Camera.Position).Normalize()

	right := forward.CrossProduct(d.Camera.Up).Normalize()

	forward = NewVec3(
		forward.X,
		0,
		forward.Z,
	).Normalize()

	up := Y
	movement := NewVec3(0, 0, 0)

	if rl.IsKeyDown(rl.KeyW) {
		movement = movement.Add(forward)
	}
	if rl.IsKeyDown(rl.KeyS) {
		movement = movement.Subtract(forward)
	}
	if rl.IsKeyDown(rl.KeyD) {
		movement = movement.Add(right)
	}
	if rl.IsKeyDown(rl.KeyA) {
		movement = movement.Subtract(right)
	}
	if rl.IsKeyDown(rl.KeyQ) {
		movement = movement.Subtract(up)
	}
	if rl.IsKeyDown(rl.KeyE) {
		movement = movement.Add(up)
	}

	if movement.Length() > 0 {
		movement = movement.Normalize().Scale(SPEED)
		d.Camera.Position = d.Camera.Position.Add(movement)
	}

	currentMousePos := Vec2FromRaylib(rl.GetMousePosition())

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		mouseMove := (currentMousePos.Subtract(d.MousePosition)).Scale(0.005)
		d.Yaw += mouseMove.X
		d.Pitch -= mouseMove.Y

		if d.Pitch < -Pi/2+1e-5 {
			d.Pitch = -Pi/2 + 1e-5
		}

		if d.Pitch >= Pi/2-1e-5 {
			d.Pitch = Pi/2 - 1e-5
		}
	}
	d.MousePosition = currentMousePos

	d.Camera.Target = d.Camera.Position.Add(NewVec3(
		math.Cos(d.Pitch)*math.Cos(d.Yaw),
		math.Sin(d.Pitch),
		math.Cos(d.Pitch)*math.Sin(d.Yaw),
	))

	mouseRay := rl.GetScreenToWorldRay(rl.Vector2{float32(currentMousePos.X), float32(currentMousePos.Y)}, d.Camera.Raylib())

	origin := Vec3FromRaylib(mouseRay.Position)
	dir := Vec3FromRaylib(mouseRay.Direction)

	if math.Abs(dir.Y) >= 1e-6 {
		t := (d.CurrentY - origin.Y) / dir.Y

		if t >= 0 {

			d.CurrentCellPos = origin.Add(dir.Scale(t))
			ix := math.Floor(d.CurrentCellPos.X)
			iz := math.Floor(d.CurrentCellPos.Z)
			fx := d.CurrentCellPos.X - ix
			fz := d.CurrentCellPos.Z - iz
			d.CurrentCellPos.X = ix
			d.CurrentCellPos.Y = d.CurrentY
			d.CurrentCellPos.Z = iz

			d.ToolWallsDirections = 0
			if fz > 0.75 {
				d.ToolWallsDirections |= TOOL_WALLS_E
			}
			if fz < 0.25 {
				d.ToolWallsDirections |= TOOL_WALLS_W
			}
			if fx > 0.75 {
				d.ToolWallsDirections |= TOOL_WALLS_N
			}
			if fx < 0.25 {
				d.ToolWallsDirections |= TOOL_WALLS_S
			}
		}
	}

	if d.Tool == MODE_FREE_TOOL_WALLS {
		if rl.IsKeyPressed(rl.KeyR) {
			if d.ToolWallsDirections&TOOL_WALLS_B != 0 {
				d.ToolWallsDirections = TOOL_WALLS_N
			} else {
				d.ToolWallsDirections = ModeFreeToolWallsDirection(d.ToolWallsDirections << 1)
			}
		}

		if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			cellRef := g.Level.GetCell(d.CurrentCellPos)

			if d.ToolWallsDirections&TOOL_WALLS_N != 0 {
				cellRef.North = d.ToolWallsSide
			}
			if d.ToolWallsDirections&TOOL_WALLS_E != 0 {
				cellRef.East = d.ToolWallsSide
			}
			if d.ToolWallsDirections&TOOL_WALLS_S != 0 {
				cellRef.South = d.ToolWallsSide
			}
			if d.ToolWallsDirections&TOOL_WALLS_W != 0 {
				cellRef.West = d.ToolWallsSide
			}
			// if d.ToolWallsDirections&TOOL_WALLS_T != 0 {
			cellRef.Top = d.ToolWallsTop
			// }
			// if d.ToolWallsDirections&TOOL_WALLS_B != 0 {
			cellRef.Bottom = d.ToolWallsBottom
			// }
		}
	}
}

func (d *ModeFree) Draw(g *Game) {
	rl.ClearBackground(rl.Black)

	BeginMode3D(d.Camera, func() {

		g.Draw3D()

		cellPos := (d.CurrentCellPos.Add(NewVec3(0.50, 0.50, 0.50)))

		if d.ToolWallsDirections&TOOL_WALLS_N != 0 {
			rl.DrawModelWiresEx(g.Models["wallDebug"], cellPos.Raylib(), Y.Raylib(), 0, XYZ.Raylib(), rl.Red)
		}
		if d.ToolWallsDirections&TOOL_WALLS_E != 0 {
			rl.DrawModelWiresEx(g.Models["wallDebug"], cellPos.Raylib(), Y.Raylib(), 270, XYZ.Raylib(), rl.Red)
		}
		if d.ToolWallsDirections&TOOL_WALLS_S != 0 {
			rl.DrawModelWiresEx(g.Models["wallDebug"], cellPos.Raylib(), Y.Raylib(), 180, XYZ.Raylib(), rl.Red)
		}
		if d.ToolWallsDirections&TOOL_WALLS_W != 0 {
			rl.DrawModelWiresEx(g.Models["wallDebug"], cellPos.Raylib(), Y.Raylib(), 90, XYZ.Raylib(), rl.Red)
		}
		if d.ToolWallsTop.Type != FaceEmpty {
			rl.DrawModelWiresEx(g.Models["wallDebug"], cellPos.Raylib(), Z.Raylib(), 90, XYZ.Raylib(), rl.Red)
		}

		rl.DrawModelWiresEx(g.Models["wallDebug"], cellPos.Raylib(), Z.Raylib(), -90, XYZ.Raylib(), rl.Red)

	})

	size := float64(30)
	line := NewLineLayout(100, 100, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), g.RenderFlags&RENDER_FLAG_NO_ENTITIES != 0) {
		g.RenderFlags |= RENDER_FLAG_NO_ENTITIES
	} else {
		g.RenderFlags &^= RENDER_FLAG_NO_ENTITIES
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_FRONT, ""), g.RenderFlags&RENDER_FLAG_NO_LEVEL != 0) {
		g.RenderFlags |= RENDER_FLAG_NO_LEVEL
	} else {
		g.RenderFlags &^= RENDER_FLAG_NO_LEVEL
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_BOX, ""), g.RenderFlags&RENDER_FLAG_CP_SHAPES != 0) {
		g.RenderFlags |= RENDER_FLAG_CP_SHAPES
	} else {
		g.RenderFlags &^= RENDER_FLAG_CP_SHAPES
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_LINK, ""), g.RenderFlags&RENDER_FLAG_CP_CONSTRAINTS != 0) {
		g.RenderFlags |= RENDER_FLAG_CP_CONSTRAINTS
	} else {
		g.RenderFlags &^= RENDER_FLAG_CP_CONSTRAINTS
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_LASER, ""), g.RenderFlags&RENDER_FLAG_CP_COLLISIONS != 0) {
		g.RenderFlags |= RENDER_FLAG_CP_COLLISIONS
	} else {
		g.RenderFlags &^= RENDER_FLAG_CP_COLLISIONS
	}

	fmt.Printf("%b\n", g.RenderFlags)

	size = float64(30)

	line = NewLineLayout(5, 20, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), d.ToolWallsSide.Type == FaceEmpty) {
		d.ToolWallsSide.Type = FaceEmpty
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_FRONT, ""), d.ToolWallsSide.Type == FaceWall) {
		d.ToolWallsSide.Type = FaceWall
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_DOOR, ""), d.ToolWallsSide.Type == FaceDoor) {
		d.ToolWallsSide.Type = FaceDoor
	}

	for x := range g.Tileset.Tiles {
		for y := range g.Tileset.Tiles {

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

			if d.ToolWallsSide.TileX == x && d.ToolWallsSide.TileY == y {
				rl.DrawRectangleLinesEx(rect, 1, rl.White)
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(g.ModeFree.MousePosition.Raylib(), rect) {

				d.ToolWallsSide.TileX = x
				d.ToolWallsSide.TileY = y

			}
		}
	}
	line.Break(size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE, ""), d.ToolWallsBottom.Type == FaceEmpty) {
		d.ToolWallsBottom.Type = FaceEmpty
	}
	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, ""), d.ToolWallsBottom.Type == FaceWall) {
		d.ToolWallsBottom.Type = FaceWall
	}
	raygui.DummyRec(line.Next(size), "")

	for x := range g.Tileset.Tiles {
		for y := range g.Tileset.Tiles {

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

			if d.ToolWallsBottom.TileX == x && d.ToolWallsBottom.TileY == y {
				rl.DrawRectangleLinesEx(rect, 1, rl.White)
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(g.ModeFree.MousePosition.Raylib(), rect) {

				d.ToolWallsBottom.TileX = x
				d.ToolWallsBottom.TileY = y

			}
		}
	}

	line.Break(size)

	rl.DrawFPS(5, 5)
}
