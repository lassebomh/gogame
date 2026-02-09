package game3

import (
	"fmt"
	"game/vec3"
	"math"

	"github.com/beefsack/go-astar"
)

var lock = 0

type PathPoint struct {
	Position vec3.Value
	Cell     *Cell
	world    *World
	points   map[vec3.Value]*PathPoint
}

var FACE_DIRECTION = []vec3.Value{
	vec3.X(-1),
	vec3.Z(1),
	vec3.X(1),
	vec3.Z(-1),
	vec3.Y(1),
}

var FACE_OPPOSITE = []FaceDirection{
	FaceEast,
	FaceSouth,
	FaceWest,
	FaceNorth,
	-1,
}

func (p *PathPoint) GetPoint(pos vec3.Value) (*PathPoint, bool) {

	point, ok := p.points[pos]
	if ok {
		// fmt.Println("exists!", pos)
		return point, true
	}

	cell, ok := p.world.GetCell(pos)
	if ok {
		// fmt.Println("new", pos)
		point := &PathPoint{
			world:    p.world,
			Position: pos,
			Cell:     cell,
			points:   p.points,
		}
		p.points[pos] = point

		// time.Sleep(100 * time.Millisecond)
		return point, true
	}

	// fmt.Println("empty", pos)

	return nil, false
}

func (p *PathPoint) PathNeighbors() []astar.Pather {

	switch p.Cell.Faces[FaceDown].Type {
	case FaceNone:
		return []astar.Pather{}
	case FaceStair:
		return []astar.Pather{}
		// prevPos := p.Position.Add(FACE_DIRECTION[p.Cell.Faces[FaceDown].Rotation])
		// nextPos := p.Position.Add(FACE_DIRECTION[FACE_OPPOSITE[p.Cell.Faces[FaceDown].Rotation]]).Add(vec3.Y(1))

		// out := make([]astar.Pather, 0, 2)

		// prevCell, ok := p.GetPoint(prevPos)
		// if ok {
		// 	out = append(out, prevCell)
		// }

		// nextCell, ok := p.GetPoint(nextPos)
		// if ok {
		// 	out = append(out, nextCell)
		// }
		// return out
	}

	neighbors := make([]astar.Pather, 0)

	for FACE := range FaceDown {
		face := &p.Cell.Faces[FACE]

		if face.Type == FaceSolid {
			continue
		}

		next, ok := p.GetPoint(p.Position.Add(FACE_DIRECTION[FACE]))
		if !ok {
			continue
		}

		if next.Cell.Faces[FACE_OPPOSITE[FACE]].Type == FaceSolid {
			continue
		}

		// if next.Cell.Faces[FaceDown].Type == FaceStair {
		// 	continue
		// }
		// if next.Cell.Faces[FaceDown].Type == FaceStair && next.Cell.Faces[FaceDown].Rotation != FACE {
		// 	continue
		// }
		neighbors = append(neighbors, next)

		// if p.Position.Y > 0 {
		// 	nextBelow, ok := p.GetPoint(p.Position.Add(FACE_DIRECTION[FACE]).Subtract(vec3.Y(1)))

		// 	if ok && nextBelow.Cell.Faces[FaceDown].Type == FaceStair { //  && nextBelow.Cell.Faces[FaceDown].Rotation == FACE_OPPOSITE[FACE] {
		// 		neighbors = append(neighbors, nextBelow)
		// 	}
		// }

	}

	return neighbors

}

func (p *PathPoint) PathNeighborCost(to astar.Pather) float64 {
	other := to.(*PathPoint)
	return p.Position.Distance(other.Position)
}

func (p *PathPoint) PathEstimatedCost(to astar.Pather) float64 {
	other := to.(*PathPoint)
	return p.Position.Distance(other.Position)
}

func FindPath(world *World, startPos vec3.Value, endPos vec3.Value) ([]vec3.Value, float64) {
	startPos = startPos.Map(math.Floor)
	endPos = endPos.Map(math.Floor)

	// fmt.Println("FROM", startPos, "TO", endPos)

	points := map[vec3.Value]*PathPoint{}

	start := &PathPoint{
		Position: startPos,
		world:    world,
		Cell:     world.UpsertCell(startPos),
		points:   points,
	}
	points[startPos] = start

	end := &PathPoint{
		Position: endPos,
		world:    world,
		Cell:     world.UpsertCell(endPos),
		points:   points,
	}
	points[endPos] = end

	pathers, length, found := astar.Path(end, start)

	if !found {
		p, ok := points[endPos]
		fmt.Println(p.Position, ok)
		return []vec3.Value{}, 0
	}

	path := make([]vec3.Value, len(pathers))

	for i, p := range pathers {
		path[i] = p.(*PathPoint).Position.AddXYZ(0.5, 0, 0.5)
	}

	return path, length
}
