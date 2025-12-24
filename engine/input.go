package engine

import (
	"math/rand/v2"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ID = uint64

var LocalPeerID ID

func init() {
	LocalPeerID = ID(rand.Uint64())
}

type InputMouse struct {
	Pos   rl.Vector2
	Left  bool
	Right bool
}

type InputKeyboard struct {
	A bool
	S bool
	D bool
	W bool
}

type Input struct {
	Time     time.Time
	Mouse    InputMouse
	Keyboard InputKeyboard
}

func GetLocalInputs() Input {
	return Input{
		Time: time.Now(),
		Mouse: InputMouse{
			Pos:   rl.GetMousePosition(),
			Left:  rl.IsMouseButtonDown(rl.MouseButtonLeft),
			Right: rl.IsMouseButtonDown(rl.MouseButtonRight),
		},
		Keyboard: InputKeyboard{
			A: rl.IsKeyDown(rl.KeyA),
			S: rl.IsKeyDown(rl.KeyS),
			D: rl.IsKeyDown(rl.KeyD),
			W: rl.IsKeyDown(rl.KeyW),
		},
	}
}
