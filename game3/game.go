package game3

import (
	"encoding/gob"
	"log"
	"os"
	"time"
)

var game *Game

type Game struct {
	Time    time.Duration
	Day     float64
	DayPrev float64

	Earth   *World
	Station *World
}

func LoadGame(path string) {
	game = &Game{}

	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
	}

	if err == nil {
		decoder := gob.NewDecoder(file)
		err := decoder.Decode(game)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err != nil {
		log.Println("Save not found.")
	}

	globalsInit()

	game.Earth.Upsert(WorldEarth)
	game.Station.Upsert(WorldStation)
	game.Earth.other = game.Station
	game.Station.other = game.Earth

	if game.Station.Player != nil {
		game.Station.Player.Spawn(game.Station)
	} else if game.Earth.Player != nil {
		game.Earth.Player.Spawn(game.Earth)
	} else {
		SpawnNewPlayer(game.Earth)
	}

	if game.Station.Monster != nil {
		game.Station.Monster.Spawn(game.Station)
	} else if game.Earth.Monster != nil {
		game.Earth.Monster.Spawn(game.Earth)
	} else {
		SpawnNewMonster(game.Earth)
	}
}

func (g *Game) Update(dt time.Duration) {
	g.Time += dt

	g.DayPrev = g.Day
	g.Day = g.Time.Seconds() / (60 * 1)

	if g.Station.Player != nil {
		g.Station.Update(dt)
	} else {
		g.Earth.Update(dt)
	}
}

func (g *Game) Draw() {
	if g.Station.Player != nil {
		g.Station.Draw()
	} else {
		g.Earth.Draw()
	}
}

func (g *Game) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(g); err != nil {
		return err
	}

	return nil
}
