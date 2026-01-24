package game2

import (
	"github.com/beefsack/go-astar"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type PathFinder struct {
	Idle          bool
	Position      Vec3
	TargetCurrent Vec3
	Target        Vec3

	Path  []Vec3
	level *Level
}

func NewPathFinder(level *Level) *PathFinder {
	return &PathFinder{
		Idle:  true,
		level: level,
	}
}

func (p *PathFinder) SetIdle(idle bool) {
	p.Idle = idle
}

func (p *PathFinder) SetPosition(position Vec3) {
	p.Position = position.Floor()
}

func (p *PathFinder) SetTarget(position Vec3) {
	p.Target = position.Floor()

	start := p.level.GetCell(p.Position)
	end := p.level.GetCell(p.Target)

	pathers, _, found := astar.Path(end, start)
	p.Idle = !found

	if !found {
		return
	}

	cells := make([]*Cell, len(pathers))

	for i, p := range pathers {
		cell := p.(*Cell)
		cells[i] = cell
	}

	path := make([]Vec3, len(pathers))

	for i, cell := range cells {
		path[i] = cell.Position.Add(NewVec3(0.5, 0, 0.5))
	}

	p.Path = path
}

func (p *PathFinder) Draw3D(g *Game) {

	for i := 0; i < len(p.Path)-2; i++ {
		from := p.Path[i]
		to := p.Path[i+1]

		rl.DrawLine3D(from.Add(Y.Scale(0.5)).Raylib(), to.Add(Y.Scale(0.5)).Raylib(), rl.Green)
	}
}
