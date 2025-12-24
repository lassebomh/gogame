package engine_test

import (
	"game/engine"
	"testing"
	"time"
)

type State struct {
	Value int
}

func (s *State) Clone() *State {
	clone := *s
	return &clone
}

func (s *State) Update(ctx *engine.UpdateContext[*State]) {
	for _, input := range engine.IterMapSorted(ctx.Inputs) {
		if input.Keyboard.A {
			s.Value -= 1
		}

		if input.Keyboard.D {
			s.Value += 1
		}
	}
}

func (s *State) Render(ctx *engine.RenderContext[*State]) {}

var _ engine.IState[*State] = (*State)(nil)

func TestRollback(t *testing.T) {

	peer_a := engine.ID(1)
	peer_b := engine.ID(2)

	t0 := time.Unix(0, 0)
	t1 := t0.Add(1 * time.Second)
	t2 := t0.Add(2 * time.Second)

	history := engine.History[*State]{
		OriginTime: t0,
		TickRate:   1,
		Items:      []engine.TickState[*State]{{Tick: 0, State: &State{}}},
		Inputs:     make(map[engine.ID][]engine.Input),
	}

	history.AddInput(peer_a, engine.Input{Time: t0, Keyboard: engine.InputKeyboard{A: true}})
	if s, _ := history.GetState(4); s.Value != -4 {
		t.Error("wrong value", s.Value)
	}
	if len(history.Items) != 5 {
		t.Error("wrong history items len", len(history.Items))
	}

	history.AddInput(peer_a, engine.Input{Time: t1})
	if len(history.Items) != 2 {
		t.Error("wrong history items len", len(history.Items))
	}
	if s, _ := history.GetState(1); s.Value != -1 {
		t.Error("wrong value", s.Value)
	}
	if s, _ := history.GetState(2); s.Value != -1 {
		t.Error("wrong value", s.Value)
	}

	history.AddInput(peer_b, engine.Input{Time: t2, Keyboard: engine.InputKeyboard{D: true}})
	if s, _ := history.GetState(3); s.Value != 0 {
		t.Error("wrong value", s.Value)
	}
	if s, _ := history.GetState(4); s.Value != 1 {
		t.Error("wrong value", s.Value)
	}
}
