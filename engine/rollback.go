package engine

import (
	"math"
	"time"
)

type Clonable[T any] interface {
	Clone() T
}

type IState[T any] interface {
	Update(ctx *UpdateContext[T])
	Render(ctx *RenderContext[T])
	Clonable[T]
}

type TickState[T IState[T]] struct {
	Tick  int
	State T
}

type UpdateContext[T any] struct {
	State    T
	Inputs   map[ID]Input
	Tick     int
	TickRate int
}

type RenderContext[T any] struct {
	Previous T
	Current  T
	Peer     ID
	Alpha    float32
	Debug    bool
}

type History[T IState[T]] struct {
	OriginTime time.Time
	TickRate   int
	Inputs     map[ID][]Input
	Items      []TickState[T]
}

func (h *History[T]) AddInput(peer ID, input Input) {
	h.Inputs[peer] = append(h.Inputs[peer], input)

	tick, _ := h.TimeToTick(input.Time)

	for i := len(h.Items) - 1; i >= 0; i-- {
		tick_state := h.Items[i]

		if tick_state.Tick == tick {
			h.Items = h.Items[:i+1]
			break
		}
	}
}

func (h *History[T]) GetState(tick int) (state T, ok bool) {
	index := len(h.Items) - 1
	current := h.Items[index]

	for {
		if current.Tick == tick {
			state = current.State
			ok = true
			return
		}

		if current.Tick < tick {
			next_state := current.State.Clone()

			inputs := make(map[ID]Input)

			for peer, peer_inputs := range h.Inputs {
				for i := len(peer_inputs) - 1; i >= 0; i-- {
					input := peer_inputs[i]
					input_tick, _ := h.TimeToTick(input.Time)
					if input_tick < tick {
						inputs[peer] = input
						break
					}
				}
			}

			ctx := &UpdateContext[T]{
				State:    next_state,
				Inputs:   inputs,
				Tick:     current.Tick + 1,
				TickRate: h.TickRate,
			}
			next_state.Update(ctx)
			current = TickState[T]{Tick: ctx.Tick, State: ctx.State}

			h.Items = append(h.Items[max(len(h.Items)-60, 0):], current)
		}

		if current.Tick > tick {
			index--
			if index < 0 {
				return
			}
			current = h.Items[index]
		}
	}
}

func (h *History[T]) TimeToTick(t time.Time) (int, float32) {
	tick, alpha := math.Modf(t.Sub(h.OriginTime).Seconds() * float64(h.TickRate))
	return int(tick), float32(alpha)
}
