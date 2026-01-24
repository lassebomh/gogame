package game2

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Player struct {
	Y         float64
	YVelocity float64
	Radius    float64
	Body      *cp.Body
	Shape     *cp.Shape

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

	newVelocity := p.Body.Velocity().Lerp(force, 0.1)
	p.Body.SetVelocity(newVelocity.X, newVelocity.Y)

	playerPos := p.Position3D()
	cell := g.Level.GetCell(playerPos)

	var groundY float64

	switch cell.Ground.Type {
	case GroundStair:
		x := playerPos.X - math.Floor(playerPos.X)
		z := playerPos.Z - math.Floor(playerPos.Z)

		switch cell.Ground.Rotation {
		case 0:
			groundY = cell.Position.Y + z
		case 90:
			groundY = cell.Position.Y + x
		case 180:
			groundY = cell.Position.Y + 1 - z
		case 270:
			groundY = cell.Position.Y + 1 - x
		}
		groundY += 0
	case GroundSolid:
		groundY = cell.Position.Y
	case GroundNone:
		groundY = 0
	}

	if p.Y > groundY {
		p.YVelocity -= g.TimeDelta.Seconds() / 6
	}

	if p.Y+p.YVelocity < groundY {
		p.Y = groundY
		p.YVelocity = 0
	}

	p.Y += p.YVelocity

	nextCell := g.Level.GetCell(playerPos.Add(Y.Scale(0.1)))

	if nextCell.Position.Y > p.Y && nextCell.Ground.Type == GroundSolid {
		p.Y = math.Ceil(p.Y)
	}

	yLevelCategory := uint(1 << uint(math.Floor(p.Y)))
	p.Shape.Filter.Categories = yLevelCategory
	p.Shape.Filter.Mask = yLevelCategory | (1 << uint(math.Floor(p.Y+0.25)))

	if math.Abs(g.MouseRayDirection.Y) >= 1e-6 {
		t := (p.Y - g.MouseRayOrigin.Y) / g.MouseRayDirection.Y

		if t >= 0 {
			p.LookPosition = g.MouseRayOrigin.Add(g.MouseRayDirection.Scale(t))
			p.LookPosition.Y = p.Y
		}
	}

	playerAngle := math.Atan2(
		playerPos.Z-p.LookPosition.Z,
		playerPos.X-p.LookPosition.X,
	)

	p.Body.SetAngle(playerAngle)
}

func (p *Player) Position3D() Vec3 {
	return Vec3From2D(Vec2FromCP(p.Body.Position()), p.Y)
}

type PlayerSave struct {
	Position Vec2
	Y        float64
}

func (p *Player) ToSave(g *Game) PlayerSave {
	return PlayerSave{
		Position: Vec2FromCP(p.Body.Position()),
		Y:        p.Y,
	}
}

func (save PlayerSave) Load(g *Game) *Player {
	p := &Player{
		Radius: 0.25,
		Y:      save.Y,
		Body:   nil,
	}

	mass := p.Radius * p.Radius * 4
	body := g.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, p.Radius, Vec2{2, 2}.CP())))
	body.SetPosition(save.Position.CP())

	p.Shape = g.Space.AddShape(cp.NewCircle(body, p.Radius, Vec2{}.CP()))
	p.Shape.SetElasticity(0)
	p.Shape.SetFriction(0)
	p.Body = body
	g.Player = p

	yLevelCategory := uint(1 << uint(math.Floor(p.Y)))
	p.Shape.Filter.Categories = yLevelCategory
	p.Shape.Filter.Mask = yLevelCategory | (1 << uint(math.Floor(p.Y+0.25)))

	return p
}

func (p *Player) Draw(g *Game) {
	rl.DrawSphere(g.Player.Position3D().Add(Y.Scale(g.Player.Radius)).Raylib(), float32(g.Player.Radius), rl.Red)
}
