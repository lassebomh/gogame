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

type EditorFocusMode = int32

const (
	EditorFocusNone = EditorFocusMode(iota)
	EditorFocusPlayer
	EditorFocusMonster
)

type EditorSettings struct {
	Play         bool
	DrawPhysics  bool
	Orthographic bool
	FocusMode    EditorFocusMode
}

type Editor struct {
	world    *World
	TimeStep time.Duration

	Position     v3.Value
	PositionSoft v3.Value

	Pitch           float64
	Yaw             float64
	ScrollY         float64
	ScrollYVelocity float64
	Scale           float64

	editorPresets *EditorPresets

	contextMenu v2.Value

	Camera   Camera3D
	Settings EditorSettings

	mousePosition      v2.Value
	mouseWorldPosition v3.Value

	mouseCellPos       v3.Value
	mouseCellDirection FaceDirection

	Tool     TOOL
	ToolCell ToolCell
}

type TOOL = int32

const (
	TOOL_CELL = TOOL(iota)
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

	if e.Settings.Orthographic {
		forward = v3.Z(1)
		right = v3.X(-1)
	}

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
	}

	if rl.IsKeyDown(rl.KeyQ) && e.contextMenu == v2.Zero {
		e.contextMenu = e.mousePosition
	}
	if rl.IsKeyReleased(rl.KeyQ) {
		e.contextMenu = v2.Zero
	}

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

	if !e.Settings.Play {

		e.Position = e.Position.Add(targetVelocity)

		currentMousePos := v2.FromRaylib(rl.GetMousePosition())

		if e.contextMenu == v2.Zero {
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

			mouseRay := rl.GetScreenToWorldRay(e.mousePosition.Raylib(), e.Camera.Raylib())

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
	}

	if e.Settings.FocusMode == EditorFocusPlayer && e.world.Player != nil {
		// e.Position = e.world.Player.Position
		e.PositionSoft = e.PositionSoft.Lerp(e.world.Player.Position.AddXYZ(0, 0.1, 0), 1-friction)
	} else if e.Settings.FocusMode == EditorFocusMonster && e.world.Monster != nil {
		// e.Position = e.world.Monster.Position
		e.PositionSoft = e.PositionSoft.Lerp(e.world.Monster.Position.AddXYZ(0, 0.1, 0), 1-friction)
	} else {
		e.PositionSoft = e.PositionSoft.Lerp(e.Position.AddXYZ(0, 0.1, 0), 1-friction)
	}

	if e.Settings.Orthographic {
		e.Camera = Camera3D{
			Position:   e.PositionSoft.AddXYZ(0, e.Scale, -0.0001),
			Target:     e.PositionSoft,
			Up:         v3.Y(1),
			Fovy:       e.Scale,
			Projection: rl.CameraOrthographic,
		}
	} else {
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
	}

	if rl.IsKeyPressed(rl.KeyOne) {
		e.Settings.Play = !e.Settings.Play
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		e.Settings.DrawPhysics = !e.Settings.DrawPhysics
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		e.Settings.Orthographic = !e.Settings.Orthographic
	}
	if rl.IsKeyPressed(rl.KeyFour) {
		switch e.Settings.FocusMode {
		case EditorFocusNone:
			e.Settings.FocusMode = EditorFocusMonster
		case EditorFocusPlayer:
			e.Settings.FocusMode = EditorFocusNone
		case EditorFocusMonster:
			e.Settings.FocusMode = EditorFocusPlayer
		}
	}

	if e.Settings.Play {
		game.Update(e.TimeStep)
	} else {
		switch e.Tool {
		case TOOL_CELL:
			e.ToolCell.Update(e)
		}
	}

}

func (e *Editor) Draw() {
	globals.Shaders.Main.FullBright.Set(1)
	rl.ClearBackground(rl.DarkGray)

	BeginMode3D(e.Camera, func() {

		for pos, chunk := range e.world.Chunks {
			for y := range min(ChunkHeight, int(e.PositionSoft.Y)+1) {
				worldPos := ChunkToWorld(pos, LocalPos{0, y, 0})
				rl.DrawModel(chunk.models[y], worldPos.Raylib(), 1, rl.White)
			}
		}

		if e.world.Player != nil && e.world.Player.Position.Y <= math.Floor(e.PositionSoft.Y) {
			e.world.Player.Draw()
		}
		if e.world.Monster != nil {
			e.world.Monster.Draw()
		}

		BeginOverlayMode(func() {

			if e.Settings.DrawPhysics {
				phys := NewPhysicsDrawer(e.PositionSoft.Y, true, true, true)

				e.world.space.BBQuery(
					cp.NewBBForCircle(e.PositionSoft.Chipmunk(), 8),
					cp.NewShapeFilter(0, Category(e.PositionSoft.Y, true, true), Category(e.PositionSoft.Y, true, true)),
					func(shape *cp.Shape, data interface{}) {
						cp.DrawShape(shape, &phys)
					},
					nil,
				)
				oldLineWidth := rl.GetLineWidth()
				rl.SetLineWidth(1)
				for _, ray := range e.world.raycastResults {
					col := color.RGBA{0, 0, 255, 255}
					if ray.Hit {
						col = color.RGBA{0, 255, 0, 255}
						rl.DrawSphere(ray.Position.Raylib(), 0.03, col)
					}
					rl.DrawLine3D(ray.Start.Raylib(), ray.End.Raylib(), col)
				}
				rl.SetLineWidth(oldLineWidth)

			}

			rl.DrawCubeWiresV(e.mouseCellPos.AddXYZ(0.5, 0, 0.5).Raylib(), v3.XYZ(1, 0, 1).Raylib(), color.RGBA{255, 255, 255, 255})
			rl.DrawSphere(e.PositionSoft.Raylib(), float32(e.Scale/400), color.RGBA{0, 255, 0, 255})

			for cpos := range e.world.Chunks {
				rl.DrawCubeWires(v3.XYZ(float64(cpos.X)*ChunkWidth+ChunkWidth*0.5, 0, float64(cpos.Z)*ChunkWidth+ChunkWidth*0.5).Raylib(), ChunkWidth, 0, ChunkWidth, color.RGBA{255, 255, 255, 10})
			}

			// if e.world.Monster != nil {
			// 	for _, arm := range e.world.Monster.arms {
			// 		DrawPath(arm.path)
			// 	}
			// }

			// pathPoint := NewPathPoint(e.world, e.PositionSoft)
			// if pathPoint != nil && e.world.Player != nil {
			// 	path, _ := pathPoint.FindPath(e.world.Player.Position)
			// 	DrawPath(path)

			// 	for _, p := range pathPoint.GetNeighborPathPoints() {
			// 		p.Draw()
			// 	}
			// }

			switch e.Tool {
			case TOOL_CELL:
				e.ToolCell.Draw3D(e)
			}
		})
	})

	globals.Shaders.Main.FullBright.Set(0)

	switch e.Tool {
	case TOOL_CELL:
		e.ToolCell.DrawHUD(e)
	}

	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_ALIGNMENT, int64(raygui.TEXT_ALIGN_LEFT))
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_ALIGNMENT_VERTICAL, int64(raygui.TEXT_ALIGN_MIDDLE))
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_PADDING, 2)

	topbar := NewCursorLayout(300, 0, 24, 24)

	e.Settings.Play = raygui.Toggle(topbar.Right().Rectangle, raygui.IconText(raygui.ICON_PLAYER_PLAY, ""), e.Settings.Play)
	e.Settings.DrawPhysics = raygui.Toggle(topbar.Right().Rectangle, raygui.IconText(raygui.ICON_LASER, ""), e.Settings.DrawPhysics)
	e.Settings.Orthographic = raygui.Toggle(topbar.Right().Rectangle, raygui.IconText(raygui.ICON_MODE_2D, ""), e.Settings.Orthographic)

	if raygui.Button(topbar.Right().Rectangle, raygui.IconText(raygui.ICON_PLAYER_NEXT, "")) {
		e.world.Update(time.Second / 60)
	}

	topbar.Right().With(func(cursor *CursorLayout) {
		var icon int32

		switch e.Settings.FocusMode {
		case EditorFocusNone:
			icon = raygui.ICON_TARGET_MOVE
		case EditorFocusPlayer:
			icon = raygui.ICON_PLAYER
		case EditorFocusMonster:
			icon = raygui.ICON_DEMON
		}

		if raygui.Button(cursor.Rectangle, raygui.IconText(icon, "")) {
			switch e.Settings.FocusMode {
			case EditorFocusNone:
				e.Settings.FocusMode = EditorFocusMonster
			case EditorFocusPlayer:
				e.Settings.FocusMode = EditorFocusNone
			case EditorFocusMonster:
				e.Settings.FocusMode = EditorFocusPlayer
			}
		}
	})

	sidebar := NewCursorLayout(0, 100, 30, 30)

	RenderPresetGroup(e, raygui.IconText(raygui.ICON_CUBE_FACE_FRONT, "Wall"), sidebar, e.editorPresets.Wall)
	RenderPresetGroup(e, raygui.IconText(raygui.ICON_CUBE_FACE_BOTTOM, "Floor"), sidebar, e.editorPresets.Floor)
	RenderPresetGroup(e, raygui.IconText(raygui.ICON_VERTICAL_BARS, "Stair"), sidebar, e.editorPresets.Stair)

	contextMenu := NewStackLayout(0, 20, 120, 24)

	if raygui.Button(contextMenu.Down(24), raygui.IconText(raygui.ICON_TARGET, "Teleport Player")) {
		player := game.Earth.Player
		if player == nil {
			player = game.Station.Player
		}

		player.Spawn(e.world)
		pos := e.Position
		player.Position.Y = pos.Y
		player.body.SetPosition(pos.Chipmunk())
		player.Update()
	}

	if raygui.Button(contextMenu.Down(24), raygui.IconText(raygui.ICON_ARROW_UP, "Up")) {
		e.Position.Y++
	}
	if raygui.Button(contextMenu.Down(24), raygui.IconText(raygui.ICON_ARROW_DOWN, "Down")) {
		e.Position.Y--
	}
}

func RenderPresetGroup(e *Editor, title string, cursor *CursorLayout, presets []EditorPreset) {

	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT, int64(raygui.TEXT_ALIGN_CENTER))
	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_ALIGNMENT_VERTICAL, int64(raygui.TEXT_ALIGN_MIDDLE))
	raygui.SetStyle(raygui.TOGGLE, raygui.TEXT_PADDING, 4)
	tileWidth := float32(30.)
	tileHorizontal := 2

	cursor.X = 0
	cursor.Width = tileWidth * float32(tileHorizontal)
	raygui.Toggle(cursor.Down().Rectangle, title, false)

	cursor.Width = tileWidth
	cursor.Height = tileWidth

	for i, preset := range presets {
		if i%tileHorizontal == 0 {
			cursor.X = -tileWidth
			cursor.Y += tileWidth
		}
		buttonRect := cursor.Right().Rectangle
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
