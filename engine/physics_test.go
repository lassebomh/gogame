package engine_test

import (
	"testing"

	. "game/engine"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestCircleVSCircle(t *testing.T) {

	a := &Body[Circle]{
		Position: rl.Vector2{X: 0, Y: 0},
		Angle:    0,
		Shape: Circle{
			Radius: 1,
		},
	}

	b := &Body[Circle]{
		Position: rl.Vector2{X: 2, Y: 0},
		Angle:    0,
		Shape: Circle{
			Radius: 1,
		},
	}

	if ok, collision := CircleVsCircle(a, b); !ok || collision.Penetration != 0 {
		t.Errorf("no contact")
	}
}

func TestCircleVsBox(t *testing.T) {
	circle := &Body[Circle]{
		Position: rl.Vector2{X: 0, Y: 0},
		Angle:    0,
		Shape: Circle{
			Radius: 1,
		},
	}

	box := &Body[Box]{
		Position: rl.Vector2{X: 2, Y: 0},
		Angle:    0,
		Shape: Box{
			Width:  2,
			Height: 2,
		},
	}

	if ok, collision := CircleVsBox(circle, box); !ok || collision.Penetration != 0 {
		t.Errorf("no contact")
	}
}

func TestCircleVsRotatedBox(t *testing.T) {
	circle := &Body[Circle]{
		Position: rl.Vector2{X: 0, Y: 0},
		Angle:    0,
		Shape: Circle{
			Radius: 1,
		},
	}

	box := &Body[Box]{
		Position: rl.Vector2{X: 2.01, Y: 0},
		Angle:    45 * rl.Deg2rad,
		Shape: Box{
			Width:  2,
			Height: 2,
		},
	}

	// for i := float32(0); i < 360*rl.Deg2rad; i += 10 * rl.Deg2rad {
	// 	box.Angle = i
	// 	_, c := CircleVsBox(circle, box)
	// 	t.Logf("%+v\n", box)
	// 	t.Logf("%+v\n", c)
	// }

	if ok, _ := CircleVsBox(circle, box); !ok {
		t.Errorf("no contact")
	}
}
