package game

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type Player struct {
	Body  *cp.Body
	Speed float64

	SpO2      float32
	Health    float32
	HealthMax float32

	Radius float32
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
		Speed:     50,
		Health:    100,
		HealthMax: 100,
		SpO2:      100,

		Items:  make([]Item, 0),
		Radius: 2,
	}

	mass := float64(player.Radius * player.Radius / 25.0)
	body := w.Space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, float64(player.Radius), cp.Vector{})))
	body.SetPosition(pos)

	shape := w.Space.AddShape(cp.NewCircle(body, float64(player.Radius), cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)
	player.Body = body

	player.Items = append(player.Items,
		&ItemOxygenTank{
			Oxygen:    100,
			OxygenMax: 100,
		},
		&ItemBattery{
			Power:    100,
			PowerMax: 100,
		},
		&ItemAirPurifier{
			Active: false,
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

	if p.SpO2 > 0 {
		p.SpO2 -= 0.5 * w.DT
	}

	for _, item := range p.Items {
		item.InventoryUpdate(w, p)
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

	rl.DrawText(fmt.Sprintf("SpO2: %.1f%%", w.Player.SpO2), int32(cursor.X), int32(cursor.Y), 20, rl.Black)
	cursor.Y += 20

	for _, item := range p.Items {
		cursor.Y += 20
		if item.RenderHud(&cursor) {
			p.RemoveItem(w, item)
			w.NewPhysicalItem(item, p.Body.Position())
		}
	}
}
