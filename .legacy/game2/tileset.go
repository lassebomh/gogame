package game2

import rl "github.com/gen2brain/raylib-go/raylib"

type Tileset struct {
	Texture       rl.Texture2D
	TextureWidth  float64
	TextureHeight float64
	Tiles         int
}

func NewTileset(path string, tilesX int) *Tileset {
	ts := &Tileset{
		Texture: rl.LoadTexture(path),
		Tiles:   tilesX,
	}
	ts.TextureWidth = float64(ts.Texture.Width)
	ts.TextureHeight = float64(ts.Texture.Height)

	return ts
}

func (ts *Tileset) GetAABB(x int, y int) (Vec2, Vec2) {
	aa := NewVec2(float64(x)/float64(ts.Tiles), float64(y)/float64(ts.Tiles))
	bb := aa.Add(NewVec2(1/float64(ts.Tiles), 1/float64(ts.Tiles)))

	return aa, bb
}
