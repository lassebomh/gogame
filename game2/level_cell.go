package game2

import (
	"math"

	"github.com/beefsack/go-astar"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const WALL_WIDTH = float64(0.15)

type FaceIndex = uint8

const (
	FACE_WEST = FaceIndex(iota)
	FACE_NORTH
	FACE_EAST
	FACE_SOUTH
	FACES
)

var WALL_VERTS = [4][]cp.Vector{
	[]cp.Vector{{1, 1}, {1, 0}, {1 - WALL_WIDTH, 0}, {1 - WALL_WIDTH, 1}},
	[]cp.Vector{{1, 1}, {0, 1}, {0, 1 - WALL_WIDTH}, {1, 1 - WALL_WIDTH}},
	[]cp.Vector{{0, 1}, {0, 0}, {WALL_WIDTH, 0}, {WALL_WIDTH, 1}},
	[]cp.Vector{{1, 0}, {0, 0}, {0, WALL_WIDTH}, {1, WALL_WIDTH}},
}

var DOOR_VERTS = []cp.Vector{
	{0.5, -WALL_WIDTH / 2},
	{-0.5, -WALL_WIDTH / 2},
	{-0.5, WALL_WIDTH / 2},
	{0.5, WALL_WIDTH / 2},
}

var FACE_DIRECTION = [4]Vec3{
	NewVec3(1, 0, 0),
	NewVec3(0, 0, 1),
	NewVec3(-1, 0, 0),
	NewVec3(0, 0, -1),
}

var FACE_OPPOSITE = [4]FaceIndex{
	FACE_EAST,
	FACE_SOUTH,
	FACE_WEST,
	FACE_NORTH,
}
var FACE_NEXT = [4]FaceIndex{
	FACE_NORTH,
	FACE_EAST,
	FACE_SOUTH,
	FACE_WEST,
}

var FACE_DEGREE = [4]float64{
	0,
	90,
	180,
	270,
}

type FaceType = uint8

const (
	FaceEmpty = FaceType(iota)
	FaceDoor
	FaceWall
)

type GroundType = uint8

const (
	GroundEmpty = GroundType(iota)
	GroundFloor
	GroundStair
)

type Ground struct {
	StairDirection FaceIndex
	TileX          int
	TileY          int
	Type           GroundType
}

type Face struct {
	Type  FaceType
	TileX int
	TileY int

	body  *cp.Body
	shape *cp.Shape
}

type Cell struct {
	Faces  [4]Face
	Ground Ground

	level    *Level
	Position Vec3
}

func (c *Cell) Wake(g *Game) {
	transform := cp.NewTransformTranslate(cp.Vector{c.Position.X, c.Position.Z})

	for FACE := range FACES {
		face := &c.Faces[FACE]

		if face.Type != FaceEmpty && face.body == nil {

			switch face.Type {
			case FaceWall:
				face.body = g.Space.StaticBody
				shape := cp.NewPolyShape(face.body, 4, WALL_VERTS[FACE], transform, 0)
				shape.Filter.Categories = 1 << uint(c.Position.Y)
				shape.Filter.Group = 1
				face.shape = g.Space.AddShape(shape)
			case FaceDoor:
				position := c.Position.AddXYZ(0.5, 0, 0.5).Subtract(FACE_DIRECTION[FACE_OPPOSITE[FACE]].Scale((1 - WALL_WIDTH) / 2))
				mass := float64(0.2)
				moment := cp.MomentForBox(mass, 1, WALL_WIDTH)
				angle := (FACE_DEGREE[FACE_NEXT[FACE]]) * rl.Deg2rad

				face.body = g.Space.AddBody(cp.NewBody(mass, moment))
				face.body.SetPosition(position.Chipmunk())
				face.body.SetAngle(angle)

				face.shape = cp.NewPolyShape(face.body, 4, DOOR_VERTS, cp.NewTransformIdentity(), 0)
				face.shape.Filter.Categories = 1 << uint(c.Position.Y)
				face.shape.Filter.Group = 1
				g.Space.AddShape(face.shape)

				pivotPos := position.Add(FACE_DIRECTION[FACE_NEXT[FACE]].Scale(0.5))
				pivot := cp.NewPivotJoint(face.body, g.Space.StaticBody, pivotPos.Chipmunk())

				stiffness := 5.0 * face.body.Moment()
				damping := 3 * math.Sqrt(stiffness*face.body.Moment())
				dampedSpring := cp.NewDampedRotarySpring(g.Space.StaticBody, face.body, -angle, stiffness, damping)

				minAngle := angle - rl.Pi/1.2
				maxAngle := angle + rl.Pi/1.2
				rotaryLimit := cp.NewRotaryLimitJoint(g.Space.StaticBody, face.body, minAngle, maxAngle)
				rotaryLimit.SetMaxForce(1e8)

				g.Space.AddConstraint(rotaryLimit)
				g.Space.AddConstraint(pivot)
				g.Space.AddConstraint(dampedSpring)
			}
		}

	}
}

func (face *Face) Draw(g *Game, cellPos Vec3, rotationAxis Vec3, rotationDegrees float32) {
	switch face.Type {
	case FaceWall:
		aa, bb := g.Tileset.GetAABB(face.TileX, face.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		center := cellPos
		rl.DrawModelEx(g.GetModel("wall"), center.Raylib(), rotationAxis.Raylib(), rotationDegrees, XYZ.Raylib(), rl.White)
	case FaceDoor:
		if face.body != nil {
			aa, bb := g.Tileset.GetAABB(face.TileX, face.TileY)
			g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
			pos := face.body.Position()
			origin := NewVec3(pos.X, cellPos.Y-0.5, pos.Y)
			angle := face.body.Angle() * rl.Rad2deg
			rl.DrawModelEx(g.GetModel("door"), origin.Raylib(), Y.Negate().Raylib(), float32(angle), XYZ.Raylib(), rl.White)
		}

	}
}

func (gr *Ground) Draw(g *Game, cellPos Vec3) {
	if gr.Type == GroundFloor {
		aa, bb := g.Tileset.GetAABB(gr.TileX, gr.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		center := cellPos.Add(NewVec3(0.5, 0.5-WALL_WIDTH+0.01, 0.5))
		rl.DrawModelEx(g.GetModel("wall"), center.Raylib(), Z.Raylib(), float32(-90), XYZ.Raylib(), rl.White)
	} else if gr.Type == GroundStair {
		aa, bb := g.Tileset.GetAABB(gr.TileX, gr.TileY)
		g.MainShader.UVClamp.Set(aa.X, aa.Y, bb.X, bb.Y)
		center := cellPos.AddXYZ(0.5, 0, 0.5)
		rl.DrawModelEx(g.GetModel("stair"), center.Raylib(), Y.Negate().Raylib(), float32(FACE_DEGREE[gr.StairDirection]-90), XYZ.Scale(0.99).Raylib(), rl.White)
	}
}

func (c *Cell) PathNeighbors() []astar.Pather {

	switch c.Ground.Type {
	case GroundEmpty:
		return []astar.Pather{}
	case GroundStair:
		prevPos := c.Position.Add(FACE_DIRECTION[FACE_OPPOSITE[c.Ground.StairDirection]])
		nextPos := c.Position.Add(FACE_DIRECTION[c.Ground.StairDirection]).Add(Y)
		return []astar.Pather{
			c.level.GetCell(prevPos),
			c.level.GetCell(nextPos),
		}
	}

	neighbors := make([]astar.Pather, 0)

	for FACE := range FACES {
		face := &c.Faces[FACE]

		if face.Type == FaceWall {
			continue
		}

		next := c.level.GetCell(c.Position.Add(FACE_DIRECTION[FACE]))

		if next.Faces[FACE_OPPOSITE[FACE]].Type == FaceWall {
			continue
		}

		if next.Ground.Type == GroundStair && next.Ground.StairDirection != FACE {
			continue
		}
		neighbors = append(neighbors, next)

		if c.Position.Y > 0 {
			nextBelow := c.level.GetCell(c.Position.Add(FACE_DIRECTION[FACE]).Subtract(Y))

			if nextBelow.Ground.Type == GroundStair && nextBelow.Ground.StairDirection == FACE_OPPOSITE[FACE] {
				neighbors = append(neighbors, nextBelow)
			}
		}

	}

	return neighbors

}

func (c *Cell) PathNeighborCost(to astar.Pather) float64 {
	other := to.(*Cell)
	return c.Position.Distance(other.Position)
}

func (c *Cell) PathEstimatedCost(to astar.Pather) float64 {
	other := to.(*Cell)
	return c.Position.Distance(other.Position)
}
