package game

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const (
	WALL_T = 1 << iota
	WALL_R
	WALL_B
	WALL_L
)

type Tile struct {
	Wall     int
	WallBody *cp.Body
}

type Tilemap struct {
	Cols  [][]Tile
	Scale float64
}

func NewTilemap(width int, height int, scale float64) *Tilemap {
	tilemap := Tilemap{
		Cols:  make([][]Tile, width),
		Scale: scale,
	}

	for x := range width {
		col := make([]Tile, height)

		for y := range height {
			col[y] = Tile{}
		}

		tilemap.Cols[x] = col
	}

	return &tilemap
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

			th := t.Scale / 8
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

			tile.WallBody = body
		}
	}
}
