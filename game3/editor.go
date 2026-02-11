package game3

import (
	v2 "game/vec2"
	v3 "game/vec3"
	"image/color"
	"math"
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type EditorDrawFlags int32

const (
	EditorDrawFlagPhysics = EditorDrawFlags(1 << iota)
	EditorDrawFlagFocusPlayer
	EditorDrawFlagFocusMonster
)

type Editor struct {
	TimeStep time.Duration

	Position     v3.Value
	PositionSoft v3.Value

	Pitch           float64
	Yaw             float64
	ScrollY         float64
	ScrollYVelocity float64
	Scale           float64

	Camera          Camera3D
	EditorDrawFlags EditorDrawFlags

	mousePosition      v2.Value
	mouseWorldPosition v3.Value

	mouseCellPos       v3.Value
	mouseCellDirection FaceDirection

	Tool     TOOL
	ToolCell ToolCell

	editorPresets *EditorPresets

	world *World
}

type TOOL = int32

const (
	TOOL_CELL = TOOL(iota)
	TOOL_PLAY
)

func (e *Editor) Upsert(world *World) {
	if e == nil {
		e = &Editor{
			TimeStep: 0,

			Position: v3.XYZ(0, 0, 0),
			ScrollY:  -30,

			Pitch: -0.8,
			Yaw:   0,
		}
	}
	e.world = world
	world.Editor = e
	e.PositionSoft = e.Position
	e.editorPresets = GetEditorPresets(e)
}

func (e *Editor) Update(timeStep time.Duration) {
	e.TimeStep = timeStep

	friction := math.Pow(0.5, timeStep.Seconds()*20)

	forward := v3.XZ(math.Cos(e.Yaw), math.Sin(e.Yaw))
	right := forward.RotateByAxisAngle(v3.Y(-1), rl.Pi/2)

	targetVelocity := v3.Zero

	if rl.IsKeyDown(rl.KeyW) {
		targetVelocity = targetVelocity.Add(forward)
	}
	if rl.IsKeyDown(rl.KeyS) {
		targetVelocity = targetVelocity.Subtract(forward)
	}
	if rl.IsKeyDown(rl.KeyD) {
		targetVelocity = targetVelocity.Add(right)
	}
	if rl.IsKeyDown(rl.KeyA) {
		targetVelocity = targetVelocity.Subtract(right)
	}

	if targetVelocity.Length() > 0 {
		targetVelocity = targetVelocity.Normalize().Scale(e.Scale * timeStep.Seconds() * 1.5)
		// e.EditorDrawFlags &= ^EditorDrawFlagFocusPlayer
		// e.EditorDrawFlags &= ^EditorDrawFlagFocusMonster
	}

	if e.Tool != TOOL_PLAY {

		scroll := float64(rl.GetMouseWheelMoveV().Y)

		if rl.IsKeyDown(rl.KeyLeftShift) {
			if scroll > 0 {
				e.Position.Y += 1
			} else if scroll < 0 {
				e.Position.Y -= 1
			}
		} else {
			e.ScrollYVelocity += scroll
		}

		e.ScrollYVelocity *= friction
		e.ScrollY += e.ScrollYVelocity * timeStep.Seconds() * 120

		e.Scale = math.Pow(2, -e.ScrollY/50)

		e.Position = e.Position.Add(targetVelocity)

		currentMousePos := v2.FromRaylib(rl.GetMousePosition())

		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			mouseMove := (currentMousePos.Subtract(e.mousePosition)).Scale(0.007)
			e.Yaw += mouseMove.X
			e.Pitch -= mouseMove.Y

			if e.Pitch < -rl.Pi/2+1e-2 {
				e.Pitch = -rl.Pi/2 + 1e-2
			}

			if e.Pitch >= rl.Pi/2-1e-2 {
				e.Pitch = rl.Pi/2 - 1e-2
			}
		}
		e.mousePosition = currentMousePos

		mouseRay := rl.GetScreenToWorldRay(rl.Vector2{float32(currentMousePos.X), float32(currentMousePos.Y)}, e.Camera.Raylib())

		origin := v3.FromRaylib(mouseRay.Position)
		dir := v3.FromRaylib(mouseRay.Direction)
		ground := math.Floor(e.Position.Y)

		if math.Abs(dir.Y) >= 1e-6 {
			t := (ground - origin.Y) / dir.Y

			if t >= 0 {
				e.mouseWorldPosition = origin.Add(dir.Scale(t))
			}
		}
		e.mouseWorldPosition.Y = ground

		ix := math.Floor(e.mouseWorldPosition.X)
		iz := math.Floor(e.mouseWorldPosition.Z)
		fx := e.mouseWorldPosition.X - ix - 0.5
		fz := e.mouseWorldPosition.Z - iz - 0.5

		e.mouseCellPos = v3.XYZ(ix, e.mouseWorldPosition.Y, iz)

		if math.Abs(fx) > math.Abs(fz) {
			if fx < 0 {
				e.mouseCellDirection = FaceEast
			}
			if fx >= 0 {
				e.mouseCellDirection = FaceWest
			}
		} else {
			if fz > 0 {
				e.mouseCellDirection = FaceNorth
			}
			if fz <= 0 {
				e.mouseCellDirection = FaceSouth
			}
		}
	}

	if e.EditorDrawFlags&EditorDrawFlagFocusPlayer != 0 && e.world.Player != nil {
		e.Position = e.world.Player.Position
	}
	if e.EditorDrawFlags&EditorDrawFlagFocusMonster != 0 && e.world.Monster != nil {
		e.Position = e.world.Monster.Position
	}

	e.PositionSoft = e.PositionSoft.Lerp(e.Position, 1-friction)

	e.Camera = Camera3D{
		Position: e.PositionSoft.Subtract(v3.XYZ(
			math.Cos(e.Pitch)*math.Cos(e.Yaw),
			math.Sin(e.Pitch),
			math.Cos(e.Pitch)*math.Sin(e.Yaw),
		).Scale(e.Scale)),
		Target:     e.PositionSoft,
		Up:         v3.Y(1),
		Fovy:       70,
		Projection: rl.CameraPerspective,
	}

	if rl.IsKeyPressed(rl.KeyOne) {
		e.Tool = TOOL_CELL
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		e.Tool = TOOL_PLAY
	}

	switch e.Tool {
	case TOOL_CELL:
		e.ToolCell.Update(e)
	case TOOL_PLAY:
		e.world.Update(e.TimeStep)
	}
}

func (e *Editor) Draw() {
	globals.Shaders.Main.FullBright.Set(1)
	rl.ClearBackground(rl.DarkGray)

	BeginMode3D(e.Camera, func() {

		for pos, chunk := range e.world.Chunks {
			for y := range min(ChunkHeight, int(e.Position.Y)+1) {
				worldPos := ChunkToWorld(pos, LocalPos{0, y, 0})
				rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
			}
		}

		if e.world.Player != nil && e.world.Player.Position.Y <= e.Position.Y {
			e.world.Player.Draw()
		}
		if e.world.Monster != nil && e.world.Monster.Position.Y <= e.Position.Y {
			e.world.Monster.Draw()
		}

		BeginOverlayMode(func() {
			// rl.DrawCube(e.mouseWorldPosition.Raylib(), 0.05, 0.05, 0.05, rl.Black)
			// rl.DrawCube(e.mouseWorldPosition.Add(v3.X(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Red)
			// rl.DrawCube(e.mouseWorldPosition.Add(v3.Y(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Green)
			// rl.DrawCube(e.mouseWorldPosition.Add(v3.Z(0.25)).Raylib(), 0.05, 0.05, 0.05, rl.Blue)

			if e.EditorDrawFlags&EditorDrawFlagPhysics != 0 {
				phys := NewPhysicsDrawer(e.Position.Y, true, true, true)

				e.world.space.BBQuery(
					cp.NewBBForCircle(e.Position.Chipmunk(), 8),
					cp.NewShapeFilter(0, Category(e.Position.Y, true, true), Category(e.Position.Y, true, true)),
					func(shape *cp.Shape, data interface{}) {
						cp.DrawShape(shape, &phys)
						body := shape.Body()
						position := body.Position()
						velocity := body.Velocity()
						diff := v3.XYZ(velocity.X, 0, velocity.Y).Scale(e.TimeStep.Seconds() * 20)

						from := v3.XYZ(position.X, e.Position.Y, position.Y)
						to := from.Add(diff)

						rl.DrawLine3D(from.Raylib(), to.Raylib(), rl.Yellow)
					},
					nil,
				)
			}

			rl.DrawCubeWiresV(e.mouseCellPos.AddXYZ(0.5, 0, 0.5).Raylib(), v3.XYZ(1, 0, 1).Raylib(), color.RGBA{255, 255, 255, 255})
			rl.DrawSphere(e.PositionSoft.Raylib(), float32(e.Scale/400), color.RGBA{0, 255, 0, 255})

			for cpos, _ := range e.world.Chunks {

				rl.DrawCubeWires(v3.XYZ(float64(cpos.X)*ChunkWidth+ChunkWidth*0.5, 0, float64(cpos.Z)*ChunkWidth+ChunkWidth*0.5).Raylib(), ChunkWidth, 0, ChunkWidth, color.RGBA{255, 255, 255, 10})
			}

			switch e.Tool {
			case TOOL_CELL:
				e.ToolCell.Draw3D(e)
			}

			if e.world.Player != nil {
				point := NewPathPoint(e.world, e.Position)
				path, _ := point.FindPath(e.world.Player.Position)
				DrawPath(path)
			}
		})

	})
	globals.Shaders.Main.FullBright.Set(0)

	switch e.Tool {
	case TOOL_CELL:
		e.ToolCell.DrawHUD(e)
	}

	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT, int64(raygui.TEXT_ALIGN_LEFT))
	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT_VERTICAL, int64(raygui.TEXT_ALIGN_MIDDLE))
	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_PADDING, 4)

	stack := NewStackLayout(10, 30, 120, 24)

	if raygui.Toggle(stack.Down(24), raygui.IconText(raygui.ICON_LASER, "Show Bodies"), e.EditorDrawFlags&EditorDrawFlagPhysics != 0) {
		e.EditorDrawFlags |= EditorDrawFlagPhysics
	} else {
		e.EditorDrawFlags &^= EditorDrawFlagPhysics
	}

	if raygui.Toggle(stack.Down(24), raygui.IconText(raygui.ICON_CAMERA, "Focus Player"), e.EditorDrawFlags&EditorDrawFlagFocusPlayer != 0) {
		e.EditorDrawFlags |= EditorDrawFlagFocusPlayer
		e.EditorDrawFlags &^= EditorDrawFlagFocusMonster
	} else {
		e.EditorDrawFlags &^= EditorDrawFlagFocusPlayer
	}

	if raygui.Toggle(stack.Down(24), raygui.IconText(raygui.ICON_CAMERA, "Focus Monster"), e.EditorDrawFlags&EditorDrawFlagFocusMonster != 0) {
		e.EditorDrawFlags |= EditorDrawFlagFocusMonster
		e.EditorDrawFlags &^= EditorDrawFlagFocusPlayer
	} else {
		e.EditorDrawFlags &^= EditorDrawFlagFocusMonster
	}

	if raygui.Button(stack.Down(24), "TP Player") {
		player := game.Earth.Player
		if player == nil {
			player = game.Station.Player
		}

		player.Spawn(e.world)
		player.body.SetPosition(e.Position.Chipmunk())
		player.Update()
	}

	RenderPresetGroup(e, raygui.IconText(raygui.ICON_CUBE_FACE_FRONT, ""), NewStackLayout(200, 200, 40, 40), e.editorPresets.Wall)
	RenderPresetGroup(e, raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, ""), NewStackLayout(240, 200, 40, 40), e.editorPresets.Floor)
	RenderPresetGroup(e, raygui.IconText(raygui.ICON_VERTICAL_BARS, ""), NewStackLayout(280, 200, 40, 40), e.editorPresets.Stair)

}

func RenderPresetGroup(e *Editor, icon string, stack *StackLayout, presets []EditorPreset) {

	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT, int64(raygui.TEXT_ALIGN_CENTER))
	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT_VERTICAL, int64(raygui.TEXT_ALIGN_MIDDLE))
	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_PADDING, 4)
	// raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, 10)

	raygui.Toggle(stack.Down(40), icon, false)

	for i := range presets {
		preset := &presets[i]
		buttonRect := stack.Down(40)
		if raygui.Button(buttonRect, "") {
			for i := range presets {
				presets[i].Active = false
			}
			preset.Activate(e)
			preset.Active = true
		}
		tileWidth := float32(globals.Textures.Atlas.Width) / float32(globals.Textures.AtlasTiles)
		destRect := rl.NewRectangle(buttonRect.X+1, buttonRect.Y+1, buttonRect.Width-2, buttonRect.Height-2)
		sourceRect := rl.NewRectangle(
			float32(preset.TileX-1)*tileWidth,
			float32(preset.TileY-1)*tileWidth,
			tileWidth,
			tileWidth,
		)
		rl.DrawTexturePro(globals.Textures.Atlas, sourceRect, destRect, v2.Zero.Raylib(), 0, rl.White)
	}
}
