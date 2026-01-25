package game2

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const VISIBILITY_VERTS = 40
const VISIBILITY_CONE_RADIANS = math.Pi / 3
const VISIBILITY_DISTANCE = 8

type Player struct {
	Y         float64
	YVelocity float64
	Radius    float64
	body      *cp.Body
	shape     *cp.Shape

	visibilityVerts [VISIBILITY_VERTS]Vec3

	ViewTexture  rl.RenderTexture2D
	LookPosition Vec3
}

func (p *Player) Update(g *Game) {
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
		force = force.Normalize().Mult(4)
	}

	newVelocity := p.body.Velocity().Lerp(force, 0.1)
	p.body.SetVelocity(newVelocity.X, newVelocity.Y)

	p.Y, p.YVelocity = UpdatePhysicsY(g, p.shape, p.Y, p.YVelocity)

	if math.Abs(g.MouseRayDirection.Y) >= 1e-6 {
		t := (p.Y - g.MouseRayOrigin.Y) / g.MouseRayDirection.Y

		if t >= 0 {
			p.LookPosition = g.MouseRayOrigin.Add(g.MouseRayDirection.Scale(t))
			p.LookPosition.Y = p.Y
		}
	}

	playerPos := p.Position3D()

	playerAngle := math.Atan2(
		playerPos.Z-p.LookPosition.Z,
		playerPos.X-p.LookPosition.X,
	)
	p.body.SetAngle(playerAngle)

	p.visibilityVerts[0] = playerPos

	from := playerPos.To2D()

	for i := range VISIBILITY_VERTS - 1 {
		f := (float64(i)/float64(VISIBILITY_VERTS-2))*2 - 1

		baseAngle := (math.Floor((playerAngle/(math.Pi*2))*VISIBILITY_VERTS) / VISIBILITY_VERTS) * math.Pi * 2
		// baseAngle := playerAngle
		angleOffset := f*VISIBILITY_CONE_RADIANS - math.Pi
		angle := baseAngle - angleOffset
		dir := NewVec2(math.Cos(angle), math.Sin(angle))
		to := from.Add(dir.Scale(VISIBILITY_DISTANCE))

		result := g.Space.SegmentQueryFirst(from.CP(), to.CP(), 0, cp.NewShapeFilter(p.shape.Filter.Group, p.shape.Filter.Categories, p.shape.Filter.Mask))

		p.visibilityVerts[i+1] = NewVec3(result.Point.X, p.Y, result.Point.Y)
	}

}

func (p *Player) Position3D() Vec3 {
	return Vec3From2D(Vec2FromCP(p.body.Position()), p.Y)
}

type PlayerSave struct {
	Position Vec2
	Y        float64
}

func (p *Player) ToSave(g *Game) PlayerSave {
	return PlayerSave{
		Position: Vec2FromCP(p.body.Position()),
		Y:        p.Y,
	}
}

func (save PlayerSave) Load(g *Game) *Player {
	p := &Player{
		Radius:      0.25,
		Y:           save.Y,
		body:        nil,
		ViewTexture: rl.LoadRenderTexture(g.MainTexture.Texture.Width, g.MainTexture.Texture.Height),
	}

	mass := p.Radius * p.Radius * 4
	body := g.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, p.Radius, Vec2{2, 2}.CP())))
	body.SetPosition(save.Position.CP())

	p.shape = g.Space.AddShape(cp.NewCircle(body, p.Radius, Vec2{}.CP()))
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.Filter.Group = 3
	p.body = body
	g.Player = p

	yLevelCategory := uint(1 << uint(math.Floor(p.Y)))
	p.shape.Filter.Categories = yLevelCategory
	p.shape.Filter.Mask = yLevelCategory | (1 << uint(math.Floor(p.Y+0.25)))

	return p
}

func (p *Player) Draw(g *Game) {
	rl.DrawSphere(g.Player.Position3D().Add(Y.Scale(g.Player.Radius)).Raylib(), float32(g.Player.Radius), rl.Red)
}

func (p *Player) RenderViewTexture(g *Game) {
	BeginTextureMode(p.ViewTexture, func() {
		BeginMode3D(g.Camera, func() {
			rl.ClearBackground(color.RGBA{})

			a := p.visibilityVerts[0]
			for i, b := range p.visibilityVerts[:len(p.visibilityVerts)-1] {
				c := p.visibilityVerts[i+1]

				rl.DrawTriangle3D(a.Raylib(), b.Raylib(), c.Raylib(), color.RGBA{0, 255, 0, 255})
			}
		})
	})
}
