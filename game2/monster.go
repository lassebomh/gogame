package game2

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Monster struct {
	Y          float64
	YVelocity  float64
	Radius     float64
	body       *cp.Body
	shape      *cp.Shape
	PathFinder *PathFinder

	SavePosition Vec2
}

func (p *Monster) Update(g *Game) {
	monsterPos := p.Position3D()

	p.PathFinder.SetPosition(p.Position3D())
	p.PathFinder.SetTarget(g.Player.Position3D())

	force := cp.Vector{}

	if !p.PathFinder.Idle && len(p.PathFinder.Path) > 2 {
		a := p.PathFinder.Path[0]
		b := p.PathFinder.Path[1]

		force3D := b.Lerp(a, 0.5).Subtract(monsterPos)
		force.X = force3D.X
		force.Y = force3D.Z
	}

	forceMag := force.Length()

	if forceMag != 0 {
		force = force.Normalize().Mult(4)
	}

	newVelocity := p.body.Velocity().Lerp(force, 0.1)
	p.body.SetVelocity(newVelocity.X, newVelocity.Y)

	cell := g.Level.GetCell(monsterPos)

	var groundY float64

	switch cell.Ground.Type {
	case GroundStair:
		x := math.Ceil(monsterPos.X) - monsterPos.X
		z := monsterPos.Z - math.Floor(monsterPos.Z)

		switch cell.Ground.StairDirection {
		case FACE_EAST:
			groundY = cell.Position.Y + x
		case FACE_NORTH:
			groundY = cell.Position.Y + z
		case FACE_WEST:
			groundY = cell.Position.Y + 1 - x
		case FACE_SOUTH:
			groundY = cell.Position.Y + 1 - z
		}
	case GroundFloor:
		groundY = cell.Position.Y
	case GroundEmpty:
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

	nextCell := g.Level.GetCell(monsterPos.Add(Y.Scale(0.1)))

	if nextCell.Position.Y > p.Y && nextCell.Ground.Type == GroundFloor {
		p.Y = math.Ceil(p.Y)
	}

	yLevelCategory := uint(1 << uint(math.Floor(p.Y)))
	p.shape.Filter.Categories = yLevelCategory
	p.shape.Filter.Mask = yLevelCategory | (1 << uint(math.Floor(p.Y+0.25)))

}

func (p *Monster) Position3D() Vec3 {
	return Vec3From2D(Vec2FromCP(p.body.Position()), p.Y)
}

func (p *Monster) ToSave(g *Game) *Monster {
	p.SavePosition = Vec2FromCP(p.body.Position())
	return p
}

func (p *Monster) Load(g *Game) *Monster {
	p.Radius = 0.3
	if p.PathFinder == nil {
		p.PathFinder = NewPathFinder(g.Level)
	}
	p.PathFinder.level = g.Level

	mass := p.Radius * p.Radius * 4
	body := g.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, p.Radius, Vec2{2, 2}.CP())))
	body.SetPosition(p.SavePosition.CP())

	p.shape = g.Space.AddShape(cp.NewCircle(body, p.Radius, Vec2{}.CP()))
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.body = body
	g.Monster = p

	yLevelCategory := uint(1 << uint(math.Floor(p.Y)))
	p.shape.Filter.Categories = yLevelCategory
	p.shape.Filter.Mask = yLevelCategory | (1 << uint(math.Floor(p.Y+0.25)))

	return p
}

func (p *Monster) Draw(g *Game) {
	rl.DrawSphere(g.Monster.Position3D().Add(Y.Scale(g.Monster.Radius)).Raylib(), float32(g.Monster.Radius), rl.Red)
}
