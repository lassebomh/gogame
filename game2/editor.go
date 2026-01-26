package game2

import (
	"image/color"
	"math"
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type EditorTool = int32

const (
	TOOL_WALLS = EditorTool(iota)
	TOOL_FLOOR
	TOOL_PLAY
)

type Editor struct {
	Camera Camera3D
	Pitch  float64
	Yaw    float64

	MousePosition Vec2
	Y             float64
	HitPos        Vec3

	Tool      EditorTool
	ToolFloor ToolFloor
	ToolWall  ToolWall
}

func NewEditor() *Editor {
	return (&Editor{
		Y: 0,
		Camera: Camera3D{
			Projection: rl.CameraPerspective,
			Position:   NewVec3(-8, 1, 0),
			Target:     NewVec3(0, 0, 0),
			Up:         Y,
			Fovy:       75,
		},
		Pitch: 0,
		Yaw:   0,
		Tool:  TOOL_WALLS,
	})
}

func (e *Editor) Update(g *Game, dt time.Duration) {
	const SPEED = 0.05

	if rl.IsKeyPressed(rl.KeyOne) {
		e.Tool = TOOL_FLOOR
	}

	if rl.IsKeyPressed(rl.KeyTwo) {
		e.Tool = TOOL_WALLS
	}

	if rl.IsKeyPressed(rl.KeyThree) {
		e.Tool = TOOL_PLAY
	}

	if e.Tool != TOOL_PLAY {

		forward := e.Camera.Target.Subtract(e.Camera.Position).Normalize()
		right := forward.CrossProduct(e.Camera.Up).Normalize()

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
			e.Camera.Position = e.Camera.Position.Add(movement)
		}

		currentMousePos := Vec2FromRaylib(rl.GetMousePosition())

		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			mouseMove := (currentMousePos.Subtract(e.MousePosition)).Scale(0.005)
			e.Yaw += mouseMove.X
			e.Pitch -= mouseMove.Y

			if e.Pitch < -Pi/2+1e-5 {
				e.Pitch = -Pi/2 + 1e-5
			}

			if e.Pitch >= Pi/2-1e-5 {
				e.Pitch = Pi/2 - 1e-5
			}
		}
		e.MousePosition = currentMousePos

		mouseRay := rl.GetScreenToWorldRay(rl.Vector2{float32(currentMousePos.X), float32(currentMousePos.Y)}, e.Camera.Raylib())

		origin := Vec3FromRaylib(mouseRay.Position)
		dir := Vec3FromRaylib(mouseRay.Direction)
		ground := math.Floor(e.Y)

		if math.Abs(dir.Y) >= 1e-6 {
			t := (ground - origin.Y) / dir.Y

			if t >= 0 {

				e.HitPos = origin.Add(dir.Scale(t))
				e.HitPos.Y = ground

			}
		}

		scroll := float64(rl.GetMouseWheelMove())
		if scroll != 0 {
			yDiff := scroll / math.Abs(scroll)

			e.Y += yDiff
		}
	}

	e.Camera.Target = e.Camera.Position.Add(NewVec3(
		math.Cos(e.Pitch)*math.Cos(e.Yaw),
		math.Sin(e.Pitch),
		math.Cos(e.Pitch)*math.Sin(e.Yaw),
	))

	switch e.Tool {
	case TOOL_WALLS:
		e.ToolWall.Update(g, e)
	case TOOL_FLOOR:
		e.ToolFloor.Update(g, e)
	case TOOL_PLAY:
		g.Update(dt)
	}
}

func (e *Editor) Draw(g *Game) {
	rl.ClearBackground(rl.Black)

	camera := e.Camera
	maxY := int(e.Y)

	if e.Tool == TOOL_PLAY {
		camera = g.Camera
		maxY = int(g.Player.Y)
	}

	BeginMode3D(camera, func() {
		g.MainShader.FullBright.Set(1)
		g.Draw3D(maxY)

		BeginOverlayMode(func() {
			g.Monster.PathFinder.Draw3D(g)
			for _, arm := range g.Monster.arms {
				tip := arm.Tip()

				rl.DrawLine3D(tip.Position3D().Raylib(), arm.tipTarget.Raylib(), rl.Blue)
			}

			if g.RenderFlags&(RENDER_FLAG_PHYSICS) != 0 {
				drawer := NewPhysicsDrawer(float64(maxY), true, true, true)

				renderBody := func(body *cp.Body) {
					body.EachConstraint(func(c *cp.Constraint) {
						cp.DrawConstraint(c, &drawer)
					})
					body.EachShape(func(s *cp.Shape) {
						if s.Filter.Categories&(1<<uint(math.Floor(float64(maxY)))) != 0 {
							cp.DrawShape(s, &drawer)
						}
					})
				}

				renderBody(g.Space.StaticBody)
				g.Space.EachBody(renderBody)
			}

			center := e.HitPos.Floor().Add(NewVec3(0.5, 0, 0.5))
			size := float64(1)

			for x := -size; x <= size; x++ {
				for z := -size; z <= size; z++ {
					pos := center.Add(NewVec3(x, 0, z))
					rl.DrawCubeWiresV(pos.Raylib(), XZ.Raylib(), color.RGBA{255, 255, 255, 50})
				}
			}

			switch e.Tool {
			case TOOL_WALLS:
				e.ToolWall.Draw3D(g, e)
			case TOOL_FLOOR:
				e.ToolFloor.Draw3D(g, e)
			}

		})

	})

	size := float64(30)
	line := NewLineLayout(90, 0, size)

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_PHOTO_CAMERA_FLASH, ""), g.RenderFlags&RENDER_FLAG_FULLBRIGHT != 0) {
		g.RenderFlags |= RENDER_FLAG_FULLBRIGHT
	} else {
		g.RenderFlags &^= RENDER_FLAG_FULLBRIGHT
	}

	if raygui.Toggle(line.Next(size), raygui.IconText(raygui.ICON_LASER, ""), g.RenderFlags&RENDER_FLAG_PHYSICS != 0) {
		g.RenderFlags |= RENDER_FLAG_PHYSICS
	} else {
		g.RenderFlags &^= RENDER_FLAG_PHYSICS
	}

	switch e.Tool {
	case TOOL_WALLS:
		e.ToolWall.DrawHUD(g, e)
	case TOOL_FLOOR:
		e.ToolFloor.DrawHUD(g, e)
	}

}
