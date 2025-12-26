package game

import (
	"image/color"
	"math"

	"github.com/beefsack/go-astar"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

var p astar.Pather = &Tile{}

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

	tilemap  *Tilemap
	wallBody *cp.Body
}

type Tilemap struct {
	Cols   [][]Tile
	Scale  float64
	Width  int
	Height int
}

func NewTilemap(width int, height int, scale float64) *Tilemap {
	tilemap := &Tilemap{
		Cols:   make([][]Tile, width),
		Scale:  scale,
		Width:  width,
		Height: height,
	}

	for x := range width {
		col := make([]Tile, height)
		for y := range height {
			col[y] = Tile{X: x, Y: y, tilemap: tilemap}
		}

		tilemap.Cols[x] = col
	}

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

func (t *Tilemap) GenerateBodies(w *World) {
	for x, rows := range t.Cols {
		for y := range rows {
			tile := &t.Cols[x][y]
			if tile.Wall == 0 {
				continue
			}

			tileX := float64(x) * t.Scale
			tileY := float64(y) * t.Scale

			th := t.Scale / 6
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

func RenderPath(path []cp.Vector, color rl.Color) {
	if len(path) < 2 {
		return
	}

	for i := 0; i < len(path)-1; i++ {
		start := path[i]
		end := path[i+1]
		rl.DrawLineV(v(start), v(end), color)
	}

	// Optional: Draw dots at each waypoint
	for _, point := range path {
		rl.DrawCircleV(v(point), 3, color)
	}
}
