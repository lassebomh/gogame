package game2

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Player struct {
	Y      float64
	Radius float64
	Body   *cp.Body
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
	player := &Player{
		Radius: 0.25,
		Y:      save.Y,
		Body:   nil,
	}

	mass := player.Radius * player.Radius * 4
	body := g.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, player.Radius, Vec2{2, 2}.CP())))
	body.SetPosition(save.Position.CP())

	shape := g.Space.AddShape(cp.NewCircle(body, player.Radius, Vec2{}.CP()))
	shape.SetElasticity(0)
	shape.SetFriction(0)
	player.Body = body
	g.Player = player

	return player
}

func (p *Player) Draw(g *Game) {
	rl.DrawSphere(g.Player.Position3D().Add(Y.Scale(g.Player.Radius)).Raylib(), float32(g.Player.Radius), rl.Red)
}
