package game

import (
	"image/color"

	. "game/engine"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	Peer     ID
	Position rl.Vector2
	Radius   float32
}

func NewPlayer(peer ID, pos rl.Vector2, radius float32) *Player {

	return &Player{
		Peer:   peer,
		Radius: radius,
	}
}

func (p Player) Clone() Player {
	return Player{
		Peer:     p.Peer,
		Position: p.Position,
		Radius:   p.Radius,
	}
}

func (p *Player) Update(ctx *UpdateContext[*State]) {
	input := ctx.Inputs[p.Peer]

	move := rl.Vector2{}

	if input.Keyboard.D {
		move.X += 1
	}
	if input.Keyboard.A {
		move.X -= 1
	}
	if input.Keyboard.W {
		move.Y -= 1
	}
	if input.Keyboard.S {
		move.Y += 1
	}

	mag := rl.Vector2Length(move)
	if mag != 0 {
		move = rl.Vector2Divide(move, rl.Vector2{X: mag, Y: mag})
	}

	move = rl.Vector2Scale(move, 20)

	p.Position = rl.Vector2Add(p.Position, move)
}

func (p *Player) Render(ctx *RenderContext[*State]) {
	prev, ok := ctx.Previous.Players[p.Peer]
	if !ok {
		prev = p
	}
	pos := Vector2Lerp(prev.Position, p.Position, ctx.Alpha)

	rl.DrawCircle(int32(pos.X), int32(pos.Y), p.Radius, color.RGBA{255, 0, 0, 255})
}
