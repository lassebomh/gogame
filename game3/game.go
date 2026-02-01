package game3

import rl "github.com/gen2brain/raylib-go/raylib"

var game *Game

type Game struct {
	Level *Level

	atlasTexture rl.Texture2D
}

func (g *Game) Init() {
	if g == nil {
		g = &Game{}
	}
	game = g
	game.Level.Init()
}

// Seperate levels for station and world?
type Level struct {
	Chunks map[ChunkPos]*Chunk
}

func (l *Level) Init() *Level {
	if l == nil {
		l = &Level{}
	}
	game.Level = l
	if l.Chunks == nil {
		l.Chunks = make(map[ChunkPos]*Chunk)
	}
	for pos, chunk := range l.Chunks {
		chunk.Init(pos)
	}

	return l
}
