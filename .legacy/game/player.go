package game

import (
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Player struct {
	Body  *cp.Body
	Speed float64

	Health    float32
	HealthMax float32

	Radius float64
	Items  []Item
}

func (p *Player) RemoveItem(w *World, target Item) {
	for i, item := range p.Items {
		if item == target {
			last := len(p.Items) - 1
			p.Items[i] = p.Items[last]
			p.Items = p.Items[:last]
			break
		}
	}
}

func (p *Player) TakeItem(w *World, pitem *PhysicalItem) {
	w.RemovePhysicalItem(pitem)
	w.Player.Items = append(w.Player.Items, pitem.Item)
}

func NewPlayer(w *World, pos cp.Vector) *Player {

	player := &Player{
		Speed:     40,
		Health:    100,
		HealthMax: 100,

		Items:  make([]Item, 0),
		Radius: 2,
	}

	mass := player.Radius * player.Radius * 4
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, player.Radius, cp.Vector{})))
	body.SetPosition(pos)

	shape := w.Space.AddShape(cp.NewCircle(body, player.Radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)
	player.Body = body

	return player
}

func (p *Player) Update(w *World) {
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
		force = force.Normalize().Mult(p.Speed)
	}

	newVelocity := p.Body.Velocity().Lerp(force, 0.1)
	p.Body.SetVelocity(newVelocity.X, newVelocity.Y)

	for _, item := range p.Items {
		item.InventoryUpdate(w, p)
	}

	mousePositionDiff := w.MouseWorldPosition.Sub(p.Body.Position())
	angleToMouse := math.Atan2(mousePositionDiff.Y, mousePositionDiff.X)
	p.Body.SetAngle(angleToMouse)
}

func (p *Player) Teleport(from *World, to *World) {
	p.Body.EachShape(func(shape *cp.Shape) {
		from.Space.RemoveShape(shape)
	})
	from.Space.RemoveBody(p.Body)

	temp := NewPlayer(to, from.Tilemap.CenterPosition.Sub(p.Body.Position()).Add(to.Tilemap.CenterPosition))
	p.Body = temp.Body
	from.Player = nil
	to.Player = p
}

func (p *Player) RenderHud(w *World, font rl.Font) {

	t := w.Day * 24 * float64(time.Hour)
	t /= (10 * float64(time.Minute))
	t = math.Floor(t)
	t *= (10 * float64(time.Minute))

	clock := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(int64(t))).Format("15:04")
	rl.DrawTextEx(font, clock, rl.NewVector2(12, 12), float32(font.BaseSize), 1, rl.Green)
}
