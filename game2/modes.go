package game2

import (
	"math"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ModeType = int32

const (
	MODE_DEFAULT = ModeType(iota)
	MODE_FREE
)

type ModeFree struct {
	TargetLayerY  float64
	TargetCellPos Vec3

	Pitch float64
	Yaw   float64

	CellPaste Cell

	LastMousePosition Vec2
	Camera            Camera3D
}

func NewModeFree() ModeFree {
	return ModeFree{
		Camera: Camera3D{
			Projection: rl.CameraPerspective,
			Position:   NewVec3(-8, 1, 0),
			Target:     NewVec3(0, 0, 0),
			Up:         Y,
			Fovy:       90,
		},
		Pitch: 0,
		Yaw:   0,
	}
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

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		mouseMove := (currentMousePos.Sub(d.LastMousePosition)).Mult(0.005)
		d.Yaw += mouseMove.X
		d.Pitch -= mouseMove.Y

		if d.Pitch < -Pi/2+1e-5 {
			d.Pitch = -Pi/2 + 1e-5
		}

		if d.Pitch >= Pi/2-1e-5 {
			d.Pitch = Pi/2 - 1e-5
		}
	}

	if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {

		mouseRay := rl.GetScreenToWorldRay(rl.Vector2{float32(currentMousePos.X), float32(currentMousePos.Y)}, d.Camera.Raylib())

		origin := Vec3FromRaylib(mouseRay.Position)
		dir := Vec3FromRaylib(mouseRay.Direction)

		if math.Abs(dir.Y) >= 1e-6 {

			t := (d.TargetLayerY - origin.Y) / dir.Y

			if t >= 0 {
				d.TargetCellPos = origin.Add(dir.Scale(t))
				d.TargetCellPos.X = math.Floor(d.TargetCellPos.X)
				d.TargetCellPos.Y = d.TargetLayerY
				d.TargetCellPos.Z = math.Floor(d.TargetCellPos.Z)

				cell := g.Level.GetCell(d.TargetCellPos.X, d.TargetCellPos.Y, d.TargetCellPos.Z)
				*cell = d.CellPaste

			}
		}

	}

	d.LastMousePosition = currentMousePos

	d.Camera.Target = d.Camera.Position.Add(NewVec3(
		math.Cos(d.Pitch)*math.Cos(d.Yaw),
		math.Sin(d.Pitch),
		math.Cos(d.Pitch)*math.Sin(d.Yaw),
	))
}

func (d *ModeFree) Draw(g *Game) {
	rl.ClearBackground(rl.Black)

	BeginMode3D(d.Camera, func() {

		g.Player.Draw(g)

		rl.DrawSphere(rl.NewVector3(0, 0, 0), 0.1, rl.Red)
		rl.DrawSphere(rl.NewVector3(1, 0, 0), 0.1, rl.Green)
		rl.DrawSphere(rl.NewVector3(0, 1, 0), 0.1, rl.Blue)
		rl.DrawSphere(rl.NewVector3(0, 0, 1), 0.1, rl.Yellow)

		rl.DrawCubeWires((d.TargetCellPos.Add(NewVec3(0.5, 0.5, 0.5))).Raylib(), 1, 1, 1, rl.White)
	})

	line := NewLineLayout(30, 30, 30)

	faceIcons := []int32{
		raygui.ICON_CUBE_FACE_BACK,
		raygui.ICON_CUBE_FACE_RIGHT,
		raygui.ICON_CUBE_FACE_FRONT,
		raygui.ICON_CUBE_FACE_LEFT,
		raygui.ICON_CUBE_FACE_TOP,
		raygui.ICON_CUBE_FACE_BOTTOM,
	}

	faces := []*Face{
		&d.CellPaste.North,
		&d.CellPaste.East,
		&d.CellPaste.South,
		&d.CellPaste.West,
		&d.CellPaste.Top,
		&d.CellPaste.Bottom,
	}

	for i := range 6 {
		face := faces[i]
		icon := faceIcons[i]

		raygui.DummyRec(line.Next(30), raygui.IconText(icon, ""))
		if raygui.Toggle(line.Next(30), raygui.IconText(raygui.ICON_NONE, ""), face.Type == FaceEmpty) {
			face.Type = FaceEmpty
		}
		if raygui.Toggle(line.Next(30), raygui.IconText(raygui.ICON_GRID_FILL, ""), face.Type == FaceWall) {
			face.Type = FaceWall
		}
		if raygui.Toggle(line.Next(30), raygui.IconText(raygui.ICON_DOOR, ""), face.Type == FaceDoor) {
			face.Type = FaceDoor
		}
		line.Break()
	}

	rl.DrawFPS(5, 5)
}
