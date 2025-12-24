package main

import (
	"time"

	. "game/engine"
	. "game/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {

	history := History[*State]{
		OriginTime: time.Now(),
		TickRate:   60,
		Items: []TickState[*State]{{Tick: 0, State: &State{
			Bodies: []*Body[Shape]{
				{Position: rl.Vector2{X: 90, Y: 30}, Shape: Circle{Radius: 20}},
				{Position: rl.Vector2{X: 30, Y: 30}, Shape: Box{Width: 20, Height: 20}},
			},
		}}},
		Inputs: make(map[ID][]Input),
	}
	rl.SetConfigFlags(rl.FlagWindowHighdpi)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(800, 700, "")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	prev_tick := 0

	for !rl.WindowShouldClose() {
		tick, alpha := history.TimeToTick(time.Now())
		if tick != prev_tick {
			input := GetLocalInputs()
			history.AddInput(LocalPeerID, input)
			prev_tick = tick
		}

		current, _ := history.GetState(tick)
		previous, ok := history.GetState(tick - 1)

		if !ok {
			previous = current
			alpha = 1
		}

		ctx := &RenderContext[*State]{
			Current:  current,
			Previous: previous,
			Peer:     LocalPeerID,
			Alpha:    alpha,
			Debug:    true,
		}

		rl.BeginDrawing()
		current.Render(ctx)
		rl.EndDrawing()
	}
}
