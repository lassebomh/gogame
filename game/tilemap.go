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
	X    int
	Y    int

	WorldPosition cp.Vector

	tilemap  *Tilemap
	wallBody *cp.Body
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
	neighbors := make([]astar.Pather, 0, 4)

	if t.Wall&WALL_L == 0 && t.X > 0 {
		left := &t.tilemap.Cols[t.X-1][t.Y]

		if left.Wall&WALL_R == 0 {
			neighbors = append(neighbors, left)
		}
	}

	if t.Wall&WALL_T == 0 && t.Y > 0 {
		TOP := &t.tilemap.Cols[t.X][t.Y-1]

		if TOP.Wall&WALL_B == 0 {
			neighbors = append(neighbors, TOP)
		}
	}

	if t.Wall&WALL_R == 0 && t.X < t.tilemap.Width-1 {
		left := &t.tilemap.Cols[t.X+1][t.Y]

		if left.Wall&WALL_L == 0 {
			neighbors = append(neighbors, left)
		}
	}

	if t.Wall&WALL_B == 0 && t.Y < t.tilemap.Height-1 {
		left := &t.tilemap.Cols[t.X][t.Y+1]

		if left.Wall&WALL_T == 0 {
			neighbors = append(neighbors, left)
		}
	}

	return neighbors
}

func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	return 1
}

func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	other := to.(*Tile)
	return math.Hypot(float64(t.X-other.X), float64(t.Y-other.Y))
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

func (t *Tilemap) FindPath(start, end cp.Vector) []cp.Vector {

	startTile := t.GetTileAtWorldPosition(start)
	endTile := t.GetTileAtWorldPosition(end)

	if startTile == nil || endTile == nil {
		return nil
	}

	path, _, found := astar.Path(endTile, startTile)
	if !found {
		return nil
	}

	worldPath := make([]cp.Vector, len(path))
	for i, p := range path {
		tile := p.(*Tile)
		worldPath[i] = cp.Vector{X: float64(tile.X) + 0.5, Y: float64(tile.Y) + 0.5}.Mult(t.Scale)
	}

	return worldPath
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
	}
	if doorDirection&WALL_R != 0 {
		t.Cols[x+width-1][y+height/2].Wall = 0
	}
}

func (t *Tilemap) GenerateBodies(w *World) {
	for x, rows := range t.Cols {
		for y := range rows {
			tile := &t.Cols[x][y]
			if tile.Wall == 0 {
				continue
			}

			tileX := float64(x) * t.Scale
			tileY := float64(y) * t.Scale

			th := t.Scale * float64(t.WallDepthRatio)
			s := t.Scale

			body := cp.NewStaticBody()

			// tile bounds (inside edges)
			left := tileX + th/2
			right := tileX + s - th/2
			top := tileY + th/2
			bottom := tileY + s - th/2

			// Top wall
			if tile.Wall&WALL_T != 0 {
				shape := cp.NewSegment(body,
					cp.Vector{X: left, Y: top},
					cp.Vector{X: right, Y: top},
					th/2,
				)
				w.Space.AddShape(shape)
			}

			// Right wall
			if tile.Wall&WALL_R != 0 {
				shape := cp.NewSegment(body,
					cp.Vector{X: right, Y: top},
					cp.Vector{X: right, Y: bottom},
					th/2,
				)
				w.Space.AddShape(shape)
			}

			// Bottom wall
			if tile.Wall&WALL_B != 0 {
				shape := cp.NewSegment(body,
					cp.Vector{X: left, Y: bottom},
					cp.Vector{X: right, Y: bottom},
					th/2,
				)
				w.Space.AddShape(shape)
			}

			// Left wall
			if tile.Wall&WALL_L != 0 {
				shape := cp.NewSegment(body,
					cp.Vector{X: left, Y: top},
					cp.Vector{X: left, Y: bottom},
					th/2,
				)
				w.Space.AddShape(shape)
			}

			tile.wallBody = body
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
}
