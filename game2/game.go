package game2

import (
	"encoding/gob"
	"image/color"
	"math"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jakecoffman/cp"
)

type RenderFlags = int32

const (
	RENDER_FLAG_NO_ENTITIES = RenderFlags(1 << iota)
	RENDER_FLAG_NO_LEVEL
	RENDER_FLAG_EFFECTS
	RENDER_FLAG_CP_SHAPES
	RENDER_FLAG_CP_CONSTRAINTS
	RENDER_FLAG_CP_COLLISIONS
)

const PHYSICS_TICKRATE = time.Second / 60

type Game struct {
	Time                   time.Duration
	TimeDelta              time.Duration
	TimePhysicsAccumulator time.Duration

	Day float64

	Player *Player
	Space  *cp.Space
	Level  *Level

	IsStation   bool
	RenderFlags RenderFlags

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

	Player PlayerSave

	Level       Level
	RenderFlags RenderFlags

	Mode     ModeType
	ModeFree *ModeFree
}

func (g *Game) ToSave() GameSave {
	return GameSave{
		Time:                   g.Time,
		TimeDelta:              g.TimeDelta,
		TimePhysicsAccumulator: g.TimePhysicsAccumulator,
		Level:                  *g.Level,
		Player:                 g.Player.ToSave(g),
		RenderFlags:            g.RenderFlags,
		Mode:                   g.Mode,
		ModeFree:               g.ModeFree,
	}
}
func (g *Game) Update(dt time.Duration) {

	g.TimeDelta = dt
	g.Time += dt
	g.TimePhysicsAccumulator += dt

	g.Day += dt.Seconds() / 10

	for g.TimePhysicsAccumulator >= PHYSICS_TICKRATE {
		g.Space.Step(PHYSICS_TICKRATE.Seconds())
		g.TimePhysicsAccumulator -= PHYSICS_TICKRATE
	}

	g.Player.Update(g)

	cellWakeX := 3
	cellWakeZ := 3
	playerPos := g.Player.Position3D()

	for ix := range cellWakeX {
		for iz := range cellWakeZ {
			cellPos := playerPos.Add(NewVec3(
				float64(ix)-float64(cellWakeX-1)/2,
				0,
				float64(iz)-float64(cellWakeZ-1)/2,
			))

			cell := g.Level.GetCell(cellPos)
			cell.Wake(g)
		}
	}

}

func (save GameSave) Load() *Game {
	g := &Game{
		Time:                   save.Time,
		TimeDelta:              save.TimeDelta,
		TimePhysicsAccumulator: save.TimePhysicsAccumulator,

		RenderFlags: save.RenderFlags,
		Mode:        save.Mode,
		ModeFree:    save.ModeFree,

		Tileset: NewTileset("./models/atlas.png", 5),

		Models:     map[string]rl.Model{},
		Textures:   map[string]rl.Texture2D{},
		MainShader: NewShader(&MainShader{}, "./glsl330/lighting.vs", "./glsl330/lighting.fs"),
	}

	g.Space = cp.NewSpace()
	g.Level = save.Level.Init()

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
	gameSave := GameSave{
		Time:                   0,
		TimeDelta:              0,
		TimePhysicsAccumulator: 0,
		Player: PlayerSave{
			Position: NewVec2(0, 0),
		},
		Mode:     MODE_DEFAULT,
		ModeFree: NewModeFree(),
		Level:    Level{},
	}

	return gameSave
}

func (g *Game) Draw() {
	rl.ClearBackground(rl.DarkGray)

	camera := Camera3D{
		Position:   g.Player.Position3D().Add(NewVec3(0, 8, -1).Normalize().Scale(10)),
		Target:     g.Player.Position3D(),
		Fovy:       10,
		Up:         Y,
		Projection: rl.CameraOrthographic,
	}

	BeginMode3D(camera, func() {
		g.Draw3D()
	})
}

var NIGHT = Vec3{-115, 0.3, .1}
var DAWN = Vec3{5, 0.5, 1}
var DAY = Vec3{55, 0.1, 1}
var HOUR_MORNING float64 = 9
var HOUR_NIGHT float64 = 21
var HOURS_TRANSITION float64 = 1

func c(x float64) float64 {
	return (1 + math.Tanh(x)) / 2
}

func (g *Game) Draw3D() {

	if g.RenderFlags&RENDER_FLAG_EFFECTS != 0 {
		g.MainShader.Ambient.SetColor(color.RGBA{255, 255, 255, 255})
	} else {
		g.MainShader.Ambient.SetColor(color.RGBA{5, 5, 5, 255})

		if !g.IsStation {

			hour := math.Mod(g.Day, 1) * 24
			day := c(hour-HOUR_MORNING) - c(hour-HOUR_NIGHT)
			transitionColor := 1 + ((c(2*(hour-HOUR_MORNING-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_MORNING+HOURS_TRANSITION/2))) + (c(2*(hour-HOUR_NIGHT-HOURS_TRANSITION/2)) - c(2*(hour-HOUR_NIGHT+HOURS_TRANSITION/2))))
			transitionAngle := 1 + ((c((hour - HOUR_MORNING - HOURS_TRANSITION/2)) - c((hour - HOUR_MORNING + HOURS_TRANSITION/2))) + (c((hour - HOUR_NIGHT - HOURS_TRANSITION/2)) - c((hour - HOUR_NIGHT + HOURS_TRANSITION/2))))

			sunColor := DAWN.Lerp(NIGHT.Lerp(DAY, (day)), (transitionColor))

			g.MainShader.LightDirectional(NewVec3((1-transitionAngle), (1-day*2), 0).Normalize(), rl.ColorFromHSV(float32(sunColor.X), float32(sunColor.Y), float32(sunColor.Z)), 0.5)

		} else {
			g.MainShader.LightDirectional(NewVec3(0, -1, 0).Normalize(), rl.White, 0.25)

		}
	}

	g.MainShader.UpdateValues()

	if g.RenderFlags&RENDER_FLAG_NO_ENTITIES == 0 {
		g.Player.Draw(g)
	}

	if g.RenderFlags&RENDER_FLAG_NO_LEVEL == 0 {
		g.Level.Draw(g)
	}

	rl.DrawRenderBatchActive()
	rl.DisableDepthTest()
	if g.RenderFlags&(RENDER_FLAG_CP_SHAPES|RENDER_FLAG_CP_CONSTRAINTS|RENDER_FLAG_CP_COLLISIONS) != 0 {
		drawer := NewPhysicsDrawer(
			g.RenderFlags&RENDER_FLAG_CP_SHAPES != 0,
			g.RenderFlags&RENDER_FLAG_CP_CONSTRAINTS != 0,
			g.RenderFlags&RENDER_FLAG_CP_COLLISIONS != 0,
		)

		cp.DrawSpace(g.Space, &drawer)
	}

	rl.DrawCube(NewVec3(0, 0, 0).Raylib(), 0.05, 0.05, 0.05, rl.White)
	rl.DrawCube(NewVec3(0, 0, 1).Raylib(), 0.05, 0.05, 0.05, rl.SkyBlue)
	rl.DrawCube(NewVec3(0, 1, 0).Raylib(), 0.05, 0.05, 0.05, rl.Yellow)
	rl.DrawCube(NewVec3(1, 0, 0).Raylib(), 0.05, 0.05, 0.05, rl.Red)

	rl.DrawRenderBatchActive()
	rl.EnableDepthTest()

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
