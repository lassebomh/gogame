package game2

import (
	"encoding/gob"
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

	Tileset *Tileset

	MainShader *MainShader

	Textures map[string]rl.Texture2D
	Models   map[string]rl.Model
}

type GameSave struct {
	Time                   time.Duration
	TimeDelta              time.Duration
	TimePhysicsAccumulator time.Duration

	Level LevelSave

	Player   PlayerSave
	Mode     ModeType
	ModeFree ModeFree
}

func (g *Game) ToSave() GameSave {
	return GameSave{
		Time:                   g.Time,
		TimeDelta:              g.TimeDelta,
		TimePhysicsAccumulator: g.TimePhysicsAccumulator,
		Level:                  g.Level.ToSave(),
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

	// g.Level.GetCell(g.Player.Body.Position().X, g.Player.Y, g.Player.Body.Position().Y)
}

func (save GameSave) Load() *Game {
	g := &Game{
		Time:                   save.Time,
		TimeDelta:              save.TimeDelta,
		TimePhysicsAccumulator: save.TimePhysicsAccumulator,
		Mode:                   save.Mode,
		ModeFree:               &save.ModeFree,

		Tileset: NewTileset("./models/atlas.png", 5),

		Models:     map[string]rl.Model{},
		Textures:   map[string]rl.Texture2D{},
		MainShader: NewShader(&MainShader{}, "./glsl330/lighting.vs", "./glsl330/lighting.fs"),
	}

	g.Space = cp.NewSpace()
	g.Level = save.Level.Load(g)

	g.Models["wallDebug"] = rl.LoadModel("./models/wallx.glb")
	g.Models["wall"] = rl.LoadModel("./models/wallx.glb")

	mats := g.Models["wall"].GetMaterials()

	for i := range mats {
		mat := &mats[i]
		mat.Shader = g.MainShader.shader
		mat.Maps.Texture = g.Tileset.Texture
	}

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
		Position:   g.Player.Position3D().Add(NewVec3(0, 8, 1).Normalize().Scale(10)),
		Target:     g.Player.Position3D(),
		Fovy:       10,
		Up:         Y,
		Projection: rl.CameraOrthographic,
	}

	BeginMode3D(camera, func() {
		g.Player.Draw(g)
		g.Level.Draw(g)
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
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(save); err != nil {
		return err
	}

	return nil
}

func (save GameSave) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(save); err != nil {
		return err
	}

	return nil
}
