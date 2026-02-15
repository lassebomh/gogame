package game3

import (
	v3 "game/vec3"
	"image/color"
	"math"

	"github.com/beefsack/go-astar"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type PathPoint struct {
	Position v3.Value
	Center   v3.Value
	Cell     *Cell
	world    *World
	points   map[v3.Value]*PathPoint
}

func (p *PathPoint) GetPoint(pos v3.Value) *PathPoint {
	if pos.Y < 0 {
		return nil
	}

	point, ok := p.points[pos]
	if ok {
		return point
	}

	cell, ok := p.world.GetCell(pos)
	if !ok {
		return nil
	}

	point = &PathPoint{
		world:    p.world,
		Position: pos,
		Center:   pos.AddXYZ(0.5, 0.5, 0.5),
		Cell:     cell,
		points:   p.points,
	}
	p.points[pos] = point

	return point
}

func (p *PathPoint) PathNeighbors() []astar.Pather {
	out := make([]astar.Pather, 0)

	if p.Cell.Faces[FaceDown].Type == FaceNone {
		below := p.GetPoint(p.Position.Subtract(v3.Y(1)))
		if below != nil && below.Cell.Faces[FaceDown].Type != FaceNone {
			out = append(out, below)
		}

		return out
	}

	if p.Cell.Faces[FaceDown].Type == FaceStair {
		forward := FaceForward[p.Cell.Faces[FaceDown].Rotation]
		prev := p.GetPoint(p.Position.Subtract(forward))
		if prev != nil && prev.Cell.Faces[FaceDown].Type != FaceNone {
			out = append(out, prev)
		} else {
			prevBelow := p.GetPoint(p.Position.Subtract(forward).SubtractXYZ(0, 1, 0))
			if prevBelow != nil && prevBelow.Cell.Faces[FaceDown].Type == FaceStair {
				out = append(out, prevBelow)
			}

		}

		nextUp := p.GetPoint(p.Position.Add(forward.Add(v3.Y(1))))
		if nextUp != nil && nextUp.Cell.Faces[FaceDown].Type != FaceNone {
			out = append(out, nextUp)
		}
	}

	if p.Cell.Faces[FaceDown].Type == FaceSolid {
		for i := range FaceDown {
			face := &p.Cell.Faces[i]
			if face.Type == FaceSolid {
				continue
			}

			nextPos := p.Position.Add(FaceForward[i])

			next := p.GetPoint(nextPos)
			if next == nil {
				continue
			}

			if next.Cell.Faces[FaceOpposite[i]].Type == FaceSolid {
				continue
			}

			if next.Cell.Faces[FaceDown].Type != FaceNone {
				out = append(out, next)
			} else {
				nextBelow := p.GetPoint(nextPos.SubtractXYZ(0, 1, 0))
				if nextBelow != nil {
					out = append(out, nextBelow)
				}
			}
		}
	}

	return out

}

func (p *PathPoint) PathNeighborCost(to astar.Pather) float64 {
	other := to.(*PathPoint)
	return p.Center.Distance(other.Center)
}

func (p *PathPoint) PathEstimatedCost(to astar.Pather) float64 {
	other := to.(*PathPoint)
	return p.Center.Distance(other.Center)
}

func NewPathPoint(world *World, pos v3.Value) *PathPoint {
	pos = pos.Map(math.Floor)

	var cell *Cell
	var ok bool

	for i := pos.Y; i >= 0; i-- {
		pos.Y = i
		cell, ok = world.GetCell(pos)
		if ok && cell.Faces[FaceDown].Type != FaceNone {
			start := &PathPoint{
				Position: pos,
				Center:   pos.AddXYZ(0.5, 0.5, 0.5),
				world:    world,
				Cell:     cell,
				points:   map[v3.Value]*PathPoint{},
			}
			start.points[pos] = start
			return start
		}
	}

	return nil
}

func (p *PathPoint) GetNeighborPathPoints() []*PathPoint {
	pathers := p.PathNeighbors()
	points := make([]*PathPoint, 0)
	for _, pather := range pathers {
		points = append(points, pather.(*PathPoint))
	}
	return points
}

func (p *PathPoint) FindPath(endPos v3.Value) ([]*PathPoint, float64) {
	if p == nil {
		return []*PathPoint{}, 0
	}
	endPos = endPos.Map(math.Floor)

	end := p.GetPoint(endPos)
	if end == nil {
		return []*PathPoint{}, 0
	}

	pathers, length, found := astar.Path(end, p)

	if !found {
		return []*PathPoint{}, 0
	}

	path := make([]*PathPoint, len(pathers))

	for i, p := range pathers {
		path[i] = p.(*PathPoint)
	}

	return path, length
}

func (p *PathPoint) Draw() {
	rl.DrawSphere(p.Center.Raylib(), 0.05, color.RGBA{255, 255, 0, 80})
}

func DrawPath(pathPoints []*PathPoint) {
	for i, p := range pathPoints {
		if i != 0 {
			prev := pathPoints[i-1]
			rl.DrawLine3D(p.Center.Raylib(), prev.Center.Raylib(), color.RGBA{255, 255, 0, 128})
		}
		p.Draw()
	}
}
