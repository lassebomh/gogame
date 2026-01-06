package game

import (
	"image/color"
	"math"

	"github.com/beefsack/go-astar"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

var p astar.Pather = &Tile{}

const (
	WALL_T = 1 << iota
	WALL_R
	WALL_B
	WALL_L
)

type Tile struct {
	Wall int
	Door int
	X    int
	Y    int

	WorldPosition cp.Vector

	tilemap  *Tilemap
	wallBody *cp.Body
	DoorBody *cp.Body
}

type Tilemap struct {
	Cols           [][]Tile
	Scale          float64
	Width          int
	Height         int
	CenterPosition cp.Vector
	WallDepthRatio float32
}

func (t *Tile) PathNeighbors() []astar.Pather {
	neighbors := make([]astar.Pather, 0, 8)

	if t.Wall&WALL_L == 0 && t.X > 0 {
		left := &t.tilemap.Cols[t.X-1][t.Y]
		if left.Wall&WALL_R == 0 {
			neighbors = append(neighbors, left)
		}
	}

	if t.Wall&WALL_T == 0 && t.Y > 0 {
		top := &t.tilemap.Cols[t.X][t.Y-1]
		if top.Wall&WALL_B == 0 {
			neighbors = append(neighbors, top)
		}
	}

	if t.Wall&WALL_R == 0 && t.X < t.tilemap.Width-1 {
		right := &t.tilemap.Cols[t.X+1][t.Y]
		if right.Wall&WALL_L == 0 {
			neighbors = append(neighbors, right)
		}
	}

	if t.Wall&WALL_B == 0 && t.Y < t.tilemap.Height-1 {
		bottom := &t.tilemap.Cols[t.X][t.Y+1]
		if bottom.Wall&WALL_T == 0 {
			neighbors = append(neighbors, bottom)
		}
	}

	if t.X > 0 && t.Y > 0 {
		left := &t.tilemap.Cols[t.X-1][t.Y]
		top := &t.tilemap.Cols[t.X][t.Y-1]
		topleft := &t.tilemap.Cols[t.X-1][t.Y-1]

		if t.Wall&WALL_L == 0 && t.Wall&WALL_T == 0 &&
			left.Wall&WALL_R == 0 && top.Wall&WALL_B == 0 {
			neighbors = append(neighbors, topleft)
		}
	}

	if t.X < t.tilemap.Width-1 && t.Y > 0 {
		right := &t.tilemap.Cols[t.X+1][t.Y]
		top := &t.tilemap.Cols[t.X][t.Y-1]
		topright := &t.tilemap.Cols[t.X+1][t.Y-1]

		if t.Wall&WALL_R == 0 && t.Wall&WALL_T == 0 &&
			right.Wall&WALL_L == 0 && top.Wall&WALL_B == 0 {
			neighbors = append(neighbors, topright)
		}
	}

	if t.X > 0 && t.Y < t.tilemap.Height-1 {
		left := &t.tilemap.Cols[t.X-1][t.Y]
		bottom := &t.tilemap.Cols[t.X][t.Y+1]
		bottomleft := &t.tilemap.Cols[t.X-1][t.Y+1]

		if t.Wall&WALL_L == 0 && t.Wall&WALL_B == 0 &&
			left.Wall&WALL_R == 0 && bottom.Wall&WALL_T == 0 {
			neighbors = append(neighbors, bottomleft)
		}
	}

	if t.X < t.tilemap.Width-1 && t.Y < t.tilemap.Height-1 {
		right := &t.tilemap.Cols[t.X+1][t.Y]
		bottom := &t.tilemap.Cols[t.X][t.Y+1]
		bottomright := &t.tilemap.Cols[t.X+1][t.Y+1]

		if t.Wall&WALL_R == 0 && t.Wall&WALL_B == 0 &&
			right.Wall&WALL_L == 0 && bottom.Wall&WALL_T == 0 {
			neighbors = append(neighbors, bottomright)
		}
	}

	return neighbors
}

func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	other := to.(*Tile)
	dist := t.WorldPosition.Distance(other.WorldPosition)

	return dist
}

func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	other := to.(*Tile)
	return t.WorldPosition.Distance(other.WorldPosition)
}
func NewTilemap(width int, height int, scale float64) *Tilemap {
	tilemap := &Tilemap{
		Cols:           make([][]Tile, width),
		Scale:          scale,
		Width:          width,
		Height:         height,
		WallDepthRatio: 0.1,
	}

	for x := range width {
		col := make([]Tile, height)
		for y := range height {
			col[y] = Tile{X: x, Y: y, tilemap: tilemap, WorldPosition: cp.Vector{X: float64(x) * tilemap.Scale, Y: float64(y) * tilemap.Scale}}
		}

		tilemap.Cols[x] = col
	}

	tilemap.CenterPosition = cp.Vector{X: float64(tilemap.Width / 2), Y: float64(tilemap.Height / 2)}.Mult(tilemap.Scale)

	return tilemap
}

func (t *Tilemap) GetTileAtWorldPosition(pos cp.Vector) *Tile {
	pos = pos.Mult(1 / t.Scale)
	x := int(pos.X)
	y := int(pos.Y)

	if x < 0 || y < 0 || x >= t.Width || y >= t.Height {
		return nil
	}

	return &t.Cols[x][y]
}

func (t *Tilemap) FindPath(start, end cp.Vector) (float64, []cp.Vector) {

	startTile := t.GetTileAtWorldPosition(start)
	endTile := t.GetTileAtWorldPosition(end)

	if startTile == nil || endTile == nil {
		return 0, nil
	}

	path, distance, found := astar.Path(endTile, startTile)
	if !found {
		return 0, nil
	}

	worldPath := make([]cp.Vector, len(path))
	for i, p := range path {
		tile := p.(*Tile)
		worldPath[i] = cp.Vector{X: float64(tile.X) + 0.5, Y: float64(tile.Y) + 0.5}.Mult(t.Scale)
	}

	return distance, worldPath
}

func (t *Tilemap) Render() {
	for x, rows := range t.Cols {
		for y, tile := range rows {
			pos := cp.Vector{X: float64(x), Y: float64(y)}.Mult(t.Scale)

			if tile.Wall != 0 {
				rl.DrawRectangle(int32(pos.X), int32(pos.Y), int32(t.Scale), int32(t.Scale), color.RGBA{255, 0, 0, 30})
			}
		}
	}
}

func (t *Tilemap) CreateRoom(x, y, width, height int, doorDirection int) {
	for w := range width {
		t.Cols[x+w][y].Wall |= WALL_T
		t.Cols[x+w][y+height-1].Wall |= WALL_B
	}
	for h := range height {
		t.Cols[x][y+h].Wall |= WALL_L
		t.Cols[x+width-1][y+h].Wall |= WALL_R
	}

	if doorDirection&WALL_T != 0 {
		t.Cols[x+width/2][y].Wall = 0
	}
	if doorDirection&WALL_L != 0 {
		t.Cols[x][y+height/2].Wall = 0
	}
	if doorDirection&WALL_B != 0 {
		t.Cols[x+width/2][y+height-1].Wall = 0
		t.Cols[x+width/2][y+height-1].Door = WALL_B
	}
	if doorDirection&WALL_R != 0 {
		t.Cols[x+width-1][y+height/2].Wall = 0
		t.Cols[x+width-1][y+height/2].Door = WALL_R
	}
}
func (t *Tilemap) GenerateBodies(w *World) {
	var shapeFilterGroup uint = 2

	addSegment := func(body *cp.Body, start, end cp.Vector, radius float64) {
		shape := cp.NewSegment(body, start, end, radius)
		w.Space.AddShape(shape)
		shape.Filter.Group = shapeFilterGroup
	}

	for x, rows := range t.Cols {
		for y := range rows {
			tile := &t.Cols[x][y]

			tileX := float64(x) * t.Scale
			tileY := float64(y) * t.Scale

			th := t.Scale * float64(t.WallDepthRatio)
			s := t.Scale

			if tile.Wall != 0 {
				body := cp.NewStaticBody()

				left := tileX + th/2
				right := tileX + s - th/2
				top := tileY + th/2
				bottom := tileY + s - th/2

				if tile.Wall&WALL_T != 0 {
					addSegment(body, cp.Vector{X: left, Y: top}, cp.Vector{X: right, Y: top}, th/2)
				}
				if tile.Wall&WALL_R != 0 {
					addSegment(body, cp.Vector{X: right, Y: top}, cp.Vector{X: right, Y: bottom}, th/2)
				}
				if tile.Wall&WALL_B != 0 {
					addSegment(body, cp.Vector{X: left, Y: bottom}, cp.Vector{X: right, Y: bottom}, th/2)
				}
				if tile.Wall&WALL_L != 0 {
					addSegment(body, cp.Vector{X: left, Y: top}, cp.Vector{X: left, Y: bottom}, th/2)
				}

				tile.wallBody = body
			}

			if tile.Door != 0 {
				mass := 20.0

				// Variables for rectangle and constraints
				var width, height float64
				var bodyPos, pivot cp.Vector
				var angle float64

				switch {
				case tile.Door&WALL_T != 0:
					// Top wall (horizontal door)
					angle = 0
					width = s - th/2
					height = th
					bodyPos = cp.Vector{X: tileX + s/2, Y: tileY + th/2}
					pivot = cp.Vector{X: tileX + th/2, Y: tileY + th/2} // left edge
				case tile.Door&WALL_R != 0:
					// Right wall (vertical door)
					angle = rl.Pi * 0.5
					width = th
					height = s - th/2
					bodyPos = cp.Vector{X: tileX + s - th/2, Y: tileY + s/2}
					pivot = cp.Vector{X: tileX + s - th/2, Y: tileY + th/2} // top edge
				case tile.Door&WALL_B != 0:
					// Bottom wall (horizontal door)
					angle = rl.Pi * 1
					width = s - th/2
					height = th
					bodyPos = cp.Vector{X: tileX + s/2, Y: tileY + s - th/2}
					pivot = cp.Vector{X: tileX + th/2, Y: tileY + s - th/2} // left edge
				case tile.Door&WALL_L != 0:
					// Left wall (vertical door)
					angle = rl.Pi * 1.5
					width = th
					height = s - th/2
					bodyPos = cp.Vector{X: tileX + th/2, Y: tileY + s/2}
					pivot = cp.Vector{X: tileX + th/2, Y: tileY + th/2} // top edge
				}

				// Create dynamic body
				moment := cp.MomentForBox(mass, width, height)
				body := cp.NewBody(mass, moment)
				body.SetPosition(bodyPos)
				body.SetAngle(angle)
				w.Space.AddBody(body)

				// Rectangle vertices centered at body
				hw := width / 2
				hh := height / 2
				verts := []cp.Vector{
					{X: -hw, Y: -hh},
					{X: hw, Y: -hh},
					{X: hw, Y: hh},
					{X: -hw, Y: hh},
				}

				// Create polygon shape
				shape := cp.NewPolyShape(body, 4, verts, cp.NewTransformRotate(angle), 0)
				shape.Filter.Group = shapeFilterGroup
				shape.SetElasticity(0)
				shape.SetFriction(0.9)
				body.AddShape(shape)
				w.Space.AddShape(shape)

				// Add pivot joint to static body
				pivotJoint := cp.NewPivotJoint(w.Space.StaticBody, body, pivot)
				pivotJoint.SetMaxForce(1e6)
				w.Space.AddConstraint(pivotJoint)

				// Add rotary limit to constrain door swing
				minAngle := angle - rl.Pi/1.2 // adjust as needed
				maxAngle := angle + rl.Pi/1.2
				rotaryLimit := cp.NewRotaryLimitJoint(w.Space.StaticBody, body, minAngle, maxAngle)
				rotaryLimit.SetMaxForce(1e8)
				w.Space.AddConstraint(rotaryLimit)

				stiffness := 50.0 * body.Moment()
				damping := 2 * math.Sqrt(stiffness*body.Moment())
				dampedSpring := cp.NewDampedRotarySpring(w.Space.StaticBody, body, -angle, stiffness, damping)
				w.Space.AddConstraint(dampedSpring)

				tile.DoorBody = body
			}

		}
	}
}

func GenerateMaze(tilemap *Tilemap, startX, startY, width, height int) {
	// 1. Fill the area with walls
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			tilemap.Cols[startX+x][startY+y].Wall = WALL_R | WALL_L | WALL_T | WALL_B
		}
	}

	type cell struct{ x, y int }
	visited := make(map[cell]bool)
	stack := []cell{{startX, startY}}
	visited[cell{startX, startY}] = true

	for len(stack) > 0 {
		current := stack[len(stack)-1]

		// Find unvisited neighbors
		var neighbors []cell
		dirs := []struct {
			dx, dy        int
			wall, oppWall int
		}{
			{0, -1, WALL_T, WALL_B}, // Top
			{0, 1, WALL_B, WALL_T},  // Bottom
			{-1, 0, WALL_L, WALL_R}, // Left
			{1, 0, WALL_R, WALL_L},  // Right
		}

		for _, d := range dirs {
			nx, ny := current.x+d.dx, current.y+d.dy
			if nx >= startX && nx < startX+width && ny >= startY && ny < startY+height {
				if !visited[cell{nx, ny}] {
					neighbors = append(neighbors, cell{nx, ny})
				}
			}
		}

		if len(neighbors) > 0 {
			// Pick random neighbor
			next := neighbors[rl.GetRandomValue(0, int32(len(neighbors)-1))]

			// Remove walls between current and next
			dx, dy := next.x-current.x, next.y-current.y
			if dx == 1 {
				tilemap.Cols[current.x][current.y].Wall &= ^WALL_R
				tilemap.Cols[next.x][next.y].Wall &= ^WALL_L
			} else if dx == -1 {
				tilemap.Cols[current.x][current.y].Wall &= ^WALL_L
				tilemap.Cols[next.x][next.y].Wall &= ^WALL_R
			} else if dy == 1 {
				tilemap.Cols[current.x][current.y].Wall &= ^WALL_B
				tilemap.Cols[next.x][next.y].Wall &= ^WALL_T
			} else if dy == -1 {
				tilemap.Cols[current.x][current.y].Wall &= ^WALL_T
				tilemap.Cols[next.x][next.y].Wall &= ^WALL_B
			}

			visited[next] = true
			stack = append(stack, next)
		} else {
			// Backtrack
			stack = stack[:len(stack)-1]
		}
	}

	tilemap.Cols[startX+width-2][startY+height-1].Wall = 0
}
