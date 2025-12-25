package game

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Item interface {
	Update(w *World)
	RenderHud(cursor *rl.Vector2) bool
	GetBody() *cp.Body
	SetBody(body *cp.Body)
}

func ItemSpawnBody(item Item, w *World, pos cp.Vector) {
	radius := 20.
	mass := radius * radius / 25.0
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	shape := w.Space.AddShape(cp.NewBox(body, radius, radius*2, 0))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)

	item.SetBody(body)
}

func ItemDespawnBody(item Item, w *World) {
	body := item.GetBody()
	body.EachShape(func(shape *cp.Shape) {
		w.Space.RemoveShape(shape)
	})
	w.Space.RemoveBody(body)
	item.SetBody(nil)
}

type Player struct {
	Body  *cp.Body
	Speed float64

	Health    float32
	HealthMax float32

	Inventory []Item
}

func (p *Player) DropItem(w *World, target Item) {
	targetIndex := -1

	for i, item := range p.Inventory {
		if item == target {
			targetIndex = i
			break
		}
	}

	if targetIndex != -1 {
		p.Inventory[targetIndex] = p.Inventory[len(p.Inventory)-1]
		p.Inventory = p.Inventory[:len(p.Inventory)-1]
		ItemSpawnBody(target, w, p.Body.Position())
		w.Items = append(w.Items, target)
	}

}

func NewPlayer(w *World) *Player {
	radius := 20.
	mass := radius * radius / 25.0
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: 200, Y: 200})

	shape := w.Space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)

	player := &Player{
		Body:      body,
		Speed:     500,
		Health:    100,
		HealthMax: 100,

		Inventory: make([]Item, 0),
	}

	player.Inventory = append(player.Inventory,
		&ItemOxygenTank{
			Oxygen:    50,
			OxygenMax: 100,
		},
		&ItemOxygenTank{
			Oxygen:    4,
			OxygenMax: 100,
		})

	return player
}

func (p *Player) Update(w *World) {
	force := cp.Vector{}

	if rl.IsKeyDown(rl.KeyA) {
		force = force.Add(cp.Vector{X: -1})
	}
	if rl.IsKeyDown(rl.KeyD) {
		force = force.Add(cp.Vector{X: 1})
	}
	if rl.IsKeyDown(rl.KeyS) {
		force = force.Add(cp.Vector{Y: 1})
	}
	if rl.IsKeyDown(rl.KeyW) {
		force = force.Add(cp.Vector{Y: -1})
	}

	forceMag := force.Length()

	if forceMag != 0 {
		force = force.Normalize().Mult(p.Speed)
	}

	newVelocity := p.Body.Velocity().Lerp(force, 0.1)
	p.Body.SetVelocity(newVelocity.X, newVelocity.Y)

	var activeOxygenTank *ItemOxygenTank

	for _, item := range p.Inventory {
		item.Update(w)

		if oxygenTank, ok := item.(*ItemOxygenTank); ok && oxygenTank.Oxygen > 0 {

			activeOxygenTank = oxygenTank
		}
	}

	if activeOxygenTank != nil && activeOxygenTank.Oxygen > 0 {
		activeOxygenTank.Oxygen = max(activeOxygenTank.Oxygen-w.DT, 0)
	}

	if activeOxygenTank == nil || activeOxygenTank.Oxygen == 0 {
		p.Health -= w.DT
	}

	// fmt.Printf("%+v\n", currentOxygenTank)

	// if p.Oxygen > 0 {
	// 	p.Oxygen -= w.DT * 4
	// }
	// if p.Oxygen < 0 {
	// 	p.Oxygen = 0
	// }

	// if p.Oxygen == 0 && p.Health > 0 {
	// 	p.Health -= w.DT
	// }
	// if p.Health < 0 {
	// 	p.Health = 0
	// }

}

func (p *Player) RenderHud(w *World) {

	cursor := rl.Vector2{X: 20, Y: 20}

	rl.DrawText(fmt.Sprintf("HP: %.0f/%.0f", w.Player.Health, w.Player.HealthMax), int32(cursor.X), int32(cursor.Y), 20, rl.Black)
	cursor.Y += 20

	for _, item := range p.Inventory {
		cursor.Y += 20
		if item.RenderHud(&cursor) {
			p.DropItem(w, item)
		}
	}

	// rl.DrawText(fmt.Sprintf("O2: %.0f/%.0f", w.Player.Oxygen, w.Player.OxygenMax), 30, 50, 20, rl.Black)
	// rl.DrawText(fmt.Sprintf("PW: %.0f/%.0f", w.Player.Power, w.Player.PowerMax), 30, 80, 20, rl.Black)
}
