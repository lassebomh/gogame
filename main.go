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
		TickRate:   10,
		Items: []TickState[*State]{{Tick: 0, State: &State{
			World: World{
				Bodies: []*Body{
					CreateBody(400, 400, 0.3, 1, Box{Width: 50, Height: 50}),
					CreateBody(400, 600, 0, 0, Box{Width: 600, Height: 50}),
				},
				Gravity: rl.Vector2{Y: 90},
			},
		}}},
		Inputs: make(map[ID][]Input),
	}
	rl.SetConfigFlags(rl.FlagWindowHighdpi)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(800, 700, "")
	defer rl.CloseWindow()

	rl.SetTargetFPS(90)

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
