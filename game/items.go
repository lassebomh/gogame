package game

import (
	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type ItemOxygenTank struct {
	Body      *cp.Body
	Oxygen    float32
	OxygenMax float32
}

func (item *ItemOxygenTank) RenderHud(cursor *rl.Vector2) (drop bool) {
	var width float32 = 200
	var height float32 = 50

	drop = gui.WindowBox(rl.NewRectangle(cursor.X, cursor.Y, width, height), "Oxygen Tank")
	gui.ProgressBar(rl.NewRectangle(cursor.X, cursor.Y+25, width, 25), "", "", item.Oxygen, 0, item.OxygenMax)

	cursor.Y += height
	return
}

func (item *ItemOxygenTank) Update(w *World) {
	if item.Body != nil {
		newVelocity := item.Body.Velocity().Mult(0.8)
		item.Body.SetVelocity(newVelocity.X, newVelocity.Y)
		item.Body.SetAngularVelocity(item.Body.AngularVelocity() * 0.8)
	}
}

func (item *ItemOxygenTank) GetBody() *cp.Body {
	return item.Body
}
func (item *ItemOxygenTank) SetBody(body *cp.Body) {
	item.Body = body
}
