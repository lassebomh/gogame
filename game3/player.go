package game3

import (
	v2 "game/vec2"
	v3 "game/vec3"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const VISIBILITY_VERTS = 240
const VISIBILITY_CONE_RADIANS = math.Pi / 3
const VISIBILITY_DISTANCE = 10

type Player struct {
	DynamicPhysicsObject

	// Position v3.Value
	// YVelocity float64
	// body      *cp.Body
	// shape     *cp.Shape
	Radius float64

	visibilityVerts [VISIBILITY_VERTS]v3.Value

	viewTexture  rl.RenderTexture2D
	lookPosition v3.Value
}

func (p *Player) Update() {
	force := cp.Vector{}

	if rl.IsKeyDown(rl.KeyA) {
		force = force.Add(cp.Vector{X: 1})
	}
	if rl.IsKeyDown(rl.KeyD) {
		force = force.Add(cp.Vector{X: -1})
	}
	if rl.IsKeyDown(rl.KeyS) {
		force = force.Add(cp.Vector{Y: -1})
	}
	if rl.IsKeyDown(rl.KeyW) {
		force = force.Add(cp.Vector{Y: 1})
	}

	forceMag := force.Length()

	if forceMag != 0 {
		force = force.Normalize()
	}

	if rl.IsKeyDown(rl.KeyLeftShift) {
		force = force.Mult(5)
	} else {
		force = force.Mult(3)

	}

	newVelocity := p.body.Velocity().Lerp(force, 0.1)
	p.body.SetVelocity(newVelocity.X, newVelocity.Y)

	p.UpdatePhysics()

	// look position
	if math.Abs(p.world.MouseRayDirection.Y) >= 1e-6 {
		t := (p.Position.Y - p.world.MouseRayOrigin.Y) / p.world.MouseRayDirection.Y

		if t >= 0 {
			p.lookPosition = p.world.MouseRayOrigin.Add(p.world.MouseRayDirection.Scale(t))
			p.lookPosition.Y = p.Position.Y
		}
	}

	// update visibility verticies
	playerAngle := math.Atan2(
		p.Position.Z-p.lookPosition.Z,
		p.Position.X-p.lookPosition.X,
	)
	p.body.SetAngle(playerAngle)

	p.visibilityVerts[0] = p.Position

	from := v2.XY(p.Position.X, p.Position.Z)

	for i := range VISIBILITY_VERTS - 1 {
		f := (float64(i)/float64(VISIBILITY_VERTS-2))*2 - 1

		baseAngle := playerAngle
		angleOffset := f*VISIBILITY_CONE_RADIANS - math.Pi
		angle := baseAngle - angleOffset
		dir := v2.XY(math.Cos(angle), math.Sin(angle))
		to := from.Add(dir.Scale(VISIBILITY_DISTANCE))

		result := p.world.space.SegmentQueryFirst(
			from.Chipmunk(),
			to.Chipmunk(),
			0,
			cp.NewShapeFilter(
				0,
				Category(p.Position.Y, true, false),
				Category(p.Position.Y, true, false),
			),
		)

		p.visibilityVerts[i+1] = v3.XYZ(result.Point.X, p.Position.Y, result.Point.Y)
	}

	// update last seen cells
	cellCoords := map[v3.Value]bool{}

	a := p.visibilityVerts[0]
	for _, b := range p.visibilityVerts[:len(p.visibilityVerts)-1] {
		step := (1 / a.Distance(b)) / 2

		for i := float64(0); i <= 1; i += step {
			cellCoords[a.Lerp(b, i).Map(math.Floor)] = true
		}
	}

	// for cellCoord := range cellCoords {
	// 	cell, _ := p.world.GetCell(cellCoord)
	// 	cell.LastSeenPlayer = g.Time
	// }

	p.UpdateView()

}

func SpawnNewPlayer(world *World) {
	p := &Player{
		Radius: 0.1,
		DynamicPhysicsObject: DynamicPhysicsObject{
			Position: v3.XYZ(0, 0, 0),
		},
	}
	p.Spawn(world)
}

func (p *Player) Spawn(world *World) {
	if p.world != nil {
		p.world.Player = nil
	}
	if p.shape != nil {
		p.world.space.RemoveShape(p.shape)
	}
	if p.body != nil {
		p.world.space.RemoveBody(p.body)
	}

	world.Player = p
	p.world = world
	p.Radius = 0.2

	mass := p.Radius * p.Radius * 4
	body := p.world.space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, p.Radius, v2.XY(2, 2).Chipmunk())))
	body.SetPosition(p.Position.Chipmunk())
	p.shape = p.world.space.AddShape(cp.NewCircle(body, p.Radius, cp.Vector{}))
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.Filter.Group = GroupPlayer
	p.body = body
	p.UpdatePhysics()

	if !rl.IsRenderTextureValid(p.viewTexture) {
		p.viewTexture = rl.LoadRenderTexture(16*40, 16*40)
	}
}

func (p *Player) Draw() {
	rl.DrawSphere(p.Position.Add(v3.Y(p.Radius)).Raylib(), float32(p.Radius), rl.Red)
}

func (p *Player) UpdateView() {
	BeginTextureMode(p.viewTexture, func() {
		camera := Camera3D{
			Position:   p.Position.Add(v3.Y(5)),
			Target:     p.Position.AddXYZ(0, 0, 0.0001),
			Fovy:       20,
			Projection: rl.CameraOrthographic,
			Up:         v3.Y(1),
		}
		BeginMode3D(camera, func() {
			rl.ClearBackground(color.RGBA{})

			// for x := float64(-camera.Fovy); x <= camera.Fovy+1; x++ {
			// 	for z := float64(-camera.Fovy); z <= camera.Fovy+1; z++ {
			// 		cell, chunk := p.world.GetCell(p.Position.AddXYZ(x, 0, z))
			// 		cellPosition := ChunkToWorld(chunk.Position, LocalPos{int(x), 0, int(z)})
			// 		pos := v3.XYZ(cell.Position.X+0.5, p.Position.Y-0.5, cell.Position.Z+0.5)
			// 		seen := float64(0)
			// 		if cell.LastSeenPlayer != 0 {
			// 			seen = Clamp((5-(game.Time-cell.LastSeenPlayer).Seconds())/4, 0, 1)
			// 		}
			// 		rl.DrawCube(pos.Raylib(), 1, 0, 1, v3.X(seen).ToColor())
			// 	}
			// }

			a := p.visibilityVerts[0]
			for i, b := range p.visibilityVerts[:len(p.visibilityVerts)-1] {
				c := p.visibilityVerts[i+1]

				rl.DrawTriangle3D(a.Raylib(), b.Raylib(), c.Raylib(), color.RGBA{0, 255, 0, 255})
			}
		})
	})

}
