package game2

import (
	"encoding/json"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

const PHYSICS_TICKRATE = time.Second / 60

type Game struct {
	Time                   time.Duration
	TimeDelta              time.Duration
	TimePhysicsAccumulator time.Duration

	Player *Player
	Space  *cp.Space
	Level  *Level

	Mode     ModeType
	ModeFree *ModeFree
}

type GameSave struct {
	Time                   time.Duration
	TimeDelta              time.Duration
	TimePhysicsAccumulator time.Duration

	Player   PlayerSave
	Mode     ModeType
	ModeFree ModeFree
}

func (g *Game) ToSave() GameSave {
	return GameSave{
		Time:                   g.Time,
		TimeDelta:              g.TimeDelta,
		TimePhysicsAccumulator: g.TimePhysicsAccumulator,
		Player:                 g.Player.ToSave(g),
		Mode:                   g.Mode,
		ModeFree:               *g.ModeFree,
	}
}
func (g *Game) Update(dt time.Duration) {

	g.TimeDelta = dt
	g.Time += dt
	g.TimePhysicsAccumulator += dt

	for g.TimePhysicsAccumulator >= PHYSICS_TICKRATE {
		g.Space.Step(PHYSICS_TICKRATE.Seconds())
		g.TimePhysicsAccumulator -= PHYSICS_TICKRATE
	}

	g.Player.Update(g)

	g.Level.GetCell(g.Player.Body.Position().X, g.Player.Y, g.Player.Body.Position().Y)
}

func (save GameSave) Load() *Game {
	g := &Game{
		Time:                   save.Time,
		TimeDelta:              save.TimeDelta,
		TimePhysicsAccumulator: save.TimePhysicsAccumulator,
		Mode:                   save.Mode,
		ModeFree:               &save.ModeFree,
	}

	g.Level = NewLevel()
	g.Space = cp.NewSpace()

	save.Player.Load(g)

	return g
}

func NewGameSave() GameSave {
	return GameSave{
		Time:                   0,
		TimeDelta:              0,
		TimePhysicsAccumulator: 0,
		Player: PlayerSave{
			Position: NewVec2(0, 0),
		},
		Mode:     MODE_DEFAULT,
		ModeFree: NewModeFree(),
	}
}

func (g *Game) Draw() {
	rl.ClearBackground(rl.Gray)

	camera := Camera3D{
		Position:   g.Player.Position3D().Add(NewVec3(-1, 10, 0).Normalize().Scale(10)),
		Target:     g.Player.Position3D(),
		Fovy:       20,
		Up:         Y,
		Projection: rl.CameraOrthographic,
	}

	BeginMode3D(camera, func() {
		g.Player.Draw(g)
	})
}

func (g *Game) ModeUpdate(mode ModeType, dt time.Duration) {
	switch g.Mode {
	case MODE_DEFAULT:
		g.Update(dt)
	case MODE_FREE:
		g.ModeFree.Update(g)
	}
}

func (g *Game) ModeDraw(mode ModeType) {
	switch mode {
	case MODE_DEFAULT:
		g.Draw()
	case MODE_FREE:
		g.ModeFree.Draw(g)
	}
}

func LoadSaveFromFile(path string, save *GameSave) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, save)
	if err != nil {
		return err
	}

	return nil
}

func (save GameSave) WriteToFile(path string) error {
	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
