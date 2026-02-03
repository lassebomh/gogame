package game3

import "game/vec3"

var game *Game

type Game struct {
	Earth   *World
	Station *World

	ActiveEditor *Editor
}

func (g *Game) Upsert() {
	if g == nil {
		g = &Game{}
	}
	game = g
	game.Earth.Upsert(WorldEarth)
	game.Station.Upsert(WorldStation)

}

type WorldType int32

const (
	WorldEarth = WorldType(iota)
	WorldStation
)

// Seperate levels for station and world?
type World struct {
	Type   WorldType
	chunks map[ChunkPos]*Chunk

	Editor *Editor
}

func (w *World) Upsert(worldType WorldType) *World {
	if w == nil {
		w = &World{}
	}
	w.Type = worldType
	if w.Type == WorldEarth {
		game.Earth = w
	} else {
		game.Station = w
	}

	if w.chunks == nil {
		w.chunks = make(map[ChunkPos]*Chunk)
	}
	for _, chunk := range w.chunks {
		chunk.Reload()
	}
	w.Editor.Upsert(w)

	return w
}

func (w *World) GetCell(pos vec3.Value) *Cell {
	cpos, lpos := WorldToChunk(pos)

	chunk, ok := w.chunks[cpos]

	if !ok {
		chunk = chunk.Upsert(cpos)
	}

	cell := &chunk.Cells[lpos.Y][lpos.X][lpos.Z]

	return cell
}
