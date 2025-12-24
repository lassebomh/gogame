package game

import (
	. "game/engine"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {
	Ticks   int
	Players map[ID]*Player
	Bodies  []*Body[Shape]
}

func (s *State) Clone() *State {
	clone := *s
	clone.Players = make(map[ID]*Player, len(clone.Players))

	for k, v := range s.Players {
		playerCopy := v.Clone()
		clone.Players[k] = &playerCopy
	}

	clone.Bodies = make([]*Body[Shape], 0)
	for _, v := range s.Bodies {
		bodyClone := *v
		clone.Bodies = append(clone.Bodies, &bodyClone)
	}

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

	for _, body := range s.Bodies {
		body.Angle += 0.05
	}

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
		for _, body := range ctx.Current.Bodies {
			body.Render()
		}
	}
}
