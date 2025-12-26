package game

import (
	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type PhysicalItem struct {
	Body *cp.Body
	Item Item
}

func (pi *PhysicalItem) Update(w *World) {
	newVelocity := pi.Body.Velocity().Mult(0.8)
	pi.Body.SetVelocity(newVelocity.X, newVelocity.Y)
	pi.Body.SetAngularVelocity(pi.Body.AngularVelocity() * 0.8)

}

type Item interface {
	PhysicalUpdate(w *World, physical *PhysicalItem)
	InventoryUpdate(w *World, player *Player)
	RenderHud(cursor *rl.Vector2) bool
}

type ItemOxygenTank struct {
	Oxygen    float32
	OxygenMax float32
	Active    bool
}

func (item *ItemOxygenTank) RenderHud(cursor *rl.Vector2) (drop bool) {
	var width float32 = 200
	var height float32 = 50

	drop = gui.WindowBox(rl.NewRectangle(cursor.X, cursor.Y, width, height), "Oxygen Tank")
	item.Active = gui.CheckBox(rl.NewRectangle(cursor.X, cursor.Y+25, 25, 25), "", item.Active)
	gui.ProgressBar(rl.NewRectangle(cursor.X+25, cursor.Y+25, width-25, 25), "", "", item.Oxygen, 0, item.OxygenMax)

	cursor.Y += height
	return
}

func (item *ItemOxygenTank) PhysicalUpdate(w *World, physical *PhysicalItem) {
}

func (item *ItemOxygenTank) InventoryUpdate(w *World, p *Player) {

	if item.Active {
		if item.Oxygen > 0 {

			transfer := (float32(rl.GetRandomValue(80, 100)) - p.SpO2) / 100
			transfer = min(item.Oxygen, transfer)

			if transfer > 0 {

				item.Oxygen -= transfer
				p.SpO2 += transfer
			}

		} else {
			item.Active = false
		}
	}
}

type ItemBattery struct {
	Power    float32
	PowerMax float32
}

func (item *ItemBattery) RenderHud(cursor *rl.Vector2) (drop bool) {
	var width float32 = 200
	var height float32 = 50

	drop = gui.WindowBox(rl.NewRectangle(cursor.X, cursor.Y, width, height), "Battery")
	gui.ProgressBar(rl.NewRectangle(cursor.X, cursor.Y+25, width, 25), "", "", item.Power, 0, item.PowerMax)

	cursor.Y += height
	return
}

func (item *ItemBattery) PhysicalUpdate(w *World, physical *PhysicalItem) {
}

func (item *ItemBattery) InventoryUpdate(w *World, p *Player) {
}

type ItemAirPurifier struct {
	Active bool
}

func (item *ItemAirPurifier) RenderHud(cursor *rl.Vector2) (drop bool) {
	var width float32 = 200
	var height float32 = 50

	drop = gui.WindowBox(rl.NewRectangle(cursor.X, cursor.Y, width, height), "Air Purifier")
	item.Active = gui.CheckBox(rl.NewRectangle(cursor.X, cursor.Y+25, 25, 25), "Active", item.Active)

	cursor.Y += height
	return
}

func (item *ItemAirPurifier) PhysicalUpdate(w *World, physical *PhysicalItem) {
}

func (item *ItemAirPurifier) InventoryUpdate(w *World, p *Player) {
	if item.Active {
		var chosenBattery *ItemBattery

		for _, item := range p.Items {
			if battery, ok := item.(*ItemBattery); ok && battery.Power > 0 {
				chosenBattery = battery
				break
			}
		}

		var chosenOxygenTank *ItemOxygenTank

		for _, item := range p.Items {
			if oxygenTank, ok := item.(*ItemOxygenTank); ok && oxygenTank.Oxygen < 100 {
				chosenOxygenTank = oxygenTank
				break
			}
		}

		if chosenBattery != nil && chosenOxygenTank != nil {
			transfer := (100 - chosenOxygenTank.Oxygen)
			transfer = min(transfer, chosenBattery.Power, 2*w.DT)

			chosenBattery.Power -= transfer
			chosenOxygenTank.Oxygen += transfer
		}
	}
}
