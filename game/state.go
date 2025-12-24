package game

import (
	. "game/engine"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {
	Ticks   int
	Players map[ID]*Player
	World   World
}

func (s *State) Clone() *State {
	clone := *s
	clone.Players = make(map[ID]*Player, len(clone.Players))

	for k, v := range s.Players {
		playerCopy := v.Clone()
		clone.Players[k] = &playerCopy
	}

	clone.World = *s.World.Clone()

	return &clone
}

func (s *State) Update(ctx *UpdateContext[*State]) {
	s.Ticks++

	if len(s.Players) == 0 {
		s.Players[LocalPeerID] = &Player{
			Peer:     LocalPeerID,
			Position: rl.Vector2{X: 5, Y: 5},
			Radius:   10,
		}
	}

	dt := 1 / float32(ctx.TickRate)
	s.World.Integrate(dt)
	contacts := s.World.GenerateContactConstraints()
	s.World.SolveConstraints(contacts, dt, 8)

	for _, player := range IterMapSorted(s.Players) {
		player.Update(ctx)
	}
}

func (s *State) Render(ctx *RenderContext[*State]) {
	rl.ClearBackground(rl.Black)

	for _, player := range IterMapSorted(s.Players) {
		player.Render(ctx)
	}

	if ctx.Debug {

		for _, body := range s.World.Bodies {
			body.Render(rl.Green)
		}

	}
}
